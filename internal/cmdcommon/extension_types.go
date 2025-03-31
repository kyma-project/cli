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
	ExtensionActionCommandsKey  = "actionCommands"
)

// map of allowed core commands in format ID: FUNC
type CoreCommandsMap map[string]func(*KymaConfig, interface{}) (*cobra.Command, error)

// map of allowed action commands in format ID: FUNC
type ActionCommandsMap map[string]func(*KymaConfig, types.ActionConfig) (*cobra.Command, error)

// allowed template commands
type TemplateCommandsList struct {
	Explain func(*templates.ExplainOptions) *cobra.Command
	Get     func(templates.KubeClientGetter, *templates.GetOptions) *cobra.Command
	Create  func(templates.KubeClientGetter, *templates.CreateOptions) *cobra.Command
	Delete  func(templates.KubeClientGetter, *templates.DeleteOptions) *cobra.Command
}

type ExtensionList []ExtensionItem

type ExtensionItem struct {
	ConfigMapName      string
	ConfigMapNamespace string
	Extension          Extension
}

func (el *ExtensionList) ContainResource(kind string) bool {
	for _, item := range *el {
		if item.Extension.Resource.Kind == kind {
			return true
		}
	}

	return false
}

type Extension struct {
	// main command of the command group
	RootCommand types.RootCommand
	// details about managed resource passed to every sub-command
	Resource *types.ResourceInfo
	// configuration of generic commands (like 'create', 'delete', 'get', ...) which implementation is provided by the cli
	// most of these commands bases on the `Resource` field
	TemplateCommands *TemplateCommands
	// configuration of buildin commands (like 'registry config') which implementation is provided by cli
	// use this command to enable feature for a module
	// DEPRECATED: use actionCommands
	CoreCommands []CoreCommandInfo
	// configuration of buildin commands (like 'registry config') which implementation is provided by cli
	// use this command to enable feature for a module
	ActionCommands []types.ActionCommand
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
	// allows to get resources based on the ResourceInfo structure
	// kyma <root_command> get
	GetCommand *types.GetCommand `yaml:"get"`
}

type CoreCommandInfo struct {
	// id of the functionality that cli will run when user use this command
	ActionID string `yaml:"actionID"`
	// additional config pass to the command
	Config interface{} `yaml:"config"`
}
