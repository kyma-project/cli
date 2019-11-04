package run

import (
	"testing"
	"time"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/kyma-project/cli/pkg/step/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
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
					Name: "TestOneProper",
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
		}, nil)
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
		mCli := octopus.NewMockedOctopusRestClient(nil, &tt.inputTestSuites, nil)
		dNames, err := listTestSuiteNames(mCli)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, dNames, tt.expectedResult)
		} else {
			require.NotEqual(t, dNames, tt.expectedResult)
		}

	}
}

func Test_WaitForTestSuite(t *testing.T) {
	tests := map[string]struct {
		expMessages      []string
		statusesSequence []oct.TestSuiteStatus
	}{
		"waiting for test suite ended with success": {
			expMessages: []string{
				"0 out of 1 test(s) have finished (Succeeded: 0, Failed: 0, Skipped: 0)...",
				"1 out of 1 test(s) have finished (Succeeded: 1, Failed: 0, Skipped: 0)...",
				"Test suite 'fix-test' execution succeeded",
			},
			statusesSequence: []oct.TestSuiteStatus{
				statusRunning(),
				statusTestSucceeded(),
				statusSuiteSucceeded(),
			},
		},
		"waiting for test suite ended with failure": {
			expMessages: []string{
				"0 out of 1 test(s) have finished (Succeeded: 0, Failed: 0, Skipped: 0)...",
				"1 out of 1 test(s) have finished (Succeeded: 0, Failed: 1, Skipped: 0)...",
				"Test suite 'fix-test' execution failed",
			},
			statusesSequence: []oct.TestSuiteStatus{
				statusRunning(),
				statusTestFailed(),
				statusSuiteFailed(),
			},
		},
		"waiting for test suite ended with error": {
			expMessages: []string{
				"0 out of 1 test(s) have finished (Succeeded: 0, Failed: 0, Skipped: 0)...",
				"1 out of 1 test(s) have finished (Succeeded: 0, Failed: 0, Skipped: 1)...",
				"Test suite 'fix-test' execution errored",
			},
			statusesSequence: []oct.TestSuiteStatus{
				statusRunning(),
				statusTestSkipped(),
				statusSuiteError(),
			},
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fixTestSuite := oct.ClusterTestSuite{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fix-test",
				},
			}

			inputTestSuites := oct.ClusterTestSuiteList{
				Items: []oct.ClusterTestSuite{
					*fixTestSuite.DeepCopy(),
				},
			}

			fakeWatcher := watch.NewFake()
			mStep := &mocks.Step{}
			mCli := octopus.NewMockedOctopusRestClient(nil, &inputTestSuites, fakeWatcher)

			// when
			waitForTestSuiteDone := make(chan struct{}, 1)
			var waitErr error
			go func() {
				waitErr = waitForTestSuite(mCli, fixTestSuite.Name, clusterTestSuiteCompleted(mStep), 5*time.Second)
				waitForTestSuiteDone <- struct{}{}
			}()

			// simulates the octopus controller - emit events with modified statuses
			for _, newStatus := range tc.statusesSequence {
				modified := fixTestSuite.DeepCopy()
				modified.Status = newStatus
				fakeWatcher.Modify(modified)

			}

			// then
			waitForChanAtMost(t, waitForTestSuiteDone, time.Second)
			require.NoError(t, waitErr)
			require.Equal(t, len(tc.expMessages), len(mStep.Statuses()))

			for idx, msg := range mStep.Statuses() {
				assert.Equal(t, tc.expMessages[idx], msg)
			}
			assert.Empty(t, mStep.Errors())
		})
	}
}

func waitForChanAtMost(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("Waiting for channel result failed in given timeout %v.", timeout)
	}
}

func statusRunning() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestRunning, oct.SuiteRunning)
}

func statusTestSucceeded() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestSucceeded, oct.SuiteRunning)
}

func statusSuiteSucceeded() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestSucceeded, oct.SuiteSucceeded)
}

func statusSuiteFailed() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestFailed, oct.SuiteFailed)
}

func statusTestFailed() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestFailed, oct.SuiteRunning)
}

func statusSuiteError() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestSkipped, oct.SuiteError)
}

func statusTestSkipped() oct.TestSuiteStatus {
	return testSuiteStatus(oct.TestSkipped, oct.SuiteRunning)
}

func testSuiteStatus(testStatus oct.TestStatus, suiteStatus oct.TestSuiteConditionType) oct.TestSuiteStatus {
	return oct.TestSuiteStatus{
		Results: []oct.TestResult{
			{Status: testStatus},
		},
		Conditions: []oct.TestSuiteCondition{
			{
				Type:   suiteStatus,
				Status: oct.StatusTrue,
			},
		},
	}
}
