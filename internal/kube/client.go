package kube

import (
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"github.com/pkg/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Client interface contains all needed cluster-oriented clients to allow cluster-wide manipulations
type Client interface {
	Static() kubernetes.Interface
	Dynamic() dynamic.Interface
	Kyma() kyma.Interface
	Btp() btp.Interface
	RootlessDynamic() rootlessdynamic.Interface
	RestClient() *rest.RESTClient
	RestConfig() *rest.Config
	APIConfig() *api.Config
}

type client struct {
	restConfig     *rest.Config
	apiConfig      *api.Config
	kymaClient     kyma.Interface
	rootlessClient rootlessdynamic.Interface
	kubeClient     kubernetes.Interface
	dynamicClient  dynamic.Interface
	btpClient      btp.Interface
	restClient     *rest.RESTClient
}

func NewClient(kubeconfig string) (Client, error) {
	client, err := newClient(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialise kubernetes client")
	}
	return client, nil
}

func newClient(kubeconfig string) (Client, error) {
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

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	kymaClient := kyma.NewClient(dynamicClient)

	discovery, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	rootlessClient := rootlessdynamic.NewClient(dynamicClient, discovery)

	btpClient := btp.NewClient(dynamicClient)

	restClientConfig := *restConfig
	err = setKubernetesDefaults(&restClientConfig)
	if err != nil {
		return nil, err
	}

	restClient, err := rest.RESTClientFor(&restClientConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		restConfig:     restConfig,
		apiConfig:      apiConfig,
		kubeClient:     kubeClient,
		kymaClient:     kymaClient,
		rootlessClient: rootlessClient,
		dynamicClient:  dynamicClient,
		btpClient:      btpClient,
		restClient:     restClient,
	}, nil
}

func (c *client) Static() kubernetes.Interface {
	return c.kubeClient
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamicClient
}

func (c *client) Kyma() kyma.Interface {
	return c.kymaClient
}

func (c *client) Btp() btp.Interface {
	return c.btpClient
}

func (c *client) RootlessDynamic() rootlessdynamic.Interface {
	return c.rootlessClient
}

func (c *client) RestClient() *rest.RESTClient {
	return c.restClient // TODO: Update schema - meanwhile can use kubeclient.Static().Corev1().RESTClient()
}
func (c *client) RestConfig() *rest.Config {
	return c.restConfig
}

func (c *client) APIConfig() *api.Config {
	return c.apiConfig
}
