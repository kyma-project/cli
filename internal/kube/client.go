package kube

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Client interface contains all needed cluster-oriented clients to allow cluster-wide manipulations
type Client interface {
	Static() kubernetes.Interface
	Dynamic() dynamic.Interface
	RestClient() *rest.RESTClient
	RestConfig() *rest.Config
	ApiConfig() *api.Config
}

type client struct {
	restConfig    *rest.Config
	apiConfig     *api.Config
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	restClient    *rest.RESTClient
}

func NewClient(kubeconfig string) (Client, clierror.Error) {
	client, err := newClient(kubeconfig)
	if err != nil {
		return nil, clierror.Wrap(err,
			clierror.New("failed to initialise kubernetes client", "Make sure that kubeconfig is proper."),
		)
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
		restConfig:    restConfig,
		apiConfig:     apiConfig,
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		restClient:    restClient,
	}, nil
}

func (c *client) Static() kubernetes.Interface {
	return c.kubeClient
}

func (c *client) Dynamic() dynamic.Interface {
	return c.dynamicClient
}

func (c *client) RestClient() *rest.RESTClient {
	return c.restClient // TODO: Update schema - can use kubeclient.Static().Corev1().RESTClient()
}

func (c *client) RestConfig() *rest.Config {
	return c.restConfig
}

func (c *client) ApiConfig() *api.Config {
	return c.apiConfig
}
