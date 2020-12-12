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

// Place this folder at the beginning of PATH env-var to ensure this
// mock-script will be used instead of a locally installed k3d tool.
func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Dir(ex), os.Getenv("PATH")))
}
func TestRunCmd(t *testing.T) {
	t.Parallel()

	output, err := RunCmd(false, 5*time.Second, "help")
	if err != nil {
		require.Fail(t, fmt.Sprintf("k3d command failed: %s", output))
	}

	if !strings.Contains(output, "--help") {
		require.Fail(t, fmt.Sprintf("Expected string '--help' is missing in k3s output: %s", output))
	}
}

func TestCheckVersion(t *testing.T) {
	output, err := CheckVersion(false, 5*time.Second)
	if err != nil {
		require.Fail(t, fmt.Sprintf("k3d version check failed: %s", err))
	}
	require.Empty(t, output, "Output of of version check has to be empty")
}
