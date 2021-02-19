package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//GetFile downloads a file
func GetFile(file, dstDir string) (string, error) {
	localFiles, err := GetFiles([]string{file}, dstDir)
	if err == nil {
		return localFiles[0], nil
	}
	return "", err
}

//GetFiles downloads a list of files
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
				dstFile := filepath.Join(dstDir, filepath.Base(urlTokens[1]))
				download(file, dstFile)
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

func download(url, dst string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
