package gcp

import "github.com/kyma-project/cli/internal/cli"

//options defines available options for the minikube provisioning command
type options struct {
	*cli.Options
	CredentialsFilePath string
	Project             string
	Location            string
	NodeCount           int
	MachineType         string
	KubernetesVersion   string
}

//NewOptions creates options with default values
func NewOptions(o *cli.Options) *options {
	return &options{Options: o}
}
