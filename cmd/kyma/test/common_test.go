package test

import (
	"testing"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"github.com/kyma-project/cli/pkg/api/octopus"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_NewTestSuite(t *testing.T) {
	t.Parallel()
	testData := []struct {
		testName         string
		shouldFail       bool
		inputTestName    string
		inputTestOptions []SuiteOption
		expectedResult   *oct.ClusterTestSuite
	}{
		{
			testName:      "create test with existing test definition",
			shouldFail:    false,
			inputTestName: "TestOneProper",
			inputTestOptions: []SuiteOption{
				WithMatchNamesSelector(oct.TestDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "kyma-test",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				}),
				WithMatchNamesSelector(oct.TestDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "kyma-system",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				}),
				WithCount(1),
				WithMaxRetries(2),
				WithConcurrency(3),
			},
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
		{
			testName:      "create test with label expressions",
			shouldFail:    false,
			inputTestName: "TestOneProper",
			inputTestOptions: []SuiteOption{
				func() SuiteOption {
					sel, err := labels.Parse("superlabel")
					if err != nil {
						t.Errorf("could not parse expression %v", err)
					}
					return WithMatchLabelsExpression(sel)
				}(),
				func() SuiteOption {
					sel, err := labels.Parse("a=b")
					if err != nil {
						t.Errorf("could not parse expression %v", err)
					}
					return WithMatchLabelsExpression(sel)
				}(),
				WithCount(1),
				WithMaxRetries(2),
				WithConcurrency(3),
			},
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
						MatchLabelExpressions: []string{
							"superlabel",
							"a=b",
						},
					},
				},
			},
		},
		{
			testName:      "create test with label expressions and match names",
			shouldFail:    false,
			inputTestName: "TestOneProper",
			inputTestOptions: []SuiteOption{
				func() SuiteOption {
					sel, err := labels.Parse("superlabel")
					if err != nil {
						t.Errorf("could not parse expression %v", err)
					}
					return WithMatchLabelsExpression(sel)
				}(),
				WithMatchNamesSelector(oct.TestDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test1",
						Namespace: "kyma-test",
					},
					TypeMeta: metav1.TypeMeta{
						APIVersion: "",
					},
				}),
				WithCount(1),
				WithMaxRetries(2),
				WithConcurrency(3),
			},
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
						MatchLabelExpressions: []string{
							"superlabel",
						},
						MatchNames: []oct.TestDefReference{
							{
								Name:      "test1",
								Namespace: "kyma-test",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range testData {
		result := NewTestSuite(
			tt.inputTestName,
			tt.inputTestOptions...,
		)
		if tt.shouldFail {
			require.NotEqual(t, result, tt.expectedResult, tt.testName)
		} else {
			require.Equal(t, result, tt.expectedResult, tt.testName)
		}
	}
}

func Test_ListTestSuitesByName(t *testing.T) {
	t.Parallel()
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
		mCli := octopus.NewMockedOctopusRestClient(nil, &tt.inputTestSuites, nil)
		dNames, err := ListTestSuitesByName(mCli, tt.inputNames)
		if !tt.shouldFail {
			require.Nil(t, err, tt.testName)
			require.Equal(t, dNames, tt.expectedResult)
		} else {
			require.NotEqual(t, dNames, tt.expectedResult)
		}

	}
}
