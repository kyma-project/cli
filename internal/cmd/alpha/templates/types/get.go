package types

type GetCommand struct {
	// short description of the command
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
	// list of additional fields to keep in output
	Parameters []Parameter `yaml:"parameters"`
}

type Parameter struct {
	// path to the field
	Path string `yaml:"path"`
	// name of the output field
	Name string `yaml:"name"`
	// type of the output value
	// Type string `yaml:"type"`
}
