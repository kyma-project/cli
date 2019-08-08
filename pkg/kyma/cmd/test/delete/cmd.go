package del

import (
	"fmt"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Use:   "delete <test-suite-1> <test-suite-2> ... <test-suite-N>",
		Short: "Deletes test suites available for a provisioned Kyma cluster.",
		Long: `Deletes test suites available for a provisioned Kyma cluster.

Provide at least one test suite name.`,
		RunE:    func(_ *cobra.Command, args []string) error { return cmd.Run(args) },
		Aliases: []string{"d"},
	}

	return cobraCmd
}

func (cmd *command) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Test suite name required")
	}

	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Make sure your kubeconfig is valid.")
	}

	testSuites := &oct.ClusterTestSuiteList{}
	tSuites := []oct.ClusterTestSuite{}
	for _, testName := range args {
		ts := test.NewTestSuite(testName)
		tSuites = append(tSuites, *ts)
	}
	testSuites.Items = tSuites
	for _, ts := range testSuites.Items {
		if err := deleteTestSuite(cmd.K8s.Octopus(), ts.GetName()); err != nil {
			return err
		}
	}

	return nil
}

func deleteTestSuite(cli octopus.OctopusInterface, testName string) error {
	if err := cli.DeleteTestSuite(test.NewTestSuite(testName).GetName(), metav1.DeleteOptions{}); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to delete test suite '%s'",
			testName))
	}
	fmt.Printf("Test suite '%s' successfully deleted\n", testName)
	return nil
}
