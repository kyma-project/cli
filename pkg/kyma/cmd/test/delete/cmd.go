package del

import (
	"fmt"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
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
		Use:     "delete <test-suite name>",
		Short:   "Delete tests on a running Kyma cluster",
		Long:    `Delete tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, args []string) error { return cmd.Run(args) },
		Aliases: []string{"d"},
	}

	return cobraCmd
}

func (cmd *command) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("test suite name required")
	}

	cli, err := client.NewTestRESTClient(10 * time.Second)
	if err != nil {
		return fmt.Errorf("unable to create test REST client. E: %s", err)
	}

	testSuites := &oct.ClusterTestSuiteList{}
	tSuites := []oct.ClusterTestSuite{}
	for _, testName := range args {
		ts := test.NewTestSuite(testName)
		tSuites = append(tSuites, *ts)
	}
	testSuites.Items = tSuites
	for _, ts := range testSuites.Items {
		if err := deleteTestSuite(cli, ts.GetName()); err != nil {
			return err
		}
	}

	return nil
}

func deleteTestSuite(cli client.TestRESTClient, testName string) error {
	if err := cli.DeleteTestSuite(test.NewTestSuite(testName)); err != nil {
		return fmt.Errorf("unable to delete test suite '%s'. E: %s",
			testName, err.Error())
	}
	fmt.Printf("test '%s' successfully deleted\n", testName)
	return nil
}
