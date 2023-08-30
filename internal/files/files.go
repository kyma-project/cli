// package files provides all functionality to manage kyma CLI local files.
package files

import (
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"
)

const kymaFolder = ".kyma"

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

// FindDirectoryContaining returns the path to the directory containing the targetFolderName
// and a bool indicating if the directory was found or not.
// If the directory was not found, the error will be os.ErrNotExist.
// The search starts at the path and continues to the root of the filesystem.
// The targetFolderName can be a relative path.
// Example:
//
//	FindDirectoryContaining("` + path + `", "` + targetFolderName + `")
func FindDirectoryContaining(path, targetFolderName string) (string, error) {
	if path == string(filepath.Separator) {
		return "", os.ErrNotExist
	}

	targetPath := filepath.Join(path, targetFolderName)
	_, err := os.Stat(targetPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return FindDirectoryContaining(filepath.Dir(path), targetFolderName)
		}
		return "", err
	}

	return targetPath, nil
}
