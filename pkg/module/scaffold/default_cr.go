package scaffold

import (
	"fmt"
	"os"
)

func (g *Generator) DefaultCRFilePath() string {
	return g.DefaultCRFile
}

func (g *Generator) DefaultCRFileExists() (bool, error) {
	return g.fileExists(g.DefaultCRFilePath())
}

func (g *Generator) GenerateDefaultCRFile() error {

	blankContents := `# This is the file that contains the defaultCR for your module, which is the Custom Resource that will be created upon module enablement.
# Make sure this file contains *ONLY* the Custom Resource (not the Custom Resource Definition, which should be a part of your module manifest)

`
	filePath := g.DefaultCRFilePath()
	err := os.WriteFile(filePath, []byte(blankContents), 0600)
	if err != nil {
		return fmt.Errorf("error while saving %s: %w", filePath, err)
	}

	return nil
}
