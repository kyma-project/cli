package diagnostics_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNewModuleCustomResourceStateCollector(t *testing.T) {
	// Given
	fakeClient := &fake.KubeClient{}
	var writer bytes.Buffer
	verbose := true

	// When
	collector := diagnostics.NewModuleCustomResourceStateCollector(fakeClient, &writer, verbose)

	// Then
	assert.NotNil(t, collector)
}

func TestModuleCustomResourceStateCollector_WriteVerboseError(t *testing.T) {
	testCases := []struct {
		name           string
		verbose        bool
		err            error
		message        string
		expectedOutput string
	}{
		{
			name:           "Should write error when verbose is true",
			verbose:        true,
			err:            errors.New("test error"),
			message:        "Test error message",
			expectedOutput: "Test error message: test error\n",
		},
		{
			name:           "Should not write error when verbose is false",
			verbose:        false,
			err:            errors.New("test error"),
			message:        "Test error message",
			expectedOutput: "",
		},
		{
			name:           "Should not write error when error is nil",
			verbose:        true,
			err:            nil,
			message:        "Test error message",
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			collector := diagnostics.NewModuleCustomResourceStateCollector(nil, &writer, tc.verbose)

			// When
			collector.WriteVerboseError(tc.err, tc.message)

			// Then
			assert.Equal(t, tc.expectedOutput, writer.String())
		})
	}
}

func TestModuleCustomResourceStateCollector_Run(t *testing.T) {
	testCases := []struct {
		name                      string
		mockCoreInstalled         []kyma.ModuleTemplate
		mockCoreInstalledErr      error
		mockCommunityInstalled    []kyma.ModuleTemplate
		mockCommunityInstalledErr error
		mockReturnList            *unstructured.UnstructuredList
		mockReturnListErr         error
		expected                  []diagnostics.ModuleCustomResourceState
		expectedVerboseOutput     string
	}{
		{
			name: "Should collect module states successfully",
			mockCoreInstalled: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test-core-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestCoreModule",
								"metadata": map[string]any{
									"name":      "test-core-module",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			mockCommunityInstalled: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test-community-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestCommunityModule",
								"metadata": map[string]any{
									"name":      "test-community-module",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			mockReturnList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]any{
							"apiVersion": "test.io/v1",
							"kind":       "TestModule",
							"status": map[string]any{
								"state": "Error",
								"conditions": []any{
									map[string]any{
										"type":               "Ready",
										"status":             "False",
										"reason":             "ModuleError",
										"message":            "Module failed",
										"lastTransitionTime": "2023-09-22T10:30:00Z",
									},
								},
							},
						},
					},
					{
						Object: map[string]any{
							"apiVersion": "test.io/v1",
							"kind":       "TestModule",
							"status": map[string]any{
								"state": "Ready",
							},
						},
					},
				},
			},
			expected: []diagnostics.ModuleCustomResourceState{
				{
					ApiVersion: "test.io/v1",
					Kind:       "TestModule",
					State:      "Error",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							Reason:             "ModuleError",
							Message:            "Module failed",
							LastTransitionTime: metav1.NewTime(time.Date(2023, 9, 22, 10, 30, 0, 0, time.UTC)),
						},
					},
				},
				{
					ApiVersion: "test.io/v1",
					Kind:       "TestModule",
					State:      "Error",
					Conditions: []metav1.Condition{
						{
							Type:               "Ready",
							Status:             metav1.ConditionFalse,
							Reason:             "ModuleError",
							Message:            "Module failed",
							LastTransitionTime: metav1.NewTime(time.Date(2023, 9, 22, 10, 30, 0, 0, time.UTC)),
						},
					},
				},
			},
		},
		{
			name:                 "Should handle core modules error",
			mockCoreInstalledErr: errors.New("failed to get core modules"),
			mockCommunityInstalled: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test-community-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestCommunityModule",
								"metadata": map[string]any{
									"name":      "test-community-module",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			mockReturnList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]any{
							"apiVersion": "test.io/v1",
							"kind":       "TestModule",
							"status": map[string]any{
								"state": "Error",
							},
						},
					},
				},
			},
			expected: []diagnostics.ModuleCustomResourceState{
				{
					ApiVersion: "test.io/v1",
					Kind:       "TestModule",
					State:      "Error",
					Conditions: []metav1.Condition{},
				},
			},
			expectedVerboseOutput: "Failed to list core modules: failed to get core modules\n",
		},
		{
			name: "Should handle community modules error",
			mockCoreInstalled: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test-core-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestCoreModule",
								"metadata": map[string]any{
									"name":      "test-core-module",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			mockCommunityInstalledErr: errors.New("failed to get community modules"),
			mockReturnList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]any{
							"apiVersion": "test.io/v1",
							"kind":       "TestModule",
							"status": map[string]any{
								"state": "Error",
							},
						},
					},
				},
			},
			expected: []diagnostics.ModuleCustomResourceState{
				{
					ApiVersion: "test.io/v1",
					Kind:       "TestModule",
					State:      "Error",
					Conditions: []metav1.Condition{},
				},
			},
			expectedVerboseOutput: "Failed to list community modules: failed to get community modules\n",
		},
		{
			name:                   "Should handle no installed modules",
			mockCoreInstalled:      []kyma.ModuleTemplate{},
			mockCommunityInstalled: []kyma.ModuleTemplate{},
			expected:               []diagnostics.ModuleCustomResourceState{},
			expectedVerboseOutput:  "",
		},
		{
			name: "Should handle data resource query error",
			mockCoreInstalled: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test-core-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestCoreModule",
								"metadata": map[string]any{
									"name":      "test-core-module",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			mockCommunityInstalled: []kyma.ModuleTemplate{},
			mockReturnListErr:      errors.New("resource query failed"),
			expected:               []diagnostics.ModuleCustomResourceState{},
			expectedVerboseOutput:  "Failed to get data resource for test-core-module module: resource query failed\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer
			fakeRootlessDynamic := &fake.RootlessDynamicClient{
				ReturnListObjs: tc.mockReturnList,
				ReturnErr:      tc.mockReturnListErr,
			}
			fakeClient := &fake.KubeClient{
				TestRootlessDynamicInterface: fakeRootlessDynamic,
			}
			fakeModuleRepo := &modulesfake.ModuleTemplatesRepo{
				ReturnCoreInstalled:      tc.mockCoreInstalled,
				CoreInstalledErr:         tc.mockCoreInstalledErr,
				ReturnCommunityInstalled: tc.mockCommunityInstalled,
				CommunityInstalledErr:    tc.mockCommunityInstalledErr,
			}

			collector := diagnostics.NewModuleCustomResourceStateCollectorWithRepo(
				fakeClient,
				fakeModuleRepo,
				&writer,
				true,
			)

			// When
			result := collector.Run(context.Background())

			// Then
			assert.Equal(t, tc.expected, result)
			if tc.expectedVerboseOutput != "" {
				assert.Contains(t, writer.String(), tc.expectedVerboseOutput)
			}
		})
	}
}

