package resources

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/kube"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	Envs            []corev1.EnvVar
}

func CreateDeployment(ctx context.Context, client kube.Client, opts CreateDeploymentOpts) error {
	deployment := buildDeployment(&opts)
	_, err := client.Static().AppsV1().Deployments(opts.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func buildDeployment(opts *CreateDeploymentOpts) *appsv1.Deployment {
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
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: opts.Name,
					Labels: map[string]string{
						"app": opts.Name,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: append(secretVolumes, configVolumes...),
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Name:  opts.Name,
							Image: opts.Image,
							Env: append(opts.Envs, corev1.EnvVar{
								Name:  "SERVICE_BINDING_ROOT",
								Value: "/bindings",
							}),
							VolumeMounts: append(secretVolumeMounts, configVolumeMounts...),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("50m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("512Mi"),
									corev1.ResourceCPU:    resource.MustParse("300m"),
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
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: opts.ImagePullSecret,
			},
		}
	}

	return deployment
}

func buildSecretVolumes(secretNames []string) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	for _, secretName := range secretNames {
		volumeName := fmt.Sprintf("secret-%s", secretName)
		mountPath := fmt.Sprintf("%s%s", SecretMountPathPrefix, secretName)
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}

	return volumes, volumeMounts
}

func buildConfigmapVolumes(configmapsNames []string) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}
	for _, configmapName := range configmapsNames {
		volumeName := fmt.Sprintf("configmap-%s", configmapName)
		mountPath := fmt.Sprintf("%s%s", ConfigmapMountPathPrefix, configmapName)
		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configmapName,
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
		})
	}
	return volumes, volumeMounts
}
