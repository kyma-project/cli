package definitions

import (
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_ListTestDefinitionNames(t *testing.T) {
	testData := []struct {
		testName         string
		shouldFail       bool
		inputDefinitions oct.TestDefinitionList
		expectedResult   []string
	}{
		{
			testName:   "correct list",
			shouldFail: false,
			inputDefinitions: oct.TestDefinitionList{
				Items: []oct.TestDefinition{
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
			expectedResult: []string{"test1", "test2"},
		},
		{
			testName:   "incorrect list",
			shouldFail: true,
			inputDefinitions: oct.TestDefinitionList{
				Items: []oct.TestDefinition{
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
			expectedResult: []string{"test1", "test3"},
		},
	}

	for _, tt := range testData {
		mCli := octopus.NewMockedOctopusRestClient(&tt.inputDefinitions, nil)
		dNames, err := listTestDefinitionNames(mCli)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, dNames, tt.expectedResult)
		} else {
			require.NotEqual(t, dNames, tt.expectedResult)
		}

	}
}
