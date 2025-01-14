package resources

import (
	"context"
	"fmt"

	"github.com/kyma-project/api-gateway/apis/gateway/v2alpha1"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/istio"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
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

const (
	SecretMountPathPrefix    = "/bindings/secret-"
	ConfigmapMountPathPrefix = "/bindings/configmap-"
)

type CreateDeploymentOpts struct {
	Name            string
	Namespace       string
	Image           string
	ImagePullSecret string
	InjectIstio     types.NullableBool
	SecretMounts    []string
	ConfigmapMounts []string
}

func CreateDeployment(ctx context.Context, client kube.Client, opts CreateDeploymentOpts) error {
	secretVolumes, secretVolumeMounts := buildSecretVolumes(opts.SecretMounts)
	configVolumes, configVolumeMounts := buildConfigmapVolumes(opts.ConfigmapMounts)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: opts.Name,
			Labels: map[string]string{
				"app.kubernetes.io/name":       opts.Name,
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": opts.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: opts.Name,
					Labels: map[string]string{
						"app": opts.Name,
					},
				},
				Spec: v1.PodSpec{
					Volumes: append(secretVolumes, configVolumes...),
					Containers: []v1.Container{
						{
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Name:  opts.Name,
							Image: opts.Image,
							Env: []v1.EnvVar{
								{
									Name:  "SERVICE_BINDING_ROOT",
									Value: "/bindings",
								},
							},
							VolumeMounts: append(secretVolumeMounts, configVolumeMounts...),
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("64Mi"),
									v1.ResourceCPU:    resource.MustParse("50m"),
								},
								Limits: v1.ResourceList{
									v1.ResourceMemory: resource.MustParse("512Mi"),
									v1.ResourceCPU:    resource.MustParse("300m"),
								},
							},
						},
					},
				},
			},
		},
	}
	if opts.InjectIstio.Value != nil {
		deployment.Spec.Template.ObjectMeta.Labels["sidecar.istio.io/inject"] = opts.InjectIstio.String()
	}

	if opts.ImagePullSecret != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = []v1.LocalObjectReference{
			{
				Name: opts.ImagePullSecret,
			},
		}
	}

	_, err := client.Static().AppsV1().Deployments(opts.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func buildSecretVolumes(secretNames []string) ([]v1.Volume, []v1.VolumeMount) {
	volumes := []v1.Volume{}
	volumeMounts := []v1.VolumeMount{}
	for _, secretName := range secretNames {
		volumeName := fmt.Sprintf("secret-%s", secretName)
		mountPath := fmt.Sprintf("%s%s", SecretMountPathPrefix, secretName)
		volumes = append(volumes, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}

	return volumes, volumeMounts
}

func buildConfigmapVolumes(configmapsNames []string) ([]v1.Volume, []v1.VolumeMount) {
	volumes := []v1.Volume{}
	volumeMounts := []v1.VolumeMount{}
	for _, configmapName := range configmapsNames {
		volumeName := fmt.Sprintf("configmap-%s", configmapName)
		mountPath := fmt.Sprintf("%s%s", ConfigmapMountPathPrefix, configmapName)
		volumes = append(volumes, v1.Volume{
			Name: volumeName,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configmapName,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}
	return volumes, volumeMounts
}

func CreateService(ctx context.Context, client kube.Client, name, namespace string, port int32) error {
	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": name,
			},
			Ports: []v1.ServicePort{
				{
					Port:       port,
					TargetPort: intstr.FromInt32(port),
				},
			},
		},
	}
	_, err := client.Static().CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	return err
}

func CreateAPIRule(ctx context.Context, client rootlessdynamic.Interface, name, namespace, host string, port uint32) error {
	apirule := v2alpha1.APIRule{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gateway.kyma-project.io/v2alpha1",
			Kind:       "APIRule",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       name,
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: v2alpha1.APIRuleSpec{
			Hosts: []*v2alpha1.Host{
				ptr.To(v2alpha1.Host(host)),
			},
			Gateway: ptr.To(fmt.Sprintf("%s/%s", istio.GatewayNamespace, istio.GatewayName)),
			Rules: []v2alpha1.Rule{
				{
					Path:    "/*",
					Methods: []v2alpha1.HttpMethod{"GET", "POST", "PUT", "DELETE", "PATCH"},
					NoAuth:  ptr.To(true),
				},
			},
			Service: &v2alpha1.Service{
				Name:      ptr.To(name),
				Namespace: ptr.To(namespace),
				Port:      &port,
			},
		},
	}

	uAPIRule, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&apirule)
	if err != nil {
		return err
	}
	return client.Apply(ctx, &unstructured.Unstructured{Object: uAPIRule})
}
