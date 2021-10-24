package deploy

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/kyma-project/cli/internal/deploy/values"
	"github.com/stretchr/testify/require"
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
		opts := Options{
			Sources: values.Sources{
				TLSKeyFile: "not-existing.key",
			},
		}
		err := opts.validateFlags()
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
		err := opts.validateFlags()
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
		err := opts.validateFlags()
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
		err := opts.validateFlags()
		require.NoError(t, err)
	})
	t.Run("Check workspace folder is not deleted when it is set", func(t *testing.T) {
		ws := path.Join("testdata", "dummyWS")
		Opts := Options{Source: "local", WorkspacePath: ws}
		err := os.Mkdir(ws, 0700)
		require.NoError(t, err)
		wsp, err := Opts.ResolveLocalWorkspacePath()
		require.NoError(t, err)
		require.Equal(t, ws, wsp)
		err = Opts.pathExists(ws, "dummy ws path")
		require.NoError(t, err)
		err = os.Remove(ws)
		require.NoError(t, err)
	})

	t.Run("When workspace empty, then expect default workspace path", func(t *testing.T) {
		opts := Options{Source: "main"}
		wsp, err := opts.ResolveLocalWorkspacePath()
		require.NoError(t, err)
		require.Equal(t, defaultWorkspacePath, wsp)
	})

	t.Run("Check kyma workspace is being deleted", func(t *testing.T) {
		ws := path.Join("testdata", "checkDeleteWS")
		opts := Options{Source: "main", WorkspacePath: ws}
		err := os.Mkdir(ws, 0700)
		require.NoError(t, err)
		wsp, err := opts.ResolveLocalWorkspacePath()
		require.NoError(t, err)
		require.Equal(t, ws, wsp)
		_, err = os.Stat(ws)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("Negative timeout should be rejected", func(t *testing.T) {
		opts := Options{Timeout: -10 * time.Minute}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout must be a positive duration")
	})
}
