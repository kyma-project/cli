package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// GetFile downloads a file. Destination directory will be created if it does not exist.
// It returns the path to the downloaded file.
// If the provided file is not a URL, it checks if it exists locally
func GetFile(file, dstDir string) (string, error) {
	localFiles, err := GetFiles([]string{file}, dstDir)
	if err == nil {
		return localFiles[0], nil
	}
	return "", err
}

// GetFiles downloads a list of files. Destination directory will be created if it does not exist.
// It returns the paths to the downloaded files.
// If the provided file is not a URL, it checks if it exists locally
func GetFiles(files []string, dstDir string) ([]string, error) {
	result := []string{}
	for _, file := range files {
		urlTokens := strings.Split(file, "://")
		switch len(urlTokens) {
		case 1:
			// In case the file provided is not a URL, check if it exists locally
			if _, err := os.Stat(file); os.IsNotExist(err) {
				return nil, fmt.Errorf("File '%s' not found", file)
			}
			result = append(result, urlTokens[0])
		case 2:
			if strings.HasPrefix(urlTokens[0], "http") {
				dstFile, err := download(file, dstDir, filepath.Base(urlTokens[1]))
				if err != nil {
					return nil, err
				}
				result = append(result, dstFile)
			} else {
				return nil, fmt.Errorf("Cannot download '%s' because schema '%s' is not supported", file, urlTokens[0])
			}
		default:
			return nil, fmt.Errorf("File path '%s' is not valid", file)
		}
	}
	return result, nil
}

// RemoteReader returns a reader to a remote file.
func RemoteReader(path string) (io.ReadCloser, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	// nolint: gosec
	resp, err := client.Get(path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}

	return nil, fmt.Errorf("Couldn't not read remote file: %s, response: %v", path, resp.Status)
}

func download(url, dstDir, dstFile string) (string, error) {
	remoteReader, err := RemoteReader(url)
	if err != nil {
		return "", err
	}
	defer remoteReader.Close()

	// Create the destination directory
	if err := createDstDir(dstDir); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf(
			"Download of file '%s' failed because destination directory '%s' could not be created",
			dstFile, dstDir))
	}

	// Create the file
	destFilePath := filepath.Join(dstDir, dstFile)
	out, err := os.Create(destFilePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, remoteReader)
	if err != nil {
		return "", err
	}

	return destFilePath, nil
}

func createDstDir(dstDir string) error {
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		return os.MkdirAll(dstDir, os.ModePerm)
	}
	return nil
}
