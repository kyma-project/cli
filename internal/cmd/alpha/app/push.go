package app

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type appPushConfig struct {
	*cmdcommon.KymaConfig

	name      string
	namespace string
	image     string
	// containerPort int
}

func NewAppPushCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := appPushConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push the application to the Kyma cluster.",
		Long:  "Use this command to push the application to the Kyma cluster.",
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runAppPush(&config))
		},
	}

	cmd.Flags().StringVar(&config.name, "name", "", "Name of the app")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace where app should be deployed")
	cmd.Flags().StringVar(&config.image, "image", "", "Name of the image to deploy")
	// cmd.Flags().IntVar(&config.containerPort, "containerPort", 80, "")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("image")

	return cmd
}

func runAppPush(cfg *appPushConfig) clierror.Error {
	client, err := cfg.GetKubeClientWithClierr()
	if err != nil {
		return err
	}
	createDeployment(cfg.Ctx, client, cfg.name, cfg.namespace, cfg.image)

	return nil
}

func createDeployment(ctx context.Context, client kube.Client, name, namespace, image string) {
	deployment := prepareDeployment(name, image)

	_, err := client.Static().AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func prepareDeployment(name, image string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"app.kubernetes.io/name":        name,
				"app.kubernetes.io/created-by:": "kyma-cli",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"app":                     name,
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("50m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("128Mi"),
									corev1.ResourceCPU:    resource.MustParse("100m"),
								},
							},
						},
					},
				},
			},
		},
	}
}
