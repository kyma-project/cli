package module

import "github.com/kyma-project/cli/internal/cli"

type Options struct {
	*cli.Options

	RegistryURL       string
	PrivateKeyPath    string
	SignatureName     string
	Credentials       string
	Token             string
	Insecure          bool
	ModPath           string
	SignedRegistryURL string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
