package scaffold

import (
	"os"

	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/pkg/errors"
)

func (cmd *command) moduleConfigFilePath() string {
	return cmd.opts.getCompleteFilePath(cmd.opts.ModuleConfigFile)
}

func (cmd *command) moduleConfigFileExists() (bool, error) {
	if _, err := os.Stat(cmd.moduleConfigFilePath()); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
}

func (cmd *command) generateModuleConfigFile() error {
	cfg := module.Config{
		Name:          cmd.opts.ModuleName,
		Version:       cmd.opts.ModuleVersion,
		Channel:       cmd.opts.ModuleChannel,
		ManifestPath:  cmd.opts.ManifestFile,
		Security:      cmd.opts.SecurityConfigFile,
		DefaultCRPath: cmd.opts.DefaultCRFile,
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return cmd.generateYamlFileFromObject(cfg, cmd.moduleConfigFilePath())
}
