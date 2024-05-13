package fake

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Fake client for testing purposes
// It implements the Client interface and returns given values only
type FakeKubeClient struct {
	TestKubernetesInterface kubernetes.Interface
	TestDynamicInterface    dynamic.Interface
	TestRestClient          *rest.RESTClient
	TestRestConfig          *rest.Config
	TestApiConfig           *api.Config
}

func (f *FakeKubeClient) Static() kubernetes.Interface {
	return f.TestKubernetesInterface
}

func (f *FakeKubeClient) Dynamic() dynamic.Interface {
	return f.TestDynamicInterface
}

func (f *FakeKubeClient) RestClient() *rest.RESTClient {
	return f.TestRestClient
}

func (f *FakeKubeClient) RestConfig() *rest.Config {
	return f.TestRestConfig
}

func (f *FakeKubeClient) ApiConfig() *api.Config {
	return f.TestApiConfig
}
