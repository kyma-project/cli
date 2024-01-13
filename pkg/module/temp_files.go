package module

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

var tempFiles []*os.File

func DownloadRemoteFileToTempFile(url, dir, filenamePattern string) (string, error) {
	bytes, err := getBytesFromURL(url)
	if err != nil {
		return "", fmt.Errorf("failed to download file from %s: %w", url, err)
	}

	tmpfile, err := os.CreateTemp(dir, filenamePattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file with pattern %s: %w", filenamePattern, err)
	}
	defer tmpfile.Close()
	tempFiles = append(tempFiles, tmpfile)
	if _, err := tmpfile.Write(bytes); err != nil {
		return "", fmt.Errorf("failed to write to temp file %s: %w", tmpfile.Name(), err)
	}

	return tmpfile.Name(), nil
}

func DeleteTempFiles() []error {
	var errors []error
	for _, file := range tempFiles {
		err := os.Remove(file.Name())
		if err != nil {
			errors = append(errors, err)
		}
	}
	tempFiles = []*os.File{}
	return errors
}

func getBytesFromURL(urlString string) ([]byte, error) {
	url, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("parseing url failed for %s: %w", urlString, err)
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
