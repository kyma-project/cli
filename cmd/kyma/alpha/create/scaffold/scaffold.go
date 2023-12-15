package scaffold

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"reflect"
	"strings"
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
		Use:   "scaffold [--module-name MODULE_NAME --module-version MODULE_VERSION --module-channel CHANNEL --module-manifest] [--directory MODULE_DIRECTORY] [flags]",
		Short: "Generates necessary files required for module creation",
		Long: `Scaffold generates the necessary files for creating a new module in Kyma. This includes setting up 
a basic directory structure and creating default files based on the provided flags.

The command generates the following files:
 - Module Config:
	File: module-config.yaml
	Generated when: always
 - Manifest:
	File: template-operator.yaml
	Generated When: The "--gen-manifest" flag is set
 - Security Scanners Config:
	File: sec-scanners-config.yaml
	Generated When: The "--gen-sec-config" flag is set
 - Default CR(s):
	File(s): config/samples/*.yaml
	Generated When: The "--gen-default-cr" flag is set
	**NOTE:** The generated default CRs might contain field values which violate integer or string pattern constraints. These are just templates that have to be modified. 

You can specify the required fields of the module config using the following CLI arguments:
--module-name [NAME]
--module-version [VERSION]
--module-channel [CHANNEL]
--module-manifest-path [MANIFEST-PATH] (cannot be used with the "--gen-manifest" flag)

**NOTE:**: If the required fields aren't provided, the module-config.yaml is not ready to use out-of-the-box. You must manually edit the file to make it usable.
Also, edit the sec-scanners-config.yaml to be able to use it.

The command is designed to streamline the module creation process in Kyma, making it easier and more 
efficient for developers to get started with new modules. It supports customization through various flags, 
allowing for a tailored scaffolding experience according to the specific needs of the module being created.`,
		Example: `Generate a simple scaffold for a module
		kyma alpha create scaffold --module-name=template-operator --module-version=1.0.0 --module-channel=regular --module-manifest-path=./template-operator.yaml
Generate a scaffold with manifest file, default CR, and security config for a module
		kyma alpha create scaffold --module-name=template-operator --module-version=1.0.0 --module-channel=regular --gen-manifest --gen-sec-config --gen-default-cr
`,
		RunE: func(cobraCmd *cobra.Command, args []string) error { return c.Run(cobraCmd.Context()) },
	}

	cmd.Flags().BoolVarP(
		&o.Overwrite, "overwrite", "o", false,
		"Specifies if the scaffold overwrites existing files",
	)
	cmd.Flags().StringVarP(
		&o.Directory, "directory", "d", "./",
		"Specifies the directory where the scaffolding shall be generated",
	)

	cmd.Flags().BoolVar(
		&o.GenerateSecurityConfig, "gen-sec-config", false,
		"Specifies if security config should be generated",
	)
	cmd.Flags().BoolVar(
		&o.GenerateManifest, "gen-manifest", false,
		"Specifies if manifest file should be generated",
	)
	cmd.Flags().BoolVar(
		&o.GenerateDefaultCR, "gen-default-cr", false,
		"Specifies if a default CR should be generated",
	)

	cmd.Flags().StringVar(
		&o.ModuleConfigName, "module-name", "",
		"Specifies the module name in the generated module config file",
	)
	cmd.Flags().StringVar(
		&o.ModuleConfigVersion, "module-version", "",
		"Specifies the module version in the generated module config file",
	)
	cmd.Flags().StringVar(
		&o.ModuleConfigChannel, "module-channel", "",
		"Specifies the module channel in the generated module config file",
	)
	cmd.Flags().StringVar(
		&o.ModuleConfigManifestPath, "module-manifest-path", "",
		"Specifies the module manifest filepath in the generated module config file",
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

	if err := cmd.opts.Validate(); err != nil {
		return err
	}

	if cmd.opts.GenerateManifest || cmd.opts.GenerateDefaultCR {
		cmd.NewStep("Generating webhook, rbac and crd objects...\n")

		err := cmd.generateControllerObjects()
		if err != nil {
			cmd.CurrentStep.Failuref("%s: %s", errObjectsCreationFailed.Error(), err.Error())
			return fmt.Errorf("%w: %s", errObjectsCreationFailed, err.Error())
		}

		cmd.CurrentStep.Successf("Generated webhook, rbac and crd objects in config/ directory")
	}

	if cmd.opts.GenerateManifest {
		cmd.NewStep("Generating manifest file...\n")

		err := cmd.generateManifest()
		if err != nil {
			cmd.CurrentStep.Failuref("%s: %s", errManifestCreationFailed.Error(), err.Error())
			return fmt.Errorf("%w: %s", errManifestCreationFailed, err.Error())
		}

		cmd.CurrentStep.Successf("Generated manifest file - %s", fileNameManifest)
	}

	if cmd.opts.GenerateSecurityConfig {
		cmd.NewStep("Generating security config file...\n")

		err := cmd.generateSecurityConfig()
		if err != nil {
			cmd.CurrentStep.Failuref("%s: %s", errSecurityConfigCreationFailed.Error(), err.Error())
			return fmt.Errorf("%w: %s", errSecurityConfigCreationFailed, err.Error())
		}

		cmd.CurrentStep.Successf("Generated security config file - %s", fileNameSecurityConfig)
	}

	if cmd.opts.GenerateDefaultCR {
		cmd.NewStep("Generating default CR file...\n")

		err := cmd.generateDefaultCR()
		if err != nil {
			cmd.CurrentStep.Failuref("%s: %s", errDefaultCRCreationFailed.Error(), err.Error())
			return fmt.Errorf("%w: %s", errDefaultCRCreationFailed, err.Error())
		}

		cmd.CurrentStep.Successf("Generated default CR file(s) in config/samples/ directory")
	}

	cmd.NewStep("Generating module config file...\n")

	err := cmd.generateModuleConfig()
	if err != nil {
		cmd.CurrentStep.Failuref("%s: %s", errModuleConfigCreationFailed.Error(), err.Error())
		return fmt.Errorf("%w: %s", errModuleConfigCreationFailed, err.Error())
	}

	cmd.CurrentStep.Successf("Generated module config file - %s", fileNameModuleConfig)
	return nil
}

func (cmd *command) generateModuleConfig() error {
	cfg := module.Config{
		Name:         cmd.opts.ModuleConfigName,
		Version:      cmd.opts.ModuleConfigVersion,
		Channel:      cmd.opts.ModuleConfigChannel,
		ManifestPath: cmd.opts.ModuleConfigManifestPath,
		Security:     chooseValue(cmd.opts.GenerateSecurityConfig, fileNameSecurityConfig, ""),
	}
	if cmd.opts.GenerateDefaultCR && len(generatedDefaultCRFiles) == 1 {
		cfg.DefaultCRPath = generatedDefaultCRFiles[0]
	}
	if cmd.opts.GenerateManifest {
		cfg.ManifestPath = fileNameManifest
	}
	return cmd.generateYamlFileFromObject(cfg, fileNameModuleConfig)
}

func (cmd *command) generateYamlFileFromObject(obj interface{}, fileName string) error {
	reflectValue := reflect.ValueOf(obj)
	var yamlBuilder strings.Builder
	generateYaml(&yamlBuilder, reflectValue, 0, "")
	yamlString := yamlBuilder.String()

	err := os.WriteFile(cmd.opts.getCompleteFilePath(fileName), []byte(yamlString), 0600)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}