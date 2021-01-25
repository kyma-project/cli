package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
