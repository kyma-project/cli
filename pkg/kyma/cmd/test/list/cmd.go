package list

import (
	"fmt"
	"os"
	"strconv"
	"time"

	client "github.com/kyma-project/cli/pkg/api/test"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/core"
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
	cli, err := client.NewTestRESTClient(10 * time.Second)
	if err != nil {
		return fmt.Errorf("unable to create test REST client. E: %s", err)
	}

	if cmd.opts.Definitions {
		if testDefs, err := test.ListTestDefinitionNames(cli); err != nil {
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

	testSuites, err := cli.ListTestSuites()
	if err != nil {
		return fmt.Errorf("unable to get list of test suites. E: %s", err.Error())
	}

	if len(testSuites.Items) == 0 {
		return fmt.Errorf("no test suites in the cluster")
		return nil
	}

	writer := test.NewTableWriter([]string{"TEST NAME", "TESTS", "STATUS"}, os.Stdout)

	for _, t := range testSuites.Items {
		var testResult string
		switch len(t.Status.Results) {
		case 0:
			testResult = "-"
			break
		case 1:
			testResult = string(t.Status.Results[0].Status)
		default:
			testResult = string(t.Status.Results[len(t.Status.Results)-1].Status)
		}
		writer.Append([]string{
			t.GetName(),
			strconv.Itoa(len(t.Spec.Selectors.MatchNames)),
			testResult,
		})
	}
	writer.Render()

	return nil
}
