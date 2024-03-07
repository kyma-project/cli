package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Client interface contains all needed cluster-oriented clients to allow cluster-wide manipulations
type Client interface {
	Static() kubernetes.Interface
	RestConfig() *rest.Config
	ApiConfig() *api.Config
}

type client struct {
	restConfig *rest.Config
	apiConfig  *api.Config
	kubeClient kubernetes.Interface
}

func NewClient(kubeconfig string) (Client, error) {
	restConfig, err := restConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	apiConfig, err := apiConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		restConfig: restConfig,
		apiConfig:  apiConfig,
		kubeClient: kubeClient,
	}, nil
}

func (c *client) Static() kubernetes.Interface {
	return c.kubeClient
}

func (c *client) RestConfig() *rest.Config {
	return c.restConfig
}

func (c *client) ApiConfig() *api.Config {
	return c.apiConfig
}
