package az

import (
	"errors"
	"fmt"
	"strings"

	prov "github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/hydroform/provision/types"
)

func newAzCmd(o *Options) *azCmd {
	return &azCmd{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}
}

type azCmd struct {
	opts *Options
	cli.Command
}

func (c *azCmd) Run() error {
	return prov.RunTemplate(c)
}

func (c *azCmd) NewCluster() *types.Cluster {
	return &types.Cluster{
		Name:              c.opts.Name,
		KubernetesVersion: c.opts.KubernetesVersion,
		DiskSizeGB:        c.opts.DiskSizeGB,
		NodeCount:         c.opts.ScalerMax,
		Location:          c.opts.Region,
		MachineType:       c.opts.MachineType,
	}
}

func (c *azCmd) NewProvider() (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         c.opts.Project,
		CredentialsFilePath: c.opts.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	if c.opts.Secret != "" {
		p.CustomConfigurations["target_secret"] = c.opts.Secret
	}

	p.CustomConfigurations["target_provider"] = "azure"
	p.CustomConfigurations["disk_type"] = c.opts.DiskType
	p.CustomConfigurations["worker_minimum"] = c.opts.ScalerMin
	p.CustomConfigurations["worker_maximum"] = c.opts.ScalerMax
	p.CustomConfigurations["worker_max_surge"] = 1
	p.CustomConfigurations["worker_max_unavailable"] = 1
	p.CustomConfigurations["vnetcidr"] = "10.250.0.0/16"
	p.CustomConfigurations["workercidr"] = "10.250.0.0/16"
	p.CustomConfigurations["networking_type"] = "calico"
	p.CustomConfigurations["machine_image_name"] = "gardenlinux"
	p.CustomConfigurations["machine_image_version"] = c.opts.GardenLinuxVersion
	p.CustomConfigurations["zones"] = c.opts.Zones
	p.CustomConfigurations["hibernation_start"] = c.opts.HibernationStart
	p.CustomConfigurations["hibernation_end"] = c.opts.HibernationEnd
	p.CustomConfigurations["hibernation_location"] = c.opts.HibernationLocation

	for _, e := range c.opts.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, fmt.Errorf("wrong format for extra configuration %s. Please provide NAME=VALUE pairs", e)
		}

		p.CustomConfigurations[v[0]] = v[1]
	}
	return p, nil
}

func (c *azCmd) ProviderName() string { return "Gardener(Azure)" }

func (c *azCmd) Attempts() uint { return c.opts.Attempts }

func (c *azCmd) KubeconfigPath() string { return c.opts.KubeconfigPath }

func (c *azCmd) ValidateFlags() error {
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
	if c.opts.Secret == "" {
		errMessage.WriteString("\nRequired flag `secret` has not been set.")
	}
	if c.opts.ScalerMin < 1 {
		errMessage.WriteString("\n Minimum node count should be at least 1 node.")
	}
	if c.opts.ScalerMin > c.opts.ScalerMax {
		errMessage.WriteString("\n Minimum node count cannot be greater than maximum number nodes.")
	}

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}

func (c *azCmd) IsVerbose() bool { return c.opts.Verbose }

func (c *azCmd) FilterErr(e error) error {
	if e != nil && strings.Contains(e.Error(), "already exists") {
		return nil
	}

	return e
}
