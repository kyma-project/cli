package gke

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kyma-incubator/hydroform/provision/types"
	prov "github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/internal/cli"
)

func newGkeCmd(o *Options) *gkeCmd {
	return &gkeCmd{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}
}

type gkeCmd struct {
	opts *Options
	cli.Command
}

func (c *gkeCmd) Run() error {
	return prov.RunTemplate(c)
}

func (c *gkeCmd) NewCluster() *types.Cluster {
	return &types.Cluster{
		Name:              c.opts.Name,
		KubernetesVersion: c.opts.KubernetesVersion,
		DiskSizeGB:        c.opts.DiskSizeGB,
		NodeCount:         c.opts.NodeCount,
		Location:          c.opts.Location,
		MachineType:       c.opts.MachineType,
	}
}

func (c *gkeCmd) NewProvider() (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.GCP,
		ProjectName:         c.opts.Project,
		CredentialsFilePath: c.opts.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	for _, e := range c.opts.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, fmt.Errorf("wrong format for extra configuration %s. Please provide NAME=VALUE pairs", e)
		}
		p.CustomConfigurations[v[0]] = v[1]
	}
	return p, nil
}

func (c *gkeCmd) ProviderName() string { return "GKE" }

func (c *gkeCmd) Attempts() uint { return c.opts.Attempts }

func (c *gkeCmd) KubeconfigPath() string { return c.opts.KubeconfigPath }

func (c *gkeCmd) ValidateFlags() error {
	var errMessage strings.Builder
	// mandatory flags
	if c.opts.Name == "" {
		errMessage.WriteString("\nRequired flag `name` has not been set.")
	}
	if c.opts.Project == "" {
		errMessage.WriteString("\nRequired flag `project` has not been set.")
	}
	if c.opts.CredentialsFile == "" {
		errMessage.WriteString("\nRequired flag `credentials` has not been set.")
	}

	if len(strings.Split(c.opts.Location, "-")) <= 2 {
		if !(c.opts.NonInteractive || c.opts.CI) {
			answer := c.CurrentStep.PromptYesNo(fmt.Sprintf("Since you chose a region (%s) instead of a zone, %d number of nodes will be created on each zone in this region.\n"+
				"You can also provide a different number of nodes or specify a zone instead.\n"+
				"Are you sure you want to continue? ", c.opts.Location, c.opts.NodeCount))

			if !answer {
				return fmt.Errorf("Aborting provisioning")
			}
		}
	}

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}

func (c *gkeCmd) IsVerbose() bool { return c.opts.Verbose }
