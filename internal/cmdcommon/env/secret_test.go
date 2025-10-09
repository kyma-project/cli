package env_test

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/env"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types/sourced"
	"github.com/kyma-project/cli.v3/internal/kube"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildEnvsFromSecret(t *testing.T) {
	tests := []struct {
		name      string
		client    kube.Client
		namespace string
		envs      types.SourcedEnvArray
		want      []corev1.EnvVar
		wantErr   string
	}{
		{
			name:      "get single env var from secret",
			namespace: "default",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:        "DB_USERNAME",
						Location:    "my-secret",
						LocationKey: "username",
					},
					{
						Name:        "DB_USERNAME_2",
						Location:    "my-secret",
						LocationKey: "username",
					},
					{
						Name:        "DB_PASSWORD",
						Location:    "my-secret",
						LocationKey: "password",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name: "DB_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
				{
					Name: "DB_USERNAME_2",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
				{
					Name: "DB_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
			},
		},
		{
			name: "missing secret name",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:        "DB_USERNAME",
						LocationKey: "username",
					},
				},
			},
			wantErr: "missing secret name in env: 'DB_USERNAME=:username'",
		},
		{
			name: "missing secret key",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:     "DB_USERNAME",
						Location: "secret-name",
					},
				},
			},
			wantErr: "missing secret key in env: 'DB_USERNAME=secret-name:'",
		},
		{
			name: "multi env vars from secret",
			client: &kube_fake.KubeClient{
				TestKubernetesInterface: fake.NewClientset(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-secret",
						Namespace: "default",
					},
					Data: map[string][]byte{
						"username": []byte("admin"),
						"password": []byte("secret"),
					},
				}),
			},
			namespace: "default",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Location:           "my-secret",
						LocationKeysPrefix: "PREFIX_",
					},
					{
						Location: "my-secret",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name: "PREFIX_username",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
				{
					Name: "PREFIX_password",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
				{
					Name: "username",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
				{
					Name: "password",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-secret",
							},
						},
					},
				},
			},
		},
		{
			name: "secret not found",
			client: &kube_fake.KubeClient{
				TestKubernetesInterface: fake.NewClientset(),
			},
			namespace: "default",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Location: "missing-secret",
					},
				},
			},
			wantErr: "while reading secret 'missing-secret': secrets \"missing-secret\" not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := env.BuildEnvsFromSecret(context.Background(), tt.client, tt.namespace, tt.envs)
			if tt.wantErr != "" {
				require.ErrorContains(t, gotErr, tt.wantErr)
			}
			require.ElementsMatch(t, tt.want, got)
		})
	}
}
