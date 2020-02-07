package gardener

import "github.com/kyma-project/cli/internal/cli"

type Options struct {
	*cli.Options

	Name              string
	Project           string
	CredentialsFile   string
	TargetProvider    string
	Secret            string
	KubernetesVersion string
	Region            string
	Zone              string
	MachineType       string
	CIDR              string
	WCIDR             string
	DiskType          string
	DiskSizeGB        int
	NodeCount         int
	ScalerMin         int
	ScalerMax         int
	Surge             int
	Unavailable       int
	Extra             []string
	NetworkType       string
	NetworkNodes      string
	NetworkPods       string
	NetworkServices   string
	MachineImageName string
	MachineImageVersion string
	ServiceEndpoints []string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
