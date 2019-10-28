package gardener

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/kyma-project/cli/internal/kube"

	hf "github.com/kyma-incubator/hydroform"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new minikube command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "gardener",
		Short: "Provisions a Gardener cluster.",
		Long:  `Use this command to provision a Gardener cluster for Kyma installation.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the cluster to provision.")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the Gardener Project where to provision the cluster in.")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the Gardener service account kubeconfig file.")
	cmd.Flags().StringVar(&o.TargetProvider, "target-provider", "gcp", "Specify the cloud provider that Gardener should use to create the cluster.")
	cmd.Flags().StringVarP(&o.Secret, "secret", "s", "", "Name of the Gardener secret to access the target provider.")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.15.4", "Kubernetes version of the cluster to provision.")
	cmd.Flags().StringVarP(&o.Region, "region", "r", "europe-west3", "Region of the cluster to provision.")
	cmd.Flags().StringVarP(&o.Zone, "zone", "z", "europe-west3-a", "Zone of the cluster to provision.")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "n1-standard-4", "Type of machine of the cluster to provision.")
	cmd.Flags().StringVar(&o.CIDR, "cidr", "10.250.0.0/19", "Gardener CIDR of the cluster to provision.")
	cmd.Flags().StringVar(&o.DiskType, "disk-type", "pd-standard", "Type of disk to use on the target provider.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 30, "Specifies the disk size in GB of the cluster to provision.")
	cmd.Flags().IntVar(&o.NodeCount, "nodes", 3, "Specifies the number of nodes of the cluster to provision.")
	cmd.Flags().IntVar(&o.ScalerMin, "scaler-min", 2, "Specifies the minimum autoscale of the cluster to provision.")
	cmd.Flags().IntVar(&o.ScalerMax, "scaler-max", 4, "Specifies the maximum autoscale of the cluster to provision.")
	cmd.Flags().IntVar(&o.Surge, "surge", 4, "Specifies maximum surge of the cluster to provision.")
	cmd.Flags().IntVarP(&o.Unavailable, "unavailable", "u", 1, "Specifies the maximum number of unavailable nodes allowed.")
	cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "Provide one or more arguments of the form NAME=VALUE to add extra configurations.")

	return cmd
}

func (c *command) Run() error {
	if err := c.validateFlags(); err != nil {
		return err
	}

	cluster := newCluster(c.opts)
	provider, err := newProvider(c.opts)
	if err != nil {
		return err
	}

	if !c.opts.Verbose {
		// discard all the noise from terraform logs if not verbose
		log.SetOutput(ioutil.Discard)
	}
	s := c.NewStep("Provisioning Gardener cluster")
	cluster, err = hf.Provision(cluster, provider)
	if err != nil {
		s.Failure()
		return err
	}
	s.Success()

	s = c.NewStep("Saving cluster state")
	if err := files.SaveClusterState(cluster, provider); err != nil {
		s.Failure()
		return err
	}
	s.Success()

	s = c.NewStep("Importing kubeconfig")
	kubeconfig, err := hf.Credentials(cluster, provider)
	if err != nil {
		s.Failure()
		return err
	}

	if err := kube.AppendConfig(kubeconfig, c.opts.KubeconfigPath); err != nil {
		s.Failure()
		return err
	}
	s.Success()
	return nil
}

func newCluster(o *Options) *types.Cluster {
	return &types.Cluster{
		Name:              o.Name,
		KubernetesVersion: o.KubernetesVersion,
		DiskSizeGB:        o.DiskSizeGB,
		NodeCount:         o.NodeCount,
		Location:          o.Region,
		MachineType:       o.MachineType,
	}
}

func newProvider(o *Options) (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.Gardener,
		ProjectName:         o.Project,
		CredentialsFilePath: o.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	if o.Secret != "" {
		p.CustomConfigurations["target_secret"] = o.Secret
	}
	p.CustomConfigurations["target_provider"] = o.TargetProvider
	p.CustomConfigurations["zone"] = o.Zone
	p.CustomConfigurations["disk_type"] = o.DiskType
	p.CustomConfigurations["autoscaler_min"] = o.ScalerMin
	p.CustomConfigurations["autoscaler_max"] = o.ScalerMax
	p.CustomConfigurations["max_surge"] = o.Surge
	p.CustomConfigurations["max_unavailable"] = o.Unavailable
	p.CustomConfigurations["cidr"] = o.CIDR

	for _, e := range o.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, fmt.Errorf("Wrong format for extra configuration %s. Please provide NAME=VALUE pairs.", e)
		}
		p.CustomConfigurations[v[0]] = v[1]
	}
	return p, nil
}

func (c *command) validateFlags() error {
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

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}
