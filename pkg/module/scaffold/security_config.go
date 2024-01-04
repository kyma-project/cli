package scaffold

import (
	"os"

	"github.com/kyma-project/cli/pkg/module"
	"github.com/pkg/errors"
)

func (g *Generator) SecurityConfigFilePath() string {
	return g.SecurityConfigFile
}

func (g *Generator) SecurityConfigFileExists() (bool, error) {
	if _, err := os.Stat(g.SecurityConfigFilePath()); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
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
