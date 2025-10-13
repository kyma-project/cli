package authorization

import (
	"context"
	"errors"
	"testing"

	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestNewResourceApplier(t *testing.T) {
	fakeClient := &kubefake.KubeClient{}
	applier := NewResourceApplier(fakeClient)

	assert.NotNil(t, applier)
	assert.Equal(t, fakeClient, applier.kubeClient)
}

func TestResourceApplier_ApplyResources(t *testing.T) {
	tests := []struct {
		name                string
		setupMocks          func() *kubefake.KubeClient
		oidcResource        *unstructured.Unstructured
		rbacResource        *unstructured.Unstructured
		expectedError       string
		expectApplyOIDCCall bool
		expectApplyRBACCall bool
	}{
		{
			name: "successful apply of both resources",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
				}
			},
			oidcResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata": map[string]any{
						"name": "test-oidc",
					},
				},
			},
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
					},
				},
			},
			expectApplyOIDCCall: true,
			expectApplyRBACCall: true,
		},
		{
			name: "error applying OIDC resource should stop execution",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: errors.New("oidc apply failed"),
					},
				}
			},
			oidcResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata": map[string]any{
						"name": "test-oidc",
					},
				},
			},
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
					},
				},
			},
			expectedError:       "failed to apply OpenIDConnect resource",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.setupMocks()
			applier := NewResourceApplier(fakeClient)

			err := applier.ApplyResources(context.Background(), tt.oidcResource, tt.rbacResource)

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.String(), tt.expectedError)
			} else {
				assert.Nil(t, err)
			}

			// Verify OIDC apply was called if expected
			if tt.expectApplyOIDCCall {
				rootlessClient := fakeClient.TestRootlessDynamicInterface.(*kubefake.RootlessDynamicClient)
				assert.True(t, len(rootlessClient.ApplyObjs) >= 1)
				assert.Equal(t, "test-oidc", rootlessClient.ApplyObjs[0].GetName())
			}

			// Verify RBAC apply was called if expected
			if tt.expectApplyRBACCall {
				rootlessClient := fakeClient.TestRootlessDynamicInterface.(*kubefake.RootlessDynamicClient)
				assert.Equal(t, 2, len(rootlessClient.ApplyObjs))
				assert.Equal(t, "test-binding", rootlessClient.ApplyObjs[1].GetName())
			}
		})
	}
}

