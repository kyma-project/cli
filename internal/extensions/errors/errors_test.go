package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ErrorList(t *testing.T) {
	t.Run("return empty list", func(t *testing.T) {
		el := NewList()
		require.Nil(t, el)
	})

	t.Run("print multiple errors", func(t *testing.T) {
		el := NewList(New("error1"), nil, New("error2"), New("error3"))
		require.Equal(t, "error1\nerror2\nerror3", el.Error())
	})
}

func Test_Wrapf(t *testing.T) {
	t.Run("wrap error", func(t *testing.T) {
		err := Wrapf(New("inner error"), "error")
		require.Equal(t, "error: inner error", err.Error())
	})

	t.Run("skip inner error", func(t *testing.T) {
		err := Wrapf(nil, "%s", "error")
		require.Equal(t, "error", err.Error())
	})

	t.Run("wrap ErrorList", func(t *testing.T) {
		el := NewList(New("error1"), New("error2"), New("error3"))
		err := Wrapf(el, "error list")
		require.Equal(t, "error list:\n  error1\n  error2\n  error3", err.Error())
	})
}

func Test_JoinWithSeparator(t *testing.T) {
	t.Run("return error on one arg", func(t *testing.T) {
		err := JoinWithSeparator("/", New("error"))
		require.Equal(t, "error", err.Error())
	})
	t.Run("join", func(t *testing.T) {
		err := JoinWithSeparator("/", New("error1"), New("error2"), New("error3"))
		require.Equal(t, "error1/error2/error3", err.Error())
	})

	t.Run("return nil on nil input", func(t *testing.T) {
		err := JoinWithSeparator("/", nil, nil, nil)
		require.Nil(t, err)
	})
}
