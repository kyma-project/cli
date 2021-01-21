package function

import (
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestFunctionFlags ensures that the provided command flags are stored in the options.
func TestFunctionFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Empty(t, o.Namespace, "Default value for the --namespace flag not as expected.")
	require.Equal(t, "", o.Filename, "Default value for the --dir flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"--dir", "/fakepath",
		"--namespace", "test-namespace",
	})

	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "/fakepath", o.Filename, "The parsed value for the --dir flag not as expected.")
	require.Equal(t, "test-namespace", o.Namespace, "The parsed value for the --namespace flag not as expected.")
}
