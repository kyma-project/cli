package cmd

import (
	"bytes"
	"fmt"

	kyma_helm "github.com/kyma-project/cli/internal/helm"
	"github.com/kyma-project/cli/internal/installer"
	"github.com/kyma-project/cli/internal/kubectl"
	"github.com/kyma-project/cli/internal/step"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	helm_release "k8s.io/helm/pkg/proto/hapi/release"
)

type TestOptions struct {
	*core.Options
	skip []string
}

func NewTestOptions(o *core.Options) *TestOptions {
	return &TestOptions{Options: o}
}

var namespacesToClean = []string{
	"kyma-system",
	"istio-system",
	"kyma-integration",
	"knative-serving",
}

//NewTestCmd creates a new install command
func NewTestCmd(o *TestOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test kyma installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmd.Flags().StringArrayVar(&o.skip, "skip", []string{}, "Skip tests for these releases")
	return cmd
}

func (o *TestOptions) Run() error {
	helmHome, err := kyma_helm.GetHelmHome()
	if err != nil {
		return err
	}

	helmConfig := &environment.EnvSettings{Debug: true, TillerConnectionTimeout: 300, TLSEnable: true, TLSVerify: false, Home: helmpath.Home(helmHome)}
	kubeConfig, err := o.GetKubeconfig()
	if err != nil {
		return err
	}

	helmClient, err := kyma_helm.New(helmConfig, kubeConfig)
	if err != nil {
		return err
	}

	components, err := installer.GetComponents(kubeConfig)
	if err != nil {
		return err
	}

	for _, namespace := range namespacesToClean {
		err := o.cleanHelmTestPods(namespace)
		if err != nil {
			return err
		}
	}

	for _, component := range components {
		release := component.Name
		s := o.NewStep(fmt.Sprintf("Testing release %s", release))
		s.Start()

		if o.skipRelease(helmClient, release) {
			s.Successf("Skipping release %s", release)
			continue
		}

		msg, success, err := o.testRelease(helmClient, s, release)
		s.Stop(success)
		if msg != "" && (!success || o.Verbose) {
			fmt.Println("--- Log ---")
			fmt.Print(msg)
			fmt.Println("---")
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *TestOptions) skipRelease(client helm.Interface, release string) bool {
	for _, skipped := range o.skip {
		if skipped == release {
			return true
		}
	}

	return false
}

func (o *TestOptions) cleanHelmTestPods(namespace string) error {
	s := o.NewStep(
		fmt.Sprintf("Cleaning up helm test pods in namespace %s", namespace),
	)
	s.Start()

	_, err := kubectl.RunCmd(o.Verbose, "delete", "pod", "-n", namespace, "-l", "helm-chart-test=true")
	if err != nil {
		s.Failure()
		return err
	}

	s.Success()
	return nil
}

func (o *TestOptions) testRelease(client helm.Interface, s step.Step, release string) (string, bool, error) {
	c, errc := client.RunReleaseTest(release, helm.ReleaseTestTimeout(600))
	out := &bytes.Buffer{}
	failed := 0
	for {
		select {
		case err := <-errc:
			if err != nil {
				return out.String(), false, err
			}
		case res, ok := <-c:
			if !ok {
				return out.String(), failed == 0, nil
			}
			if res.Status == helm_release.TestRun_FAILURE {
				failed++
			}
			out.WriteString(res.Msg)
			out.WriteString("\n")
			s.Status(res.Msg)
		}
	}
}
