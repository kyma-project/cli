package cmd

import (
	"bytes"
	"fmt"

	"github.com/kyma-incubator/kymactl/internal"
	kyma_helm "github.com/kyma-incubator/kymactl/internal/helm"
	"github.com/kyma-incubator/kymactl/internal/step"
	"github.com/kyma-incubator/kymactl/pkg/installer"
	"github.com/kyma-incubator/kymactl/pkg/kyma/core"
	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
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

func (opts *TestOptions) Run() error {
	helmConfig := &environment.EnvSettings{TillerConnectionTimeout: 300}
	kubeConfig, err := opts.GetKubeconfig()
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
		err := opts.cleanHelmTestPods(namespace)
		if err != nil {
			return err
		}
	}

	for _, component := range components {
		release := component.Name
		s := opts.NewStep(
			fmt.Sprintf("Testing release %s", release),
		)
		s.Start()

		if opts.skipRelease(helmClient, release) {
			s.Successf("Skipping release %s", release)
			continue
		}

		msg, success, err := opts.testRelease(helmClient, s, release)
		s.Stop(success)
		if msg != "" && (!success || opts.Verbose) {
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

func (opts *TestOptions) skipRelease(client helm.Interface, release string) bool {
	for _, skipped := range opts.skip {
		if skipped == release {
			return true
		}
	}

	return false
}

func (opts *TestOptions) cleanHelmTestPods(namespace string) error {
	s := opts.NewStep(
		fmt.Sprintf("Cleaning up helm test pods in namespace %s", namespace),
	)
	s.Start()

	_, err := internal.RunKubectlCmd([]string{"delete", "pod", "-n", namespace, "-l", "helm-chart-test=true"})
	if err != nil {
		s.Failure()
		return err
	}

	s.Success()
	return nil
}

func (opts *TestOptions) testRelease(client helm.Interface, s step.Step, release string) (string, bool, error) {
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
