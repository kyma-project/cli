package status

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
		Use:     "status <test-suite name>",
		Short:   "Status of tests on a running Kyma cluster",
		Long:    `Status of tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, args []string) error { return cmd.Run(args) },
		Aliases: []string{"s"},
	}

	cobraCmd.Flags().StringVarP(&o.OutputFormat, "output", "o", "wide",
		"Output format. One of: json|yaml|wide")
	return cobraCmd
}

func (cmd *command) Run(args []string) error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Please make sure that you have a valid kubeconfig.")
	}

	switch len(args) {
	case 1:
		testSuite, err := cmd.K8s.Octopus().GetTestSuiteByName(args[0])
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("unable to get test suite '%s'",
				args[0]))
		}
		return cmd.printTestSuiteStatus(testSuite, cmd.opts.OutputFormat)
	case 0:
		testList, err := cmd.K8s.Octopus().ListTestSuites()
		if err != nil {
			return errors.Wrap(err, "unable to list test suites. E: %s")
		}

		if len(testList.Items) == 0 {
			fmt.Println("no test suites in the cluster")
			return nil
		}

		for _, t := range testList.Items {
			if err := cmd.printTestSuiteStatus(&t, cmd.opts.OutputFormat); err != nil {
				return err
			}
		}
	default:
		testsList, err := listTestSuitesByName(cmd.K8s.Octopus(), args)
		if err != nil {
			return errors.Wrap(err, "unable to list test suites")
		}

		for _, t := range testsList {
			if err := cmd.printTestSuiteStatus(&t, cmd.opts.OutputFormat); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd *command) printTestSuiteStatus(testSuite *oct.ClusterTestSuite, outputFormat string) error {
	switch strings.ToLower(outputFormat) {
	case "yaml":
		d, err := yaml.Marshal(testSuite)
		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf("unable to marshal test suite '%s' to yaml",
					testSuite.GetName()))
		}
		fmt.Println(string(d))
		return nil
	case "json":
		d, err := json.MarshalIndent(testSuite, "", "\t")
		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf("unable to marshal test suite '%s' to json",
					testSuite.GetName()))
		}
		fmt.Println(string(d))
		return nil
	case "wide":
		fmt.Printf("Name:\t\t%s\r\n", testSuite.GetName())
		fmt.Printf("Concurrency:\t%d\r\n", testSuite.Spec.Concurrency)
		fmt.Printf("MaxRetries:\t%d\r\n", testSuite.Spec.MaxRetries)
		if testSuite.Status.StartTime != nil {
			fmt.Printf("StartTime:\t%s\r\n", testSuite.Status.StartTime.String())
		} else {
			fmt.Printf("StartTime:\t%s\r\n", "not started yet")
		}
		if testSuite.Status.CompletionTime != nil {
			fmt.Printf("EndTime:\t%s\r\n", testSuite.Status.CompletionTime)
		} else {
			fmt.Printf("EndTime:\t%s\r\n", "not finished yet")
		}

		fmt.Printf("Condition:\t%s\r\n", testSuite.Status.Conditions[len(testSuite.Status.Conditions)-1].Type)

		writer := test.NewTableWriter([]string{}, os.Stdout)
		for _, t := range testSuite.Status.Results {
			writer.Append([]string{t.Name, string(t.Status)})
		}
		fmt.Printf("Tests finished:\t%d/%d\r\n",
			test.GetNumberOfFinishedTests(testSuite), len(testSuite.Status.Results))
		writer.Render()
	}

	return nil
}

func listTestSuitesByName(cli octopus.OctopusInterface, names []string) ([]oct.ClusterTestSuite, error) {
	suites, err := cli.ListTestSuites()
	if err != nil {
		return nil, fmt.Errorf("unable to list test suites. E: %s", err.Error())
	}

	result := []oct.ClusterTestSuite{}
	for _, suite := range suites.Items {
		for _, tName := range names {
			if suite.ObjectMeta.Name == tName {
				result = append(result, suite)
			}
		}
	}
	return result, nil
}
