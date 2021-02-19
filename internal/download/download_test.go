package download

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testDir string = "tmp"

func TestMain(m *testing.M) {
	//create and erase tmp-folder
	if err := os.MkdirAll(testDir, 0700); err != nil {
		panic(err)
	}
	exitVal := m.Run()
	if err := os.RemoveAll(testDir); err != nil {
		panic(err)
	}
	os.Exit(exitVal)
}

func Test_GetFile(t *testing.T) {
	t.Parallel()

	t.Run("Retrieve local file", func(t *testing.T) {
		_, currFile, _, _ := runtime.Caller(1)
		file, err := GetFile(currFile, testDir)
		assert.NoError(t, err)
		assert.Equal(t, currFile, file, "Local files should not be copied to the dst-folder")
	})
	t.Run("Retrieve remote file", func(t *testing.T) {
		file, err := GetFile("https://raw.githubusercontent.com/kyma-project/cli/master/LICENCE", testDir)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(testDir, "LICENCE"), file, "Remote files should be copied to the dst-folder")
	})
}

func Test_GetFiles(t *testing.T) {
	_, currFile, _, _ := runtime.Caller(1)
	files, err := GetFiles([]string{currFile, "https://raw.githubusercontent.com/kyma-project/kyma/master/LICENSE"}, testDir)
	assert.NoError(t, err)
	assert.Equal(t, []string{currFile, filepath.Join(testDir, "LICENSE")}, files, "Retrieved files differ in names")

}
