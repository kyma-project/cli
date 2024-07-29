package fake

import (
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Fake client for testing purposes
// It implements the Client interface and returns given values only
type FakeKubeClient struct {
	TestKubernetesInterface      kubernetes.Interface
	TestDynamicInterface         dynamic.Interface
	TestKymaInterface            kyma.Interface
	TestBtpInterface             btp.Interface
	TestRestClient               *rest.RESTClient
	TestRestConfig               *rest.Config
	TestAPIConfig                *api.Config
	TestRootlessDynamicInterface rootlessdynamic.Interface
}

func (f *FakeKubeClient) Static() kubernetes.Interface {
	return f.TestKubernetesInterface
}

func (f *FakeKubeClient) Dynamic() dynamic.Interface {
	return f.TestDynamicInterface
}

func (f *FakeKubeClient) Btp() btp.Interface {
	return f.TestBtpInterface
}

func (f *FakeKubeClient) RestClient() *rest.RESTClient {
	return f.TestRestClient
}

func (f *FakeKubeClient) RestConfig() *rest.Config {
	return f.TestRestConfig
}

func (f *FakeKubeClient) APIConfig() *api.Config {
	return f.TestAPIConfig
}

func (f *FakeKubeClient) Kyma() kyma.Interface {
	return f.TestKymaInterface
}

func (f *FakeKubeClient) RootlessDynamic() rootlessdynamic.Interface {
	return f.TestRootlessDynamicInterface
}
