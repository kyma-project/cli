package run

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/client"
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
		Use:     "run",
		Short:   "Run tests on a running Kyma cluster",
		Long:    `Run tests on a running Kyma cluster`,
		RunE:    func(_ *cobra.Command, _ []string) error { return cmd.Run() },
		Aliases: []string{"r"},
	}

	cobraCmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name for the new test suite")
	cobraCmd.Flags().StringVarP(&o.Tests, "tests", "s", "", "Test names to execute. Example: --tests=cluster-users-test,test-api-controller-acceptance")
	cobraCmd.Flags().BoolVarP(&o.Wait, "wait", "w", false, "Wait for test execution to finish")
	cobraCmd.Flags().IntVarP(&o.Timeout, "timeout", "t", 120, "Timeout for test execution (in seconds)")
	return cobraCmd
}

func (cmd *command) Run() error {
	var err error

	cli, err := client.NewTestRESTClient(10 * time.Second)
	if err != nil {
		return fmt.Errorf("unable to create test REST client. E: %s", err)
	}

	var testSuiteName string
	if len(cmd.opts.Name) > 0 {
		testSuiteName = cmd.opts.Name
	} else {
		rand.Seed(time.Now().UTC().UnixNano())
		rnd := rand.Int31()
		testSuiteName = fmt.Sprintf("test-%d", rnd)
	}

	tNotExists, err := verifyIfTestNotExists(testSuiteName, cli)
	if err != nil {
		return err
	}
	if !tNotExists {
		return fmt.Errorf("test suite '%s' already exists\n", testSuiteName)
	}

	var testDefToApply []oct.TestDefinition

	clusterTestDefs, err := cli.ListTestDefinitions()
	if err != nil {
		return fmt.Errorf("unable to get list of test definitions. E: %s",
			err.Error())
	}

	if cmd.opts.Tests != "" {
		testDefNames := strings.Split(cmd.opts.Tests, ",")

		var err error
		if testDefToApply, err = matchTestDefinitionNames(testDefNames,
			clusterTestDefs.Items); err != nil {
			return err
		}
	}

	testResource := generateTestsResource(testSuiteName, testDefToApply)
	if err != nil {
		return err
	}

	if err := cli.CreateTestSuite(testResource); err != nil {
		return err
	}

	fmt.Printf("Test '%s' successfully created\r\n", testSuiteName)
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

func verifyIfTestNotExists(suiteName string,
	cli client.TestRESTClient) (bool, error) {
	tests, err := test.ListTestSuiteNames(cli)
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
