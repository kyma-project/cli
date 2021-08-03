package k3d

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractPortsFromFlag(t *testing.T) {
	rawPorts := []string{"8000:80@loadbalancer", "8443:443@loadbalancer"}
	res, err := extractPortsFromFlag(rawPorts)
	require.NoError(t, err)
	require.Equal(t, []int{8000, 8443}, res)
}
