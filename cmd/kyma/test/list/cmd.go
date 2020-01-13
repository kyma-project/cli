package list

import (
	"fmt"
	"os"

	"github.com/kyma-project/cli/cmd/kyma/test"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type command struct {
	opts *Options
	cli.Command
}

func NewCmd(o *Options) *cobra.Command {
	cmd := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cobraCmd := &cobra.Command{
		Use:     "list",
		Short:   "Lists test suites available for a provisioned Kyma cluster.",
		Long:    `Use this command to list test suites available for a provisioned Kyma cluster.`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"l"},
	}
	return cobraCmd
}

func (cmd *command) Run() error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure that your kubeconfig is valid.")
	}

	testSuites, err := cmd.K8s.Octopus().ListTestSuites(metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "Unable to get list of test suites")
	}

	if len(testSuites.Items) == 0 {
		fmt.Println("No test suites found")
		return nil
	}

	writer := test.NewTableWriter([]string{"TEST SUITE", "COMPLETED", "STATUS"}, os.Stdout)

	for idx := range testSuites.Items {
		ts := testSuites.Items[idx]
		var testResult string
		switch len(ts.Status.Results) {
		case 0:
			testResult = "-"
		case 1:
			testResult = string(ts.Status.Results[0].Status)
		default:
			testResult = string(ts.Status.Conditions[len(ts.Status.Conditions)-1].Type)
		}
		writer.Append([]string{
			ts.GetName(),
			fmt.Sprintf("%d/%d", test.GetNumberOfFinishedTests(&ts), len(ts.Status.Results)),
			testResult,
		})
	}
	writer.Render()

	return nil
}
