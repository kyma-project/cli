package list

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
	return cobraCmd
}

func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Please make sure that you have a valid kubeconfig.")
	}

	if err != nil {
		return errors.Wrap(err, "unable to create test REST client")
	}

	if cmd.opts.Definitions {
		if testDefs, err := test.ListTestDefinitionNames(cmd.K8s.Octopus()); err != nil {
			return err
		} else {
			if len(testDefs) == 0 {
				fmt.Errorf("no test definitions in the cluster")
			}
			for _, t := range testDefs {
				fmt.Printf("%s\r\n", t)
			}
		}
		return nil
	}

	testSuites, err := cmd.K8s.Octopus().ListTestSuites()
	if err != nil {
		return errors.Wrap(err, "unable to get list of test suites")
	}

	if len(testSuites.Items) == 0 {
		fmt.Println("no test suites in the cluster")
		return nil
	}

	writer := test.NewTableWriter([]string{"TEST NAME", "TESTS", "FINISHED", "STATUS"}, os.Stdout)

	for _, t := range testSuites.Items {
		var testResult string
		switch len(t.Status.Results) {
		case 0:
			testResult = "-"
			break
		case 1:
			testResult = string(t.Status.Results[0].Status)
		default:
			testResult = string(t.Status.Conditions[len(t.Status.Conditions)-1].Type)
		}
		writer.Append([]string{
			t.GetName(),
			strconv.Itoa(len(t.Spec.Selectors.MatchNames)),
			fmt.Sprintf("%d/%d", test.GetNumberOfFinishedTests(&t), len(t.Status.Results)),
			testResult,
		})
	}
	writer.Render()

	return nil
}
