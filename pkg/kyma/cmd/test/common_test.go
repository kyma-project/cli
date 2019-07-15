package test

import (
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_NewTestSuite(t *testing.T) {
	testData := []struct {
		testName       string
		inputName      string
		expectedResult *oct.ClusterTestSuite
	}{
		{
			testName:  "correct test suite name",
			inputName: "test1",
			expectedResult: &oct.ClusterTestSuite{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "testing.kyma-project.io/v1alpha1",
					Kind:       "ClusterTestSuite",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: NamespaceForTests,
				},
			},
		},
	}

	for _, tt := range testData {
		result := NewTestSuite(tt.inputName)
		require.Equal(t, result, tt.expectedResult, tt.testName)
	}
}
