package extension

const (
	extensionLabelKey           = "kyma-cli/extension"
	extensionResourceLabelValue = "resource"

	extensionRootCommandKey     = "rootCommand"
	extensionResourceInfoKey    = "managedResource"
	extensionGenericCommandsKey = "genericCommands"
)

type Extension struct {
	// main command of the command group
	RootCommand RootCommand
	// details about managed resource passed to every sub-command
	ManagedResource *ResourceInfo
	// configuration of generic commands (like 'create', 'delete', 'get', ...) which implementation is provided by the cli
	// most of these commands bases on the `ManagedResource` field
	GenericCommands *GenericCommands
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
	// test that will be printed after running the `explain` command
	ExplainOutput string `yaml:"explainText"`
}

type GenericCommands struct {
	// allows to explaining command to the commands group in format:
	// kyma <root_command> explain
	ExplainCommand *ExplainCommand `yaml:"explain"`
}
