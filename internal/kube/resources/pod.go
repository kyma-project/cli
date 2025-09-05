package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodForSelector(ctx context.Context, client kubernetes.Interface, namespace string, labelSelector map[string]string) (*corev1.Pod, error) {
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: LabelSelectorFor(labelSelector),
	})
	if err != nil {
		return nil, err
	}
	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pod found for selector %s in namespace %s", labelSelector, namespace)
	}

	readyPod, err := GetReadyPod(podList.Items)
	if err != nil {
		return nil, err
	}

	return readyPod, nil
}

func LabelSelectorFor(labels map[string]string) string {
	labelSelectors := []string{}
	for key, value := range labels {
		labelSelectors = append(labelSelectors, fmt.Sprintf("%s=%s", key, value))
	}

	return strings.Join(labelSelectors, ",")
}

func GetReadyPod(pods []corev1.Pod) (*corev1.Pod, error) {
	for _, registryPod := range pods {
		if IsPodReady(registryPod) {
			return &registryPod, nil
		}
	}

	return nil, errors.New("no running registry pod found")
}

func IsPodReady(pod corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.ContainersReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
