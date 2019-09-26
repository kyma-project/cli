package uninstall

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestUninstallFlags ensures that the provided command flags are stored in the options.
func TestUninstallFlags(t *testing.T) {
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	// test default flag values
	require.Equal(t, 30*time.Minute, o.Timeout, "Incorrect default uninstall time-out")

	// test passing flags
	c.ParseFlags([]string{"--timeout=60m0s"})
	require.Equal(t, 60*time.Minute, o.Timeout, "Incorrect specified uninstall time-out")
}

func TestUninstallSubcommands(t *testing.T) {
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)
	c.SetOutput(ioutil.Discard) // not interested in the command's output

	sub := c.Commands()

	require.Equal(t, 0, len(sub), "Number of uninstall subcommands not as expected")
}
