package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIfNil(t *testing.T) {
	t.Run("two arguments, last is nil", func(t *testing.T) {
		result, err := ifNil("default", nil)
		assert.NoError(t, err)
		assert.Equal(t, "default", result)
	})
	t.Run("two arguments, last is not nil", func(t *testing.T) {
		result, err := ifNil("default", "value")
		assert.NoError(t, err)
		assert.Equal(t, "default", result)
	})
	t.Run("three arguments, last is bool true", func(t *testing.T) {
		result, err := ifNil("notNil", "nil", true)
		assert.NoError(t, err)
		assert.Equal(t, "notNil", result)
	})
	t.Run("three arguments, last is bool false", func(t *testing.T) {
		result, err := ifNil("notNil", "nil", false)
		assert.NoError(t, err)
		assert.Equal(t, "notNil", result)
	})
	t.Run("three arguments, last is nil", func(t *testing.T) {
		result, err := ifNil("notNil", "nil", nil)
		assert.NoError(t, err)
		assert.Equal(t, "nil", result)
	})
	t.Run("four arguments, last is bool true", func(t *testing.T) {
		result, err := ifNil("true", "false", "nil", true)
		assert.NoError(t, err)
		assert.Equal(t, "false", result)
	})
	t.Run("four arguments, last is bool false", func(t *testing.T) {
		result, err := ifNil("true", "false", "nil", false)
		assert.NoError(t, err)
		assert.Equal(t, "false", result)
	})
	t.Run("four arguments, last is nil", func(t *testing.T) {
		result, err := ifNil("true", "false", "nil", nil)
		assert.NoError(t, err)
		assert.Equal(t, "nil", result)
	})
	t.Run("invalid number of arguments", func(t *testing.T) {
		_, err := ifNil("onlyOneArg")
		assert.Error(t, err)
	})
	t.Run("invalid number of arguments for bool", func(t *testing.T) {
		_, err := ifNil("arg1", "arg2", "arg3", "arg4", true)
		assert.Error(t, err)
		assert.Equal(t, "ifNil requires at least three arguments for type bool", err.Error())
	})
	t.Run("invalid number of arguments for string", func(t *testing.T) {
		_, err := ifNil("arg1", "arg2", "arg3")
		assert.Error(t, err)
		assert.Equal(t, "ifNil requires exactly two arguments for type string", err.Error())
	})
	t.Run("invalid number of arguments for int", func(t *testing.T) {
		_, err := ifNil(1, 2, 3)
		assert.Error(t, err)
		assert.Equal(t, "ifNil requires exactly two arguments for type int", err.Error())
	})
	t.Run("invalid number of arguments for map", func(t *testing.T) {
		_, err := ifNil(map[string]interface{}{"key": "value"}, map[string]interface{}{"key": "value"}, map[string]interface{}{"key": "value"})
		assert.Error(t, err)
		assert.Equal(t, "ifNil requires exactly two arguments for type map[string]interface {}", err.Error())
	})
	t.Run("invalid number of arguments (many) for string", func(t *testing.T) {
		_, err := ifNil("arg1", "arg2", "arg3", "arg4", "arg5")
		assert.Error(t, err)
		assert.Equal(t, "ifNil requires exactly two arguments for type string", err.Error())
	})
}
