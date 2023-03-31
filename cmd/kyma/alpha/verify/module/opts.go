package module

import (
	"github.com/kyma-project/cli/internal/cli"
)

// Options defines available options for the version command
type Options struct {
	*cli.Options

	Name, Version   string
	RegistryURL     string
	NameMappingMode string
	PublicKeyPath   string
	Credentials     string
	Token           string
	Insecure        bool
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
