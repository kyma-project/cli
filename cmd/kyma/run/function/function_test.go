package function

import (
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

// TestFunctionFlags ensures that the provided command flags are stored in the options.
func TestFunctionFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{})
	c := NewCmd(o)

	// test default flag values
	require.Equal(t, "", o.Filename, "Default value for the --filename flag not as expected.")
	require.Equal(t, "", o.ImageName, "Default value for the --imageName flag not as expected.") //"Default value for the --namespace flag not as expected.")
	require.Equal(t, "", o.ContainerName, "Default value for the --containerName flag not as expected.")
	require.Equal(t, "8080", o.FuncPort, "Default value for the --port flag not as expected.")
	require.Equal(t, []string{}, o.Envs, "Default value for the --env flag not as expected.")
	require.Equal(t, time.Duration(0), o.Timeout, "Default value for the --timeout flag not as expected.")
	require.Equal(t, false, o.Detach, "Default value for the --detach flag not as expected.")

	// test passing flags
	err := c.ParseFlags([]string{
		"--filename", "test-name",
		"--imageName", "test/ima:ge",
		"--containerName", "test-container",
		"--port", "9090",
		"--env", "TEST1=ENV1", "--env", "TEST2=ENV2",
		"--timeout", "5s",
		"--detach", "true",
	})

	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "test-name", o.Filename, "The parsed value for the --filename flag not as expected.")
	require.Equal(t, "test/ima:ge", o.ImageName, "The parsed value for the --imageName flag not as expected.") //"Default value for the --namespace flag not as expected.")
	require.Equal(t, "test-container", o.ContainerName, "The parsed value for the --containerName flag not as expected.")
	require.Equal(t, "9090", o.FuncPort, "The parsed value for the --port flag not as expected.")
	require.Equal(t, []string{"TEST1=ENV1", "TEST2=ENV2"}, o.Envs, "The parsed value for the --env flag not as expected.")
	require.Equal(t, time.Second*5, o.Timeout, "The parsed value for the --timeout flag not as expected.")
	require.Equal(t, true, o.Detach, "The parsed value for the --detach flag not as expected.")

	err = c.ParseFlags([]string{
		"-f", "test-name",
		"-p", "9091",
		"-e", "TEST3=ENV3", "-e", "TEST4=ENV4",
		"-t", "10s",
		"-d", "false",
	})

	require.NoError(t, err, "Parsing flags should not return an error")
	require.Equal(t, "test-name", o.Filename, "The parsed value for the --filename flag not as expected.")
	require.Equal(t, "test/ima:ge", o.ImageName, "The parsed value for the --imageName flag not as expected.") //"Default value for the --namespace flag not as expected.")
	require.Equal(t, "test-container", o.ContainerName, "The parsed value for the --containerName flag not as expected.")
	require.Equal(t, "9091", o.FuncPort, "The parsed value for the --port flag not as expected.")
	require.Equal(t, []string{"TEST1=ENV1", "TEST2=ENV2", "TEST3=ENV3", "TEST4=ENV4"}, o.Envs, "The parsed value for the --env flag not as expected.")
	require.Equal(t, time.Second*10, o.Timeout, "The parsed value for the --timeout flag not as expected.")
	require.Equal(t, true, o.Detach, "The parsed value for the --detach flag not as expected.")
}
