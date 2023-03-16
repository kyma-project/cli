package deploy

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOptsValidation(t *testing.T) {
	t.Run("Negative timeout should be rejected", func(t *testing.T) {
		opts := Options{Timeout: -10 * time.Minute}
		err := opts.validateTimeout()
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout must be a positive duration")
	})

	t.Run("Invalid targets should be rejected", func(t *testing.T) {
		opts := Options{Target: "some-target"}
		err := opts.validateTarget()
		require.Error(t, err)
		require.Contains(t, err.Error(), "target must be either 'control-plane' or 'remote'")
	})
}
