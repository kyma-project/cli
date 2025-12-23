package resources

import (
	"context"
	"strings"

	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	RoleKind        string = "Role"
	ClusterRoleKind string = "ClusterRole"
)

func EnsureServiceAccount(ctx context.Context, client kube.Client, name, namespace string) error {
	sa := buildServiceAccount(name, namespace)
	_, err := client.Static().CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func CreateServiceAccountToken(ctx context.Context, client kube.Client, name, namespace string) error {
	secret := buildServiceAccountToken(name, namespace)
	_, err := client.Static().CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func CreateClusterRoleBinding(ctx context.Context, client kube.Client, name, namespace, clusterRole string) error {
	// Check if the cluster role to bind to exists
	_, err := client.Static().RbacV1().ClusterRoles().Get(ctx, clusterRole, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Create clusterRoleBinding
	cRoleBinding := buildClusterRoleBinding(name, namespace, clusterRole)
	_, err = client.Static().RbacV1().ClusterRoleBindings().Create(ctx, cRoleBinding, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func CreateRoleBindingToClusterRole(ctx context.Context, client kube.Client, name, namespace, clusterRole string) error {

	// Check if the cluster role to bind to exists
	_, err := client.Static().RbacV1().ClusterRoles().Get(ctx, clusterRole, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Create roleBinding
	roleBinding := buildRoleBinding(name, namespace, clusterRole, ClusterRoleKind)
	_, err = client.Static().RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func CreateRoleBindingToRole(ctx context.Context, client kube.Client, name, namespace, role string) error {

	// Check if the role to bind to exists
	_, err := client.Static().RbacV1().Roles(namespace).Get(ctx, role, metav1.GetOptions{})
	if err != nil {
		return err
	}
	// Create roleBinding
	roleBinding := buildRoleBinding(name, namespace, role, RoleKind)
	_, err = client.Static().RbacV1().RoleBindings(namespace).Create(ctx, roleBinding, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func BindingExists(ctx context.Context, client kube.Client, binding *unstructured.Unstructured) bool {
	name := binding.GetName()
	kind := binding.GetKind()
	namespace := binding.GetNamespace()

	var err error

	switch kind {
	case "ClusterRoleBinding":
		_, err = client.Static().RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
	case "RoleBinding":
		if namespace == "" {
			return false
		}
		_, err = client.Static().RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
	default:
		return false
	}

	return err == nil
}

func buildServiceAccountToken(name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": name,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}
}

func buildClusterRoleBinding(name, namespace, clusterRole string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-clusterrole-" + clusterRole + "-binding",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			}},

		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: clusterRole,
		},
	}
}

func buildRoleBinding(name, namespace, role, roleKind string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-" + strings.ToLower(roleKind) + "-" + role + "-binding",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: namespace,
			}},

		RoleRef: rbacv1.RoleRef{
			Kind: roleKind,
			Name: role,
		},
	}
}

func buildServiceAccount(name, namespace string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
}
