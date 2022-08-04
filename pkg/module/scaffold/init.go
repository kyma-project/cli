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

// createResources is where you define all resources to be created.
func createResources(rb resourceBuilder) error {
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
func InitEmpty(fsys vfs.FileSystem, name, parentDir string) error {

	fileInfo, err := fsys.Stat(parentDir)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			//create directory tree for the user and continue
			err = fsys.MkdirAll(parentDir, directoryMode)
			if err != nil {
				return fmt.Errorf("Error while creating directory %q: %w", parentDir, err)
			}
		} else {
			return fmt.Errorf("unable to access module's parent directory %q: %w", parentDir, err)
		}
	} else {
		if !fileInfo.IsDir() {
			return fmt.Errorf("unable to create module: %q is not a directory", parentDir)
		}
	}

	moduleDir := path.Join(parentDir, name)
	_, err = os.Stat(moduleDir)
	if err == nil {
		return fmt.Errorf("unable to create module: file or directory %q already exists", moduleDir)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("unable to create module: %w", err)
	}

	fswd, err := cwdfs.New(fsys, parentDir)
	if err != nil {
		return fmt.Errorf("unable to create module: error accessing parent directory %q: %w", parentDir, err)
	}

	return createEmptyModule(fswd, name)

}

func createEmptyModule(fsys vfs.FileSystemWithWorkingDirectory, name string) error {

	err := fsys.Mkdir(name, directoryMode)
	if err != nil {
		return fmt.Errorf("unable to create module directory: %w", err)
	}
	err = fsys.Chdir(name)
	if err != nil {
		return fmt.Errorf("unable to chdir into module directory: %w", err)
	}

	rb := resourceBuilder{
		targetFs: fsys,
		opts:     builderOptions{name},
	}
	return createResources(rb)
}
