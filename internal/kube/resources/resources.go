package resources

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateServiceAccount(ctx context.Context, client kube.Client, name, namespace string) error {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	_, err := client.Static().CoreV1().ServiceAccounts(namespace).Create(ctx, &sa, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func CreateServiceAccountToken(ctx context.Context, client kube.Client, name, namespace string) error {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": name,
			},
		},
		Type: v1.SecretTypeServiceAccountToken,
	}

	_, err := client.Static().CoreV1().Secrets(namespace).Create(ctx, &secret, metav1.CreateOptions{})
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
	cRoleBinding := rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-binding",
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
	_, err = client.Static().RbacV1().ClusterRoleBindings().Create(ctx, &cRoleBinding, metav1.CreateOptions{})
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func CreateDeployment(ctx context.Context, client kube.Client, name, namespace, image string) error {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"app":                     name,
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name,
							Image: image,
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("64Mi"),
									v1.ResourceCPU:    resource.MustParse("50m"),
								},
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("128Mi"),
									v1.ResourceCPU:    resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := client.Static().AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}
