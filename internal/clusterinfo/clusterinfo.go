package clusterinfo

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	k8sConfigMap string = "kyma-cluster-info"
	k8sNamespace string = "kube-system"
)

// ClusterInfo contains data about the current cluster
type ClusterInfo struct {
	k8sClient   kubernetes.Interface
	initialized bool
	isLocal     bool
	provider    string
}

// NewClusterInfo creates a new cluster info instance
func NewClusterInfo(k8sClient kubernetes.Interface) *ClusterInfo {
	return &ClusterInfo{k8sClient: k8sClient}
}

// Exists verify whether the cluster provides cluster information data
func (c *ClusterInfo) Exists() (bool, error) {
	cm, err := c.k8sClient.CoreV1().ConfigMaps(k8sNamespace).Get(context.Background(), k8sConfigMap, metav1.GetOptions{})
	// verify whether error occurred which is not related ot the lookup
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return (cm != nil), nil
}

// Write cluster information into cluster
func (c *ClusterInfo) Write(provider string, isLocal bool) error {
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
			"provider": provider,
			"isLocal":  strconv.FormatBool(isLocal),
		},
	}, metav1.CreateOptions{})

	// remember state
	c.provider = provider
	c.isLocal = isLocal
	c.initialized = true

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

	cm, err := c.k8sClient.CoreV1().ConfigMaps(k8sNamespace).Get(context.Background(), k8sConfigMap, metav1.GetOptions{})
	if err != nil {
		return err
	}

	c.provider = cm.Data["provider"]
	c.isLocal, err = strconv.ParseBool(cm.Data["isLocal"])
	if err != nil {
		return err
	}
	c.initialized = true

	return nil
}

//IsLocal returns true if kubernetes cluster runs locally
func (c *ClusterInfo) IsLocal() (bool, error) {
	if err := c.isInitialized(); err != nil {
		return false, err
	}
	return c.isLocal, nil
}

//GetProvider returns the cluster provider
func (c *ClusterInfo) GetProvider() (string, error) {
	if err := c.isInitialized(); err != nil {
		return "", err
	}
	return c.provider, nil
}

func (c *ClusterInfo) isInitialized() error {
	if !c.initialized {
		return fmt.Errorf("ClusterInfo not initialized. Either write or read it from cluster first")
	}
	return nil
}
