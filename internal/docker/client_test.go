package docker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {

	t.Run("creates docker client successfully", func(t *testing.T) {
		cli, err := NewClient()

		require.NoError(t, err)
		require.NotNil(t, cli.APIClient)
	})
}
