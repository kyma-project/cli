package authorization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-project/cli.v3/internal/authorization"
	kubefake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

func TestNewResourceApplier(t *testing.T) {
	fakeClient := &kubefake.KubeClient{}
	applier := authorization.NewResourceApplier(fakeClient)

	assert.NotNil(t, applier)
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
			name: "successful apply with ClusterRoleBinding referencing existing ClusterRole",
			setupMocks: func() *kubefake.KubeClient {
				clusterrole := &rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-clusterrole",
					},
				}

				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(clusterrole),
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
					"roleRef": map[string]any{
						"apiGroup": "rbac.authorization.k8s.io",
						"kind":     "ClusterRole",
						"name":     "test-clusterrole",
					},
				},
			},
			expectApplyOIDCCall: true,
			expectApplyRBACCall: true,
		},
		{
			name: "successful apply with RoleBinding referencing existing Role",
			setupMocks: func() *kubefake.KubeClient {
				role := &rbacv1.Role{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "editor",
						Namespace: "default",
					},
				}

				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(role),
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
					"kind":       "RoleBinding",
					"metadata": map[string]any{
						"name":      "test-role-binding",
						"namespace": "default",
					},
					"roleRef": map[string]any{
						"kind":     "Role",
						"name":     "editor",
						"apiGroup": "rbac.authorization.k8s.io",
					},
				},
			},
			expectApplyOIDCCall: true,
			expectApplyRBACCall: true,
		},
		{
			name: "error applying OIDC resource should stop execution",
			setupMocks: func() *kubefake.KubeClient {
				clusterRole := &rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-clusterrole",
					},
				}

				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: errors.New("oidc apply failed"),
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(clusterRole),
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
					"roleRef": map[string]any{
						"apiGroup": "rbac.authorization.k8s.io",
						"kind":     "ClusterRole",
						"name":     "test-clusterrole",
					},
				},
			},
			expectedError:       "failed to apply OpenIDConnect resource",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
		{
			name: "error when ClusterRole does not exist should stop execution",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(),
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
					"roleRef": map[string]any{
						"kind":     "ClusterRole",
						"name":     "non-existent-cluster-role",
						"apiGroup": "rbac.authorization.k8s.io",
					},
				},
			},
			expectedError:       "ClusterRole 'non-existent-cluster-role' does not exist",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
		{
			name: "error when Role does not exist should stop execution",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(),
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
					"kind":       "RoleBinding",
					"metadata": map[string]any{
						"name":      "test-role-binding",
						"namespace": "default",
					},
					"roleRef": map[string]any{
						"kind":     "Role",
						"name":     "non-existent-role",
						"apiGroup": "rbac.authorization.k8s.io",
					},
				},
			},
			expectedError:       "Role 'non-existent-role' does not exist in namespace 'default'",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
		{
			name: "error when namespace is missing for Role validation",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(),
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
					"kind":       "RoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
						// Note: no namespace specified
					},
					"roleRef": map[string]any{
						"kind":     "Role",
						"name":     "test-role",
						"apiGroup": "rbac.authorization.k8s.io",
					},
				},
			},
			expectedError:       "namespace is required for Role validation",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
		{
			name: "error when roleRef kind is unsupported",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(),
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
					"kind":       "RoleBinding",
					"metadata": map[string]any{
						"name":      "test-binding",
						"namespace": "default",
					},
					"roleRef": map[string]any{
						"kind":     "UnsupportedRole",
						"name":     "test-role",
						"apiGroup": "rbac.authorization.k8s.io",
					},
				},
			},
			expectedError:       "unsupported roleRef kind: UnsupportedRole",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
		{
			name: "error when roleRef is missing",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{
						ReturnErr: nil,
					},
					TestKubernetesInterface: k8sfake.NewSimpleClientset(),
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
					"kind":       "RoleBinding",
					"metadata": map[string]any{
						"name":      "test-binding",
						"namespace": "default",
					},
				},
			},
			expectedError:       "roleRef not found in RBAC resource",
			expectApplyOIDCCall: true,
			expectApplyRBACCall: false,
		},
		{
			name: "successful apply when OIDC resource exists with identical config",
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
				clusterRole := &rbacv1.ClusterRole{
					ObjectMeta: metav1.ObjectMeta{
						Name: "cluster-admin",
					},
				}
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(scheme, existingOIDC),
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(clusterRole),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
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
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
					},
					"roleRef": map[string]any{
						"kind":     "ClusterRole",
						"name":     "cluster-admin",
						"apiGroup": "rbac.authorization.k8s.io",
					},
				},
			},
			expectApplyOIDCCall: true,
			expectApplyRBACCall: true,
		},
		{
			name: "error when OIDC resource exists with different config",
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
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(scheme, existingOIDC),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
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
			rbacResource: &unstructured.Unstructured{
				Object: map[string]any{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRoleBinding",
					"metadata": map[string]any{
						"name": "test-binding",
					},
				},
			},
			expectedError:       "configuration conflict detected - operation aborted",
			expectApplyOIDCCall: false,
			expectApplyRBACCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.setupMocks()
			applier := authorization.NewResourceApplier(fakeClient)

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
				assert.Equal(t, tt.rbacResource.GetName(), rootlessClient.ApplyObjs[1].GetName())
			}
		})
	}
}
