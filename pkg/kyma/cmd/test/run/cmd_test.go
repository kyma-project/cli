package run

import (
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/kyma-project/cli/pkg/kyma/cmd/test"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_matchTestDefinitionNames(t *testing.T) {
	testData := []struct {
		testName        string
		shouldFail      bool
		testNames       []string
		testDefinitions []oct.TestDefinition
		result          []oct.TestDefinition
	}{
		{
			testName:   "match all tests",
			shouldFail: false,
			testNames:  []string{"test1", "test2"},
			testDefinitions: []oct.TestDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				},
			},
			result: []oct.TestDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test1",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				},
			},
		},
	}

	for _, tt := range testData {
		result, err := matchTestDefinitionNames(tt.testNames, tt.testDefinitions)
		if tt.shouldFail {
			require.NotNil(t, err, tt.testName)
		} else {
			require.Nil(t, err, tt.testName)
			require.Equal(t, result, tt.result, tt.testName)
		}
	}
}

func Test_generateTestsResource(t *testing.T) {
	testData := []struct {
		testName             string
		shouldFail           bool
		inputTestName        string
		inputTestDefinitions []oct.TestDefinition
		inputExecutionCount  int64
		inputMaxRetires      int64
		inputConcurrency     int64
		expectedResult       *oct.ClusterTestSuite
	}{
		{
			testName:      "create test with existing test definition",
			shouldFail:    false,
			inputTestName: "TestOneProper",
			inputTestDefinitions: []oct.TestDefinition{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "kyma-test",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "kyma-system",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				},
			},
			inputExecutionCount: 1,
			inputMaxRetires:     2,
			inputConcurrency:    3,
			expectedResult: &oct.ClusterTestSuite{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "testing.kyma-project.io/v1alpha1",
					Kind:       "ClusterTestSuite",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "TestOneProper",
					Namespace: test.NamespaceForTests,
				},
				Spec: oct.TestSuiteSpec{
					Count:       1,
					MaxRetries:  2,
					Concurrency: 3,
					Selectors: oct.TestsSelector{
						MatchNames: []oct.TestDefReference{
							{
								Name:      "test1",
								Namespace: "kyma-test",
							},
							{
								Name:      "test2",
								Namespace: "kyma-system",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testData {
		result := generateTestsResource(
			tt.inputTestName,
			tt.inputExecutionCount,
			tt.inputMaxRetires,
			tt.inputConcurrency,
			tt.inputTestDefinitions,
		)
		if tt.shouldFail {
			require.NotEqual(t, result, tt.expectedResult, tt.testName)
		} else {
			require.Equal(t, result, tt.expectedResult, tt.testName)
		}
	}
}

func Test_verifyIfTestNotExists(t *testing.T) {
	testData := []struct {
		testName       string
		inputSuiteName string
		inputSuites    []oct.ClusterTestSuite
		expectedExists bool
	}{
		{
			testName:       "verify existing test",
			inputSuiteName: "test1",
			inputSuites: []oct.ClusterTestSuite{
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
			expectedExists: false,
		},
		{
			testName:       "verify non-existing test",
			inputSuiteName: "test1",
			inputSuites: []oct.ClusterTestSuite{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test2",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test3",
					},
				},
			},
			expectedExists: true,
		},
	}

	for _, tt := range testData {
		mCli := octopus.NewMockedOctopusRestClient(nil, &oct.ClusterTestSuiteList{
			Items: tt.inputSuites,
		})
		tExists, _ := verifyIfTestNotExists(tt.inputSuiteName, mCli)
		require.Equal(t, tExists, tt.expectedExists)
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
		dNames, err := listTestSuiteNames(mCli)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, dNames, tt.expectedResult)
		} else {
			require.NotEqual(t, dNames, tt.expectedResult)
		}

	}
}
