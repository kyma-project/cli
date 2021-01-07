package clusterinfo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	k8sConfigMap string = "kyma-cluster-info"
	k8sNamespace string = "kube-system"
)

// ClusterInfo contains data about the current cluster
type ClusterInfo struct {
	K8sClient kubernetes.Interface
	IsLocal   bool
	Provider  string
}

// Exists verify whether the cluster provides cluster information data
func (c *ClusterInfo) Exists() (bool, error) {
	cm, err := c.K8sClient.CoreV1().ConfigMaps(k8sNamespace).Get(context.Background(), k8sConfigMap, metav1.GetOptions{})
	// verify whether error occurred which is not related ot the lookup
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return false, err
	}
	return (cm != nil), nil
}

// Write cluster information into cluster
func (c *ClusterInfo) Write() error {
	if c.Provider == "" {
		return fmt.Errorf("Please defined a cluster provider before updating the cluster info")
	}

	// write config-map
	_, err := c.K8sClient.CoreV1().ConfigMaps(k8sNamespace).Create(context.Background(), &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   k8sConfigMap,
			Labels: map[string]string{"app": "kyma"},
		},
		Data: map[string]string{
			"provider": c.Provider,
			"isLocal":  strconv.FormatBool(c.IsLocal),
		},
	}, metav1.CreateOptions{})

	return err
}

// Read cluster information from cluster. Method will fail if cluster information doesn't exist.
func (c *ClusterInfo) Read() error {
	exists, err := c.Exists()
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("Cluster information not found")
	}

	cm, err := c.K8sClient.CoreV1().ConfigMaps(k8sNamespace).Get(context.Background(), k8sConfigMap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	c.Provider = cm.Data["provider"]
	c.IsLocal, err = strconv.ParseBool(cm.Data["isLocal"])

	return err
}
