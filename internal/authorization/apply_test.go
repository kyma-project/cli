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

func TestResourceApplier_ApplyOIDC(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func() *kubefake.KubeClient
		oidcResource  *unstructured.Unstructured
		expectedError string
		expectApplied bool
	}{
		{
			name: "successful create new OIDC",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
				}
			},
			oidcResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata":   map[string]any{"name": "test-oidc"},
				"spec": map[string]any{
					"issuerURL":      "https://token.actions.githubusercontent.com",
					"clientID":       "test-client",
					"usernameClaim":  "repository",
					"usernamePrefix": "github:",
					"requiredClaims": map[string]any{"repository": "owner/repo"},
				},
			}},
			expectApplied: true,
		},
		{
			name: "success when identical OIDC exists",
			setupMocks: func() *kubefake.KubeClient {
				scheme := runtime.NewScheme()
				scheme.AddKnownTypes(kubeconfig.OpenIdConnectGVR.GroupVersion())
				existing := &unstructured.Unstructured{Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata":   map[string]any{"name": "test-oidc"},
					"spec": map[string]any{
						"issuerURL":      "https://token.actions.githubusercontent.com",
						"clientID":       "test-client",
						"usernameClaim":  "repository",
						"usernamePrefix": "github:",
						"requiredClaims": map[string]any{"repository": "owner/repo"},
					},
				}}
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(scheme, existing),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
				}
			},
			oidcResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata":   map[string]any{"name": "test-oidc"},
				"spec": map[string]any{
					"issuerURL":      "https://token.actions.githubusercontent.com",
					"clientID":       "test-client",
					"usernameClaim":  "repository",
					"usernamePrefix": "github:",
					"requiredClaims": map[string]any{"repository": "owner/repo"},
				},
			}},
			expectApplied: true,
		},
		{
			name: "conflict when OIDC exists with different spec",
			setupMocks: func() *kubefake.KubeClient {
				scheme := runtime.NewScheme()
				scheme.AddKnownTypes(kubeconfig.OpenIdConnectGVR.GroupVersion())
				existing := &unstructured.Unstructured{Object: map[string]any{
					"apiVersion": "authentication.gardener.cloud/v1alpha1",
					"kind":       "OpenIDConnect",
					"metadata":   map[string]any{"name": "test-oidc"},
					"spec": map[string]any{
						"issuerURL":      "https://different.issuer.com",
						"clientID":       "different-client",
						"usernameClaim":  "repository",
						"requiredClaims": map[string]any{"repository": "different/repo"},
					},
				}}
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(scheme, existing),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
				}
			},
			oidcResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata":   map[string]any{"name": "test-oidc"},
				"spec": map[string]any{
					"issuerURL":      "https://token.actions.githubusercontent.com",
					"clientID":       "test-client",
					"usernameClaim":  "repository",
					"requiredClaims": map[string]any{"repository": "owner/repo"},
				},
			}},
			expectedError: "configuration conflict detected - operation aborted",
			expectApplied: false,
		},
		{
			name: "apply error for new OIDC",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: errors.New("apply failed")},
				}
			},
			oidcResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "authentication.gardener.cloud/v1alpha1",
				"kind":       "OpenIDConnect",
				"metadata":   map[string]any{"name": "test-oidc"},
			}},
			expectedError: "failed to apply OpenIDConnect resource",
			expectApplied: true, // attempt was made
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.setupMocks()
			applier := authorization.NewResourceApplier(fakeClient)

			err := applier.ApplyOIDC(context.Background(), tt.oidcResource)

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.String(), tt.expectedError)
			} else {
				assert.Nil(t, err)
			}

			rootlessClient := fakeClient.TestRootlessDynamicInterface.(*kubefake.RootlessDynamicClient)
			if tt.expectApplied {
				assert.GreaterOrEqual(t, len(rootlessClient.ApplyObjs), 1)
				assert.Equal(t, "test-oidc", rootlessClient.ApplyObjs[0].GetName())
			} else {
				assert.Equal(t, 0, len(rootlessClient.ApplyObjs))
			}
		})
	}
}

