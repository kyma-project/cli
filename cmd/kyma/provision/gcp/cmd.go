package gcp

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
		Use:   "gcp",
		Short: "Provisions a GCP cluster.",
		Long:  `Use this command to provision GCP for Kyma installation.`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the cluster to provision.")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the GCP Project where to provision the cluster in.")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the GCP service account key file.")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.14.6", "Kubernetes version of the cluster to provision.")
	cmd.Flags().StringVarP(&o.Location, "location", "l", "europe-west3-a", "Location of the cluster to provision.")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "n1-standard-4", "Type of machine of the cluster to provision.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 30, "Specifies the disk size in GB of the cluster to provision.")
	cmd.Flags().IntVar(&o.CPUS, "cpus", 4, "Specifies the number of CPUs of the cluster to provision.")
	cmd.Flags().IntVar(&o.NodeCount, "nodes", 1, "Specifies the number of nodes of the cluster to provision.")
	cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "Provide one or more arguments of the form NAME=VALUE to add extra configurations.")

	return cmd
}

func (c *command) Run() error {
	cluster := newCluster(c.opts)
	provider, err := newProvider(c.opts)
	if err != nil {
		return err
	}

	if !c.opts.Verbose {
		// discard all the noise from terraform logs if not verbose
		log.SetOutput(ioutil.Discard)
	}
	s := c.NewStep("Provisioning GCP cluster")
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
		CPU:               o.CPUS,
		DiskSizeGB:        o.DiskSizeGB,
		NodeCount:         o.NodeCount,
		Location:          o.Location,
		MachineType:       o.MachineType,
	}
}

func newProvider(o *Options) (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.GCP,
		ProjectName:         o.Project,
		CredentialsFilePath: o.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	for _, e := range o.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, errors.New(fmt.Sprintf("Wrong format for extra configuration %s. Please provide NAME=VALUE pairs.", e))
		}
		p.CustomConfigurations[v[0]] = v[1]
	}
	return p, nil
}
