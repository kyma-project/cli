package dashboard

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptsValidation(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		opts := &Options{
			ContainerName: "valid-container-name",
		}
		err := opts.validateFlags()
		require.NoError(t, err)
	})
	t.Run("Invalid container name", func(t *testing.T) {
		opts := &Options{
			ContainerName: "",
		}
		err := opts.validateFlags()
		require.Error(t, err)
	})
}
