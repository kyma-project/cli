package cmdcommon

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/spf13/cobra"
)

const (
	ExtensionLabelKey           = "kyma-cli/extension"
	ExtensionCommandsLabelValue = "commands"
	ExtensionResourceLabelValue = "resource"

	ExtensionResourceInfoKey    = "resource"
	ExtensionRootCommandKey     = "rootCommand"
	ExtensionGenericCommandsKey = "templateCommands"
	ExtensionActionCommandsKey  = "actionCommands"
	ExtensionCommandsKey        = "kyma-commands-extension.yaml"
)

// map of allowed action commands in format ID: FUNC
type ActionCommandsMap map[string]func(*KymaConfig, types.ActionConfig) *cobra.Command

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

// TODO: add config validation

type CommandConfig = map[string]interface{}

type ConfigFieldType string

var (
	StringCustomFlagType ConfigFieldType = "string"
	PathCustomFlagType   ConfigFieldType = "path"
	IntCustomFlagType    ConfigFieldType = "int64"
	BoolCustomFlagType   ConfigFieldType = "bool"
	// TODO: support other types e.g. float and stringArray
)

type CommandMetadata struct {
	// name of the command group
	Name string `yaml:"name"`
	// short description of the command group
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
}

type CommandArgs struct {
	// type of the argument and config field
	// TODO: support many args by adding new type, like `stringArray`
	Type ConfigFieldType `yaml:"type"`
	// mark if args are required to run command
	Optional bool `yaml:"optional"`
	// path to the config fild that will be updated with value from args
	ConfigPath string `yaml:"configPath"`
}

type CommandFlags struct {
	// type of the flag and config field
	Type ConfigFieldType `yaml:"type"`
	// name of the flag
	Name string `yaml:"name"`
	// description of the flag
	Description string `yaml:"description"`
	// optional shorthand of the flag
	Shorthand string `yaml:"shorthand"`
	// path to the config fild that will be updated with value from the flag
	ConfigPath string `yaml:"configPath"`
	// default value for the flag
	DefaultValue interface{} `yaml:"default"`
	// mark if flag is required
	Required bool `yaml:"required"`
}

type CommandExtension struct {
	// metadata (name, descriptions) for the command
	Metadata CommandMetadata `yaml:"metadata"`
	// id of the functionality that cli will run when user use this command
	Uses string `yaml:"uses"`
	// custom flags used to build command and set values for specific fields in config
	Flags []CommandFlags `yaml:"flags"`
	// additional config pass to the command
	Config CommandConfig `yaml:"config"`
	// list of sub commands
	SubCommands []CommandExtension `yaml:"subCommands"`
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
