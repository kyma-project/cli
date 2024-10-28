package cmdcommon

import (
	"context"
	"errors"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
	pkgerrors "github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func BuildExtensions(config *KymaConfig) []*cobra.Command {
	cmds := make([]*cobra.Command, len(config.Extensions))

	for i, extension := range config.Extensions {
		cmds[i] = buildCommandFromExtension(&extension)
	}

	return cmds
}

func buildCommandFromExtension(extension *Extension) *cobra.Command {
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
		addGenericCommands(cmd, extension.TemplateCommands)
	}

	return cmd
}

func addGenericCommands(cmd *cobra.Command, genericCommands *TemplateCommands) {
	if genericCommands.ExplainCommand != nil {
		cmd.AddCommand(templates.BuildExplainCommand(&templates.ExplainOptions{
			Short:  genericCommands.ExplainCommand.Description,
			Long:   genericCommands.ExplainCommand.DescriptionLong,
			Output: genericCommands.ExplainCommand.Output,
		}))
	}
}

func ListExtensions(ctx context.Context, client kubernetes.Interface) (ExtensionList, error) {
	labelSelector := fmt.Sprintf("%s==%s", ExtensionLabelKey, ExtensionResourceLabelValue)
	cms, err := client.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
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

	resourceInfo, err := parseOptionalField[ResourceInfo](cmData, ExtensionResourceInfoKey)
	if err != nil {
		return nil, err
	}

	genericCommands, err := parseOptionalField[TemplateCommands](cmData, ExtensionGenericCommandsKey)
	if err != nil {
		return nil, err
	}

	return &Extension{
		RootCommand:      *rootCommand,
		Resource:         resourceInfo,
		TemplateCommands: genericCommands,
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

func parseOptionalField[T any](cmData map[string]string, cmKey string) (*T, error) {
	dataBytes, ok := cmData[cmKey]
	if !ok {
		// skip because field is not required
		return nil, nil
	}

	var data T
	err := yaml.Unmarshal([]byte(dataBytes), &data)
	return &data, err
}
