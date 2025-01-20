package cmdcommon

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/spf13/cobra"
)

const (
	ExtensionLabelKey           = "kyma-cli/extension"
	ExtensionResourceLabelValue = "resource"

	ExtensionResourceInfoKey    = "resource"
	ExtensionRootCommandKey     = "rootCommand"
	ExtensionGenericCommandsKey = "templateCommands"
	ExtensionCoreCommandsKey    = "coreCommands"
)

// map of allowed core commands in format ID: FUNC
type CoreCommandsMap map[string]func(*KymaConfig) *cobra.Command

// allowed template commands
type TemplateCommandsList struct {
	Explain func(*templates.ExplainOptions) *cobra.Command
	Create  func(templates.KubeClientGetter, *templates.CreateOptions) *cobra.Command
	Delete  func(templates.KubeClientGetter, *templates.DeleteOptions) *cobra.Command
}

type ExtensionList []Extension

func (el *ExtensionList) ContainResource(kind string) bool {
	for _, extension := range *el {
		if extension.Resource.Kind == kind {
			return true
		}
	}

	return false
}

type Extension struct {
	// main command of the command group
	RootCommand RootCommand
	// details about managed resource passed to every sub-command
	Resource *types.ResourceInfo
	// configuration of generic commands (like 'create', 'delete', 'get', ...) which implementation is provided by the cli
	// most of these commands bases on the `Resource` field
	TemplateCommands *TemplateCommands
	// configuration of buildin commands (like 'registry config') which implementation is provided by cli
	// use this command to enable feature for a module
	CoreCommands []CoreCommandInfo
}

type RootCommand struct {
	// name of the command group
	Name string `yaml:"name"`
	// short description of the command group
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
}

type TemplateCommands struct {
	// allows to explaining command to the commands group in format:
	// kyma <root_command> explain
	ExplainCommand *types.ExplainCommand `yaml:"explain"`
	// allows to create resource based on the ResourceInfo structure
	// kyma <root_command> create
	CreateCommand *types.CreateCommand `yaml:"create"`
	// allows to delete resource based on the ResourceInfo structure
	// kyma <root_command> delete
	DeleteCommand *types.DeleteCommand `yaml:"delete"`
}

type CoreCommandInfo struct {
	// id of the functionality that cli will run when user use this command
	ActionID string `yaml:"actionID"`
}
