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
	Location          string
	Zone              string
	MachineType       string
	CIDR              string
	DiskType          string
	CPUS              int
	DiskSizeGB        int
	NodeCount         int
	ScalerMin         int
	ScalerMax         int
	Surge             int
	Unavailable       int
	Extra             []string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
