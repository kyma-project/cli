package module

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInvalidPolicy(t *testing.T) {
	t.Run("Invalid policy should be rejected", func(t *testing.T) {
		opts := Options{Policy: "Unknown"}
		err := opts.validatePolicy()
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("policy must be either %s or %s", customResourcePolicyCreateAndDelete, customResourcePolicyIgnore))
	})
}
