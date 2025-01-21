package types

type ExplainCommand struct {
	// short description of the command
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
	// text that will be printed after running the `explain` command
	Output string `yaml:"output"`
}
