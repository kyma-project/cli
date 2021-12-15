package deploy

import (
	"path"
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/stretchr/testify/require"
)

func TestOptsValidation(t *testing.T) {
	t.Run("unknown profile", func(t *testing.T) {
		opts := Options{Profile: "fancy"}
		err := opts.validateProfile()
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown profile: fancy")
	})

	t.Run("supported profiles", func(t *testing.T) {
		profiles := []string{"", "evaluation", "production"}
		for _, p := range profiles {
			opts := Options{Profile: p}
			err := opts.validateProfile()
			require.NoError(t, err)
		}
	})

	t.Run("tls key not found", func(t *testing.T) {
		opts := Options{
			Sources: values.Sources{
				TLSKeyFile: "not-existing.key",
			},
		}
		err := opts.validateTLSCertAndKey()
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls key not found")
	})

	t.Run("tls key exists but crt not found", func(t *testing.T) {
		dummyFilePath := path.Join("testdata", "dummy.txt")
		opts := Options{
			Sources: values.Sources{
				TLSKeyFile: dummyFilePath,
				TLSCrtFile: "not-existing.crt",
			},
		}
		err := opts.validateTLSCertAndKey()
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls cert not found")
	})

	t.Run("tls crt exists but key not found", func(t *testing.T) {
		dummyFilePath := path.Join("testdata", "dummy.txt")
		opts := Options{
			Sources: values.Sources{
				TLSKeyFile: "not-existing.crt",
				TLSCrtFile: dummyFilePath,
			},
		}
		err := opts.validateTLSCertAndKey()
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls key not found")
	})

	t.Run("tls key and crt found", func(t *testing.T) {
		dummyFilePath := path.Join("testdata", "dummy.txt")
		opts := Options{
			Sources: values.Sources{
				TLSKeyFile: dummyFilePath,
				TLSCrtFile: dummyFilePath,
			},
		}
		err := opts.validateTLSCertAndKey()
		require.NoError(t, err)
	})

	t.Run("Negative timeout should be rejected", func(t *testing.T) {
		opts := Options{Timeout: -10 * time.Minute}
		err := opts.validateTimeout()
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout must be a positive duration")
	})
}
