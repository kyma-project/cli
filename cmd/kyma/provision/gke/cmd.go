package gke

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new minikube command
func NewCmd(o *Options) *cobra.Command {
	c := newGkeCmd(o)

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
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.19", "Kubernetes version of the cluster.")
	cmd.Flags().StringVarP(&o.Location, "location", "l", "europe-west3-a", "Location of the cluster.")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "n1-standard-4", "Machine type used for the cluster.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 50, "Disk size (in GB) of the cluster.")
	cmd.Flags().IntVar(&o.NodeCount, "nodes", 3, "Number of cluster nodes.")
	// Temporary disabled flag. To be enabled when hydroform supports TF modules
	//cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "Provide one or more arguments of the form NAME=VALUE to add extra configurations.")
	cmd.Flags().UintVar(&o.Attempts, "attempts", 3, "Maximum number of attempts to provision the cluster.")

	return cmd
}
