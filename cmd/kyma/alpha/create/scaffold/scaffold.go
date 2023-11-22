package scaffold

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli/cmd/kyma/alpha/create/module"
	"github.com/kyma-project/cli/internal/cli"
	pkgmodule "github.com/kyma-project/cli/pkg/module"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"os/exec"
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
 - Module Config - module-config.yaml (always generated)
 - Manifest - template-operate.yaml (generated when the "--gen-manifest" flag is set)
 - Security Scanners Config - sec-scanners-config.yaml (generated when the "--gen-sec-config" flag is set)
 - Default CR - config/samples/operator.kyma-project.io_v1alpha1_sample.yaml (generated when the "--gen-default-cr" is flag set)

You must specify the required fields of the module config using the following CLI arguments:
--module-name [NAME]
--module-version [VERSION]
--module-channel [CHANNEL]
--module-manifest-path [MANIFEST-PATH] (cannot be used with the "--gen-manifest" flag)

**NOTE:**: If the required fields aren't provided, the module-config.yaml is not ready to use out-of-the-box. You must manually edit the file to make it usable.
Also, edit the sec-scanners-config.yaml to be able to use it.

The command is designed to streamline the module creation process in Kyma, making it easier and more 
efficient for developers to get started with new modules. It supports customization through various flags, 
allowing for a tailored scaffolding experience according to the specific needs of the module being created.`,
		Example: `Examples:
Generate a simple scaffold for a module
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

	if cmd.opts.GenerateManifest {
		cmd.NewStep("Generating manifest file...")

		err := cmd.generateManifest()
		if err != nil {
			cmd.CurrentStep.Failuref("%v: %w", errManifestCreationFailed, err)
			return fmt.Errorf("%v: %w", errManifestCreationFailed, err)
		}

		cmd.CurrentStep.Successf("Generated %s", fileNameManifest)
	}

	if cmd.opts.GenerateSecurityConfig {
		cmd.NewStep("Generating security config file...")

		err := cmd.generateSecurityConfig()
		if err != nil {
			cmd.CurrentStep.Failuref("%w: %v", errSecurityConfigCreationFailed, err)
			return fmt.Errorf("%w: %v", errSecurityConfigCreationFailed, err)
		}

		cmd.CurrentStep.Successf("Generated %s", fileNameSecurityConfig)
	}

	if cmd.opts.GenerateDefaultCR {
		cmd.NewStep("Generating default CR file...")

		err := cmd.generateDefaultCR()
		if err != nil {
			cmd.CurrentStep.Failuref("%w: %v", errDefaultCRCreationFailed, err)
			return fmt.Errorf("%w: %v", errDefaultCRCreationFailed, err)
		}

		cmd.CurrentStep.Successf("Generated %s", fileNameDefaultCR)
	}

	cmd.NewStep("Generating module config file...")

	err := cmd.generateModuleConfig()
	if err != nil {
		cmd.CurrentStep.Failuref("%w: %v", errModuleConfigCreationFailed, err)
		return fmt.Errorf("%w: %v", errModuleConfigCreationFailed, err)
	}

	cmd.CurrentStep.Successf("Generated %s", fileNameModuleConfig)
	return nil
}

func (cmd *command) generateManifest() error {
	makeCmd := exec.Command("make", "build-manifests")
	makeCmd.Dir = cmd.opts.Directory
	err := makeCmd.Run()
	return err
}

func (cmd *command) generateSecurityConfig() error {
	cfg := pkgmodule.SecurityScanCfg{
		ModuleName: cmd.opts.ModuleConfigName,
		Protecode: []string{"europe-docker.pkg.dev/kyma-project/prod/myimage:1.2.3",
			"europe-docker.pkg.dev/kyma-project/prod/external/ghcr.io/mymodule/anotherimage:4.5.6"},
		WhiteSource: pkgmodule.WhiteSourceSecCfg{
			Exclude: []string{"**/test/**", "**/*_test.go"},
		},
	}
	err := cmd.generateYamlFileFromObject(cfg, fileNameSecurityConfig)
	return err
}

func (cmd *command) generateDefaultCR() error {
	content := `apiVersion: operator.kyma-project.io/v1alpha1
kind: Sample
metadata:
  name: sample-yaml
spec:
  resourceFilePath: "./module-data/yaml"`
	return os.WriteFile(cmd.opts.getCompleteFilePath(fileNameDefaultCR), []byte(content), 0600)
}

func (cmd *command) generateModuleConfig() error {
	cfg := module.Config{
		Name:          cmd.opts.ModuleConfigName,
		Version:       cmd.opts.ModuleConfigVersion,
		Channel:       cmd.opts.ModuleConfigChannel,
		DefaultCRPath: chooseValue(cmd.opts.GenerateDefaultCR, fileNameDefaultCR, ""),
		Security:      chooseValue(cmd.opts.GenerateSecurityConfig, fileNameSecurityConfig, ""),
	}
	if cmd.opts.GenerateManifest {
		cfg.ManifestPath = fileNameManifest
	} else if cmd.opts.ModuleConfigManifestPath != "" {
		cfg.ManifestPath = cmd.opts.ModuleConfigManifestPath
	}
	return cmd.generateYamlFileFromObject(cfg, fileNameModuleConfig)
}

func (cmd *command) generateYamlFileFromObject(obj interface{}, fileName string) error {
	reflectValue := reflect.ValueOf(obj)
	var yamlBuilder strings.Builder
	generateYaml(&yamlBuilder, reflectValue, 0, "")
	yaml := yamlBuilder.String()

	err := os.WriteFile(cmd.opts.getCompleteFilePath(fileName), []byte(yaml), 0600)
	if err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	return nil
}

func generateYaml(yamlBuilder *strings.Builder, reflectValue reflect.Value, indentLevel int, commentPrefix string) {
	t := reflectValue.Type()

	indentPrefix := strings.Repeat("  ", indentLevel)
	originalCommentPrefix := commentPrefix
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := reflectValue.Field(i)
		tag := field.Tag.Get("yaml")
		comment := field.Tag.Get("comment")

		if value.IsZero() && !strings.Contains(comment, "required") {
			commentPrefix = "# "
		}

		if value.Kind() == reflect.Struct {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, tag, comment))
			generateYaml(yamlBuilder, value, indentLevel+1, commentPrefix)
			continue
		}

		if value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
			yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: # %s\n", commentPrefix, indentPrefix, tag, comment))
			for j := 0; j < value.Len(); j++ {
				valueStr := getValueStr(value.Index(j))
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  - %s\n", commentPrefix, indentPrefix, valueStr))
			}
			if value.Len() == 0 {
				yamlBuilder.WriteString(fmt.Sprintf("%s%s  - \n", commentPrefix, indentPrefix))
			}
			continue
		}

		valueStr := getValueStr(value)
		yamlBuilder.WriteString(fmt.Sprintf("%s%s%s: %s # %s\n", commentPrefix, indentPrefix,
			tag, valueStr, comment))

		commentPrefix = originalCommentPrefix
	}
}

func getValueStr(value reflect.Value) string {
	valueStr := ""
	if value.Kind() == reflect.String {
		valueStr = fmt.Sprintf("\"%v\"", value.Interface())
	} else {
		valueStr = fmt.Sprintf("%v", value.Interface())
	}
	return valueStr
}

func chooseValue(condition bool, trueValue, falseValue string) string {
	if condition {
		return trueValue
	}
	return falseValue
}
