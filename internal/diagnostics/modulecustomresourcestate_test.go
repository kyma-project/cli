package diagnostics_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/cli.v3/internal/diagnostics"
	"github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	modulesfake "github.com/kyma-project/cli.v3/internal/modules/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sfake "k8s.io/client-go/kubernetes/fake"
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

func TestNewModuleCustomResourceStateCollectorWithRepo(t *testing.T) {
	// Given
	fakeClient := &fake.KubeClient{}
	fakeRepo := &modulesfake.ModuleTemplatesRepo{}
	var writer bytes.Buffer
	verbose := true

	// When
	collector := diagnostics.NewModuleCustomResourceStateCollectorWithRepo(fakeClient, fakeRepo, &writer, verbose)

	// Then
	assert.NotNil(t, collector)
}

func TestModuleCustomResourceStateCollector_HandleInvalidData(t *testing.T) {
	testCases := []struct {
		name                  string
		moduleData            map[string]any
		expectedVerboseOutput string
	}{
		{
			name: "Should handle missing apiVersion",
			moduleData: map[string]any{
				"kind": "TestModule",
				"metadata": map[string]any{
					"name":      "test",
					"namespace": "test-namespace",
				},
			},
			expectedVerboseOutput: "Failed to get data resource for test-module module",
		},
		{
			name: "Should handle missing kind",
			moduleData: map[string]any{
				"apiVersion": "test.io/v1",
				"metadata": map[string]any{
					"name":      "test",
					"namespace": "test-namespace",
				},
			},
			expectedVerboseOutput: "Failed to get data resource for test-module module",
		},
		{
			name: "Should handle missing metadata",
			moduleData: map[string]any{
				"apiVersion": "test.io/v1",
				"kind":       "TestModule",
			},
			expectedVerboseOutput: "Failed to get data resource for test-module module",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer

			fakeModule := kyma.ModuleTemplate{
				Spec: kyma.ModuleTemplateSpec{
					ModuleName: "test-module",
					Data: unstructured.Unstructured{
						Object: tc.moduleData,
					},
				},
			}

			fakeKubeClient := k8sfake.NewSimpleClientset()
			fakeClient := &fake.KubeClient{
				TestKubernetesInterface: fakeKubeClient,
				TestRootlessDynamicInterface: &fake.RootlessDynamicClient{
					ReturnListObjs: &unstructured.UnstructuredList{},
				},
			}

			fakeModuleRepo := &modulesfake.ModuleTemplatesRepo{
				ReturnCore:      []kyma.ModuleTemplate{fakeModule},
				ReturnCommunity: []kyma.ModuleTemplate{},
			}

			collector := diagnostics.NewModuleCustomResourceStateCollectorWithRepo(
				fakeClient,
				fakeModuleRepo,
				&writer,
				true, // verbose enabled
			)

			// When
			result := collector.Run(context.Background())

			// Then
			assert.Empty(t, result)
			assert.Contains(t, writer.String(), tc.expectedVerboseOutput)
		})
	}
}

func TestModuleCustomResourceStateCollector_Run(t *testing.T) {
	testCases := []struct {
		name                       string
		coreModules                []kyma.ModuleTemplate
		communityModules           []kyma.ModuleTemplate
		mockResourceList           *unstructured.UnstructuredList
		enableVerboseLogging       bool
		expectedResultCount        int
		expectedVerboseOutput      string
		shouldRegisterAPIResources bool
	}{
		{
			name: "Should collect non-ready module states successfully",
			coreModules: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "test-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestModule",
								"metadata": map[string]any{
									"name":      "test-resource",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			communityModules: []kyma.ModuleTemplate{},
			mockResourceList: &unstructured.UnstructuredList{
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
										"reason":             "InstallationFailed",
										"message":            "Module installation failed",
										"lastTransitionTime": "2023-01-01T00:00:00Z",
									},
								},
							},
						},
					},
				},
			},
			enableVerboseLogging:       false,
			expectedResultCount:        1,
			shouldRegisterAPIResources: true,
		},
		{
			name: "Should skip ready modules",
			coreModules: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "ready-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "test.io/v1",
								"kind":       "TestModule",
								"metadata": map[string]any{
									"name":      "ready-resource",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			communityModules: []kyma.ModuleTemplate{},
			mockResourceList: &unstructured.UnstructuredList{
				Items: []unstructured.Unstructured{
					{
						Object: map[string]any{
							"apiVersion": "test.io/v1",
							"kind":       "TestModule",
							"status": map[string]any{
								"state": "Ready", // This should be skipped
							},
						},
					},
				},
			},
			enableVerboseLogging:       false,
			expectedResultCount:        0,
			shouldRegisterAPIResources: true,
		},
		{
			name: "Should handle resource not registered error",
			coreModules: []kyma.ModuleTemplate{
				{
					Spec: kyma.ModuleTemplateSpec{
						ModuleName: "unregistered-module",
						Data: unstructured.Unstructured{
							Object: map[string]any{
								"apiVersion": "unregistered.io/v1",
								"kind":       "UnregisteredModule",
								"metadata": map[string]any{
									"name":      "unregistered-resource",
									"namespace": "test-namespace",
								},
							},
						},
					},
				},
			},
			communityModules:           []kyma.ModuleTemplate{},
			mockResourceList:           &unstructured.UnstructuredList{},
			enableVerboseLogging:       true,
			expectedResultCount:        0,
			expectedVerboseOutput:      "Failed to get data resource for unregistered-module module",
			shouldRegisterAPIResources: false, // Don't register this resource
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var writer bytes.Buffer

			// Create a fake Kubernetes client
			fakeKubeClient := k8sfake.NewSimpleClientset()

			// Register API resources if needed
			if tc.shouldRegisterAPIResources {
				// Add the API resource to the discovery client
				fakeKubeClient.Resources = []*metav1.APIResourceList{
					{
						GroupVersion: "test.io/v1",
						APIResources: []metav1.APIResource{
							{
								Name:       "testmodules",
								Kind:       "TestModule",
								Namespaced: true,
							},
						},
					},
				}
			}

			fakeRootlessDynamic := &fake.RootlessDynamicClient{
				ReturnListObjs: tc.mockResourceList,
			}

			fakeClient := &fake.KubeClient{
				TestKubernetesInterface:      fakeKubeClient,
				TestRootlessDynamicInterface: fakeRootlessDynamic,
			}

			fakeModuleRepo := &modulesfake.ModuleTemplatesRepo{
				ReturnCore:      tc.coreModules,
				ReturnCommunity: tc.communityModules,
			}

			collector := diagnostics.NewModuleCustomResourceStateCollectorWithRepo(
				fakeClient,
				fakeModuleRepo,
				&writer,
				tc.enableVerboseLogging,
			)

			// When
			result := collector.Run(context.Background())

			// Then
			assert.Len(t, result, tc.expectedResultCount)

			if tc.expectedVerboseOutput != "" {
				assert.Contains(t, writer.String(), tc.expectedVerboseOutput)
			}

			if tc.expectedResultCount > 0 {
				for _, state := range result {
					assert.NotEmpty(t, state.ApiVersion)
					assert.NotEmpty(t, state.Kind)
					assert.NotEmpty(t, state.State)
					assert.NotEqual(t, "Ready", state.State) // Should not collect ready states
				}
			}
		})
	}
}
