package types

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestEnvMap(t *testing.T) {
	t.Run("set values", func(t *testing.T) {
		envMap := Map{}
		require.NoError(t, envMap.Set("TEST1=1"))
		require.NoError(t, envMap.Set("TEST2=2"))

		expectedMap := map[string]string{
			"TEST1": "1",
			"TEST2": "2",
		}
		expectedNullableMap := map[string]*string{
			"TEST1": ptr.To("1"),
			"TEST2": ptr.To("2"),
		}
		require.Equal(t, expectedMap, envMap.Values)
		require.Equal(t, expectedNullableMap, envMap.GetNullableMap())
		require.Equal(t, "TEST1=1,TEST2=2", envMap.String())
	})

	t.Run("get type", func(t *testing.T) {
		envMap := Map{}
		require.Equal(t, "stringArray", envMap.Type())
	})

	t.Run("set values validation error", func(t *testing.T) {
		envMap := Map{}

		err := envMap.Set("not valid value")
		require.ErrorContains(t, err, "failed to parse value 'not valid value', should be in format KEY=VALUE")
	})
}
