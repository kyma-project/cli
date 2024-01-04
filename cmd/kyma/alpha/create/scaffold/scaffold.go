package scaffold

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli/internal/cli"
	scaffgen "github.com/kyma-project/cli/pkg/module/scaffold"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	defaultCRFlagName           = "gen-default-cr"
	defaultCRFlagDefault        = "default-cr.yaml"
	manifestFileFlagName        = "gen-manifest"
	manifestFileFlagDefault     = "manifest.yaml"
	moduleConfigFileFlagName    = "module-config"
	moduleConfigFileFlagDefault = "scaffold-module-config.yaml"
	securityConfigFlagName      = "gen-security-config"
	securityConfigFlagDefault   = "sec-scanners-config.yaml"
)

type command struct {
	cli.Command
	opts *Options
}

func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "scaffold [--module-name MODULE_NAME --module-version MODULE_VERSION --module-channel CHANNEL] [--directory MODULE_DIRECTORY] [flags]",
		Short: "Generates necessary files required for module creation",
		Long: `Scaffold generates or configures the necessary files for creating a new module in Kyma. This includes setting up 
a basic directory structure and creating default files based on the provided flags.

The command is designed to streamline the module creation process in Kyma, making it easier and more 
efficient for developers to get started with new modules. It supports customization through various flags, 
allowing for a tailored scaffolding experience according to the specific needs of the module being created.

The command generates or uses the following files:
 - Module Config:
	Enabled: Always
	Adjustable with flag: --module-config=VALUE
	Generated when: The file doesn't exist or the --overwrite=true flag is provided
	Default file name: scaffold-module-config.yaml
 - Manifest:
	Enabled: Always
	Adjustable with flag: --gen-manifest=VALUE
	Generated when: The file doesn't exist. If the file exists, it's name is used in the generated module configuration file
	Default file name: manifest.yaml
 - Default CR(s):
	Enabled: When the flag --gen-default-cr is provided with or without value
	Adjustable with flag: --gen-default-cr[=VALUE], if provided without an explicit VALUE, the default value is used
	Generated when: The file doesn't exist. If the file exists, it's name is used in the generated module configuration file
	Default file name: default-cr.yaml
 - Security Scanners Config:
	Enabled: When the flag --gen-security-config is provided with or without value
	Adjustable with flag: --gen-security-config[=VALUE], if provided without an explicit VALUE, the default value is used
	Generated when: The file doesn't exist. If the file exists, it's name is used in the generated module configuration file
	Default file name: sec-scanners-config.yaml

**NOTE:**: To protect the user from accidental file overwrites, this command by default doesn't overwrite any files.
Only the module config file may be force-overwritten when the --overwrite=true flag is used.

You can specify the required fields of the module config using the following CLI flags:
--module-name=NAME
--module-version=VERSION
--module-channel=CHANNEL

**NOTE:**: If the required fields aren't provided, the defaults are applied and the module-config.yaml is not ready to be used. You must manually edit the file to make it usable.
Also, edit the sec-scanners-config.yaml to be able to use it.
`,
		Example: `Generate a minimal scaffold for a module - only a blank manifest file and module config file is generated using defaults
                kyma alpha create scaffold
Generate a scaffold providing required values explicitly
                kyma alpha create scaffold --module-name="kyma-project.io/module/testmodule" --module-version="0.1.1" --module-channel=fast
Generate a scaffold with a manifest file, default CR and security-scanners config for a module
                kyma alpha create scaffold --gen-default-cr --gen-security-config
Generate a scaffold with a manifest file, default CR and security-scanners config for a module, overriding default values
                kyma alpha create scaffold --gen-manifest="my-manifest.yaml" --gen-default-cr="my-cr.yaml" --gen-security-config="my-seccfg.yaml"

`,
		RunE: func(cobraCmd *cobra.Command, args []string) error { return c.Run(cobraCmd.Context()) },
	}
	cmd.Flags().StringVar(
		&o.ModuleName, "module-name", "kyma-project.io/module/mymodule",
		"Specifies the module name in the generated module config file",
	)
	cmd.Flags().StringVar(
		&o.ModuleVersion, "module-version", "0.0.1",
		"Specifies the module version in the generated module config file",
	)
	cmd.Flags().StringVar(
		&o.ModuleChannel, "module-channel", "regular",
		"Specifies the module channel in the generated module config file",
	)
	cmd.Flags().StringVar(
		&o.ModuleConfigFile, moduleConfigFileFlagName, moduleConfigFileFlagDefault,
		"Specifies the name for the generated module configuration file",
	)
	cmd.Flags().Lookup(moduleConfigFileFlagName).NoOptDefVal = moduleConfigFileFlagDefault

	cmd.Flags().StringVar(
		&o.ManifestFile, manifestFileFlagName, manifestFileFlagDefault,
		"Specifies the manifest in the generated module config. A blank manifest file is generated if it doesn't exist",
	)
	cmd.Flags().Lookup(manifestFileFlagName).NoOptDefVal = manifestFileFlagDefault

	cmd.Flags().StringVar(
		&o.DefaultCRFile, defaultCRFlagName, "",
		"Specifies the defaultCR in the generated module config. A blank defaultCR file is generated if it doesn't exist",
	)
	cmd.Flags().Lookup(defaultCRFlagName).NoOptDefVal = defaultCRFlagDefault

	cmd.Flags().StringVar(
		&o.SecurityConfigFile, securityConfigFlagName, "",
		"Specifies the security file in the generated module config. A scaffold security config file is generated if it doesn't exist",
	)
	cmd.Flags().Lookup(securityConfigFlagName).NoOptDefVal = securityConfigFlagDefault

	cmd.Flags().BoolVarP(
		&o.Overwrite, "overwrite", "o", false,
		"Specifies if the command overwrites an existing module configuration file",
	)
	cmd.Flags().StringVarP(
		&o.Directory, "directory", "d", "./",
		"Specifies the directory where the scaffolding shall be generated",
	)

	return cmd
}

