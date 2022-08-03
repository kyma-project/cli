package scaffold

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/mandelsoft/vfs/pkg/cwdfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

const directoryMode os.FileMode = 0755

type resourcesToCreate struct{}

func (r resourcesToCreate) start(rb resourceBuilder) error {
	//paths may be absolute, they are resolved against the current working directory anyway
	return rb.createDirectory("/charts").
		createDirectory("/crds").
		createDirectory("/operator").
		createDirectory("/profiles").
		createDirectory("/channels").
		createFileFromTemplate("config.yaml").
		createFileFromTemplate("default.yaml").
		createFileFromTemplate("README.md").
		result()
}

// Init initializes an empty module directory structure with the given name inside given parentDir
func InitEmpty(fs vfs.FileSystem, name, parentDir string) error {

	fileInfo, err := fs.Stat(parentDir)

	if err != nil {
		return fmt.Errorf("unable to read module's parent directory %q: %w", parentDir, err)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("unable to create module: %q is not a directory", parentDir)
	}

	moduleDir := path.Join(parentDir, name)
	_, err = os.Stat(moduleDir)
	if err == nil {
		return fmt.Errorf("unable to create module: file or directory %q already exists", moduleDir)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to create module: %w", err)
	}

	fswd, err := cwdfs.New(fs, parentDir)
	if err != nil {
		return fmt.Errorf("unable to create module: error accessing parent directory %q: %w", parentDir, err)
	}

	return createEmptyModule(fswd, name)

}

func createEmptyModule(fs vfs.FileSystemWithWorkingDirectory, name string) error {

	err := fs.Mkdir(name, directoryMode)
	if err != nil {
		return fmt.Errorf("unable to create module directory: %w", err)
	}
	err = fs.Chdir(name)
	if err != nil {
		return fmt.Errorf("unable to chdir into module directory: %w", err)
	}

	//Inside the  module directory
	return createResources(fs, builderOptions{name})
}

func createResources(fs vfs.FileSystemWithWorkingDirectory, opts builderOptions) error {

	rb := resourceBuilder{
		targetFs: fs,
		opts:     opts,
	}

	return resourcesToCreate{}.start(rb)
}
