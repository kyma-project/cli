package list

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/octopus/pkg/apis"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/client"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes/scheme"
)

type command struct {
	opts *options
	core.Command
}

func NewCmd(o *options) *cobra.Command {
	cmd := command{
		Command: core.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "list",
		Short:   "Show available tests on a running Kyma cluster",
		Long:    `Show available tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"l"},
	}

	cobraCmd.Flags().BoolVarP(&o.Definitions, "definitions", "d", false, "Show test definitions only")
	cobraCmd.Flags().BoolVarP(&o.Tests, "tests", "t", false, "Show test-suites only")
	return cobraCmd
}

func (cmd *command) Run() error {
	var err error
	apis.AddToScheme(scheme.Scheme)

	cli, err := client.NewTestRESTClient(10 * time.Second)
	if err != nil {
		return fmt.Errorf("unable to create test REST client. E: %s", err)
	}

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Please make sure that you have a valid kubeconfig.")
	}
	if !cmd.opts.Tests && !cmd.opts.Definitions {
		cmd.opts.Tests = true
		cmd.opts.Definitions = true
	}
	if cmd.opts.Tests {
		fmt.Println("Test suites:")
		if testSuites, err := test.ListTestSuiteNames(cli); err != nil {
			return err
		} else {
			for _, t := range testSuites {
				fmt.Printf("\t%s\r\n", t)
			}
		}
	}
	if cmd.opts.Definitions {
		fmt.Println("Test definitions:")
		if testDefs, err := test.ListTestDefinitionNames(cli); err != nil {
			return err
		} else {
			for _, t := range testDefs {
				fmt.Printf("\t%s\r\n", t)
			}
		}
	}
	return nil
}
