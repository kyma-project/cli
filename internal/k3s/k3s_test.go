package k3s

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	mockDir = "src/github.com/kyma-project/cli/internal/k3s/mock"
)

func TestMain(m *testing.M) {
	if !setup() {
		fmt.Println("Setup of k3s test failed: test case for k3s can' be executed")
		return
	}
	code := m.Run()
	//shutdown()
	os.Exit(code)
}

// Place this folder at the beginning of PATH env-var to ensure this
// mock-script will be used instead of a locally installed k3d tool.
func setup() bool {
	if os.Getenv("GOPATH") == "" {
		fmt.Println("Could not inject k3s mock directory into PATH: env-var GOPATH is undefined")
		return false
	}
	os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(os.Getenv("GOPATH"), mockDir), os.Getenv("PATH")))
	return true
}

// function to verify output of k3d tool
type testFunc func(output string, err error)

func TestRunCmd(t *testing.T) {
	tests := []struct {
		cmd      []string
		verifyer testFunc
	}{
		{
			cmd: []string{"cluster", "list"},
			verifyer: testFunc(func(output string, err error) {
				if !strings.Contains(output, "kyma-cluster") {
					require.Fail(t, fmt.Sprintf("Expected string 'kyma-cluster' is missing in k3d output: %s", output))
				}
			}),
		},
		{
			cmd: []string{"cluster", "xyz"},
			verifyer: testFunc(func(output string, err error) {
				require.NotEmpty(t, err, "Error object expected")
			}),
		},
	}

	for testID, testCase := range tests {
		output, err := RunCmd(false, 5*time.Second, testCase.cmd...)
		require.NotNilf(t, testCase.verifyer, "Verifyer function missing for test #'%d'", testID)
		testCase.verifyer(output, err)
	}

}

func TestCheckVersion(t *testing.T) {
	err := CheckVersion(false)
	require.NoError(t, err)
}

func TestInitialize(t *testing.T) {
	err := Initialize(false)
	require.NoError(t, err)
}

func TestInitializeFailed(t *testing.T) {
	pathPrev := os.Getenv("PATH")
	os.Setenv("PATH", "/usr/bin")

	err := Initialize(false)
	require.Error(t, err)

	os.Setenv("PATH", pathPrev)
}

func TestStartCluster(t *testing.T) {
	err := StartCluster(false, 5*time.Second, "kyma", 1)
	require.NoError(t, err)
}

func TestDeleteCluster(t *testing.T) {
	err := DeleteCluster(false, 5*time.Second, "kyma")
	require.NoError(t, err)
}

func TestClusterExists(t *testing.T) {
	os.Setenv("K3D_MOCK_DUMPFILE", "cluster_list_exists.json")
	exists, err := ClusterExists(false, "kyma")
	require.NoError(t, err)
	require.True(t, exists)
}
