package kubeconfig

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-project/cli.v3/internal/clierror"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/clientcmd/api"
)

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
					"context": {
						Cluster:   "cluster",
						AuthInfo:  "username",
						Namespace: "default",
					},
				},
				CurrentContext: "context",
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
					"context": {
						Cluster:   "cluster",
						AuthInfo:  "username",
						Namespace: "default",
					},
				},
				CurrentContext: "context",
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
					"context": {
						Cluster: "cluster",
					},
				},
				CurrentContext: "context",
			}

			staticClient := k8s_fake.NewSimpleClientset(
				&serviceAccount,
				&secret,
			)
			kubeClient := &kube_fake.FakeKubeClient{
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
			kubeClient := &kube_fake.FakeKubeClient{
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
