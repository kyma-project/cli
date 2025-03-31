package types

type CreateCommand struct {
	// short description of the command
	Description string `yaml:"description"`
	// long description of the command group
	DescriptionLong string `yaml:"descriptionLong"`
	// custom flags used to build command and set values for specific fields
	CustomFlags []CustomFlag `yaml:"customFlags"`
}
