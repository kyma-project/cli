package run

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/internal/kube"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	if cmd.K8s, err = kube.NewFromConfig("", cmd.KubeconfigPath); err != nil {
		return errors.Wrap(err, "Could not initialize the Kubernetes client. Please make sure that you have a valid kubeconfig.")
	}

	var testSuiteName string
	if cmd.opts.Name == "" {
		testSuiteName = cmd.opts.Name
	} else {
		rnd := rand.Int31()
		testSuiteName = fmt.Sprintf("test-%d", rnd)
	}

	tNotExists, err := cmd.verifyIfTestNotExists()
	if err != nil {
		return err
	}
	if !tNotExists {
		return fmt.Errorf("test suite '%s' already exists\r\n", testSuiteName)
	}

	testDefNames := strings.Split(cmd.opts.Tests, ",")
	if err := cmd.verifyTestNames(testDefNames); err != nil {
		return err
	}

	testResource := cmd.generateTestsResources(cmd.opts.Name, testDefNames)
	if err != nil {
		return err
	}

	if err := cmd.applyTestResource(testResource); err != nil {
		return err
	}

	//TODO: remote the next line
	fmt.Println(testResource)

	return nil
}

func (cmd *command) applyTestResource(resource *oct.ClusterTestSuite) error {
	errorF := func(err error) error {
		return fmt.Errorf("unable to apply test resource. E: %s", err.Error())
	}

	deploymentResource := schema.GroupVersionResource{
		Group:    "testing.kyma-project.io",
		Version:  "v1alpha1",
		Resource: "ClusterTestSuite",
	}

	fmt.Println(deploymentResource)
	marshalled, err := json.Marshal(resource)
	if err != nil {
		return errorF(err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(marshalled, &data); err != nil {
		return errorF(err)
	}
	fmt.Println(data)
	deployment := &unstructured.Unstructured{
		Object: data,
	}

	_, err = cmd.K8s.Dynamic().Resource(deploymentResource).Namespace("kyma-system").Create(deployment, metav1.CreateOptions{})
	if err != nil {
		return errorF(err)
	}
	fmt.Printf("Test successfully added")

	return nil
}

func (cmd *command) verifyTestNames(testsNames []string) error {
	clusterTestDefNames, err := cmd.getClusterTestDefinitionNames()
	if err != nil {
		return err
	}

	for _, tName := range testsNames {
		for _, tDefName := range clusterTestDefNames {
			if strings.ToLower(tName) != strings.ToLower(tDefName) {
				return fmt.Errorf("give test defintion '%s' not found in the list of cluster test definitions\r\n", tName)
			}
		}
	}
	return nil
}

func (cmd *command) newTestSuite(name string) *oct.ClusterTestSuite {
	return &oct.ClusterTestSuite{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "testing.kyma-project.io/v1alpha1",
			Kind:       "ClusterTestSuite",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kyma-system",
		},
	}
}

func (cmd *command) generateTestsResources(testName string, testsNames []string) *oct.ClusterTestSuite {
	octTestDefs := cmd.newTestSuite(testName)
	matchNames := []oct.TestDefReference{}
	for _, tName := range testsNames {
		matchNames = append(matchNames, oct.TestDefReference{
			Name:      tName,
			Namespace: "kyma-system",
		})
	}
	octTestDefs.Spec.MaxRetries = 1
	octTestDefs.Spec.Concurrency = 1
	octTestDefs.Spec.Selectors.MatchNames = matchNames

	return octTestDefs
}

func (cmd *command) getClusterTestDefinitionNames() ([]string, error) {
	return nil, nil
}

func (cmd *command) verifyIfTestNotExists() (bool, error) {

	res, err := cmd.Kubectl().RunCmd("-n", "kyma-system", "get", test.TestCrdDefinition)
	if err != nil {
		return false, err
	}
	if strings.Contains(res, "alredy exists") {
		return false, nil
	}
	return true, nil
}
