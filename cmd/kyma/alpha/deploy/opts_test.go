package deploy

import (
	"github.com/stretchr/testify/require"
	"path"
	"testing"
)

func TestOptsValidation(t *testing.T) {
	t.Run("unknown profile", func(t *testing.T) {
		opts := Options{Profile: "fancy"}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown profile: fancy")
	})

	t.Run("supported profiles", func(t *testing.T) {
		profiles := []string{"", "evaluation", "production"}
		for _, p := range profiles {
			opts := Options{Profile: p}
			err := opts.validateFlags()
			require.NoError(t, err)
		}
	})

	t.Run("tls key not found", func(t *testing.T) {
		opts := Options{TLSKeyFile: "not-existing.key"}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls key not found")
	})

	t.Run("tls key exists but crt not found", func(t *testing.T) {
		dummyFilePath := path.Join("testdata", "dummy.txt")
		opts := Options{TLSKeyFile: dummyFilePath, TLSCrtFile: "not-existing.crt"}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls cert not found")
	})

	t.Run("tls crt exists but key not found", func(t *testing.T) {
		dummyFilePath := path.Join("testdata", "dummy.txt")
		opts := Options{TLSKeyFile: "not-existing.crt", TLSCrtFile: dummyFilePath}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls key not found")
	})

	t.Run("tls key and crt found", func(t *testing.T) {
		dummyFilePath := path.Join("testdata", "dummy.txt")
		opts := Options{TLSKeyFile: dummyFilePath, TLSCrtFile: dummyFilePath}
		err := opts.validateFlags()
		require.NoError(t, err)
	})
}
