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

type K3dTestSuite struct {
	suite.Suite
	client         Client
	mockCmdRunner  *mocks.CmdRunner
	mockPathLooker *mocks.PathLooker
}

func (suite *K3dTestSuite) SetupTest() {
	suite.mockCmdRunner = &mocks.CmdRunner{}
	suite.mockPathLooker = &mocks.PathLooker{}
	suite.client = NewClient(suite.mockCmdRunner, suite.mockPathLooker, testClusterName, true, testTimeout)
}

func (suite *K3dTestSuite) TestVerifyStatus() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v5.0.0\nk3s version v1.21.5-k3s2 (default)", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "list").Return("", nil)

	err := suite.client.VerifyStatus()
	suite.Nil(err)
}

func (suite *K3dTestSuite) TestCheckVersionIncompMinor() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v4.4.8\nk3s version v1.21.3-k3s1 (default)", nil)

	err := suite.client.VerifyStatus()
	suite.Error(err)
}

func (suite *K3dTestSuite) TestCheckVersionIncompMajor() {
	suite.mockPathLooker.On("Look", "k3d").Return("", nil)

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "version").Return("k3d version v6.1.0\nk3s version latest (default)", nil)

	err := suite.client.VerifyStatus()
	suite.Error(err)
}

func (suite *K3dTestSuite) TestClusterExistsTrue() {
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

func (suite *K3dTestSuite) TestClusterExistsFalse() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "list", "-o", "json").
		Return("[]", nil)

	exists, err := suite.client.ClusterExists()
	suite.False(exists)
	suite.Nil(err)
}

func (suite *K3dTestSuite) TestRegistryExistsTrue() {
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

func (suite *K3dTestSuite) TestRegistryExistsFalse() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "list", "-o", "json").
		Return("[]", nil)

	exists, err := suite.client.RegistryExists()
	suite.False(exists)
	suite.Nil(err)
}

func (suite *K3dTestSuite) TestCreateCluster() {
	settings := CreateClusterSettings{
		Args:              []string{"--verbose"},
		KubernetesVersion: "1.20.11",
		PortMapping:       []string{"80:80@loadbalancer", "443:443@loadbalancer"},
		Workers:           0,
		K3sArgs:           []string{},
		UseRegistry:       []string{"k3d-own-registry:5001"},
	}

	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "create", testClusterName,
		"--kubeconfig-update-default",
		"--timeout", "5s",
		"--agents", "0",
		"--image", "rancher/k3s:v1.20.11-k3s1",
		"--kubeconfig-switch-context",
		"--k3s-arg", "--disable=traefik@server:0",
		"--k3s-arg", "--kubelet-arg=containerd=/run/k3s/containerd/containerd.sock@all:*",
		"--registry-use", "k3d-own-registry:5001",
		"--port", "80:80@loadbalancer",
		"--port", "443:443@loadbalancer",
		"--verbose").Return("", nil)

	err := suite.client.CreateCluster(settings)
	suite.Nil(err)
}

func (suite *K3dTestSuite) TestCreateRegistry() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "create", "kyma-registry",
		"--port", "5001").Return("", nil)

	registryName, err := suite.client.CreateRegistry("5001", []string{})
	suite.Nil(err)
	suite.Equal("kyma-registry:5001", registryName)
}

func (suite *K3dTestSuite) TestDeleteCluster() {
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "cluster", "delete", testClusterName).Return("", nil)

	err := suite.client.DeleteCluster()
	suite.Nil(err)
}

func (suite *K3dTestSuite) TestDeleteRegistry() {
	registryName := fmt.Sprintf(defaultRegistryNamePattern, testClusterName)
	suite.mockCmdRunner.On("Run", mock.Anything, "k3d", "registry", "delete", fmt.Sprintf("k3d-%s", registryName)).Return("", nil)

	err := suite.client.DeleteRegistry()
	suite.Nil(err)
}

func TestK3dSuite(t *testing.T) {
	suite.Run(t, new(K3dTestSuite))
}

func TestArgConstruction(t *testing.T) {
	rawPorts := []string{"8000:80@loadbalancer", "8443:443@loadbalancer"}
	res := constructArgs("-p", rawPorts)
	require.Equal(t, []string{"-p", "8000:80@loadbalancer", "-p", "8443:443@loadbalancer"}, res)
}
