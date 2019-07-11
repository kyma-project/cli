package status

import (
	"reflect"
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ListTestSuitesByName(t *testing.T) {
	testData := []struct {
		testName        string
		shouldFail      bool
		inputTestSuites oct.ClusterTestSuiteList
		inputNames      []string
		expectedResult  []oct.ClusterTestSuite
	}{
		{
			testName:   "correct list",
			shouldFail: false,
			inputNames: []string{"test1", "test2"},
			inputTestSuites: oct.ClusterTestSuiteList{
				Items: []oct.ClusterTestSuite{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test2",
						},
					},
				},
			},
			expectedResult: []oct.ClusterTestSuite{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
					},
				},
			},
		},
		{
			testName:   "incorrect list",
			shouldFail: true,
			inputNames: []string{"test1", "test3"},
			inputTestSuites: oct.ClusterTestSuiteList{
				Items: []oct.ClusterTestSuite{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test2",
						},
					},
				},
			},
			expectedResult: []oct.ClusterTestSuite{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
					},
				},
			},
		},
	}
	for _, tt := range testData {
		mCli := octopus.NewMockedOctopusRestClient(nil, &tt.inputTestSuites)
		dNames, err := listTestSuitesByName(mCli, tt.inputNames)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.True(t, reflect.DeepEqual(dNames, tt.expectedResult))
		} else {
			require.False(t, reflect.DeepEqual(dNames, tt.expectedResult))
		}

	}
}
