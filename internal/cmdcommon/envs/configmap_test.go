package envs

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types/sourced"
	"github.com/kyma-project/cli.v3/internal/kube"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBuildEnvsFromConfigmap(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		client    kube.Client
		namespace string
		envs      types.SourcedEnvArray
		want      []corev1.EnvVar
		wantErr   string
	}{
		{
			name:      "get single env var from cm",
			namespace: "default",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:        "DB_USERNAME",
						Location:    "my-cm",
						LocationKey: "username",
					},
					{
						Name:        "DB_USERNAME_2",
						Location:    "my-cm",
						LocationKey: "username",
					},
					{
						Name:        "DB_PASSWORD",
						Location:    "my-cm",
						LocationKey: "password",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name: "DB_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
				{
					Name: "DB_USERNAME_2",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
				{
					Name: "DB_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
			},
		},
		{
			name: "missing cm name",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:        "DB_USERNAME",
						LocationKey: "username",
					},
				},
			},
			wantErr: "missing configmap name in env: 'DB_USERNAME=:username'",
		},
		{
			name: "missing cm key",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Name:     "DB_USERNAME",
						Location: "cm-name",
					},
				},
			},
			wantErr: "missing configmap key in env: 'DB_USERNAME=cm-name:'",
		},
		{
			name: "multi env vars from cm",
			client: &kube_fake.KubeClient{
				TestKubernetesInterface: fake.NewClientset(&corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-cm",
						Namespace: "default",
					},
					Data: map[string]string{
						"username": "admin",
						"password": "cm",
					},
				}),
			},
			namespace: "default",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Location:           "my-cm",
						LocationKeysPrefix: "PREFIX_",
					},
					{
						Location: "my-cm",
					},
				},
			},
			want: []corev1.EnvVar{
				{
					Name: "PREFIX_username",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
				{
					Name: "PREFIX_password",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
				{
					Name: "username",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
				{
					Name: "password",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "my-cm",
							},
						},
					},
				},
			},
		},
		{
			name: "cm not found",
			client: &kube_fake.KubeClient{
				TestKubernetesInterface: fake.NewClientset(),
			},
			namespace: "default",
			envs: types.SourcedEnvArray{
				Values: []sourced.Env{
					{
						Location: "missing-cm",
					},
				},
			},
			wantErr: "while reading configmap 'missing-cm': configmaps \"missing-cm\" not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := BuildFromConfigmap(context.Background(), tt.client, tt.namespace, tt.envs)
			if tt.wantErr != "" {
				require.ErrorContains(t, gotErr, tt.wantErr)
			}
			require.ElementsMatch(t, tt.want, got)
		})
	}
}
