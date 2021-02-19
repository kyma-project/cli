package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

//GetFile downloads a file. Destination directory will be create if it does not exist.
func GetFile(file, dstDir string) (string, error) {
	localFiles, err := GetFiles([]string{file}, dstDir)
	if err == nil {
		return localFiles[0], nil
	}
	return "", err
}

//GetFiles downloads a list of files. Destination directory will be create if it does not exist.
func GetFiles(files []string, dstDir string) ([]string, error) {
	result := []string{}
	for _, file := range files {
		urlTokens := strings.Split(file, "://")
		switch len(urlTokens) {
		case 1:
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

func download(url, dstDir, dstFile string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return destFilePath, nil
}

func createDstDir(dstDir string) error {
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		return os.MkdirAll(dstDir, 0700)
	}
	return nil
}
