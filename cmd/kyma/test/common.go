package test

import (
	"io"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/kyma-project/cli/pkg/api/octopus"
)

type SuiteOption func(suite *oct.ClusterTestSuite)

func NewTestSuite(name string, options ...SuiteOption) *oct.ClusterTestSuite {
	cts := &oct.ClusterTestSuite{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "testing.kyma-project.io/v1alpha1",
			Kind:       "ClusterTestSuite",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	for _, opt := range options {
		opt(cts)
	}
	return cts
}

func WithMaxRetries(maxRetries int64) SuiteOption {
	return func(suite *oct.ClusterTestSuite) {
		suite.Spec.MaxRetries = maxRetries
	}
}

func WithConcurrency(concurrency int64) SuiteOption {
	return func(suite *oct.ClusterTestSuite) {
		suite.Spec.Concurrency = concurrency
	}
}

func WithCount(count int64) SuiteOption {
	return func(suite *oct.ClusterTestSuite) {
		suite.Spec.Count = count
	}
}

func WithMatchNamesSelector(testDefinition oct.TestDefinition) SuiteOption {
	return func(suite *oct.ClusterTestSuite) {
		suite.Spec.Selectors.MatchNames = append(suite.Spec.Selectors.MatchNames, oct.TestDefReference{
			Name:      testDefinition.GetName(),
			Namespace: testDefinition.GetNamespace(),
		})
	}
}

func WithMatchLabelsExpression(selector labels.Selector) SuiteOption {
	return func(suite *oct.ClusterTestSuite) {
		suite.Spec.Selectors.MatchLabelExpressions = append(suite.Spec.Selectors.MatchLabelExpressions, selector.String())
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
			result++
		}
	}
	return result
}

func ListTestSuitesByName(cli octopus.Interface, names []string) ([]oct.ClusterTestSuite, error) {
	suites, err := cli.ListTestSuites(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to list test suites")
	}

	indexedNames := map[string]struct{}{}
	for _, n := range names {
		indexedNames[n] = struct{}{}
	}

	var result []oct.ClusterTestSuite
	for _, suite := range suites.Items {
		if _, found := indexedNames[suite.Name]; found {
			result = append(result, suite)
		}
	}

	return result, nil
}
