package k3d

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/k3d/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testClusterName = "kyma"
	testTimeout     = 5 * time.Second
)

type V5TestSuite struct {
	suite.Suite
	client         Client
	mockCmdRunner  *mocks.CmdRunner
	mockPathLooker *mocks.PathLooker
}

func (suite *V5TestSuite) SetupTest() {
	suite.mockCmdRunner = &mocks.CmdRunner{}
	suite.mockPathLooker = &mocks.PathLooker{}
	suite.client = NewClient(suite.mockCmdRunner, suite.mockPathLooker, testClusterName, true, testTimeout)
}

func (suite *V5TestSuite) TestVerifyStatus() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v5.0.0\nk3s version v1.21.5-k3s2 (default)", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "list").Return("", nil)

	err := suite.client.VerifyStatus(true)
	suite.Nil(err)
}

func (suite *V5TestSuite) TestCheckVersionIncompMinor() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v4.4.8\nk3s version v1.21.3-k3s1 (default)", nil)

	err := suite.client.VerifyStatus(true)
	suite.Error(err)
}

func (suite *V5TestSuite) TestCheckVersionIncompMajor() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v6.1.0\nk3s version latest (default)", nil)

	err := suite.client.VerifyStatus(true)
	suite.Error(err)
}

func (suite *V5TestSuite) TestClusterExistsTrue() {
	clusterExistsOutput := `  [
    {
      "name": "kyma",
      "nodes": [
        {
          "name": "k3d-kyma-serverlb",
          "role": "loadbalancer",
          "State": {
            "Running": true,
            "Status": "running"
          }
        },
        {
          "name": "k3d-kyma-server-0",
          "role": "server",
          "State": {
            "Running": true,
            "Status": "running"
          }
        }
      ]
    }
  ]`
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "list", "-o", "json").
		Return(clusterExistsOutput, nil)

	exists, err := suite.client.ClusterExists()
	suite.True(exists)
	suite.Nil(err)
}

func (suite *V5TestSuite) TestClusterExistsFalse() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "list", "-o", "json").
		Return("[]", nil)

	exists, err := suite.client.ClusterExists()
	suite.False(exists)
	suite.Nil(err)
}

func (suite *V5TestSuite) TestRegistryExistsTrue() {
	registryExistsOutput := `[
  {
    "name": "k3d-kyma-registry",
    "role": "registry",
    "State": {
      "Running": true,
      "Status": "running",
      "Started": ""
    }
  }
]`
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "list", "-o", "json").
		Return(registryExistsOutput, nil)

	exists, err := suite.client.RegistryExists()
	suite.True(exists)
	suite.Nil(err)
}

func (suite *V5TestSuite) TestRegistryExistsFalse() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "list", "-o", "json").
		Return("[]", nil)

	exists, err := suite.client.RegistryExists()
	suite.False(exists)
	suite.Nil(err)
}

func (suite *V5TestSuite) TestCreateCluster() {
	settings := CreateClusterSettings{
		Args:              []string{"--verbose"},
		KubernetesVersion: "1.20.11",
		PortMapping:       []string{"80:80@loadbalancer", "443:443@loadbalancer"},
		Workers:           0,
		V5Settings: V5CreateClusterSettings{
			K3sArgs:     []string{},
			UseRegistry: []string{"k3d-own-registry:5001"},
		},
	}

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "create", testClusterName,
		"--kubeconfig-update-default",
		"--timeout", "5s",
		"--agents", "0",
		"--image", "rancher/k3s:v1.20.11-k3s1",
		"--kubeconfig-switch-context",
		"--k3s-arg", "--disable=traefik@server:0",
		"--registry-use", "k3d-own-registry:5001",
		"--port", "80:80@loadbalancer",
		"--port", "443:443@loadbalancer",
		"--verbose").Return("", nil)

	err := suite.client.CreateCluster(settings, true)
	suite.Nil(err)
}

func (suite *V5TestSuite) TestCreateRegistry() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "create", "kyma-registry",
		"--port", "5001").Return("", nil)

	registryName, err := suite.client.CreateRegistry()
	suite.Nil(err)
	suite.Equal("kyma-registry:5001", registryName)
}

func (suite *V5TestSuite) TestDeleteCluster() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "delete", testClusterName).Return("", nil)

	err := suite.client.DeleteCluster()
	suite.Nil(err)
}

func (suite *V5TestSuite) TestDeleteRegistry() {
	registryName := fmt.Sprintf(v5DefaultRegistryNamePattern, testClusterName)
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "delete", fmt.Sprintf("k3d-%s", registryName)).Return("", nil)

	err := suite.client.DeleteRegistry()
	suite.Nil(err)
}

func TestV5TestSuite(t *testing.T) {
	suite.Run(t, new(V5TestSuite))
}

type V4TestSuite struct {
	suite.Suite
	client         Client
	mockCmdRunner  *mocks.CmdRunner
	mockPathLooker *mocks.PathLooker
}

func (suite *V4TestSuite) SetupTest() {
	suite.mockCmdRunner = &mocks.CmdRunner{}
	suite.mockPathLooker = &mocks.PathLooker{}
	suite.client = NewClient(suite.mockCmdRunner, suite.mockPathLooker, testClusterName, true, testTimeout)
}

func (suite *V4TestSuite) TestCheckVersionIncompMinor() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v3.4.0\nk3s version v1.21.3-k3s1 (default)", nil)

	err := suite.client.VerifyStatus(false)
	suite.Error(err)
}

func (suite *V4TestSuite) TestCheckVersionIncompMajor() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v5.1.0\nk3s version latest (default)", nil)

	err := suite.client.VerifyStatus(false)
	suite.Error(err)
}

func (suite *V4TestSuite) TestCreateCluster() {
	settings := CreateClusterSettings{
		Args:              []string{"--verbose"},
		KubernetesVersion: "1.20.11",
		PortMapping:       []string{"80:80@loadbalancer", "443:443@loadbalancer"},
		Workers:           1,
		V4Settings: V4CreateClusterSettings{
			ServerArgs: []string{"--alsologtostderr"},
			AgentArgs:  []string{"--no-rollback"},
		},
	}

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "create", testClusterName,
		"--kubeconfig-update-default",
		"--timeout", "5s",
		"--agents", "1",
		"--image", "rancher/k3s:v1.20.11-k3s1",
		"--registry-create",
		"--k3s-server-arg", "--disable",
		"--k3s-server-arg", "traefik",
		"--k3s-server-arg", "--alsologtostderr",
		"--k3s-agent-arg", "--no-rollback",
		"--port", "80:80@loadbalancer",
		"--port", "443:443@loadbalancer",
		"--verbose").Return("", nil)

	err := suite.client.CreateCluster(settings, false)
	suite.Nil(err)
}

func TestV4TestSuite(t *testing.T) {
	suite.Run(t, new(V4TestSuite))
}

func TestArgConstruction(t *testing.T) {
	rawPorts := []string{"8000:80@loadbalancer", "8443:443@loadbalancer"}
	res := constructArgs("-p", rawPorts)
	require.Equal(t, []string{"-p", "8000:80@loadbalancer", "-p", "8443:443@loadbalancer"}, res)

}
