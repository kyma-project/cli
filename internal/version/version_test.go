package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	t.Run("Create NoVersion", func(t *testing.T) {
		version := NewNoVersion()
		assert.True(t, version.HasNoVersion())
	})

	t.Run("Create from Kyma 2 version", func(t *testing.T) {
		version, err := NewKymaVersion("2.01")
		assert.NoError(t, err)
		assert.True(t, version.IsKyma2())
	})

	t.Run("Create from Kyma 1 version", func(t *testing.T) {
		version, err := NewKymaVersion("1.024")
		assert.NoError(t, err)
		assert.True(t, version.IsKyma1())
	})

	t.Run("Create from PR version", func(t *testing.T) {
		version, err := NewKymaVersion("PR-123")
		assert.NoError(t, err)
		assert.False(t, version.IsReleasedVersion())
	})

	t.Run("Returns same version as string", func(t *testing.T) {
		version := "123"
		kymaVersion, err := NewKymaVersion(version)
		assert.NoError(t, err)
		assert.Equal(t, version, kymaVersion.String())
	})
}