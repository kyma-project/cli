package clusterinfo

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ClusterType string

const (
	k8sConfigMap string = "kyma-cluster-info"
	k8sNamespace string = "kube-system"
)

// ClusterInfo contains data about the current cluster
type ClusterInfo struct {
	k8sClient   kubernetes.Interface
	initialized bool
	local       bool
	provider    ClusterType
}

// New creates a new cluster info instance
func New(k8sClient kubernetes.Interface) *ClusterInfo {
	return &ClusterInfo{k8sClient: k8sClient}
}

// Write cluster information into cluster
func (c *ClusterInfo) Write(provider ClusterType, local bool) error {
	if provider == "" {
		return fmt.Errorf("Cluster provider cannot be empty")
	}

	// write config-map
	_, err := c.k8sClient.CoreV1().ConfigMaps(k8sNamespace).Create(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   k8sConfigMap,
			Labels: map[string]string{"app": "kyma"},
		},
		Data: map[string]string{
			"provider": string(provider),
			"local":    strconv.FormatBool(local),
		},
	}, metav1.CreateOptions{})

	// remember state
	c.provider = provider
	c.local = local
	c.initialized = true

	return err
}
