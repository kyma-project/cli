package resources

import (
	"context"
	"testing"

	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
)

func Test_CreateClusterRoleBinding(t *testing.T) {
	t.Parallel()
	t.Run("create ClusterRoleBinding", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixClusterRole()),
		}
		err := CreateClusterRoleBinding(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.NoError(t, err)

		binding, err := kubeClient.Static().RbacV1().ClusterRoleBindings().Get(context.Background(), "test-name-binding", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixClusterRoleBinding(), binding)
	})

	t.Run("ClusterRole not found error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(),
		}
		err := CreateClusterRoleBinding(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.ErrorContains(t, err, `clusterroles.rbac.authorization.k8s.io "clusterRole" not found`)
	})

	t.Run("ignore already exists error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixClusterRole(), fixClusterRoleBinding()),
		}
		err := CreateClusterRoleBinding(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.NoError(t, err)
	})
}

func Test_CreateServiceAccountToken(t *testing.T) {
	t.Parallel()
	t.Run("create Secret with token", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(),
		}
		err := CreateServiceAccountToken(context.Background(), kubeClient, "test-name", "default")
		require.NoError(t, err)

		secret, err := kubeClient.Static().CoreV1().Secrets("default").Get(context.Background(), "test-name", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixTokenSecret(), secret)
	})

	t.Run("ignore already exists error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixTokenSecret()),
		}
		err := CreateServiceAccountToken(context.Background(), kubeClient, "test-name", "default")
		require.NoError(t, err)
	})
}

func Test_CreateServiceAccount(t *testing.T) {
	t.Run("create ServiceAccount", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(),
		}
		err := CreateServiceAccount(context.Background(), kubeClient, "test-name", "default")
		require.NoError(t, err)

		serviceAccount, err := kubeClient.Static().CoreV1().ServiceAccounts("default").Get(context.Background(), "test-name", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixServiceAccount(), serviceAccount)
	})

	t.Run("ignore already exists error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixServiceAccount()),
		}
		err := CreateServiceAccount(context.Background(), kubeClient, "test-name", "default")
		require.NoError(t, err)
	})
}

func fixTokenSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "default",
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": "test-name",
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
}

func fixClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-name-binding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test-name",
				Namespace: "default",
			}},

		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "clusterRole",
		},
	}
}

func fixServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "default",
		},
	}
}

func fixClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "clusterRole",
		},
	}
}
