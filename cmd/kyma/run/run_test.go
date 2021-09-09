package run

import (
	"io/ioutil"
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestSubcommands(t *testing.T) {
	t.Parallel()
	c := NewCmd(&cli.Options{
		KubeconfigPath: "/fakepath",
	})
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.NoError(t, c.Execute(), "Command execution must not fail")

	sub := c.Commands()

	require.Equal(t, 3, len(sub), "Number of created subcommands not as expected")
}
