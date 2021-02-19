package download

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetFile(t *testing.T) {
	t.Run("Retrieve local file", func(t *testing.T) {
		_, currFile, _, _ := runtime.Caller(1)
		file, err := GetFile(currFile, ".")
		assert.NoError(t, err)
		assert.Equal(t, currFile, file, "Local files should not be copied to the dst-folder")
	})
}
