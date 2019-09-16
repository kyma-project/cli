package kyma

import (
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestKymaFlags ensures that the provided command flags are stored in the options.
func TestKymaFlags(t *testing.T) {
	o := &cli.Options{}

	// test default flag values
	// KUBECONFIG is set
	os.Setenv("KUBECONFIG", "/my/kube/path")
	c := NewCmd(o)
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	require.NoError(t, c.Execute(), "Command execution must not fail")
	require.Equal(t, "/my/kube/path", o.KubeconfigPath, "kubeconfig path must have the value of the KUBECONFIG environment variable")
	require.False(t, o.Verbose, "Verbose flag must be false")
	require.False(t, o.NonInteractive, "Non-interactive flag must be false")

	// KUBECONFIG is not set
	os.Setenv("KUBECONFIG", "")
	c = NewCmd(o)
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	require.NoError(t, c.Execute(), "Command execution must not fail")
	require.Equal(t, clientcmd.RecommendedHomeFile, o.KubeconfigPath, "kubeconfig path must have the value of the k8s recommended path")
	require.False(t, o.Verbose, "Verbose flag must be false")
	require.False(t, o.NonInteractive, "Non-interactive flag must be false")

	// test passing flags
	c.SetArgs([]string{"--kubeconfig=/some/file", "--non-interactive=true", "--verbose=true"})

	require.NoError(t, c.Execute(), "Command execution must not fail")
	require.Equal(t, "/some/file", o.KubeconfigPath, "kubeconfig path must be the same as the flag provided")
	require.True(t, o.Verbose, "Verbose flag must be true")
	require.True(t, o.NonInteractive, "Non-interactive flag must be true")
}

func TestKymaSubcommands(t *testing.T) {
	c := NewCmd(&cli.Options{})
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.NoError(t, c.Execute(), "Command execution must not fail")

	sub := c.Commands()

	require.Equal(t, 8, len(sub), "Number of Kyma subcommands not as expected")
}
