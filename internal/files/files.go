// package files provides all functionality to manage kyma CLI local files.
package files

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mandelsoft/vfs/pkg/vfs"
)

const kymaFolder = ".kyma"

var customSkipSearch = errors.New("found .git directory, skipping search")

func KymaHome() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(u.HomeDir, kymaFolder)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.MkdirAll(p, 0700)
		if err != nil {
			return "", err
		}
	}
	return p, nil
}

// isDir determines if a file represented by `path` is a directory or not
func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), err
}

func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// FileType returns the mimetype of a file.
func FileType(fs vfs.FileSystem, path string) (string, error) {
	file, err := fs.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// Only the first 512 bytes are used to sniff the content type.
	buf := make([]byte, 512)
	_, err = file.Read(buf)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	return http.DetectContentType(buf), nil
}

// SearchForTargetDirByName walks the given root path and searches for a directory with the given name.
// If the directory is found the function returns the path to the directory and nil as error.
// If the directory is not found the function returns an empty string and an error.
func SearchForTargetDirByName(root string, targetFolderName string) (gitFolderPath string, walkErr error) {
	walkErr = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error while walking the path %q: %v", path, err)
		}
		if info.IsDir() && info.Name() == targetFolderName {
			gitFolderPath = path
			return customSkipSearch
		}
		return nil
	})

	if walkErr == customSkipSearch {
		return gitFolderPath, nil
	}
	return
}

// IsFileExists checks if a file exists.
// It returns an error if the file does not exist or if the file path is empty.
//   - If the file exists, it returns nil.
//   - If the file does not exist, it returns an error with the message: "file <filePath> does not exist".
//   - If the file path is empty, it returns an error with the message: "file path is empty".
//
// Note: This function does not check if the file path is valid.
func IsFileExists(filePath string) bool {
	if filePath == "" {
		return false
	}
	isAbsolute := filepath.IsAbs(filePath)
	if !isAbsolute {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return false
		}
		filePath = absPath
	}
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
