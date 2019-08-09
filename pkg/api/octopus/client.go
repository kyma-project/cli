package octopus

import (
	"github.com/kyma-incubator/octopus/pkg/apis"
	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const NamespaceForTests = "kyma-system"

type OctopusInterface interface {
	ListTestDefinitions(opts metav1.ListOptions) (result *oct.TestDefinitionList, err error)
	ListTestSuites(opts metav1.ListOptions) (result *oct.ClusterTestSuiteList, err error)
	CreateTestSuite(cts *oct.ClusterTestSuite) (result *oct.ClusterTestSuite, err error)
	DeleteTestSuite(name string, options metav1.DeleteOptions) error
	GetTestSuite(name string, options metav1.GetOptions) (result *oct.ClusterTestSuite, err error)
}

type OctopusRestClient struct {
	restClient rest.Interface
}

func NewFromConfig(config *rest.Config) (OctopusInterface, error) {
	apis.AddToScheme(scheme.Scheme)
	setConfigDefaults(config)

	cl, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return &OctopusRestClient{
		restClient: cl,
	}, nil
}

func (t *OctopusRestClient) ListTestDefinitions(opts metav1.ListOptions) (result *oct.TestDefinitionList, err error) {
	result = &oct.TestDefinitionList{}
	err = t.restClient.Get().
		Namespace(NamespaceForTests).
		Resource("testdefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

func (t *OctopusRestClient) ListTestSuites(opts metav1.ListOptions) (result *oct.ClusterTestSuiteList, err error) {
	result = &oct.ClusterTestSuiteList{}
	err = t.restClient.Get().
		Namespace(NamespaceForTests).
		Resource("clustertestsuites").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

func (t *OctopusRestClient) CreateTestSuite(cts *oct.ClusterTestSuite) (result *oct.ClusterTestSuite, err error) {
	result = &oct.ClusterTestSuite{}
	err = t.restClient.Post().
		Namespace(NamespaceForTests).
		Resource("clustertestsuites").
		Body(cts).
		Do().
		Into(result)
	return
}

func (t *OctopusRestClient) DeleteTestSuite(name string, options metav1.DeleteOptions) error {
	return t.restClient.Delete().
		Namespace(NamespaceForTests).
		Resource("clustertestsuites").
		Name(name).
		// Reenable this when deleteing supports options
		//Body(options).
		Do().
		Error()
}

func (t *OctopusRestClient) GetTestSuite(name string, options metav1.GetOptions) (result *oct.ClusterTestSuite, err error) {
	result = &oct.ClusterTestSuite{}
	err = t.restClient.Get().
		Namespace(NamespaceForTests).
		Resource("clustertestsuites").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

func setConfigDefaults(config *rest.Config) error {
	gv := oct.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}