func TestResourceApplier_ApplyRBAC(t *testing.T) {
	tests := []struct {
		name            string
		setupMocks      func() *kubefake.KubeClient
		rbacResource    *unstructured.Unstructured
		expectedError   string
		expectRBACApply bool
	}{
		{
			name: "successful ClusterRoleBinding",
			setupMocks: func() *kubefake.KubeClient {
				clusterrole := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "test-clusterrole"}}
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(clusterrole),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRoleBinding",
				"metadata":   map[string]any{"name": "test-binding"},
				"roleRef": map[string]any{
					"apiGroup": "rbac.authorization.k8s.io",
					"kind":     "ClusterRole",
					"name":     "test-clusterrole",
				},
			}},
			expectRBACApply: true,
		},
		{
			name: "successful RoleBinding",
			setupMocks: func() *kubefake.KubeClient {
				role := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "editor", Namespace: "default"}}
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(role),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "RoleBinding",
				"metadata":   map[string]any{"name": "test-role-binding", "namespace": "default"},
				"roleRef": map[string]any{
					"kind":     "Role",
					"name":     "editor",
					"apiGroup": "rbac.authorization.k8s.io",
				},
			}},
			expectRBACApply: true,
		},
		{
			name: "missing ClusterRole",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRoleBinding",
				"metadata":   map[string]any{"name": "test-binding"},
				"roleRef": map[string]any{
					"kind":     "ClusterRole",
					"name":     "non-existent-cluster-role",
					"apiGroup": "rbac.authorization.k8s.io",
				},
			}},
			expectedError:   "ClusterRole 'non-existent-cluster-role' does not exist",
			expectRBACApply: false,
		},
		{
			name: "missing Role",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "RoleBinding",
				"metadata":   map[string]any{"name": "test-role-binding", "namespace": "default"},
				"roleRef": map[string]any{
					"kind":     "Role",
					"name":     "non-existent-role",
					"apiGroup": "rbac.authorization.k8s.io",
				},
			}},
			expectedError:   "Role 'non-existent-role' does not exist in namespace 'default'",
			expectRBACApply: false,
		},
		{
			name: "missing namespace for RoleBinding",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "RoleBinding",
				"metadata":   map[string]any{"name": "test-binding"},
				"roleRef": map[string]any{
					"kind":     "Role",
					"name":     "test-role",
					"apiGroup": "rbac.authorization.k8s.io",
				},
			}},
			expectedError:   "namespace is required for Role validation",
			expectRBACApply: false,
		},
		{
			name: "unsupported roleRef kind",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "RoleBinding",
				"metadata":   map[string]any{"name": "test-binding", "namespace": "default"},
				"roleRef": map[string]any{
					"kind":     "UnsupportedRole",
					"name":     "test-role",
					"apiGroup": "rbac.authorization.k8s.io",
				},
			}},
			expectedError:   "unsupported roleRef kind: UnsupportedRole",
			expectRBACApply: false,
		},
		{
			name: "missing roleRef",
			setupMocks: func() *kubefake.KubeClient {
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: nil},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "RoleBinding",
				"metadata":   map[string]any{"name": "test-binding", "namespace": "default"},
			}},
			expectedError:   "roleRef not found in RBAC resource",
			expectRBACApply: false,
		},
		{
			name: "apply error for RBAC resource",
			setupMocks: func() *kubefake.KubeClient {
				clusterrole := &rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "test-clusterrole"}}
				return &kubefake.KubeClient{
					TestDynamicInterface:         dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
					TestRootlessDynamicInterface: &kubefake.RootlessDynamicClient{ReturnErr: errors.New("apply failed")},
					TestKubernetesInterface:      k8sfake.NewSimpleClientset(clusterrole),
				}
			},
			rbacResource: &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "rbac.authorization.k8s.io/v1",
				"kind":       "ClusterRoleBinding",
				"metadata":   map[string]any{"name": "test-binding"},
				"roleRef": map[string]any{
					"apiGroup": "rbac.authorization.k8s.io",
					"kind":     "ClusterRole",
					"name":     "test-clusterrole",
				},
			}},
			expectedError:   "failed to apply RBAC resource",
			expectRBACApply: true, // attempt made
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := tt.setupMocks()
			applier := authorization.NewResourceApplier(fakeClient)

			err := applier.ApplyRBAC(context.Background(), tt.rbacResource)

			if tt.expectedError != "" {
				assert.NotNil(t, err)
				assert.Contains(t, err.String(), tt.expectedError)
			} else {
				assert.Nil(t, err)
			}

			rootlessClient := fakeClient.TestRootlessDynamicInterface.(*kubefake.RootlessDynamicClient)
			if tt.expectRBACApply {
				assert.GreaterOrEqual(t, len(rootlessClient.ApplyObjs), 1)
				assert.Equal(t, tt.rbacResource.GetName(), rootlessClient.ApplyObjs[len(rootlessClient.ApplyObjs)-1].GetName())
			} else {
				assert.Equal(t, 0, len(rootlessClient.ApplyObjs))
			}
		})
	}
}
