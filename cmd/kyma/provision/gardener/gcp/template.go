package gcp

import (
	"errors"
	"fmt"
	"strings"

	prov "github.com/kyma-project/cli/cmd/kyma/provision"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/hydroform/provision/types"
)

func newGcpCmd(o *Options) *gcpCmd {
	return &gcpCmd{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}
}

type gcpCmd struct {
	opts *Options
	cli.Command
}

func (c *gcpCmd) Run() error {
	return prov.RunTemplate(c)
}

func (c *gcpCmd) NewCluster() *types.Cluster {
	return &types.Cluster{
		Name:              c.opts.Name,
		KubernetesVersion: c.opts.KubernetesVersion,
		DiskSizeGB:        c.opts.DiskSizeGB,
		NodeCount:         c.opts.ScalerMax,
		Location:          c.opts.Region,
		MachineType:       c.opts.MachineType,
	}
}

func (c *gcpCmd) NewProvider() (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         c.opts.Project,
		CredentialsFilePath: c.opts.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	if c.opts.Secret != "" {
		p.CustomConfigurations["target_secret"] = c.opts.Secret
	}

	p.CustomConfigurations["gcp_control_plane_zone"] = c.opts.Zones[0]
	p.CustomConfigurations["target_provider"] = "gcp"
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

func (c *gcpCmd) ProviderName() string { return "Gardener(GCP)" }

func (c *gcpCmd) Attempts() uint { return c.opts.Attempts }

func (c *gcpCmd) KubeconfigPath() string { return c.opts.KubeconfigPath }

func (c *gcpCmd) ValidateFlags() error {
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

	for _, zone := range c.opts.Zones {
		if !strings.HasPrefix(zone, c.opts.Region) {
			errMessage.WriteString(
				fmt.Sprintf(
					"\n Provided zone %s and region %s do not match. Please provide the right region for the zone.",
					zone, c.opts.Region,
				),
			)
		}
	}

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}

func (c *gcpCmd) IsVerbose() bool { return c.opts.Verbose }

func (c *gcpCmd) FilterErr(e error) error {
	if e != nil && strings.Contains(e.Error(), "already exists") {
		return nil
	}

	return e
}
