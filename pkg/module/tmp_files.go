package module

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type TmpFilesManager struct {
	tmpFiles []*os.File
}

func NewTmpFilesManager() *TmpFilesManager {
	return &TmpFilesManager{tmpFiles: []*os.File{}}
}

func (manager *TmpFilesManager) DownloadRemoteFileToTmpFile(url, dir, filenamePattern string) (string, error) {
	bytes, err := getBytesFromURL(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file from %s: %w", url, err)
	}

	tmpFile, err := os.CreateTemp(dir, filenamePattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file with pattern %s: %w", filenamePattern, err)
	}
	defer tmpFile.Close()
	manager.tmpFiles = append(manager.tmpFiles, tmpFile)
	if _, err := tmpFile.Write(bytes); err != nil {
		return "", fmt.Errorf("failed to write to temp file %s: %w", tmpFile.Name(), err)
	}

	return tmpFile.Name(), nil
}

func (manager *TmpFilesManager) DeleteTmpFiles() []error {
	var errors []error
	for _, file := range manager.tmpFiles {
		err := os.Remove(file.Name())
		if err != nil {
			errors = append(errors, err)
		}
	}
	manager.tmpFiles = []*os.File{}
	return errors
}

func getBytesFromURL(urlString string) ([]byte, error) {
	ok, url := ParseURL(urlString)
	if !ok {
		return nil, fmt.Errorf("parseing url failed for %s", urlString)
	}
	resp, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("http GET request failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status for GET request to %s: %s", url, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return data, nil
}

func ParseURL(s string) (bool, *url.URL) {
	u, err := url.Parse(s)
	if err == nil && u.Scheme != "" && u.Host != "" {
		return true, u
	}
	return false, nil
}
