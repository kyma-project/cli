package resources

import (
	"context"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	kube_fake "github.com/kyma-project/cli.v3/internal/kube/fake"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func Test_CreateDeployment(t *testing.T) {
	t.Parallel()
	t.Run("create deployment", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(),
		}
		trueValue := true

		// Create MountArray for secrets (backward compatibility - just names)
		secretMounts := types.MountArray{}
		_ = secretMounts.Set("test-name")

		// Create MountArray for configmaps (backward compatibility - just names)
		configMounts := types.MountArray{}
		_ = configMounts.Set("test-name")

		err := CreateDeployment(context.Background(), kubeClient, CreateDeploymentOpts{
			Name:                       "test-name",
			Namespace:                  "default",
			Image:                      "test:image",
			ImagePullSecret:            "test-image-pull-secret",
			InjectIstio:                types.NullableBool{Value: &trueValue},
			SecretMounts:               secretMounts,
			ConfigmapMounts:            configMounts,
			ServiceBindingSecretMounts: types.ServiceBindingSecretArray{}, // empty
			Envs:                       fixDeploymentCustomEnvs(),
		})
		require.NoError(t, err)

		deploy, err := kubeClient.Static().AppsV1().Deployments("default").Get(context.Background(), "test-name", metav1.GetOptions{})
		require.NoError(t, err)
		require.Equal(t, fixDeployment(), deploy)
	})

	t.Run("already exists error", func(t *testing.T) {
		kubeClient := &kube_fake.KubeClient{
			TestKubernetesInterface: k8s_fake.NewSimpleClientset(fixDeployment()),
		}
		trueValue := true

		// Create MountArray for secrets (backward compatibility - just names)
		secretMounts := types.MountArray{}
		_ = secretMounts.Set("test-name")

		// Create MountArray for configmaps (backward compatibility - just names)
		configMounts := types.MountArray{}
		_ = configMounts.Set("test-name")

		err := CreateDeployment(context.Background(), kubeClient, CreateDeploymentOpts{
			Name:                       "test-name",
			Namespace:                  "default",
			Image:                      "test:image",
			ImagePullSecret:            "test-image-pull-secret",
			InjectIstio:                types.NullableBool{Value: &trueValue},
			SecretMounts:               secretMounts,
			ConfigmapMounts:            configMounts,
			ServiceBindingSecretMounts: types.ServiceBindingSecretArray{}, // empty
			Envs:                       fixDeploymentCustomEnvs(),
		})
		require.ErrorContains(t, err, `deployments.apps "test-name" already exists`)
	})
}

func fixDeploymentCustomEnvs() []corev1.EnvVar {
	return []corev1.EnvVar{
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
				ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
					Key: "password",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "my-cm",
					},
				},
			},
		},
		{
			Name:  "test",
			Value: "value",
		},
	}
}

func fixDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-name",
			Namespace: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name":       "test-name",
				"app.kubernetes.io/created-by": "kyma-cli",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-name",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-name",
					Labels: map[string]string{
						"app":                     "test-name",
						"sidecar.istio.io/inject": "true",
					},
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: ptr.To(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    ptr.To(int64(1000)),
						RunAsGroup:   ptr.To(int64(3000)),
						RunAsNonRoot: ptr.To(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
						AppArmorProfile: &corev1.AppArmorProfile{
							Type: corev1.AppArmorProfileTypeRuntimeDefault,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "tmp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "secret-test-name-0",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "test-name",
								},
							},
						},
						{
							Name: "configmap-test-name-0",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-name",
									},
								},
							},
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{
							Name: "test-image-pull-secret",
						},
					},
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Name:  "test-name",
							Image: "test:image",
							Env: append(fixDeploymentCustomEnvs(), corev1.EnvVar{
								Name:  "SERVICE_BINDING_ROOT",
								Value: "/bindings",
							}),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tmp",
									MountPath: "/tmp",
								},
								{
									Name:      "secret-test-name-0",
									MountPath: "/bindings/secret-test-name",
									ReadOnly:  false,
								},
								{
									Name:      "configmap-test-name-0",
									MountPath: "/bindings/configmap-test-name",
									ReadOnly:  false,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged:               ptr.To(false),
								AllowPrivilegeEscalation: ptr.To(false),
								RunAsNonRoot:             ptr.To(true),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{
										"All",
									},
								},
								ReadOnlyRootFilesystem: ptr.To(true),
							},
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
}
