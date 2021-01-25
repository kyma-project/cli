package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCompare(t *testing.T) {
	t.Parallel()

	t.Run("Happy path - two equal versions", func(t *testing.T) {
		err := checkCompatibility("1.17.1", "1.17.1")
		assert.NoError(t, err)
	})

	t.Run("Happy path", func(t *testing.T) {
		err := checkCompatibility("1.17.1", "1.18.1-rc2")
		assert.NoError(t, err)
	})

	t.Run("Current version is not a release", func(t *testing.T) {
		err := checkCompatibility("aa6asdf32", "1.18.1")
		assert.Error(t, err)
		assert.Equal(t, err.Error(), currentVersionNoReleaseError.with("aa6asdf32", "1.18.1").Error())
	})

	t.Run("Next version is not a release", func(t *testing.T) {
		err := checkCompatibility("1.18.0", "master")
		assert.Error(t, err)
		assert.Equal(t, err.Error(), nextVersionNoReleaseError.with("1.18", "master").Error())
	})

	t.Run("Next version is lower as current version", func(t *testing.T) {
		err := checkCompatibility("1.18.6", "1.17.5-rc2")
		assert.Error(t, err)
		assert.Equal(t, err.Error(), nextVersionLowerError.with("1.18.6", "1.17.5-rc2").Error())
	})

	t.Run("Next version is too far away from current version", func(t *testing.T) {
		err := checkCompatibility("1.15.6", "1.18.5")
		assert.Error(t, err)
		assert.Equal(t, err.Error(), nextVersionTooGreatError.with("1.15.6", "1.18.5").Error())
	})

}
