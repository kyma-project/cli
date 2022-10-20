package kubebuilder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const projectYaml = `layout:
- go.kubebuilder.io/v3
projectName: example
repo: sigs.k8s.io/kubebuilder/example`

func TestParseProject(t *testing.T) {
	// invalid path
	_, err := ParseProject("/wrong/path")
	require.Error(t, err)

	//valid project
	path := testKubebuilderProject(t)
	defer func() {
		err := os.RemoveAll(path)
		require.NoError(t, err)
	}()

	p, err := ParseProject(path)
	require.NoError(t, err)

	expected := &Project{
		Name:   "example",
		Layout: []string{"go.kubebuilder.io/v3"},
		Repo:   "sigs.k8s.io/kubebuilder/example",
		path:   path,
	}
	require.Equal(t, expected, p)
}

func testKubebuilderProject(t *testing.T) string {
	d, err := os.MkdirTemp("", "kubebuilder-project")
	require.NoError(t, err)
	err = os.Mkdir(filepath.Join(d, "operator"), os.ModePerm)
	require.NoError(t, err)

	f, err := os.Create(filepath.Join(d, projectFile))
	defer func() {
		err := f.Close()
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	_, err = f.WriteString(projectYaml)
	require.NoError(t, err)

	return d
}
