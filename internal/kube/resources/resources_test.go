package resources

import (
	"context"
	"testing"

	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
)

func Test_CreateClusterRoleBinding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		username    string
		namespace   string
		clusterRole string
		wantErr     bool
	}{
		{
			name:        "create cluster role binding",
			username:    "username",
			namespace:   "default",
			clusterRole: "clusterRole",
			wantErr:     false,
		},
		{
			name:        "create existing cluster role binding",
			username:    "existing",
			namespace:   "default",
			clusterRole: "clusterRole",
			wantErr:     false,
		},
		{
			name:        "non-existent clusterRole",
			username:    "username",
			namespace:   "default",
			clusterRole: "missing",
			wantErr:     true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		username := tt.username
		namespace := tt.namespace
		clusterRole := tt.clusterRole
		wantErr := tt.wantErr

		t.Run(tt.name, func(t *testing.T) {
			serviceAccount := corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing",
					Namespace: "default",
				},
			}

			ClusterRoleBinding := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing",
				},
			}
			existingClusterRole := rbacv1.ClusterRole{
				ObjectMeta: metav1.ObjectMeta{
					Name: "clusterRole",
				},
			}
			staticClient := k8s_fake.NewSimpleClientset(
				&serviceAccount,
				&ClusterRoleBinding,
				&existingClusterRole,
			)
			kubeClient := &kube_fake.FakeKubeClient{
				TestKubernetesInterface: staticClient,
			}
			err := CreateClusterRoleBinding(ctx, kubeClient, username, namespace, clusterRole)
			if wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func Test_CreateDeployment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		deploymentName string
		namespace      string
		image          string
		wantErr        bool
	}{
		{
			name:           "create deployment",
			deploymentName: "deployment",
			namespace:      "default",
			image:          "nginx",
			wantErr:        false,
		},
		{
			name:           "do not allow creating existing deployment",
			deploymentName: "existing",
			namespace:      "default",
			image:          "nginx",
			wantErr:        true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		deploymentName := tt.deploymentName
		namespace := tt.namespace
		image := tt.image
		wantErr := tt.wantErr

		t.Run(tt.name, func(t *testing.T) {
			existingDeployment := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "existing",
					Namespace: "default",
				},
			}
			staticClient := k8s_fake.NewSimpleClientset(
				&existingDeployment,
			)
			kubeClient := &kube_fake.FakeKubeClient{
				TestKubernetesInterface: staticClient,
			}

			err := CreateDeployment(ctx, kubeClient, deploymentName, namespace, image)
			if wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
