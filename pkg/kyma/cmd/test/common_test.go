package test

import (
	"reflect"
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
		dNames, err := ListTestDefinitionNames(mCli)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.True(t, reflect.DeepEqual(dNames, tt.expectedResult))
		} else {
			require.False(t, reflect.DeepEqual(dNames, tt.expectedResult))
		}

	}
}

func Test_ListTestSuiteNames(t *testing.T) {
	testData := []struct {
		testName        string
		shouldFail      bool
		inputTestSuites oct.ClusterTestSuiteList
		expectedResult  []string
	}{
		{
			testName:   "correct list",
			shouldFail: false,
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
			expectedResult: []string{"test1", "test2"},
		},
		{
			testName:   "incorrect list",
			shouldFail: true,
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
			expectedResult: []string{"test1", "test3"},
		},
	}
	for _, tt := range testData {
		mCli := octopus.NewMockedOctopusRestClient(nil, &tt.inputTestSuites)
		dNames, err := ListTestSuiteNames(mCli)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.True(t, reflect.DeepEqual(dNames, tt.expectedResult))
		} else {
			require.False(t, reflect.DeepEqual(dNames, tt.expectedResult))
		}

	}
}

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
		dNames, err := ListTestSuitesByName(mCli, tt.inputNames)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.True(t, reflect.DeepEqual(dNames, tt.expectedResult))
		} else {
			require.False(t, reflect.DeepEqual(dNames, tt.expectedResult))
		}

	}
}

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
					Labels: map[string]string{
						"requires-testing-bundle": "true",
						"requires-test-user":      "true",
					},
				},
			},
		},
	}

	for _, tt := range testData {
		result := NewTestSuite(tt.inputName)
		require.True(t, reflect.DeepEqual(result, tt.expectedResult), tt.testName)
	}
}
