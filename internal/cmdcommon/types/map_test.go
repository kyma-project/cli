package types

import (
	"strings"
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

		// expect TEST1=1,TEST2=2 or TEST2=2,TEST1=1
		stringElems := strings.Split(envMap.String(), ",")
		require.Len(t, stringElems, 2)
		require.Contains(t, stringElems, "TEST1=1")
		require.Contains(t, stringElems, "TEST2=2")
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
