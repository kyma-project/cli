package k3d

import (
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestExtractPortsFromFlag(t *testing.T) {
	rawPorts := []string{"8000:80@loadbalancer", "8443:443@loadbalancer"}
	res, err := extractPortsFromFlag(rawPorts)
	require.NoError(t, err)
	require.Equal(t, []int{8000, 8443}, res)
}

func TestPromptUserToDeleteExistingCluster(t *testing.T) {
	options := cli.NewOptions()
	options.NonInteractive = true
	c := command{
		Command: cli.Command{Options: options},
		opts:    NewOptions(options),
	}

	answer := c.PromptUserToDeleteExistingCluster()

	require.True(t, answer)
}
