package deploy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	crtFile = filepath.Join(".", "testCrt.crt")
	keyFile = filepath.Join(".", "testCrt.key")
)

func TestMain(m *testing.M) {
	setUp()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

func setUp() {
	deleteCrtFiles()
	crt, err := base64.StdEncoding.DecodeString(defaultTLSCrtEnc)
	if err != nil {
		panic(errors.Wrap(err, "Failed to decode base64-encoded certificate key string"))
	}
	if err := ioutil.WriteFile(crtFile, crt, 0644); err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("Failed to write certificate file '%s'", crtFile)))
	}

	key, err := base64.StdEncoding.DecodeString(defaultTLSKeyEnc)
	if err != nil {
		panic(errors.Wrap(err, "Failed to decode base64-encoded certificate string"))
	}
	if err := ioutil.WriteFile(keyFile, key, 0644); err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("Failed to write certificate key file '%s'", keyFile)))
	}
}

func tearDown() {
	deleteCrtFiles()
}

func deleteCrtFiles() {
	os.Remove(crtFile)
	os.Remove(keyFile)
}

func TestCertAsFile(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: keyFile,
		}
		tlsCrt, err := opts.tlsCrtEnc()
		require.NoError(t, err)
		require.Equal(t, defaultTLSCrtEnc, tlsCrt)

		tlsKey, err := opts.tlsKeyEnc()
		require.NoError(t, err)
		require.Equal(t, defaultTLSKeyEnc, tlsKey)
	})
	t.Run("Invalid cert file", func(t *testing.T) {
		crtFileInvalid := fmt.Sprintf("%s.xyz", crtFile)
		opts := &Options{
			TLSCrtFile: crtFileInvalid,
			TLSKeyFile: keyFile,
		}
		_, err := opts.tlsCrtEnc()
		require.True(t, os.IsNotExist(err))
	})
	t.Run("Invalid key file", func(t *testing.T) {
		keyFileInvalid := fmt.Sprintf("%s.xyz", keyFile)
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: keyFileInvalid,
		}
		_, err := opts.tlsKeyEnc()
		require.True(t, os.IsNotExist(err))
	})
	t.Run("Ensure proper base64 encoding", func(t *testing.T) {
		ex, err := os.Executable()

		fileContent, err := ioutil.ReadFile(ex)
		require.NoError(t, err)
		expected := base64.StdEncoding.EncodeToString(fileContent)

		require.NoError(t, err)
		opts := &Options{
			TLSCrtFile: ex,
			TLSKeyFile: ex,
		}
		keyEnc, err := opts.tlsKeyEnc()
		require.NoError(t, err)
		require.Equal(t, expected, keyEnc)

		crtEnc, err := opts.tlsCrtEnc()
		require.NoError(t, err)
		require.Equal(t, expected, crtEnc)
	})
}

func TestOptsValidation(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: keyFile,
		}
		err := opts.validateFlags()
		require.NoError(t, err)
	})
	t.Run("Key is missing", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "key is empty")
	})
	t.Run("Cert is missing", func(t *testing.T) {
		opts := &Options{
			TLSKeyFile: keyFile,
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate is empty")
	})
	t.Run("Wrong cert path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: "/abc/test.yaml",
			TLSKeyFile: keyFile,
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
	t.Run("Wrong key path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: "/do/not/exist.key",
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
}

func TestComponentFile(t *testing.T) {
	t.Run("Test with non-changed workspace path", func(t *testing.T) {
		opts := &Options{
			WorkspacePath: defaultWorkspacePath,
		}
		assert.Equal(t, defaultComponentsFile, opts.ResolveComponentsFile())
	})
	t.Run("Test with changed component file", func(t *testing.T) {
		opts := &Options{
			WorkspacePath:  defaultWorkspacePath,
			ComponentsFile: "/xyz/comp.yaml",
		}
		assert.Equal(t, "/xyz/comp.yaml", opts.ResolveComponentsFile())
	})
	t.Run("Test with changed workspace path", func(t *testing.T) {
		opts := &Options{
			WorkspacePath: "/xyz/abc",
		}
		assert.Equal(t, "/xyz/abc/installation/resources/components.yaml", opts.ResolveComponentsFile())
	})
	t.Run("Test with changed workspace path ad componets file path", func(t *testing.T) {
		opts := &Options{
			WorkspacePath:  "/xyz/abc",
			ComponentsFile: "/some/where/components.yaml",
		}
		assert.Equal(t, "/some/where/components.yaml", opts.ResolveComponentsFile())
	})
}
