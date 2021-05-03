package aks

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/avast/retry-go"
	"github.com/kyma-project/cli/internal/kube"

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

//NewCmd creates a new minikube command
func NewCmd(o *Options) *cobra.Command {
	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "aks",
		Short: "Provisions an Azure Kubernetes Service (AKS) cluster on Azure.",
		Long: `Use this command to provision an AKS cluster on Azure for Kyma installation. Use the flags to specify cluster details. 
	NOTE: To provision and access the provisioned cluster, make sure you get authenticated by using the Azure CLI. To do so,run ` + "`az login`" + ` and log in with your Azure credentials.`,

		RunE: func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the AKS cluster to provision. (required)")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the Azure Resource Group where you provision the AKS cluster. (required)")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the TOML file containing the Azure Subscription ID (SUBSCRIPTION_ID), Tenant ID (TENANT_ID), Client ID (CLIENT_ID) and Client Secret (CLIENT_SECRET). (required)")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.19.7", "Kubernetes version of the cluster.")
	cmd.Flags().StringVarP(&o.Location, "location", "l", "westeurope", "Location of the cluster.")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "Standard_D4_v3", "Machine type used for the cluster.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 50, "Disk size (in GB) of the cluster.")
	cmd.Flags().IntVar(&o.NodeCount, "nodes", 3, "Number of cluster nodes.")
	// Temporary disabled flag. To be enabled when hydroform supports TF modules
	//cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "Provide one or more arguments of the form NAME=VALUE to add extra configurations.")
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
	s := c.NewStep("Provisioning AKS cluster")
	home, err := files.KymaHome()
	if err != nil {
		s.Failure()
		return err
	}

	err = retry.Do(
		func() error {
			cluster, err = hf.Provision(cluster, provider, types.WithDataDir(home), types.Persistent(), types.Verbose(c.opts.Verbose))
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

	fmt.Printf("\nAKS cluster installed\nKubectl correctly configured: pointing to %s\n\nHappy AKS-ing! :)\n", cluster.Name)
	return nil
}

func newCluster(o *Options) *types.Cluster {
	return &types.Cluster{
		Name:              o.Name,
		KubernetesVersion: o.KubernetesVersion,
		DiskSizeGB:        o.DiskSizeGB,
		NodeCount:         o.NodeCount,
		Location:          o.Location,
		MachineType:       o.MachineType,
	}
}

func newProvider(o *Options) (*types.Provider, error) {
	p := &types.Provider{
		Type:                types.Azure,
		ProjectName:         o.Project,
		CredentialsFilePath: o.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
	for _, e := range o.Extra {
		v := strings.Split(e, "=")

		if len(v) != 2 {
			return p, fmt.Errorf("wrong format for extra configuration %s, please provide NAME=VALUE pairs", e)
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

	if errMessage.Len() != 0 {
		return errors.New(errMessage.String())
	}
	return nil
}
