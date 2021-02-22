package gke

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
		Use:   "gke",
		Short: "Provisions a Google Kubernetes Engine (GKE) cluster on Google Cloud Platform (GCP).",
		Long: `Use this command to provision a GKE cluster on GCP for Kyma installation. Use the flags to specify cluster details.
NOTE: To access the provisioned cluster, make sure you get authenticated by Google Cloud SDK. To do so,run ` + "`gcloud auth application-default login`" + ` and log in with your Google Cloud credentials.`,

		RunE: func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the GKE cluster to provision. (required)")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the GCP Project where you provision the GKE cluster. (required)")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the GCP service account key file. (required)")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.16", "Kubernetes version of the cluster.")
	cmd.Flags().StringVarP(&o.Location, "location", "l", "europe-west3-a", "Location of the cluster.")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "n1-standard-4", "Machine type used for the cluster.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 50, "Disk size (in GB) of the cluster.")
	cmd.Flags().IntVar(&o.NodeCount, "nodes", 3, "Number of cluster nodes.")
	// Temporary disabled flag. To be enabled when hydroform supports TF modules
	//cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "Provide one or more arguments of the form NAME=VALUE to add extra configurations.")
	cmd.Flags().UintVar(&o.Attempts, "attempts", 3, "Maximum number of attempts to provision the cluster.")

	return cmd
}

func (c *command) Run() error {
	s := c.NewStep("Validating flags")
	if err := c.validateFlags(); err != nil {
		s.Failure()
		return err
	}
	s.Success()

	cluster := newCluster(c.opts)
	provider, err := newProvider(c.opts)
	if err != nil {
		return err
	}

	if !c.opts.Verbose {
		// discard all the noise from terraform logs if not verbose
		log.SetOutput(ioutil.Discard)
	}
	s = c.NewStep("Provisioning GKE cluster")
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

	fmt.Printf("\nGKE cluster installed\nKubectl correctly configured: pointing to %s\n\nHappy GKE-ing! :)\n", cluster.Name)
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
		Type:                types.GCP,
		ProjectName:         o.Project,
		CredentialsFilePath: o.CredentialsFile,
	}

	p.CustomConfigurations = make(map[string]interface{})
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
