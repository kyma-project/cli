package files

import (
	"os"
	"path/filepath"
	"testing"
)

const gitFolderName = ".git"

func TestSearchForTargetDirByName(t *testing.T) {
	t.Run("git folder exists", func(t *testing.T) {
		tmpDir := createTempDirWithGit(t, gitFolderName)

		gitPath, err := SearchForTargetDirByName(tmpDir, gitFolderName)
		if err != nil {
			t.Errorf("SearchForTargetDirByName() error = %v", err)
		}

		expectedGitPath := filepath.Join(tmpDir, ".git")
		if gitPath != expectedGitPath {
			t.Errorf("SearchForTargetDirByName() gitPath = %v, want = %v", gitPath, expectedGitPath)
		}
	})

	t.Run("git folder does not exist", func(t *testing.T) {
		tmpDir := createTempDir(t)

		_, err := SearchForTargetDirByName(tmpDir, gitFolderName)
		if err != nil {
			t.Errorf("SearchForTargetDirByName() expected no error, got = %v", err)
		}
	})

	t.Run("input path does not exist", func(t *testing.T) {
		_, err := SearchForTargetDirByName("/no/such/dir", gitFolderName)
		if err == nil {
			t.Errorf("SearchForTargetDirByName() expected error, got nil")
		}
	})

	t.Run("git folder exists in nested dir", func(t *testing.T) {
		tmpDir := createTempNestedDirWithGit(t)

		gitPath, err := SearchForTargetDirByName(tmpDir, gitFolderName)
		if err != nil {
			t.Errorf("SearchForTargetDirByName() error = %v", err)
		}

		expectedGitPath := filepath.Join(tmpDir, "a/b/.git")
		if gitPath != expectedGitPath {
			t.Errorf("SearchForTargetDirByName() gitPath = %v, want = %v", gitPath, expectedGitPath)
		}
	})
}

func TestIsFileExists(t *testing.T) {
	tests := []struct {
		filePath     string
		expectedBool bool
	}{
		{"", false},
		{"nonexistentfile.txt", false},
		{"existingfile.txt", true},
		{"/invalid/path", false},
	}

	// Create a temporary file for testing.
	tempFile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	for _, test := range tests {
		t.Run(test.filePath, func(t *testing.T) {
			exists := IsFileExists(test.filePath)
			if exists != test.expectedBool {
				t.Errorf("Expected IsFileExists(%s) to be %v, but got %v", test.filePath, test.expectedBool, exists)
			}
		})
	}
}

func createTempDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "SearchForTargetDirByName")
	if err != nil {
		t.Fatalf("could not create temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	return tmpDir
}

func createTempNestedDirWithGit(t *testing.T) string {
	t.Helper()

	tmpDir := createTempDir(t)

	nestedDir := filepath.Join(tmpDir, "a/b/")

	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("could not create temp nested dir: %v", err)
	}

	err = os.Mkdir(filepath.Join(nestedDir, ".git"), 0755)
	if err != nil {
		t.Fatalf("could not create temp .git nested dir: %v", err)
	}

	return tmpDir
}

func createTempDirWithGit(t *testing.T, path string) string {
	t.Helper()

	tmpDir := createTempDir(t)

	err := os.Mkdir(filepath.Join(tmpDir, path), 0755)
	if err != nil {
		t.Fatalf("could not create temp .git dir: %v", err)
	}

	return tmpDir
}
