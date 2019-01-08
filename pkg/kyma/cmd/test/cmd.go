package test

import (
	"bytes"
	"fmt"
	"github.com/kyma-incubator/kymactl/internal"
	kyma_helm "github.com/kyma-incubator/kymactl/internal/helm"
	"github.com/kyma-incubator/kymactl/internal/step"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	helm_release "k8s.io/helm/pkg/proto/hapi/release"
)

type Options struct {
	step.Factory
	skip []string
	verbose bool
}

var testNamespaces = []string{
	"kyma-system",
	"istio-system",
	"knative-serving",
	"kyma-integration",
}

var testReleases = map[string]bool{
	"core": true,
	"monitoring": false,
	"logging": false,
	"istio": true,
	"knative": true,
	"application-connector": true,
}

//NewCmd creates a new install command
func NewCmd() *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test kyma installation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.Run()
		},
	}

	cmd.Flags().StringArrayVar(&opts.skip, "skip", []string{}, "Skip tests for these releases")
	cmd.Flags().BoolVar(&opts.NonInteractive, "non-interactive", false, "Do not use spinners")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "Print additional output")
	return cmd
}

func (opts *Options) Run() error {
	helmConfig := &environment.EnvSettings{TillerConnectionTimeout: 300}
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		return err
	}

	helmClient, err := kyma_helm.New(helmConfig, kubeConfig)
	if err != nil {
		return err
	}

	for _, ns := range testNamespaces {
		err := opts.cleanHelmTestPods(ns)
		if err != nil {
			return err
		}
	}

	for release := range testReleases {
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
		if msg != "" && (!success || opts.verbose) {
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

func (opts *Options) skipRelease(client helm.Interface, release string) bool {
	for _, skipped := range opts.skip {
		if skipped == release {
			return true
		}
	}

	if !testReleases[release] {
		_, err := client.ReleaseStatus(release)
		if err != nil {
			return true
		}
	}

	return false
}

func (opts *Options) cleanHelmTestPods(namespace string) error {
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

func (opts *Options) testRelease(client helm.Interface, s step.Step, release string) (string, bool, error) {
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
