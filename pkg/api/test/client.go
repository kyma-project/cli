package client

import (
	"context"
	"time"

	"github.com/kyma-incubator/octopus/pkg/apis"
	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sRestClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type TestRESTClient interface {
	ListTestDefinitions() (*oct.TestDefinitionList, error)
	ListTestSuites() (*oct.ClusterTestSuiteList, error)
	CreateTestSuite(cts *oct.ClusterTestSuite) error
	DeleteTestSuite(cts *oct.ClusterTestSuite) error
	GetTestSuiteByName(name string) (*oct.ClusterTestSuite, error)
}

type testRestClient struct {
	cli         k8sRestClient.Client
	callTimeout time.Duration
}

func (t *testRestClient) ListTestDefinitions() (*oct.TestDefinitionList, error) {
	result := &oct.TestDefinitionList{}
	ctx, cancelF := context.WithTimeout(context.Background(), t.callTimeout)
	defer cancelF()
	err := t.cli.List(ctx, &client.ListOptions{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *testRestClient) ListTestSuites() (*oct.ClusterTestSuiteList, error) {
	result := &oct.ClusterTestSuiteList{}
	ctx, cancelF := context.WithTimeout(context.Background(), t.callTimeout)
	defer cancelF()
	err := t.cli.List(ctx, &client.ListOptions{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (t *testRestClient) CreateTestSuite(cts *oct.ClusterTestSuite) error {
	ctx, cancelF := context.WithTimeout(context.Background(), t.callTimeout)
	defer cancelF()
	return t.cli.Create(ctx, cts)
}

func (t *testRestClient) DeleteTestSuite(cts *oct.ClusterTestSuite) error {
	ctx, cancelF := context.WithTimeout(context.Background(), t.callTimeout)
	defer cancelF()
	return t.cli.Delete(ctx, cts)
}

func (t *testRestClient) GetTestSuiteByName(name string) (*oct.ClusterTestSuite, error) {
	ctx, cancelF := context.WithTimeout(context.Background(), t.callTimeout)
	defer cancelF()
	result := &oct.ClusterTestSuite{}
	err := t.cli.Get(ctx,
		types.NamespacedName{Name: name},
		result)
	return result, err
}

func NewTestRESTClient(callTimeout time.Duration) (TestRESTClient, error) {
	apis.AddToScheme(scheme.Scheme)
	cli, err := k8sRestClient.New(config.GetConfigOrDie(), k8sRestClient.Options{})
	if err != nil {
		return nil, err
	}

	return &testRestClient{
		cli:         cli,
		callTimeout: callTimeout,
	}, nil
}
