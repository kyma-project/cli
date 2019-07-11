package run

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/api/octopus"
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
		Use:     "run <test-definition-1> <test-defintion-2> ... <test-definition-N>",
		Short:   "Run tests on a running Kyma cluster",
		Long:    `Run tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, args []string) error { return cmd.Run(args) },
		Aliases: []string{"r"},
	}

	cobraCmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the new test suite")
	cobraCmd.Flags().BoolVarP(&o.Wait, "wait", "w", false, "Wait for test execution to finish")
	cobraCmd.Flags().DurationVarP(&o.Timeout, "timeout", "t", 120*time.Second, "Time-out for test execution (in seconds)")
	return cobraCmd
}

func (cmd *command) Run(args []string) error {
	var err error
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "could not initialize the Kubernetes client. Make sure that your kubeconfig is valid.")
	}

	var testSuiteName string
	if len(cmd.opts.Name) > 0 {
		testSuiteName = cmd.opts.Name
	} else {
		rand.Seed(time.Now().UTC().UnixNano())
		rnd := rand.Int31()
		testSuiteName = fmt.Sprintf("test-%d", rnd)
	}

	tNotExists, err := verifyIfTestNotExists(testSuiteName, cmd.K8s.Octopus())
	if err != nil {
		return err
	}
	if !tNotExists {
		return fmt.Errorf("test suite '%s' already exists\n", testSuiteName)
	}

	clusterTestDefs, err := cmd.K8s.Octopus().ListTestDefinitions()
	if err != nil {
		return errors.Wrap(err, "unable to get list of test definitions")
	}

	var testDefToApply []oct.TestDefinition
	if len(args) == 0 {
		testDefToApply = clusterTestDefs.Items
	} else {
		if testDefToApply, err = matchTestDefinitionNames(args,
			clusterTestDefs.Items); err != nil {
			return err
		}
	}

	testResource := generateTestsResource(testSuiteName, testDefToApply)
	if err != nil {
		return err
	}

	if err := cmd.K8s.Octopus().CreateTestSuite(testResource); err != nil {
		return err
	}

	fmt.Printf("test '%s' successfully created\r\n", testSuiteName)
	return nil
}

func matchTestDefinitionNames(testNames []string,
	testDefs []oct.TestDefinition) ([]oct.TestDefinition, error) {
	result := []oct.TestDefinition{}
	for _, tName := range testNames {
		found := false
		for _, tDef := range testDefs {
			if strings.ToLower(tName) == strings.ToLower(tDef.GetName()) {
				found = true
				result = append(result, tDef)
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("test defintion '%s' not found in the list of cluster test definitions\n", tName)
		}
	}
	return result, nil
}

func generateTestsResource(testName string,
	testDefinitions []oct.TestDefinition) *oct.ClusterTestSuite {

	octTestDefs := test.NewTestSuite(testName)
	matchNames := []oct.TestDefReference{}
	for _, td := range testDefinitions {
		matchNames = append(matchNames, oct.TestDefReference{
			Name:      td.GetName(),
			Namespace: td.GetNamespace(),
		})
	}
	octTestDefs.Spec.MaxRetries = 1
	octTestDefs.Spec.Concurrency = 1
	octTestDefs.Spec.Selectors.MatchNames = matchNames

	return octTestDefs
}

func listTestSuiteNames(cli octopus.OctopusInterface) ([]string, error) {
	suites, err := cli.ListTestSuites()
	if err != nil {
		return nil, fmt.Errorf("unable to list test suites. E: %s", err.Error())
	}

	var result = make([]string, len(suites.Items))
	for i := 0; i < len(suites.Items); i++ {
		result[i] = suites.Items[i].GetName()
	}
	return result, nil
}

func verifyIfTestNotExists(suiteName string,
	cli octopus.OctopusInterface) (bool, error) {
	tests, err := listTestSuiteNames(cli)
	if err != nil {
		return false, err
	}
	for _, t := range tests {
		if t == suiteName {
			return false, nil
		}
	}
	return true, nil
}
