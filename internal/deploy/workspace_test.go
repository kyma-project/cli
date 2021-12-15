package deploy

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveWorkspace(t *testing.T) {
	t.Run("Check workspace folder is not deleted when it is set", func(t *testing.T) {
		ws := path.Join("dummyWS")
		err := os.Mkdir(ws, 0700)
		require.NoError(t, err)
		wsp, err := ResolveLocalWorkspacePath(ws, true)
		require.NoError(t, err)
		require.Equal(t, ws, wsp)
		err = pathExists(ws, "dummy ws path")
		require.NoError(t, err)
		err = os.Remove(ws)
		require.NoError(t, err)
	})

	t.Run("When workspace empty, then expect default workspace path", func(t *testing.T) {
		wsp, err := ResolveLocalWorkspacePath("", false)
		require.NoError(t, err)
		require.Equal(t, getDefaultWorkspacePath(), wsp)
	})

	t.Run("Check kyma workspace is being deleted", func(t *testing.T) {
		ws := path.Join("checkDeleteWS")
		err := os.Mkdir(ws, 0700)
		require.NoError(t, err)
		wsp, err := ResolveLocalWorkspacePath(ws, false)
		require.NoError(t, err)
		require.Equal(t, ws, wsp)
		_, err = os.Stat(ws)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no such file or directory")
	})
}
