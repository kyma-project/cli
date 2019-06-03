package net

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetAvailablePort(t *testing.T) {
	p, err := GetAvailablePort()

	require.True(t, p > 0)
	require.True(t, p <= 65535)
	require.NoError(t, err)
}
