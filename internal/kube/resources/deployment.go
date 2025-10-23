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
	"k8s.io/utils/ptr"
)

const (
	SecretMountPathPrefix    = "/bindings/secret-"
	ConfigmapMountPathPrefix = "/bindings/configmap-"
)

type CreateDeploymentOpts struct {
	Name                       string
	Namespace                  string
	Image                      string
	ImagePullSecret            string
	InjectIstio                types.NullableBool
	SecretMounts               types.MountArray
	ConfigmapMounts            types.MountArray
	ServiceBindingSecretMounts types.ServiceBindingSecretArray
	Envs                       []corev1.EnvVar
	Insecure                   bool
}

func CreateDeployment(ctx context.Context, client kube.Client, opts CreateDeploymentOpts) error {
	deployment := buildDeployment(&opts)
	_, err := client.Static().AppsV1().Deployments(opts.Namespace).Create(ctx, deployment, metav1.CreateOptions{})
	return err
}

func buildDeployment(opts *CreateDeploymentOpts) *appsv1.Deployment {
	secretVolumes, secretVolumeMounts := buildSecretVolumes(opts.SecretMounts)
	configVolumes, configVolumeMounts := buildConfigmapVolumes(opts.ConfigmapMounts)
	serviceBindingVolumes, serviceBindingVolumeMounts := buildServiceBindingSecretVolumes(opts.ServiceBindingSecretMounts)

	volumes := []corev1.Volume{
		{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
	volumes = append(volumes, secretVolumes...)
	volumes = append(volumes, configVolumes...)
	volumes = append(volumes, serviceBindingVolumes...)

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "tmp",
			MountPath: "/tmp",
		},
	}
	volumeMounts = append(volumeMounts, secretVolumeMounts...)
	volumeMounts = append(volumeMounts, configVolumeMounts...)
	volumeMounts = append(volumeMounts, serviceBindingVolumeMounts...)

	podSecCtx, secCtx := buildSecurityContext(opts.Insecure)

	// Build environment variables - only add SERVICE_BINDING_ROOT if service binding secrets are mounted
	envVars := opts.Envs
	if len(opts.ServiceBindingSecretMounts.Names) > 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "SERVICE_BINDING_ROOT",
			Value: "/bindings",
		})
	}

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
					Volumes:                      volumes,
					AutomountServiceAccountToken: ptr.To(false),
					SecurityContext:              podSecCtx,
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Name:            opts.Name,
							Image:           opts.Image,
							Env:             envVars,
							VolumeMounts:    volumeMounts,
							SecurityContext: secCtx,
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

// buildSecretVolumes builds volumes and volume mounts for secrets using the MountArray type
func buildSecretVolumes(mountArray types.MountArray) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	for i, mount := range mountArray.Mounts {
		volumeName := fmt.Sprintf("secret-%s-%d", mount.Name, i)

		// Use custom path if specified, otherwise use default path
		mountPath := mount.Path
		if mountPath == "" {
			mountPath = fmt.Sprintf("%s%s", SecretMountPathPrefix, mount.Name)
		}

		secretVolumeSource := &corev1.SecretVolumeSource{
			SecretName: mount.Name,
		}

		// If a specific key is used, mount only that key
		if mount.Key != "" {
			secretVolumeSource.Items = []corev1.KeyToPath{
				{
					Key:  mount.Key,
					Path: mount.Key,
				},
			}
		}

		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: secretVolumeSource,
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  mount.ReadOnly,
		})
	}

	return volumes, volumeMounts
}

func buildSecurityContext(insecure bool) (*corev1.PodSecurityContext, *corev1.SecurityContext) {
	if insecure {
		return nil, nil
	}

	secCtx := &corev1.SecurityContext{
		Privileged:               ptr.To(false),
		AllowPrivilegeEscalation: ptr.To(false),
		RunAsNonRoot:             ptr.To(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{
				"All",
			},
		},
		ReadOnlyRootFilesystem: ptr.To(true),
	}

	podSecCtx := &corev1.PodSecurityContext{
		RunAsUser:    ptr.To(int64(1000)),
		RunAsGroup:   ptr.To(int64(3000)),
		RunAsNonRoot: ptr.To(true),
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
		AppArmorProfile: &corev1.AppArmorProfile{
			Type: corev1.AppArmorProfileTypeRuntimeDefault,
		},
	}

	return podSecCtx, secCtx
}

// buildConfigmapVolumes builds volumes and volume mounts for configmaps using the MountArray type
func buildConfigmapVolumes(mountArray types.MountArray) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	for i, mount := range mountArray.Mounts {
		volumeName := fmt.Sprintf("configmap-%s-%d", mount.Name, i)

		// Use custom path if specified, otherwise use default path
		mountPath := mount.Path
		if mountPath == "" {
			mountPath = fmt.Sprintf("%s%s", ConfigmapMountPathPrefix, mount.Name)
		}

		configMapVolumeSource := &corev1.ConfigMapVolumeSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: mount.Name,
			},
		}

		// If a specific key is requested, mount only that key
		if mount.Key != "" {
			configMapVolumeSource.Items = []corev1.KeyToPath{
				{
					Key:  mount.Key,
					Path: mount.Key,
				},
			}
		}

		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: configMapVolumeSource,
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: mountPath,
			ReadOnly:  mount.ReadOnly,
		})
	}

	return volumes, volumeMounts
}

// buildServiceBindingSecretVolumes builds volumes and volume mounts for service binding secrets
func buildServiceBindingSecretVolumes(serviceBindingSecrets types.ServiceBindingSecretArray) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	for _, secretName := range serviceBindingSecrets.Names {
		volumeName := fmt.Sprintf("service-binding-secret-%s", secretName)
		mountPath := fmt.Sprintf("/bindings/secret-%s", secretName)

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
			ReadOnly:  true, // Service binding secrets are always read-only
		})
	}

	return volumes, volumeMounts
}
