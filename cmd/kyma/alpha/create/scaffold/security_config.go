package scaffold

import (
	"os"

	"github.com/kyma-project/cli/pkg/module"
	"github.com/pkg/errors"
)

func (cmd *command) securityConfigFilePath() string {
	return cmd.opts.getCompleteFilePath(cmd.opts.SecurityConfigFile)
}

func (cmd *command) securityConfigFileExists() (bool, error) {
	if _, err := os.Stat(cmd.securityConfigFilePath()); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil

	} else {
		return false, err
	}
}

func (cmd *command) generateSecurityConfigFile() error {
	cfg := module.SecurityScanCfg{
		ModuleName: cmd.opts.ModuleName,
		Protecode: []string{"europe-docker.pkg.dev/kyma-project/prod/myimage:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6"},
		WhiteSource: module.WhiteSourceSecCfg{
			Exclude: []string{"**/test/**", "**/*_test.go"},
		},
	}
	err := cmd.generateYamlFileFromObject(cfg, cmd.securityConfigFilePath())
	return err
}
