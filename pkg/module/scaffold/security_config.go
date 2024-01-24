package scaffold

import (
	"path"

	"github.com/kyma-project/cli/pkg/module"
)

func (g *Generator) SecurityConfigFilePath() string {
	return path.Join(g.Directory, g.SecurityConfigFile)
}

func (g *Generator) SecurityConfigFileExists() (bool, error) {
	return g.fileExists(g.SecurityConfigFilePath())
}

func (g *Generator) GenerateSecurityConfigFile() error {
	cfg := module.SecurityScanCfg{
		ModuleName: g.ModuleName,
		Protecode: []string{"europe-docker.pkg.dev/kyma-project/prod/myimage:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6"},
		WhiteSource: module.WhiteSourceSecCfg{
			Exclude: []string{"**/test/**", "**/*_test.go"},
		},
	}
	err := g.generateYamlFileFromObject(cfg, g.SecurityConfigFilePath())
	return err
}
