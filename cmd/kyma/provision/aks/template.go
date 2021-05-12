package aks

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-incubator/hydroform/provision/types"
	prov "github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/internal/cli"
)

func newAksCmd(o *Options) *aksCmd {
	return &aksCmd{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}
}

type aksCmd struct {
	opts *Options
	cli.Command
}

func (c *aksCmd) Run() error {
	return prov.RunTemplate(c)
}

func (c *aksCmd) NewCluster() *types.Cluster {
	return &types.Cluster{
		Name:              c.opts.Name,
		KubernetesVersion: c.opts.KubernetesVersion,
		DiskSizeGB:        c.opts.DiskSizeGB,
		NodeCount:         c.opts.NodeCount,
		Location:          c.opts.Location,
		MachineType:       c.opts.MachineType,
	}
}

func (c *aksCmd) NewProvider() (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.Azure,
		ProjectName:         c.opts.Project,
		CredentialsFilePath: c.opts.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	for _, e := range c.opts.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, fmt.Errorf("wrong format for extra configuration %s, please provide NAME=VALUE pairs", e)
		}
		p.CustomConfigurations[v[0]] = v[1]
	}
	return p, nil
}

func (c *aksCmd) ProviderName() string { return "AKS" }

func (c *aksCmd) Attempts() uint { return c.opts.Attempts }

func (c *aksCmd) KubeconfigPath() string { return c.opts.KubeconfigPath }

func (c *aksCmd) ValidateFlags() error {
	var errMessage strings.Builder
	// mandatory flags]
	if c.opts.Name == "" {
		errMessage.WriteString("\nRequired flag `name` has not been set.")
	}
	if c.opts.Project == "" {
		errMessage.WriteString("\nRequired flag `project` has not been set.")
	}
	if c.opts.CredentialsFile == "" {
		errMessage.WriteString("\nRequired flag `credentials` has not been set.")
	}

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}

func (c *aksCmd) IsVerbose() bool { return c.opts.Verbose }
