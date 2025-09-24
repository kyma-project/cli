package kubeconfig

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-test/deep"
	"github.com/kyma-project/cli.v3/internal/clierror"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamic_fake "k8s.io/client-go/dynamic/fake"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd/api"
)

func Test_PrepareWithToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		kubeconfig *api.Config
		token      string
		want       *api.Config
	}{
		{
			name: "change kubeconfig authInfo",
			kubeconfig: &api.Config{
				Clusters: map[string]*api.Cluster{
					"cluster": {},
				},
				AuthInfos: map[string]*api.AuthInfo{
					"user": {
						Username:  "user",
						ClientKey: "remove",
					},
				},
				Contexts: map[string]*api.Context{
					"context": {
						AuthInfo: "user",
					},
				},
				CurrentContext: "context",
			},
			token: "token",
			want: &api.Config{
				Kind:       "Config",
				APIVersion: "v1",
				Clusters: map[string]*api.Cluster{
					"cluster": {},
				},
				AuthInfos: map[string]*api.AuthInfo{
					"user": {
						Token: "token",
					},
				},
				Contexts: map[string]*api.Context{
					"context": {
						AuthInfo: "user",
					},
				},
				CurrentContext: "context",
			},
		},
	}

	for _, tt := range tests {
		kubeconfig := tt.kubeconfig
		token := tt.token
		want := tt.want
		t.Run(tt.name, func(t *testing.T) {
			got := PrepareWithToken(kubeconfig, token)
			if diff := deep.Equal(got, want); diff != nil {
				t.Errorf("createKubeconfig() = %s", diff)
			}
		})
	}
}

func Test_Prepare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		username       string
		namespace      string
		time           string
		output         string
		permanent      bool
		expectedConfig *api.Config
		expectedErr    clierror.Error
	}{
		{
			name:      "create temporary kubeconfig",
			username:  "username",
			namespace: "default",
			time:      "1h",
			output:    "",
			permanent: false,
			expectedConfig: &api.Config{
				Kind:        "Config",
				APIVersion:  "v1",
				Preferences: api.Preferences{},
				AuthInfos: map[string]*api.AuthInfo{
					"username": {},
				},
				Clusters: map[string]*api.Cluster{
					"cluster": {
						Server:                   "https://localhost:8080",
						CertificateAuthorityData: []byte("certificate"),
					},
				},
				Contexts: map[string]*api.Context{
					"cluster": {
						Cluster:   "cluster",
						AuthInfo:  "username",
						Namespace: "default",
					},
				},
				CurrentContext: "cluster",
			},
			expectedErr: nil,
		},
		{
			name:      "create permanent kubeconfig",
			username:  "username",
			namespace: "default",
			time:      "1h",
			output:    "",
			permanent: true,
			expectedConfig: &api.Config{
				Kind:        "Config",
				APIVersion:  "v1",
				Preferences: api.Preferences{},
				AuthInfos: map[string]*api.AuthInfo{
					"username": {
						Token: "token",
					},
				},
				Clusters: map[string]*api.Cluster{
					"cluster": {
						Server:                   "https://localhost:8080",
						CertificateAuthorityData: []byte("secretCertificate"),
					},
				},
				Contexts: map[string]*api.Context{
					"cluster": {
						Cluster:   "cluster",
						AuthInfo:  "username",
						Namespace: "default",
					},
				},
				CurrentContext: "cluster",
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		username := tt.username
		namespace := tt.namespace
		time := tt.time
		output := tt.output
		permanent := tt.permanent
		expectedConfig := tt.expectedConfig
		expectedErr := tt.expectedErr

		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceAccount := corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "username",
					Namespace: "default",
				},
			}
			secret := corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "username",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"token":  []byte("token"),
					"ca.crt": []byte("secretCertificate"),
				},
			}

			apiConfig := &api.Config{
				Clusters: map[string]*api.Cluster{
					"cluster": {
						Server:                   "https://localhost:8080",
						CertificateAuthorityData: []byte("certificate"),
					},
				},
				Contexts: map[string]*api.Context{
					"cluster": {
						Cluster: "cluster",
					},
				},
				CurrentContext: "cluster",
			}

			staticClient := k8s_fake.NewSimpleClientset(
				&serviceAccount,
				&secret,
			)
			kubeClient := &kube_fake.KubeClient{
				TestAPIConfig:           apiConfig,
				TestKubernetesInterface: staticClient,
			}

			config, err := Prepare(ctx, kubeClient, username, namespace, time, output, permanent)

			// require.Equal(t, expected, config)
			require.Equal(t, expectedErr, err)
			if err == nil {
				require.Equal(t, expectedConfig, config)
			}
		})
	}
}

