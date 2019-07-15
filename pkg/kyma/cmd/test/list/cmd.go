package list

import (
	"fmt"
	"os"

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
		Short:   "List test suites available for a provisioned Kyma cluster",
		Long:    `List test suites available for a provisioned Kyma cluster`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"l"},
	}

	return cobraCmd
}

func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "could not initialize the Kubernetes client. Make sure that your kubeconfig is valid.")
	}

	testSuites, err := cmd.K8s.Octopus().ListTestSuites()
	if err != nil {
		return errors.Wrap(err, "unable to get list of test suites")
	}

	if len(testSuites.Items) == 0 {
		fmt.Println("no test suites found")
		return nil
	}

	writer := test.NewTableWriter([]string{"TEST SUITE", "COMPLETED", "STATUS"}, os.Stdout)

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
			fmt.Sprintf("%d/%d", test.GetNumberOfFinishedTests(&t), len(t.Status.Results)),
			testResult,
		})
	}
	writer.Render()

	return nil
}
