package cmdcommon

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
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
}

type ExtensionList []Extension

type Extension struct {
	// main command of the command group
	RootCommand RootCommand
	// details about managed resource passed to every sub-command
	Resource *ResourceInfo
	// configuration of generic commands (like 'create', 'delete', 'get', ...) which implementation is provided by the cli
	// most of these commands bases on the `Resource` field
	TemplateCommands *TemplateCommands
	// configuration of buildin commands (like 'registry config') which implementation is provided by cli
	// use this command to enable feature for a module
	CoreCommands []CoreCommandInfo
}

type Scope string

const (
	ClusterScope    Scope = "cluster"
	NamespacedScope Scope = "namespace"
)

type RootCommand struct {
	// name of the command group
	Name string `yaml:"name"`
	// short description of the command group
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
}

type ResourceInfo struct {
	Scope    Scope  `yaml:"scope"`
	Kind     string `yaml:"kind"`
	Group    string `yaml:"group"`
	Version  string `yaml:"version"`
	Singular string `yaml:"singular"`
	Plural   string `yaml:"plural"`
}

type ExplainCommand struct {
	// short description of the command
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
	// text that will be printed after running the `explain` command
	Output string `yaml:"output"`
}

type TemplateCommands struct {
	// allows to explaining command to the commands group in format:
	// kyma <root_command> explain
	ExplainCommand *ExplainCommand `yaml:"explain"`
}

type CoreCommandInfo struct {
	// id of the functionality that cli will run when user use this command
	ActionID string `yaml:"actionID"`
}