func Test_getServiceAccountToken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		username      string
		namespace     string
		time          string
		expectedError clierror.Error
	}{
		{
			name:          "get token",
			username:      "username",
			namespace:     "default",
			time:          "1h",
			expectedError: nil,
		},
		{
			name:          "incorrect time",
			username:      "username",
			namespace:     "default",
			time:          "incorrect",
			expectedError: clierror.New("failed to convert the token duration", "Make sure to use h for hours and d for days"),
		},
		{
			name:          "incorrect username / namespace",
			username:      "",
			namespace:     "",
			time:          "1h",
			expectedError: clierror.Wrap(fmt.Errorf("serviceaccounts \"\" not found"), clierror.New("failed to create token")),
		},
	}

	for _, tt := range tests {
		username := tt.username
		namespace := tt.namespace
		time := tt.time
		expectedError := tt.expectedError

		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			serviceAccount := corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "username",
					Namespace: "default",
				},
			}
			staticClient := k8s_fake.NewSimpleClientset(
				&serviceAccount,
			)
			kubeClient := &kube_fake.KubeClient{
				TestKubernetesInterface: staticClient,
			}

			_, err := getServiceAccountToken(ctx, kubeClient, username, namespace, time)

			require.Equal(t, expectedError, err)

		})
	}
}

func Test_parseExpirationTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		time            string
		expectedSeconds int64
		expectedError   bool
	}{
		{
			name:            "1 hour",
			time:            "1h",
			expectedSeconds: 3600,
		},
		{
			name:            "1 day",
			time:            "1d",
			expectedSeconds: 86400,
		},
		{
			name:            "1 day 1 hour",
			time:            "1d1h",
			expectedSeconds: 0,
			expectedError:   true,
		},
		{
			name:            "empty string",
			time:            "",
			expectedSeconds: 0,
			expectedError:   true,
		},
	}

	for _, tt := range tests {
		time := tt.time
		expectedError := tt.expectedError
		expectedSeconds := tt.expectedSeconds
		t.Run(tt.name, func(t *testing.T) {
			seconds, err := parseExpirationTime(time)
			if expectedError {
				require.NotNil(t, err)
				return
			} else {
				require.Nil(t, err)
			}
			require.Equal(t, expectedSeconds, seconds)
		})
	}
}

