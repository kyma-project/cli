package test

import (
	"io"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NamespaceForTests = "kyma-system"

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
