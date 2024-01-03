package scaffold

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

func (cmd *command) manifestFilePath() string {
	return cmd.opts.getCompleteFilePath(cmd.opts.ManifestFile)
}

func (cmd *command) manifestFileExists() (bool, error) {
	if _, err := os.Stat(cmd.manifestFilePath()); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
}

func (cmd *command) generateManifest() error {

	blankContents := `# This file holds the Manifest of your module, encompassing all resources installed in the cluster once the module is activated.
# It should include the Custom Resource Definition for your module's default CustomResource, if it exists.

`
	filePath := cmd.manifestFilePath()
	err := os.WriteFile(filePath, []byte(blankContents), 0600)
	if err != nil {
		return fmt.Errorf("error while saving %s: %w", filePath, err)
	}

	return nil
}
