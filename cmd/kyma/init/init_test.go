package init

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestSubcommands(t *testing.T) {
	c := NewCmd(&cli.Options{})
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.NoError(t, c.Execute(), "Command execution must not fail")

	sub := c.Commands()

	require.Equal(t, 2, len(sub), "Number of create subcommands not as expected")
}