func TestModuleCustomResourceStateCollector_Run_Integration(t *testing.T) {
	t.Run("Should handle mixed ready and non-ready modules from multiple sources", func(t *testing.T) {
		// Given
		var writer bytes.Buffer

		// Mock data with both core and community modules
		coreModule := kyma.ModuleTemplate{
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: "core-module",
				Data: unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "core.io/v1",
						"kind":       "CoreModule",
						"metadata": map[string]any{
							"name":      "core-module",
							"namespace": "core-namespace",
						},
					},
				},
			},
		}

		communityModule := kyma.ModuleTemplate{
			Spec: kyma.ModuleTemplateSpec{
				ModuleName: "community-module",
				Data: unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "community.io/v1",
						"kind":       "CommunityModule",
						"metadata": map[string]any{
							"name":      "community-module",
							"namespace": "community-namespace",
						},
					},
				},
			},
		}

		// Mock resources with different states
		mockResourceList := &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{
				{
					Object: map[string]any{
						"apiVersion": "test.io/v1",
						"kind":       "TestModule",
						"status": map[string]any{
							"state": "Processing",
							"conditions": []any{
								map[string]any{
									"type":               "Ready",
									"status":             "False",
									"reason":             "StillProcessing",
									"message":            "Module is still processing",
									"lastTransitionTime": "2023-09-22T10:30:00Z",
								},
							},
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "test.io/v1",
						"kind":       "ReadyModule",
						"status": map[string]any{
							"state": "Ready",
						},
					},
				},
				{
					Object: map[string]any{
						"apiVersion": "test.io/v1",
						"kind":       "ErrorModule",
						"status": map[string]any{
							"state": "Error",
							"conditions": []any{
								map[string]any{
									"type":               "Ready",
									"status":             "False",
									"reason":             "ModuleFailed",
									"message":            "Module installation failed",
									"lastTransitionTime": "2023-09-22T10:25:00Z",
								},
								map[string]any{
									"type":               "Available",
									"status":             "False",
									"reason":             "ServiceUnavailable",
									"message":            "Service is not available",
									"lastTransitionTime": "2023-09-22T10:20:00Z",
								},
							},
						},
					},
				},
			},
		}

		fakeRootlessDynamic := &fake.RootlessDynamicClient{
			ReturnListObjs: mockResourceList,
		}
		fakeClient := &fake.KubeClient{
			TestRootlessDynamicInterface: fakeRootlessDynamic,
		}
		fakeModuleRepo := &modulesfake.ModuleTemplatesRepo{
			ReturnCoreInstalled:      []kyma.ModuleTemplate{coreModule},
			ReturnCommunityInstalled: []kyma.ModuleTemplate{communityModule},
		}

		collector := diagnostics.NewModuleCustomResourceStateCollectorWithRepo(
			fakeClient,
			fakeModuleRepo,
			&writer,
			false, // Test with verbose disabled
		)

		// When
		result := collector.Run(context.Background())

		// Then - should collect only non-ready modules (Processing and Error states)
		require.Len(t, result, 4) // 2 modules × 2 non-ready resources each

		// Verify Processing module states
		processingStates := 0
		errorStates := 0
		for _, state := range result {
			if state.State == "Processing" {
				processingStates++
				assert.Len(t, state.Conditions, 1)
				assert.Equal(t, "Ready", state.Conditions[0].Type)
				assert.Equal(t, metav1.ConditionFalse, state.Conditions[0].Status)
				assert.Equal(t, "StillProcessing", state.Conditions[0].Reason)
			}
			if state.State == "Error" {
				errorStates++
				assert.Len(t, state.Conditions, 2)
			}
		}

		assert.Equal(t, 2, processingStates) // Both core and community modules have Processing resources
		assert.Equal(t, 2, errorStates)      // Both core and community modules have Error resources

		// Verify no verbose output since verbose is disabled
		assert.Empty(t, writer.String())
	})
}
