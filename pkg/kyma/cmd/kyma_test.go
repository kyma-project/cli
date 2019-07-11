package cmd

import (
	"io/ioutil"
	"testing"

	"github.com/kyma-project/cli/pkg/kyma/core"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
)

// TestKymaFlags ensures that the provided command flags are stored in the options.
func TestKymaFlags(t *testing.T) {
	o := &core.Options{}
	c := NewKymaCmd(o)
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.NoError(t, c.Execute(), "Command execution should not fail")
	require.Equal(t, clientcmd.RecommendedHomeFile, o.KubeconfigPath, "kubeconfig path should have the default flag value")
	require.False(t, o.Verbose, "Verbose flag should be false")
	require.False(t, o.NonInteractive, "non-interactive flag should be false")

	// test passing flags
	c.SetArgs([]string{"--kubeconfig=/some/file", "--non-interactive=true", "--verbose=true"})

	require.NoError(t, c.Execute(), "Command execution should not fail")
	require.Equal(t, "/some/file", o.KubeconfigPath, "kubeconfig path should be the same as the flag provided one")
	require.True(t, o.Verbose, "Verbose flag should be true")
	require.True(t, o.NonInteractive, "non-interactive flag should be true")
}

func TestKymaSubcommands(t *testing.T) {
	c := NewKymaCmd(&core.Options{})
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.NoError(t, c.Execute(), "Command execution should not fail")

	sub := c.Commands()

	require.Equal(t, 7, len(sub), "Number of Kyma subcommands not as expected")
}
