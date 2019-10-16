package status

import (
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
			require.Equal(t, dNames, tt.expectedResult)
		} else {
			require.NotEqual(t, dNames, tt.expectedResult)
		}

	}
}

func Test_generateRerunCommand(t *testing.T) {
	testData := []struct {
		testName string
		input    oct.ClusterTestSuite
		expected string
	}{
		{
			testName: "1 test failed",
			input: oct.ClusterTestSuite{
				Status: oct.TestSuiteStatus{
					Results: []oct.TestResult{
						{
							Status: oct.TestFailed,
							Name:   "1",
						},
					},
				},
				Spec: oct.TestSuiteSpec{
					Concurrency: 1,
					Count:       1,
					MaxRetries:  1,
				},
			},
			expected: "kyma test run 1",
		},
		{
			testName: "1 test failed with different concurrency",
			input: oct.ClusterTestSuite{
				Status: oct.TestSuiteStatus{
					Results: []oct.TestResult{
						{
							Status: oct.TestFailed,
							Name:   "test-1",
						},
					},
				},
				Spec: oct.TestSuiteSpec{
					Concurrency: 2,
					Count:       1,
					MaxRetries:  1,
				},
			},
			expected: "kyma test run test-1 --concurrency=2",
		},
		{
			testName: "1 test failed with different max retries",
			input: oct.ClusterTestSuite{
				Status: oct.TestSuiteStatus{
					Results: []oct.TestResult{
						{
							Status: oct.TestFailed,
							Name:   "test-1",
						},
					},
				},
				Spec: oct.TestSuiteSpec{
					Concurrency: 1,
					Count:       1,
					MaxRetries:  2,
				},
			},
			expected: "kyma test run test-1 --max-retries=2",
		},
		{
			testName: "2 tests failed with different concurrency, count, max-retries",
			input: oct.ClusterTestSuite{
				Status: oct.TestSuiteStatus{
					Results: []oct.TestResult{
						{
							Status: oct.TestFailed,
							Name:   "test-1",
						},
						{
							Status: oct.TestFailed,
							Name:   "test-2",
						},
					},
				},
				Spec: oct.TestSuiteSpec{
					Concurrency: 2,
					Count:       2,
					MaxRetries:  3,
				},
			},
			expected: "kyma test run test-1 test-2 --concurrency=2 --max-retries=3 --count=2",
		},
	}

	for _, tt := range testData {
		rc := generateRerunCommand(&tt.input)
		require.Equal(t, tt.expected, rc, tt.testName)
	}
}
