package function

import (
	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestFunctionFlags ensures that the provided command flags are stored in the options.
func TestFunctionFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, "", o.Filename, "Default value for the --filename flag not as expected.")
	require.Equal(t, "", o.Dir, "Default value for the --dir flag not as expected.")
	require.Equal(t, "", o.ContainerName, "Default value for the --containerName flag not as expected.")
	require.Equal(t, "8080", o.FuncPort, "Default value for the --port flag not as expected.")
	require.Equal(t, false, o.Detach, "Default value for the --detach flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"--filename", "test-name",
		"--sourceDir", "/test/folder",
		"--containerName", "test-container",
		"--port", "9090",
		"--detach", "true",
	})

	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "test-name", o.Filename, "The parsed value for the --filename flag not as expected.")
	require.Equal(t, "/test/folder", o.Dir, "The parsed value for the --dir flag not as expected.")
	require.Equal(t, "test-container", o.ContainerName, "The parsed value for the --containerName flag not as expected.")
	require.Equal(t, "9090", o.FuncPort, "The parsed value for the --port flag not as expected.")
	require.Equal(t, true, o.Detach, "The parsed value for the --detach flag not as expected.")

	err = c.ParseFlags([]string{
		"-f", "test-name",
		"-p", "9091",
		"-d", "/new/test/dir",
	})

	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "test-name", o.Filename, "The parsed value for the --filename flag not as expected.")
	require.Equal(t, "/new/test/dir", o.Dir, "The parsed value for the --dir flag not as expected.")
	require.Equal(t, "test-container", o.ContainerName, "The parsed value for the --containerName flag not as expected.")
	require.Equal(t, "9091", o.FuncPort, "The parsed value for the --port flag not as expected.")
}
