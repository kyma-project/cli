package del

import (
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_deleteTestSuite(t *testing.T) {
	testData := []struct {
		testName              string
		shouldFail            bool
		testSuitesAvailable   *oct.ClusterTestSuiteList
		testSuiteNameToDelete string
	}{
		{
			testName:              "remove existing test",
			shouldFail:            false,
			testSuiteNameToDelete: "TEST_NAME_1",
			testSuitesAvailable: &oct.ClusterTestSuiteList{
				Items: []oct.ClusterTestSuite{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "TEST_NAME_1",
						},
						TypeMeta: metav1.TypeMeta{
							APIVersion: "",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "TEST_NAME_2",
						},
						TypeMeta: metav1.TypeMeta{
							APIVersion: "",
						},
					},
				},
			},
		},
		{
			testName:              "remove not existing test",
			shouldFail:            true,
			testSuiteNameToDelete: "TEST_NAME_43",
			testSuitesAvailable: &oct.ClusterTestSuiteList{
				Items: []oct.ClusterTestSuite{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "TEST_NAME_1",
						},
						TypeMeta: metav1.TypeMeta{
							APIVersion: "",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "TEST_NAME_2",
						},
						TypeMeta: metav1.TypeMeta{
							APIVersion: "",
						},
					},
				},
			},
		},
	}

	for _, tt := range testData {
		mCli := octopus.NewMockedOctopusRestClient(nil, tt.testSuitesAvailable)
		err := deleteTestSuite(mCli, tt.testSuiteNameToDelete)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
		} else {
			require.NotNil(t, err, tt.testName)
		}
	}
}
