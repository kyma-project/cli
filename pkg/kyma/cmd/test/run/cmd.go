package run

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test/client"
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

	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Please make sure that you have a valid kubeconfig.")
	}

	var testSuiteName string
	if len(cmd.opts.Name) > 0 {
		testSuiteName = cmd.opts.Name
	} else {
		rnd := rand.Int31()
		testSuiteName = fmt.Sprintf("test-%d", rnd)
	}

	tNotExists, err := cmd.verifyIfTestNotExists(testSuiteName, cli)
	if err != nil {
		return err
	}
	if !tNotExists {
		return fmt.Errorf("test suite '%s' already exists\n", testSuiteName)
	}

	testDefNames := strings.Split(cmd.opts.Tests, ",")
	if err := cmd.verifyTestNames(cli, testDefNames); err != nil {
		return err
	}

	testResource := cmd.generateTestsResource(cmd.opts.Name, testDefNames)
	if err != nil {
		return err
	}

	if err := cli.CreateTestSuite(testResource); err != nil {
		return err
	}

	fmt.Printf("Test '%s' successfully created", cmd.opts.Name)
	return nil
}

func (cmd *command) verifyTestNames(cli client.TestRESTClient, testsNames []string) error {
	clusterTestDefNames, err := test.ListTestDefinitionNames(cli)
	if err != nil {
		return err
	}

	for _, tName := range testsNames {
		found := false
		for _, tDefName := range clusterTestDefNames {
			if strings.ToLower(tName) == strings.ToLower(tDefName) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("test defintion '%s' not found in the list of cluster test definitions\n", tName)
		}
	}
	return nil
}

func (cmd *command) generateTestsResource(testName string, testsNames []string) *oct.ClusterTestSuite {
	octTestDefs := test.NewTestSuite(testName)
	matchNames := []oct.TestDefReference{}
	for _, tName := range testsNames {
		matchNames = append(matchNames, oct.TestDefReference{
			Name:      tName,
			Namespace: test.TestNamespace,
		})
	}
	octTestDefs.Spec.MaxRetries = 1
	octTestDefs.Spec.Concurrency = 1
	octTestDefs.Spec.Selectors.MatchNames = matchNames

	return octTestDefs
}

func (cmd *command) verifyIfTestNotExists(suiteName string,
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
