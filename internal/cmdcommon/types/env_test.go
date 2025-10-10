package types_test

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/stretchr/testify/require"
)

func TestEnvMap_Set(t *testing.T) {
	t.Run("should set env map value", func(t *testing.T) {
		envMap := &types.EnvMap{}
		err := envMap.Set("MY_ENV=my_value")
		require.NoError(t, err)
		err = envMap.Set("name=ANOTHER_ENV,value=another_value")
		require.NoError(t, err)

		require.Equal(t, map[string]interface{}{
			"MY_ENV":      "my_value",
			"ANOTHER_ENV": "another_value",
		}, envMap.Values)
	})

	t.Run("should return error on invalid format", func(t *testing.T) {
		envMap := &types.EnvMap{Map: &types.Map{Values: map[string]interface{}{}}}
		err := envMap.Set("invalid_format")
		require.ErrorIs(t, err, types.ErrInvalidEnvFormat)
	})

	t.Run("should return error on unknown field", func(t *testing.T) {
		envMap := &types.EnvMap{Map: &types.Map{Values: map[string]interface{}{}}}
		err := envMap.Set("name=MY_ENV,unknown_field=value")
		require.ErrorIs(t, err, types.ErrUnknownEnvField)
	})

	t.Run("should return error on empty name", func(t *testing.T) {
		envMap := &types.EnvMap{Map: &types.Map{Values: map[string]interface{}{}}}
		err := envMap.Set("value=my_value,name")
		require.ErrorIs(t, err, types.ErrInvalidEnvFormat)
	})

	t.Run("should ignore empty value", func(t *testing.T) {
		envMap := &types.EnvMap{Map: &types.Map{Values: map[string]interface{}{}}}
		err := envMap.Set("")
		require.NoError(t, err)
		require.Empty(t, envMap.Values)
	})

	t.Run("should ignore nil value", func(t *testing.T) {
		envMap := &types.EnvMap{Map: &types.Map{Values: map[string]interface{}{}}}
		err := envMap.SetValue(nil)
		require.NoError(t, err)
		require.Empty(t, envMap.Values)
	})
}
