package function

import (
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// TestFunctionFlags ensures that the provided command flags are stored in the options.
func TestFunctionFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, false, o.DryRun, "Default value for the --dry-run flag not as expected.")
	require.Equal(t, "", o.Filename, "Default value for the --filename flag not as expected.")
	require.Equal(t, "nothing", o.OnError.String(), "The parsed value for the --onerror flag not as expected.")
	require.Equal(t, "text", o.Output.String(), "The parsed value for the --output flag not as expected.")
	require.Equal(t, time.Duration(0), o.Timeout, "Default value for the --timeout flag not as expected.")
	require.Equal(t, false, o.Watch, "Default value for the --watch flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"--filename", "/fakepath/config.yaml",
		"--dry-run", "true",
		"--onerror", "purge",
		"--output", "json",
		"--timeout", "15s",
		"--watch", "true",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/fakepath/config.yaml", o.Filename, "The parsed value for the --filename flag not as expected.")
	require.Equal(t, true, o.DryRun, "The parsed value for the --dry-run flag not as expected.")
	require.Equal(t, "purge", o.OnError.String(), "The parsed value for the --onerror flag not as expected.")
	require.Equal(t, "json", o.Output.String(), "The parsed value for the --output flag not as expected.")
	require.Equal(t, time.Duration(15)*time.Second, o.Timeout, "Default value for the --timeout flag not as expected.")
	require.Equal(t, true, o.Watch, "Default value for the --watch flag not as expected.")

	err = c.ParseFlags([]string{
		"-f", "/config.yaml",
		"-o", "yaml",
		"-t", "5s",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/config.yaml", o.Filename, "The parsed value for the -f flag not as expected.")
	require.Equal(t, "yaml", o.Output.String(), "The parsed value for the -o flag not as expected.")
	require.Equal(t, time.Duration(5)*time.Second, o.Timeout, "Default value for the --timeout flag not as expected.")
}
