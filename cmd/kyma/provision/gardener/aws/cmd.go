package aws

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/kyma-project/cli/internal/kube"

	retry "github.com/avast/retry-go"
	hf "github.com/kyma-incubator/hydroform/provision"
	"github.com/kyma-incubator/hydroform/provision/types"
	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/cli/internal/files"
	"github.com/spf13/cobra"
)

type command struct {
	opts *Options
	cli.Command
}

//NewCmd creates a new az command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "aws",
		Short: "Provisions a Kubernetes cluster using Gardener on Amazon Web Services (AWS).",
		Long: `Use this command to provision Kubernetes clusters with Gardener on AWS for Kyma installation. 
To successfully provision a cluster on AWS, you must first create a service account to pass its details as one of the command parameters. 
Check the roles and create a service account using instructions at https://gardener.cloud/050-tutorials/content/howto/gardener_aws/.
Use service account details to create a Secret and store it in Gardener.`,

		RunE: func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the cluster to provision. (required)")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the Gardener project where you provision the cluster. (required)")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the kubeconfig file of the Gardener service account for AWS. (required)")
	cmd.Flags().StringVarP(&o.Secret, "secret", "s", "", "Name of the Gardener secret used to access AWS. (required)")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.16", "Kubernetes version of the cluster.")
	cmd.Flags().StringVarP(&o.Region, "region", "r", "eu-west-3", "Region of the cluster.")
	cmd.Flags().StringSliceVarP(&o.Zones, "zones", "z", []string{"eu-west-3a"}, "Zones specify availability zones that are used to evenly distribute the worker pool. eg. --zones=\"europe-west3-a,europe-west3-b\"")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "m5.xlarge", "Machine type used for the cluster.")
	cmd.Flags().StringVar(&o.DiskType, "disk-type", "gp2", "Type of disk to use on AWS.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 50, "Disk size (in GB) of the cluster.")
	cmd.Flags().IntVar(&o.ScalerMin, "scaler-min", 2, "Minimum autoscale value of the cluster.")
	cmd.Flags().IntVar(&o.ScalerMax, "scaler-max", 3, "Maximum autoscale value of the cluster.")
	cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "One or more arguments provided as the `NAME=VALUE` key-value pairs to configure additional cluster settings. You can use this flag multiple times or enter the key-value pairs as a comma-separated list.")
	cmd.Flags().UintVar(&o.Attempts, "attempts", 3, "Maximum number of attempts to provision the cluster.")

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
	s := c.NewStep("Provisioning Gardener cluster on AWS")

	home, err := files.KymaHome()
	if err != nil {
		s.Failure()
		return err
	}

	err = retry.Do(
		func() error {
			cluster, err = hf.Provision(cluster, provider, types.WithDataDir(home), types.Persistent())
			return err
		},
		retry.Attempts(c.opts.Attempts), retry.LastErrorOnly(!c.opts.Verbose))

	if err != nil {
		s.Failure()
		return err
	}
	s.Success()

	s = c.NewStep("Importing kubeconfig")
	kubeconfig, err := hf.Credentials(cluster, provider, types.WithDataDir(home), types.Persistent())
	if err != nil {
		s.Failure()
		return err
	}

	if err := kube.AppendConfig(kubeconfig, c.opts.KubeconfigPath); err != nil {
		s.Failure()
		return err
	}
	s.Success()

	fmt.Printf("\nGardener cluster installed\nKubectl correctly configured: pointing to %s\n\nHappy Garden-ing! :)\n", cluster.Name)
	return nil
}

func newCluster(o *Options) *types.Cluster {
	return &types.Cluster{
		Name:              o.Name,
		KubernetesVersion: o.KubernetesVersion,
		DiskSizeGB:        o.DiskSizeGB,
		NodeCount:         o.ScalerMax,
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

	p.CustomConfigurations["target_provider"] = "aws"
	p.CustomConfigurations["disk_type"] = o.DiskType
	p.CustomConfigurations["worker_minimum"] = o.ScalerMin
	p.CustomConfigurations["worker_maximum"] = o.ScalerMax
	p.CustomConfigurations["worker_max_surge"] = 1
	p.CustomConfigurations["worker_max_unavailable"] = 1
	p.CustomConfigurations["vnetcidr"] = "10.250.0.0/16"
	p.CustomConfigurations["workercidr"] = "10.250.0.0/16"
	p.CustomConfigurations["networking_type"] = "calico"
	p.CustomConfigurations["machine_image_name"] = "gardenlinux"
	p.CustomConfigurations["machine_image_version"] = "27.1.0"
	p.CustomConfigurations["zones"] = o.Zones

	for _, e := range o.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, fmt.Errorf("wrong format for extra configuration %s. Please provide NAME=VALUE pairs", e)
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
	if c.opts.ScalerMin < 1 {
		errMessage.WriteString("\n Minimum node count should be at least 1 node.")
	}
	if c.opts.ScalerMin > c.opts.ScalerMax {
		errMessage.WriteString("\n Minimum node count cannot be greater than maximum number nodes.")
	}

	for _, zone := range c.opts.Zones {
		if !strings.HasPrefix(zone, c.opts.Region) {
			errMessage.WriteString(fmt.Sprintf("\n Provided zone %s and region %s do not match. Please provide the right region for the zone.", zone, c.opts.Region))
		}
	}

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}
