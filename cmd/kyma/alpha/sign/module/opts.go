package module

import (
	"github.com/kyma-project/cli/internal/cli"
)

type Options struct {
	*cli.Options

	Name, Version   string
	RegistryURL     string
	NameMappingMode string
	PrivateKeyPath  string
	Credentials     string
	Token           string
	Insecure        bool
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
