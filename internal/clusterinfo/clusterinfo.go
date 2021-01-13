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

//ClusterProvider is a unique identifier for a cluster provider
type ClusterProvider string

const (
	k8sConfigMap string = "kyma-cluster-info"
	k8sNamespace string = "kube-system"

	//ClusterProviderK3s indicates that K3s is used as cluster provider
	ClusterProviderK3s ClusterProvider = "k3s"
	//ClusterProviderGardener indicates that Gardener is used as cluster provider
	ClusterProviderGardener ClusterProvider = "gardener"
	//ClusterProviderAzure indicates that Azure is used as cluster provider
	ClusterProviderAzure ClusterProvider = "azure"
	//ClusterProviderGcp indicates that GCP is used as cluster provider
	ClusterProviderGcp ClusterProvider = "gcp"
	//ClusterProviderAws indicates that AWS is used as cluster provider
	ClusterProviderAws ClusterProvider = "aws"
)

// ClusterInfo contains data about the current cluster
type ClusterInfo struct {
	k8sClient   kubernetes.Interface
	initialized bool
	local       bool
	provider    ClusterProvider
}

// New creates a new cluster info instance
func New(k8sClient kubernetes.Interface) *ClusterInfo {
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
func (c *ClusterInfo) Write(provider ClusterProvider, local bool) error {
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

	c.provider = ClusterProvider(cm.Data["provider"])
	c.local, err = strconv.ParseBool(cm.Data["local"])
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
	return c.local, nil
}

//Provider returns the cluster provider
func (c *ClusterInfo) Provider() (ClusterProvider, error) {
	if err := c.isInitialized(); err != nil {
		return "", err
	}
	return ClusterProvider(c.provider), nil
}

func (c *ClusterInfo) isInitialized() error {
	if !c.initialized {
		return fmt.Errorf("ClusterInfo not initialized. Either write or read it from cluster first")
	}
	return nil
}
