package gke

import "github.com/kyma-project/cli/internal/cli"

type Options struct {
	*cli.Options

	Name              string
	Project           string
	CredentialsFile   string
	KubernetesVersion string
	Location          string
	MachineType       string
	DiskSizeGB        int
	NodeCount         int
	Extra             []string
	Attempts          uint
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
