package resources

import (
	"context"
	"strings"
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

		binding, err := kubeClient.Static().RbacV1().ClusterRoleBindings().Get(context.Background(), "test-name-clusterrole-clusterRole-binding", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixClusterRoleBinding(), binding)
	})

	t.Run("create RoleBinding to ClusterRole", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixClusterRole()),
		}
		err := CreateRoleBindingToClusterRole(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.NoError(t, err)

		binding, err := kubeClient.Static().RbacV1().RoleBindings("default").Get(context.Background(), "test-name-clusterrole-clusterRole-binding", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixRoleBinding("clusterRole", ClusterRoleKind), binding)
	})

	t.Run("create RoleBinding to Role", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixRole()),
		}
		err := CreateRoleBindingToRole(context.Background(), kubeClient, "test-name", "default", "role")
		require.NoError(t, err)

		binding, err := kubeClient.Static().RbacV1().RoleBindings("default").Get(context.Background(), "test-name-role-role-binding", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixRoleBinding("role", RoleKind), binding)
	})

	t.Run("ClusterRole not found error for ClusterRoleBinding creation", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(),
		}
		err := CreateClusterRoleBinding(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.ErrorContains(t, err, `clusterroles.rbac.authorization.k8s.io "clusterRole" not found`)
	})

	t.Run("ClusterRole not found error for RoleBinding creation", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(),
		}
		err := CreateRoleBindingToClusterRole(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.ErrorContains(t, err, `clusterroles.rbac.authorization.k8s.io "clusterRole" not found`)

		err = CreateRoleBindingToRole(context.Background(), kubeClient, "test-name", "default", "role")
		require.ErrorContains(t, err, `roles.rbac.authorization.k8s.io "role" not found`)
	})

	t.Run("ignore already exists error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixClusterRole(), fixRoleBinding("clusterRole", ClusterRoleKind)),
		}
		err := CreateClusterRoleBinding(context.Background(), kubeClient, "test-name", "default", "clusterRole")
		require.NoError(t, err)
	})
}

func Test_EnsureServiceAccountToken(t *testing.T) {
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
		err := EnsureServiceAccount(context.Background(), kubeClient, "test-name", "default")
		require.NoError(t, err)

		serviceAccount, err := kubeClient.Static().CoreV1().ServiceAccounts("default").Get(context.Background(), "test-name", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixServiceAccount(), serviceAccount)
	})

	t.Run("ignore already exists error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixServiceAccount()),
		}
		err := EnsureServiceAccount(context.Background(), kubeClient, "test-name", "default")
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
			Name: "test-name-clusterrole-clusterRole-binding",
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

func fixRoleBinding(role, roleKind string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name-" + strings.ToLower(roleKind) + "-" + role + "-binding",
			Namespace: "default",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      "test-name",
				Namespace: "default",
			}},

		RoleRef: rbacv1.RoleRef{
			Kind: roleKind,
			Name: role,
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

func fixRole() *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role",
			Namespace: "default",
		},
	}
}
