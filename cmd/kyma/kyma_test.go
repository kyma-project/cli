package kyma

import (
	"io/ioutil"
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestKymaFlags ensures that the provided command flags are stored in the options.
func TestKymaFlags(t *testing.T) {
	o := &cli.Options{}
	c := NewCmd(o)
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.Equal(t, "", o.KubeconfigPath, "kubeconfig path must be empty when default")
	require.False(t, o.Verbose, "Verbose flag must be false")
	require.False(t, o.NonInteractive, "Non-interactive flag must be false")

	// test passing flags
	err := c.ParseFlags([]string{"--kubeconfig=/some/file", "--non-interactive=true", "--verbose=true"})
	require.NoError(t, err)
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

	require.Equal(t, 7, len(sub), "Number of Kyma subcommands not as expected")
}
