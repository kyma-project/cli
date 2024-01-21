package net

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAvailablePort(t *testing.T) {
	t.Parallel()
	p, err := GetAvailablePort()

	require.True(t, p > 0)
	require.True(t, p <= 65535)
	require.NoError(t, err)
}

func TestDoGet(t *testing.T) {
	t.Parallel()
	// Happy path
	sc, err := DoGet("http://google.com")
	require.NoError(t, err)
	require.Equal(t, 301, sc)

	// Non existing URL
	_, err = DoGet("http://fake-url-which-will-never-exist.com")
	require.Error(t, err)

	// BAD URL
	_, err = DoGet("this-is%not_a=URL")
	require.Error(t, err)
}
