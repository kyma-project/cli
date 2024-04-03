package registry

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	RegistrySecretName  = "serverless-registry-config-default"
	serverlessNamespace = "kyma-system"
)

type RegistryConfig struct {
	DockerConfigJson string
	Username         string
	Password         string
	PullRegAddr      string
	PushRegAddr      string
	IsInternal       bool
}

func GetConfig(ctx context.Context, client kubernetes.Interface) (*RegistryConfig, error) {
	registrySecret, err := client.CoreV1().Secrets(serverlessNamespace).
		Get(ctx, RegistrySecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	isInternal, err := strconv.ParseBool(string(registrySecret.Data["isInternal"]))
	if err != nil {
		return nil, err
	}

	return &RegistryConfig{
		DockerConfigJson: string(registrySecret.Data[".dockerconfigjson"]),
		Username:         string(registrySecret.Data["username"]),
		Password:         string(registrySecret.Data["password"]),
		PullRegAddr:      string(registrySecret.Data["pullRegAddr"]),
		PushRegAddr:      string(registrySecret.Data["pushRegAddr"]),
		IsInternal:       isInternal,
	}, nil
}

type RegistryPodMeta struct {
	Name      string
	Namespace string
	Port      string
}

func GetWorkloadMeta(ctx context.Context, client kubernetes.Interface, config *RegistryConfig) (*RegistryPodMeta, error) {
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
