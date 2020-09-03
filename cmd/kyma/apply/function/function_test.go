package function

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestUpgradeFlags ensures that the provided command flags are stored in the options.
func TestUpgradeFlags(t *testing.T) {
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, false, o.DryRun, "Default value for the --dry-run flag not as expected.")
	require.Equal(t, "", o.Filename, "Default value for the --filename flag not as expected.")
	require.Equal(t, "nothing", o.OnError, "The parsed value for the --onerror flag not as expected.")
	require.Equal(t, "text", o.Output, "The parsed value for the --output flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"--filename", "/fakepath/config.yaml",
		"--dry-run", "true",
		"--onerror", "purge",
		"--output", "json",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/fakepath/config.yaml", o.Filename, "The parsed value for the --filename flag not as expected.")
	require.Equal(t, true, o.DryRun, "The parsed value for the --dry-run flag not as expected.")
	require.Equal(t, "purge", o.OnError, "The parsed value for the --onerror flag not as expected.")
	require.Equal(t, "json", o.Output, "The parsed value for the --output flag not as expected.")

	err = c.ParseFlags([]string{
		"-f", "/config.yaml",
		"-o", "yaml",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/config.yaml", o.Filename, "The parsed value for the -f flag not as expected.")
	require.Equal(t, "yaml", o.Output, "The parsed value for the -o flag not as expected.")
}
