package client

import (
	"fmt"

	oct "github.com/kyma-incubator/octopus/pkg/apis/testing/v1alpha1"
)

type MockedTestRestClient struct {
	testDefs   *oct.TestDefinitionList
	testSuites *oct.ClusterTestSuiteList
}

func NewMockedTestRestClient(testDefs *oct.TestDefinitionList, testSuites *oct.ClusterTestSuiteList) *MockedTestRestClient {
	return &MockedTestRestClient{
		testDefs:   testDefs,
		testSuites: testSuites,
	}
}

func (m *MockedTestRestClient) ListTestDefinitions() (*oct.TestDefinitionList, error) {
	return m.testDefs, nil
}

func (m *MockedTestRestClient) ListTestSuites() (*oct.ClusterTestSuiteList, error) {
	return m.testSuites, nil
}

func (m *MockedTestRestClient) CreateTestSuite(cts *oct.ClusterTestSuite) error {
	m.testSuites.Items = append(m.testSuites.Items, *cts)
	return nil
}

func (m *MockedTestRestClient) DeleteTestSuite(cts *oct.ClusterTestSuite) error {
	for i := 0; i < len(m.testSuites.Items); i++ {
		if m.testSuites.Items[i].GetName() == cts.GetName() {
			m.testSuites.Items = append(m.testSuites.Items[i:],
				m.testSuites.Items[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("test not found")
}

func (m *MockedTestRestClient) GetTestSuiteByName(name string) (*oct.ClusterTestSuite, error) {
	for i := 0; i < len(m.testSuites.Items); i++ {
		if m.testSuites.Items[i].GetName() == name {
			return &m.testSuites.Items[i], nil
		}
	}
	return nil, fmt.Errorf("not found")
}
