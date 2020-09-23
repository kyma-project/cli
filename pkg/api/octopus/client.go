package octopus

import (
	"context"
	"time"

	"github.com/kyma-incubator/octopus/pkg/apis"
	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type Interface interface {
	ListTestDefinitions(opts metav1.ListOptions) (result *oct.TestDefinitionList, err error)
	ListTestSuites(opts metav1.ListOptions) (result *oct.ClusterTestSuiteList, err error)
	CreateTestSuite(cts *oct.ClusterTestSuite) (result *oct.ClusterTestSuite, err error)
	DeleteTestSuite(name string, options metav1.DeleteOptions) error
	GetTestSuite(name string, options metav1.GetOptions) (result *oct.ClusterTestSuite, err error)
	WatchTestSuite(opts metav1.ListOptions) (watch.Interface, error)
}

type RestClient struct {
	restClient rest.Interface
}

func NewFromConfig(config *rest.Config) (Interface, error) {
	if err := apis.AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	setConfigDefaults(config)

	cl, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return &RestClient{
		restClient: cl,
	}, nil
}

func (t *RestClient) ListTestDefinitions(opts metav1.ListOptions) (result *oct.TestDefinitionList, err error) {
	result = &oct.TestDefinitionList{}
	err = t.restClient.Get().
		Resource("testdefinitions").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	return
}

func (t *RestClient) ListTestSuites(opts metav1.ListOptions) (result *oct.ClusterTestSuiteList, err error) {
	result = &oct.ClusterTestSuiteList{}
	err = t.restClient.Get().
		Resource("clustertestsuites").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	return
}

func (t *RestClient) CreateTestSuite(cts *oct.ClusterTestSuite) (result *oct.ClusterTestSuite, err error) {
	result = &oct.ClusterTestSuite{}
	err = t.restClient.Post().
		Resource("clustertestsuites").
		Body(cts).
		Do(context.Background()).
		Into(result)
	return
}

func (t *RestClient) DeleteTestSuite(name string, options metav1.DeleteOptions) error {
	return t.restClient.Delete().
		Resource("clustertestsuites").
		Name(name).
		// Reenable this when deleting supports options
		//Body(options).
		Do(context.Background()).
		Error()
}

func (t *RestClient) GetTestSuite(name string, options metav1.GetOptions) (result *oct.ClusterTestSuite, err error) {
	result = &oct.ClusterTestSuite{}
	err = t.restClient.Get().
		Resource("clustertestsuites").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(context.Background()).
		Into(result)
	return
}

func (t *RestClient) WatchTestSuite(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return t.restClient.Get().
		Resource("clustertestsuites").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(context.Background())
}

func setConfigDefaults(config *rest.Config) {
	gv := oct.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
}
