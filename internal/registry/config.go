package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type RegistryConfig struct {
	SecretName string
	SecretData *SecretData
	PodMeta    *RegistryPodMeta
}

func GetConfig(ctx context.Context, client kube.Client) (*RegistryConfig, clierror.Error) {
	config, err := getConfig(ctx, client)
	if err != nil {
		return nil, clierror.Wrap(err,
			clierror.New("failed to load in-cluster registry configuration",
				"make sure cluster is available and properly configured",
				"make sure the Docker Registry is installed and in Ready/Warning state.",
			),
		)
	}

	return config, nil
}

func getConfig(ctx context.Context, client kube.Client) (*RegistryConfig, error) {
	dockerRegistry, err := getDockerRegistry(ctx, client.Dynamic())
	if err != nil {
		return nil, err
	}

	secretConfig, err := getRegistrySecretData(ctx, client.Static(), dockerRegistry.Status.InternalAccess.SecretName, dockerRegistry.GetNamespace())
	if err != nil {
		return nil, err
	}

	podMeta, err := getWorkloadMeta(ctx, client.Static(), secretConfig)
	if err != nil {
		return nil, err
	}

	return &RegistryConfig{
		SecretName: dockerRegistry.Status.InternalAccess.SecretName,
		SecretData: secretConfig,
		PodMeta:    podMeta,
	}, nil
}

type SecretData struct {
	DockerConfigJSON string
	Username         string
	Password         string
	PullRegAddr      string
	PushRegAddr      string
}

func getRegistrySecretData(ctx context.Context, client kubernetes.Interface, secretName, secretNamespace string) (*SecretData, error) {
	registrySecret, err := client.CoreV1().Secrets(secretNamespace).
		Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return &SecretData{
		DockerConfigJSON: string(registrySecret.Data[".dockerconfigjson"]),
		Username:         string(registrySecret.Data["username"]),
		Password:         string(registrySecret.Data["password"]),
		PullRegAddr:      string(registrySecret.Data["pullRegAddr"]),
		PushRegAddr:      string(registrySecret.Data["pushRegAddr"]),
	}, nil
}

type RegistryPodMeta struct {
	Name      string
	Namespace string
	Port      string
}

func getWorkloadMeta(ctx context.Context, client kubernetes.Interface, config *SecretData) (*RegistryPodMeta, error) {
	// expected pushRegAddr format - serverless-docker-registry.kyma-system.svc.cluster.local:5000
	hostPort := strings.Split(config.PushRegAddr, ":")

	hostElems := strings.Split(hostPort[0], ".")
	svcName := hostElems[0]
	svcNamespace := hostElems[1]

	registrySvc, err := client.CoreV1().Services(svcNamespace).Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	registryPods, err := client.CoreV1().Pods(svcNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelectorFor(registrySvc.Spec.Selector),
	})
	if err != nil {
		return nil, err
	}

	readyRegistryPod, err := getReadyPod(registryPods.Items)
	if err != nil {
		return nil, err
	}

	return &RegistryPodMeta{
		Name:      readyRegistryPod.GetName(),
		Namespace: readyRegistryPod.GetNamespace(),
		Port:      registrySvc.Spec.Ports[0].TargetPort.String(),
	}, nil
}

func labelSelectorFor(labels map[string]string) string {
	labelSelectors := []string{}
	for key, value := range labels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(labelSelectors, ",")
}

func getReadyPod(pods []corev1.Pod) (*corev1.Pod, error) {
	for _, registryPod := range pods {
		if isPodReady(registryPod) {
			return &registryPod, nil
		}
	}

	return nil, errors.New("no ready registry pod found")
}

func isPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.ContainersReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

func getDockerRegistry(ctx context.Context, c dynamic.Interface) (*DockerRegistry, error) {
	list, err := c.Resource(DockerRegistryGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		var dockerRegistry DockerRegistry
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, &dockerRegistry)
		if err != nil {
			return nil, err
		}

		if dockerRegistry.Status.State == "Ready" || dockerRegistry.Status.State == "Warning" {
			return &dockerRegistry, nil
		}
	}

	return nil, errors.New("no installed docker registry found")
}
