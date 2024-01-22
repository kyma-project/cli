package scaffold

import (
	"path"

	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
)

func (g *Generator) ModuleConfigFilePath() string {
	return path.Join(g.Directory, g.ModuleConfigFile)
}

func (g *Generator) ModuleConfigFileExists() (bool, error) {
	return g.fileExists(g.ModuleConfigFilePath())
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
