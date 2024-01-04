package scaffold

import (
	"os"

	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/pkg/errors"
)

func (g *Generator) ModuleConfigFilePath() string {
	return g.ModuleConfigFile
}

func (g *Generator) ModuleConfigFileExists() (bool, error) {
	if _, err := os.Stat(g.ModuleConfigFilePath()); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
}

func (g *Generator) GenerateModuleConfigFile() error {
	cfg := module.Config{
		Name:          g.ModuleName,
		Version:       g.ModuleVersion,
		Channel:       g.ModuleChannel,
		ManifestPath:  g.ManifestFile,
		Security:      g.SecurityConfigFile,
		DefaultCRPath: g.DefaultCRFile,
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return g.generateYamlFileFromObject(cfg, g.ModuleConfigFilePath())
}
