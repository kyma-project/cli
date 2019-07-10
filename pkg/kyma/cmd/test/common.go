package test

import (
	"fmt"
	"io"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NamespaceForTests = "kyma-system"

func ListTestDefinitionNames(cli octopus.OctopusInterface) ([]string, error) {
	defs, err := cli.ListTestDefinitions()
	if err != nil {
		return nil, fmt.Errorf("unable to list test definitions. E: %s", err.Error())
	}

	var result = make([]string, len(defs.Items))
	for i := 0; i < len(defs.Items); i++ {
		result[i] = defs.Items[i].GetName()
	}
	return result, nil
}

func ListTestSuiteNames(cli octopus.OctopusInterface) ([]string, error) {
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

func ListTestSuitesByName(cli octopus.OctopusInterface, names []string) ([]oct.ClusterTestSuite, error) {
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

func NewTestSuite(name string) *oct.ClusterTestSuite {
	return &oct.ClusterTestSuite{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "testing.kyma-project.io/v1alpha1",
			Kind:       "ClusterTestSuite",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: NamespaceForTests,
		},
	}
}

func NewTableWriter(columns []string, out io.Writer) *tablewriter.Table {
	writer := tablewriter.NewWriter(out)
	writer.SetBorder(false)
	writer.SetHeader(columns)
	writer.SetAlignment(tablewriter.ALIGN_LEFT)
	writer.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	writer.SetHeaderLine(false)
	writer.SetRowSeparator("")
	writer.SetCenterSeparator("")
	writer.SetColumnSeparator("")
	return writer
}

func GetNumberOfFinishedTests(testSuite *oct.ClusterTestSuite) int {
	result := 0
	for _, t := range testSuite.Status.Results {
		if t.Status == oct.TestFailed || t.Status == oct.TestSucceeded || t.Status == oct.TestSkipped {
			result += 1
		}
	}
	return result
}
