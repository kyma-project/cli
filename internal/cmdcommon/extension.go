package cmdcommon

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
	pkgerrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KymaExtensionsConfig struct {
	kymaConfig *KymaConfig
	extensions ExtensionList
}

func newExtensionsConfig(config *KymaConfig, cmd *cobra.Command) *KymaExtensionsConfig {
	extensions, err := loadExtensionsFromCluster(config.Ctx, config.KubeClientConfig)
	if err != nil && shouldShowExtensionsError() {
		// print error as warning if expected and continue
		fmt.Printf("Extensions Warning:\n%s\n\n", err.Error())
	} else if err != nil {
		fmt.Print("Extensions Warning:\nfailed to fetch all extensions from the cluster. Use the '--show-extensions-error' flag to see more details.\n\n")
	}

	extensionsConfig := &KymaExtensionsConfig{
		kymaConfig: config,
		extensions: extensions,
	}
	extensionsConfig.addFlag(cmd)

	return extensionsConfig
}

func (kec *KymaExtensionsConfig) addFlag(cmd *cobra.Command) {
	// this flag is not operational. it's only to print help description and help cobra with validation
	_ = cmd.PersistentFlags().Bool("show-extensions-error", false, "Print possible error when fetching extensions failed.")
}

func (kec *KymaExtensionsConfig) BuildExtensions(availableTemplateCommands *TemplateCommandsList, availableCoreCommands CoreCommandsMap) []*cobra.Command {
	cmds := make([]*cobra.Command, len(kec.kymaConfig.extensions))

	for i, extension := range kec.kymaConfig.extensions {
		cmds[i] = buildCommandFromExtension(kec.kymaConfig, &extension, availableTemplateCommands, availableCoreCommands)
	}

	return cmds
}

func loadExtensionsFromCluster(ctx context.Context, clientConfig *KubeClientConfig) ([]Extension, error) {
	client, clientErr := clientConfig.GetKubeClient()
	if clientErr != nil {
		return nil, clientErr
	}

	labelSelector := fmt.Sprintf("%s==%s", ExtensionLabelKey, ExtensionResourceLabelValue)
	cms, err := client.Static().CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, pkgerrors.Wrapf(err, "failed to load ConfigMaps from cluster with label %s", labelSelector)
	}

	var extensions []Extension
	var parseErr error
	for _, cm := range cms.Items {
		extension, err := parseResourceExtension(cm.Data)
		if err != nil {
			// if the parse failed add an error to the errors list to take another extension
			// corrupted extension should not stop parsing the rest of the extensions
			parseErr = errors.Join(
				parseErr,
				pkgerrors.Wrapf(err, "failed to parse configmap '%s/%s'", cm.GetNamespace(), cm.GetName()),
			)
			continue
		}

		extensions = append(extensions, *extension)
	}

	return extensions, parseErr
}

func parseResourceExtension(cmData map[string]string) (*Extension, error) {
	rootCommand, err := parseRequiredField[RootCommand](cmData, ExtensionRootCommandKey)
	if err != nil {
		return nil, err
	}

	resourceInfo, err := parseOptionalField[*ResourceInfo](cmData, ExtensionResourceInfoKey)
	if err != nil {
		return nil, err
	}

	genericCommands, err := parseOptionalField[*TemplateCommands](cmData, ExtensionGenericCommandsKey)
	if err != nil {
		return nil, err
	}

	coreCommands, err := parseOptionalField[[]CoreCommandInfo](cmData, ExtensionCoreCommandsKey)
	if err != nil {
		return nil, err
	}

	return &Extension{
		RootCommand:      *rootCommand,
		Resource:         resourceInfo,
		TemplateCommands: genericCommands,
		CoreCommands:     coreCommands,
	}, nil
}

func parseRequiredField[T any](cmData map[string]string, cmKey string) (*T, error) {
	dataBytes, ok := cmData[cmKey]
	if !ok {
		return nil, fmt.Errorf("missing .data.%s field", cmKey)
	}

	var data T
	err := yaml.Unmarshal([]byte(dataBytes), &data)
	return &data, err
}

func parseOptionalField[T any](cmData map[string]string, cmKey string) (T, error) {
	var data T
	dataBytes, ok := cmData[cmKey]
	if !ok {
		// skip because field is not required
		return data, nil
	}

	err := yaml.Unmarshal([]byte(dataBytes), &data)
	return data, err
}

func buildCommandFromExtension(config *KymaConfig, extension *Extension, availableTemplateCommands *TemplateCommandsList, availableCoreCommands CoreCommandsMap) *cobra.Command {
	cmd := &cobra.Command{
		Use:   extension.RootCommand.Name,
		Short: extension.RootCommand.Description,
		Long:  extension.RootCommand.DescriptionLong,
		Run: func(cmd *cobra.Command, _ []string) {
			if err := cmd.Help(); err != nil {
				_ = err
			}
		},
	}

	if extension.TemplateCommands != nil {
		addGenericCommands(cmd, extension.TemplateCommands, availableTemplateCommands)
	}

	addCoreCommands(cmd, config, extension.CoreCommands, availableCoreCommands)

	return cmd
}

func addGenericCommands(cmd *cobra.Command, genericCommands *TemplateCommands, availableTemplateCommands *TemplateCommandsList) {
	if genericCommands.ExplainCommand != nil {
		cmd.AddCommand(availableTemplateCommands.Explain(&templates.ExplainOptions{
			Short:  genericCommands.ExplainCommand.Description,
			Long:   genericCommands.ExplainCommand.DescriptionLong,
			Output: genericCommands.ExplainCommand.Output,
		}))
	}
}

func addCoreCommands(cmd *cobra.Command, config *KymaConfig, extensionCoreCommands []CoreCommandInfo, availableCoreCommands CoreCommandsMap) {
	for _, expectedCoreCommand := range extensionCoreCommands {
		command, ok := availableCoreCommands[expectedCoreCommand.ActionID]
		if !ok {
			// commands doesn't exist in this version of cli and we will not process it
			continue
		}

		cmd.AddCommand(command(config))
	}
}

// search os.Args manually to find if user pass --show-extensions-error and return its value
func shouldShowExtensionsError() bool {
	for i, arg := range os.Args {
		//example: --show-extensions-error true
		if arg == "--show-extensions-error" && len(os.Args) > i+1 {

			value, err := strconv.ParseBool(os.Args[i+1])
			if err == nil {
				return value
			}
		}

		// example: --show-extensions-error or --show-extensions-error=true
		if strings.HasPrefix(arg, "--show-extensions-error") && !strings.Contains(arg, "false") {
			return true
		}
	}

	return false
}