func (cmd *command) Run(_ context.Context) error {

	if cmd.opts.CI {
		cmd.Factory.NonInteractive = true
	}
	if cmd.opts.Verbose {
		cmd.Factory.UseLogger = true
	}

	l := cli.NewLogger(cmd.opts.Verbose).Sugar()
	undo := zap.RedirectStdLog(l.Desugar())
	defer undo()

	if !cmd.opts.NonInteractive {
		cli.AlphaWarn()
	}

	cmd.NewStep("Validating")
	if err := cmd.opts.Validate(); err != nil {
		cmd.CurrentStep.Failuref("%s", err.Error())
		return fmt.Errorf("%w", err)
	}

	sgen := cmd.scaffoldGeneratorFromOptions()

	moduleConfigExists, err := sgen.ModuleConfigFileExists()
	if err != nil {
		cmd.CurrentStep.Failuref("%s", err.Error())
		return fmt.Errorf("%w", err)
	}
	if moduleConfigExists && !cmd.opts.Overwrite {
		cmd.CurrentStep.Failuref("%s", errModuleConfigExists.Error())
		return fmt.Errorf("%w", errModuleConfigExists)
	}
	cmd.CurrentStep.Success()

	manifestFileExists, err := sgen.ManifestFileExists()
	if err != nil {
		return err
	}
	cmd.NewStep("Configuring manifest file...\n")
	if manifestFileExists {
		cmd.CurrentStep.Successf("The manifest file exists, reusing: %s", sgen.ManifestFilePath())
	} else {
		cmd.CurrentStep.Status("Generating the manifest file")
		err := sgen.GenerateManifest()
		if err != nil {
			cmd.CurrentStep.Failuref("%s: %s", errManifestCreationFailed.Error(), err.Error())
			return fmt.Errorf("%w: %s", errManifestCreationFailed, err.Error())
		}

		cmd.CurrentStep.Successf("Generated a blank manifest file: %s", sgen.ManifestFilePath())
	}

	if cmd.opts.generateDefaultCRFile() {
		defaultCRFileExists, err := sgen.DefaultCRFileExists()
		if err != nil {
			return err
		}
		cmd.NewStep("Configuring defaultCR file...\n")
		if defaultCRFileExists {
			cmd.CurrentStep.Successf("The defaultCR file exists, reusing: %s", sgen.DefaultCRFilePath())
		} else {
			cmd.CurrentStep.Status("Generating the default CR file")
			err := sgen.GenerateDefaultCRFile()
			if err != nil {
				cmd.CurrentStep.Failuref("%s: %s", errDefaultCRCreationFailed.Error(), err.Error())
				return fmt.Errorf("%w: %s", errDefaultCRCreationFailed, err.Error())
			}

			cmd.CurrentStep.Successf("Generated a blank defaultCR file: %s", sgen.DefaultCRFilePath())
		}
	}

	if cmd.opts.generateSecurityConfigFile() {
		secCfgFileExists, err := sgen.SecurityConfigFileExists()
		if err != nil {
			return err
		}
		cmd.NewStep("Configuring security-scanners config file...\n")
		if secCfgFileExists {
			cmd.CurrentStep.Successf("The security-scanners config file exists, reusing: %s", sgen.SecurityConfigFilePath())
		} else {
			cmd.CurrentStep.Status("Generating security-scanners config file")
			err := sgen.GenerateSecurityConfigFile()
			if err != nil {
				cmd.CurrentStep.Failuref("%s: %s", errSecurityConfigCreationFailed.Error(), err.Error())
				return fmt.Errorf("%w: %s", errSecurityConfigCreationFailed, err.Error())
			}

			cmd.CurrentStep.Successf("Generated security-scanners config file - %s", sgen.SecurityConfigFilePath())
		}
	}

	cmd.NewStep("Generating module config file...\n")

	err = sgen.GenerateModuleConfigFile()
	if err != nil {
		cmd.CurrentStep.Failuref("%s: %s", errModuleConfigCreationFailed.Error(), err.Error())
		return fmt.Errorf("%w: %s", errModuleConfigCreationFailed, err.Error())
	}

	cmd.CurrentStep.Successf("Generated module config file: %s", sgen.ModuleConfigFilePath())

	return nil
}

func (cmd *command) scaffoldGeneratorFromOptions() *scaffgen.Generator {

	toFullPath := func(file string) string {
		if file == "" {
			return ""
		}
		return cmd.opts.getCompleteFilePath(file)
	}

	res := scaffgen.Generator{
		ModuleName:         cmd.opts.ModuleName,
		ModuleVersion:      cmd.opts.ModuleVersion,
		ModuleChannel:      cmd.opts.ModuleChannel,
		ModuleConfigFile:   toFullPath(cmd.opts.ModuleConfigFile),
		ManifestFile:       toFullPath(cmd.opts.ManifestFile),
		SecurityConfigFile: toFullPath(cmd.opts.SecurityConfigFile),
		DefaultCRFile:      toFullPath(cmd.opts.DefaultCRFile),
	}

	return &res
}