func Test_PrepareFromOpenIDConnectorResource(t *testing.T) {
	newFakeClient := func(oidcRes *unstructured.Unstructured) *kube_fake.KubeClient {
		scheme := runtime.NewScheme()
		scheme.AddKnownTypes(OpenIdConnectGVR.GroupVersion())
		dynamicClient := dynamic_fake.NewSimpleDynamicClient(scheme, oidcRes)
		apiConfig := &api.Config{
			Clusters: map[string]*api.Cluster{
				"cluster": {
					Server:                   "https://localhost:8080",
					CertificateAuthorityData: []byte("certificate"),
				},
			},
			Contexts: map[string]*api.Context{
				"context": {
					Cluster: "cluster",
				},
			},
			CurrentContext: "context",
		}
		return &kube_fake.KubeClient{
			TestDynamicInterface: dynamicClient,
			TestAPIConfig:        apiConfig,
		}
	}

	baseOIDCRes := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "authentication.gardener.cloud/v1alpha1",
			"kind":       "OpenIDConnect",
			"metadata": map[string]any{
				"labels": map[string]any{
					"operator.kyma-project.io/managed-by": "infrastructure-manager",
				},
				"name": "kyma-oidc",
			},
			"spec": map[string]any{
				"issuerURL": "http://super.issuer.com",
				"clientID":  "10033344-6d08-46cd-80b9-37e18cd856c1",
			},
		},
	}

	t.Run("OIDC resource is missing", func(t *testing.T) {
		fakeClient := newFakeClient(baseOIDCRes)
		result, clierr := PrepareFromOpenIDConnectorResource(context.Background(), fakeClient, "invalid-oidc-name")
		require.Nil(t, result)
		require.NotNil(t, clierr)
		require.Equal(
			t,
			"Error:\n  failed to get invalid-oidc-name oidc resource\n\nError Details:\n  openidconnects.authentication.gardener.cloud \"invalid-oidc-name\" not found\n\n",
			clierr.String(),
		)
	})

	t.Run("returns kubeconfig", func(t *testing.T) {
		fakeClient := newFakeClient(baseOIDCRes)
		result, clierr := PrepareFromOpenIDConnectorResource(context.Background(), fakeClient, "kyma-oidc")
		require.Nil(t, clierr)

		expected := &api.Config{
			Kind:       "Config",
			APIVersion: "v1",
			Clusters: map[string]*api.Cluster{
				"cluster": {
					Server:                   "https://localhost:8080",
					CertificateAuthorityData: []byte("certificate"),
				},
			},
			AuthInfos: map[string]*api.AuthInfo{
				"kyma-oidc": {
					Exec: &api.ExecConfig{
						APIVersion: "client.authentication.k8s.io/v1beta1",
						Command:    "kubectl-oidc_login",
						Args: []string{
							"get-token",
							"--oidc-issuer-url=http://super.issuer.com",
							"--oidc-client-id=10033344-6d08-46cd-80b9-37e18cd856c1",
							"--oidc-extra-scope=email",
							"--oidc-extra-scope=openid",
						},
					},
				},
			},
			Contexts: map[string]*api.Context{
				"cluster": {
					Cluster:  "cluster",
					AuthInfo: "kyma-oidc",
				},
			},
			CurrentContext: "cluster",
			Extensions:     nil,
		}
		require.Equal(t, expected, result)
	})

	t.Run("kubeconfig without current context", func(t *testing.T) {
		apiConfig := &api.Config{
			Clusters: map[string]*api.Cluster{
				"cluster": {
					Server:                   "https://localhost:8080",
					CertificateAuthorityData: []byte("certificate"),
				},
			},
			Contexts: map[string]*api.Context{
				"context": {
					Cluster: "cluster",
				},
			},
			CurrentContext: "",
		}
		fakeClient := newFakeClient(baseOIDCRes)
		fakeClient.TestAPIConfig = apiConfig

		result, clierr := PrepareFromOpenIDConnectorResource(context.Background(), fakeClient, "kyma-oidc")
		require.Nil(t, clierr)

		expected := &api.Config{
			Kind:       "Config",
			APIVersion: "v1",
			Clusters: map[string]*api.Cluster{
				"cluster": {
					Server:                   "https://localhost:8080",
					CertificateAuthorityData: []byte("certificate"),
				},
			},
			AuthInfos: map[string]*api.AuthInfo{
				"kyma-oidc": {
					Exec: &api.ExecConfig{
						APIVersion: "client.authentication.k8s.io/v1beta1",
						Command:    "kubectl-oidc_login",
						Args: []string{
							"get-token",
							"--oidc-issuer-url=http://super.issuer.com",
							"--oidc-client-id=10033344-6d08-46cd-80b9-37e18cd856c1",
							"--oidc-extra-scope=email",
							"--oidc-extra-scope=openid",
						},
					},
				},
			},
			Contexts: map[string]*api.Context{
				"cluster": {
					Cluster:  "cluster",
					AuthInfo: "kyma-oidc",
				},
			},
			CurrentContext: "cluster",
			Extensions:     nil,
		}
		require.Equal(t, expected, result)
	})
}
