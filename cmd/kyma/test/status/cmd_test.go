package status

import (
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/stretchr/testify/require"
)

func Test_generateRerunCommand(t *testing.T) {
	t.Parallel()
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
