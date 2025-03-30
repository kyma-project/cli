package types

type CreateCustomFlagType string

var (
	StringCustomFlagType CreateCustomFlagType = "string"
	PathCustomFlagType   CreateCustomFlagType = "path"
	IntCustomFlagType    CreateCustomFlagType = "int64"
	// TODO: support other types e.g. bool, float and stringArray
)

type CustomFlag struct {
	// type of the custom flag
	Type CreateCustomFlagType `yaml:"type"`
	// name of the custom flag
	Name string `yaml:"name"`
	// description of the custom flag
	Description string `yaml:"description"`
	// optional shorthand of the custom flag
	Shorthand string `yaml:"shorthand"`
	// path to the resource fild that will be updated with value from the flag
	Path string `yaml:"path"`
	// default value for the flag
	DefaultValue interface{} `yaml:"default"`
	// mark if flag is required
	Required bool `yaml:"required"`
}
