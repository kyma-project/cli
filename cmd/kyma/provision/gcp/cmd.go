package gcp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/spf13/cobra"

	terraformCli "github.com/kyma-project/cli/pkg/api/terraform"
	"github.com/terraform-providers/terraform-provider-google/google"
)

type command struct {
	opts *options
	cli.Command
}

//NewCmd creates a new gcp command
func NewCmd(o *options) *cobra.Command {

	c := command{
		Command: cli.Command{Options: o.Options},
		opts:    o,
	}

	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Provisions on GCP",
		Long:  `Provisions a Kubernetes cluster on Google Cloud Platform for Kyma installation`,
		RunE:  func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVar(&o.CredentialsFilePath, "credentials-file", "", "Credentials to provision a cluster on GKE")
	cmd.Flags().StringVar(&o.Project, "project", "", "GCP project name")
	cmd.Flags().StringVar(&o.Location, "location", "europe-west3-a", "GCP Region or Zone")
	cmd.Flags().IntVar(&o.NodeCount, "node-count", 3, "Cluster node count")
	cmd.Flags().StringVar(&o.MachineType, "machine-type", "n1-standard-4", "GCP node machine type")
	cmd.Flags().StringVar(&o.KubernetesVersion, "version", "1.12", "GCP Kubernetes cluster version")
	return cmd
}

//Run runs the command
func (c *command) Run() error {
	s := c.NewStep("Provisioning the cluster")
	if err := c.run(); err != nil {
		s.Failure()
		return err
	}
	s.Successf("Cluster provisioned")

	return nil
}

func (c *command) run() error {
	log.SetOutput(ioutil.Discard)
	stateFilename := "simple.tfstate"
	clusterName := "terraform-test-cluster"

	code := `
  variable "node_count"    {}
  variable "cluster_name"  {}
  variable "credentials_file_path" {}
  variable "project"       {}
  variable "location"      {}
  variable "machine_type"  {}
  variable "kubernetes_version"   {}
  provider "google" {
    credentials   = "${file("${var.credentials_file_path}")}"
	project       = "${var.project}"
  }
  resource "google_container_cluster" "gke_test_cluster" {
    name          = "${var.cluster_name}"
    location       = "${var.location}"
    initial_node_count = "${var.node_count}"
    min_master_version = "${var.kubernetes_version}"
    node_version = "${var.kubernetes_version}"
    
    node_config {
      machine_type = "${var.machine_type}"
    }

	  maintenance_policy {
      daily_maintenance_window {
        start_time = "03:00"
      }
    }
  }
`

	platform, err := terraformCli.NewPlatform(code).
		AddProvider("google", google.Provider()).
		Var("cluster_name", clusterName).
		Var("node_count", c.opts.NodeCount).
		Var("machine_type", c.opts.MachineType).
		Var("kubernetes_version", c.opts.KubernetesVersion).
		Var("credentials_file_path", c.opts.CredentialsFilePath).
		Var("project", c.opts.Project).
		Var("location", c.opts.Location).
		ReadStateFromFile(stateFilename)

	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[DEBUG] state file %s does not exists", stateFilename)
		} else {
			return fmt.Errorf("Fail to load the initial state of the platform from file %s. %s", stateFilename, err)
		}
	}

	terminate := false
	if err := platform.Apply(terminate); err != nil {
		return fmt.Errorf("Fail to apply the changes to the platform. %s", err)
	}

	if _, err := platform.WriteStateToFile(stateFilename); err != nil {
		return fmt.Errorf("Fail to save the final state of the platform to file %s. %s", stateFilename, err)
	}
	return nil
}
