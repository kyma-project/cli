package scaffold

import "github.com/kyma-project/cli/pkg/module"

func (cmd *command) generateSecurityConfig() error {
	cfg := module.SecurityScanCfg{
		ModuleName: cmd.opts.ModuleConfigName,
		Protecode: []string{"europe-docker.pkg.dev/kyma-project/prod/myimage:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6"},
		WhiteSource: module.WhiteSourceSecCfg{
			Exclude: []string{"**/test/**", "**/*_test.go"},
		},
	}
	err := cmd.generateYamlFileFromObject(cfg, fileNameSecurityConfig)
	return err
}
