package scaffold

import (
	"fmt"
	"os"
	"path"
)

func (g *Generator) ManifestFilePath() string {
	return path.Join(g.Directory, g.ManifestFile)
}

func (g *Generator) ManifestFileExists() (bool, error) {
	return g.fileExists(g.ManifestFilePath())
}

func (g *Generator) GenerateManifest() error {

	blankContents := `# This file holds the Manifest of your module, encompassing all resources installed in the cluster once the module is activated.
# It should include the Custom Resource Definition for your module's default CustomResource, if it exists.

`
	filePath := g.ManifestFilePath()
	err := os.WriteFile(filePath, []byte(blankContents), 0600)
	if err != nil {
		return fmt.Errorf("error while saving %s: %w", filePath, err)
	}

	return nil
}
