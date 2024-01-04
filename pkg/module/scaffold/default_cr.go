package scaffold

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func (g *Generator) DefaultCRFilePath() string {
	return g.DefaultCRFile
}

func (g *Generator) DefaultCRFileExists() (bool, error) {
	if _, err := os.Stat(g.DefaultCRFilePath()); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
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
