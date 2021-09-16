package deploy

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOptsValidation(t *testing.T) {
	t.Run("unknown profile", func(t *testing.T) {
		opts := Options{Profile: "fancy"}
		err := opts.validateFlags()

		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown profile: fancy")
	})

	t.Run("supported profiles", func(t *testing.T) {
		profiles := []string{"", "evaluation", "production"}
		for _, p := range profiles {
			opts := Options{Profile: p}
			err := opts.validateFlags()
			require.NoError(t, err)
		}
	})
}
