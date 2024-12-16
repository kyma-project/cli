package fake

import (
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/kyma-project/cli.v3/internal/kube/istio"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/kube/rootlessdynamic"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Fake client for testing purposes
// It implements the Client interface and returns given values only
type KubeClient struct {
	TestKubernetesInterface      kubernetes.Interface
	TestDynamicInterface         dynamic.Interface
	TestIstioInterface           istio.Interface
	TestKymaInterface            kyma.Interface
	TestBtpInterface             btp.Interface
	TestRestClient               *rest.RESTClient
	TestRestConfig               *rest.Config
	TestAPIConfig                *api.Config
	TestRootlessDynamicInterface rootlessdynamic.Interface
}

func (f *KubeClient) Static() kubernetes.Interface {
	return f.TestKubernetesInterface
}

func (f *KubeClient) Dynamic() dynamic.Interface {
	return f.TestDynamicInterface
}

func (f *KubeClient) Istio() istio.Interface {
	return f.TestIstioInterface
}

func (f *KubeClient) Btp() btp.Interface {
	return f.TestBtpInterface
}

func (f *KubeClient) RestClient() *rest.RESTClient {
	return f.TestRestClient
}

func (f *KubeClient) RestConfig() *rest.Config {
	return f.TestRestConfig
}

func (f *KubeClient) APIConfig() *api.Config {
	return f.TestAPIConfig
}

func (f *KubeClient) Kyma() kyma.Interface {
	return f.TestKymaInterface
}

func (f *KubeClient) RootlessDynamic() rootlessdynamic.Interface {
	return f.TestRootlessDynamicInterface
}
