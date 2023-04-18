package az

import "github.com/kyma-project/cli/internal/cli"

type Options struct {
	*cli.Options

	Name                string
	Project             string
	CredentialsFile     string
	Secret              string
	KubernetesVersion   string
	GardenLinuxVersion  string
	Region              string
	Zones               []string
	MachineType         string
	DiskType            string
	DiskSizeGB          int
	ScalerMin           int
	ScalerMax           int
	Extra               []string
	Attempts            uint
	HibernationStart    string
	HibernationEnd      string
	HibernationLocation string
}

// NewOptions creates options with default values
func NewOptions(o *cli.Options) *Options {
	return &Options{Options: o}
}
