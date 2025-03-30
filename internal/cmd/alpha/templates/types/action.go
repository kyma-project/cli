package types

type ActionCommand struct {
	// name of the command group
	Name string `yaml:"name"`
	// short description of the command group
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
	// action details
	Action Action `yaml:"action"`
}

type Action struct {
	// id of the functionality that cli will run when user use this command
	FunctionID string `yaml:"functionID"`
	// custom flags used to build command and set values for specific fields
	CustomFlags []CustomFlag `yaml:"customFlags"`
	// additional config pass to the command
	Config interface{} `yaml:"config"`
}