func TestResourceApplier_applyOIDCResource(t *testing.T) {
	tests := []struct {
		name            string
		setupMocks      func() *kubefake.KubeClient
		oidcResource    *unstructured.Unstructured
		expectedError   string
		expectApplyCall bool
	}{
		{
			name: "successful apply when resource doesn't exist",
			setupMocks: func() *kubefake.KubeClient {
				scheme := runtime.NewScheme()
				scheme.AddKnownTypes(kubeconfig.OpenIdConnectGVR.GroupVersion())
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(scheme),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
				}
			},
			oidcResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata": map[string]any{
						"name": "test-oidc",
					},
				},
			},
			expectApplyCall: true,
		},
		{
			name: "successful apply when resource exists with identical config",
			setupMocks: func() *kubefake.KubeClient {
				scheme := runtime.NewScheme()
				scheme.AddKnownTypes(kubeconfig.OpenIdConnectGVR.GroupVersion())
				existingOIDC := &unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "authentication.gardener.cloud/v1alpha1",
						"kind":       "OpenIDConnect",
						"metadata": map[string]any{
							"name": "test-oidc",
						},
						"spec": map[string]any{
							"issuerURL":      "https://token.actions.githubusercontent.com",
							"clientID":       "test-client",
							"usernameClaim":  "repository",
							"usernamePrefix": "github:",
							"requiredClaims": map[string]any{
								"repository": "owner/repo",
							},
						},
					},
				}
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(scheme, existingOIDC),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
				}
			},
			oidcResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata": map[string]any{
						"name": "test-oidc",
					},
					"spec": map[string]any{
						"issuerURL":      "https://token.actions.githubusercontent.com",
						"clientID":       "test-client",
						"usernameClaim":  "repository",
						"usernamePrefix": "github:",
						"requiredClaims": map[string]any{
							"repository": "owner/repo",
						},
					},
				},
			},
			expectApplyCall: true,
		},
		{
			name: "error when resource exists with different config",
			setupMocks: func() *kubefake.KubeClient {
				scheme := runtime.NewScheme()
				scheme.AddKnownTypes(kubeconfig.OpenIdConnectGVR.GroupVersion())
				existingOIDC := &unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "authentication.gardener.cloud/v1alpha1",
						"kind":       "OpenIDConnect",
						"metadata": map[string]any{
							"name": "test-oidc",
						},
						"spec": map[string]any{
							"issuerURL":     "https://different.issuer.com",
							"clientID":      "different-client",
							"usernameClaim": "repository",
							"requiredClaims": map[string]any{
								"repository": "different/repo",
							},
						},
					},
				}
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(scheme, existingOIDC),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
				}
			},
			oidcResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata": map[string]any{
						"name": "test-oidc",
					},
					"spec": map[string]any{
						"issuerURL":     "https://token.actions.githubusercontent.com",
						"clientID":      "test-client",
						"usernameClaim": "repository",
						"requiredClaims": map[string]any{
							"repository": "owner/repo",
						},
					},
				},
			},
			expectedError:   "configuration conflict detected - operation aborted",
			expectApplyCall: false,
		},
		{
			name: "error applying resource",
			setupMocks: func() *kubefake.KubeClient {
				scheme := runtime.NewScheme()
				scheme.AddKnownTypes(kubeconfig.OpenIdConnectGVR.GroupVersion())
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(scheme),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: errors.New("apply failed"),
					},
				}
			},
			oidcResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata": map[string]any{
						"name": "test-oidc",
					},
				},
			},
			expectedError:   "failed to apply OpenIDConnect resource",
			expectApplyCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.setupMocks()
			applier := NewResourceApplier(fakeClient)

			err := applier.applyOIDCResource(context.Background(), tt.oidcResource)

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.String(), tt.expectedError)
			} else {
				assert.Nil(t, err)
			}

			if tt.expectApplyCall {
				rootlessClient := fakeClient.TestRootlessDynamicInterface.(*kubefake.RootlessDynamicClient)
				assert.Len(t, rootlessClient.ApplyObjs, 1)
				assert.Equal(t, "test-oidc", rootlessClient.ApplyObjs[0].GetName())
			}
		})
	}
}

func TestResourceApplier_applyRBACResource(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func() *kubefake.KubeClient
		rbacResource  *unstructured.Unstructured
		expectedError string
	}{
		{
			name: "successful ClusterRoleBinding apply",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
				}
			},
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
					},
				},
			},
		},
		{
			name: "successful RoleBinding apply",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
				}
			},
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "RoleBinding",
					"metadata": map[string]any{
						"name":      "test-role-binding",
						"namespace": "default",
					},
				},
			},
		},
		{
			name: "error applying RBAC resource",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: errors.New("apply failed"),
					},
				}
			},
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
					},
				},
			},
			expectedError: "failed to apply RBAC resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.setupMocks()
			applier := NewResourceApplier(fakeClient)

			err := applier.applyRBACResource(context.Background(), tt.rbacResource)

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.String(), tt.expectedError)
			} else {
				assert.Nil(t, err)

				// Verify apply was called
				rootlessClient := fakeClient.TestRootlessDynamicInterface.(*kubefake.RootlessDynamicClient)
				require.Len(t, rootlessClient.ApplyObjs, 1)
				assert.Equal(t, tt.rbacResource.GetName(), rootlessClient.ApplyObjs[0].GetName())
				assert.Equal(t, tt.rbacResource.GetKind(), rootlessClient.ApplyObjs[0].GetKind())
			}
		})
	}
}
