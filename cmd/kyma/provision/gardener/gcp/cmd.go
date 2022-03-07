package gcp

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new az command
func NewCmd(o *Options) *cobra.Command {
	c := newGcpCmd(o)

	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "Provisions a Kubernetes cluster using Gardener on Google Cloud Platform (GCP).",
		Long: `Use this command to provision Kubernetes clusters with Gardener on GCP for Kyma installation. 
To successfully provision a cluster on GCP, you must first create a service account to pass its details as one of the command parameters. 
Check the roles and create a service account using instructions at https://gardener.cloud/050-tutorials/content/howto/gardener_gcp/.
Use service account details to create a Secret and import it in Gardener.`,

		RunE: func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the cluster to provision. (required)")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the Gardener project where you provision the cluster. (required)")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the kubeconfig file of the Gardener service account for GCP. (required)")
	cmd.Flags().StringVarP(&o.Secret, "secret", "s", "", "Name of the Gardener secret used to access GCP. (required)")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.21", "Kubernetes version of the cluster.")
	cmd.Flags().StringVarP(&o.Region, "region", "r", "europe-west3", "Region of the cluster.")
	cmd.Flags().StringSliceVarP(&o.Zones, "zones", "z", []string{"europe-west3-a"}, "Zones specify availability zones that are used to evenly distribute the worker pool. eg. --zones=\"europe-west3-a,europe-west3-b\"")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "n1-standard-4", "Machine type used for the cluster.")
	cmd.Flags().StringVar(&o.DiskType, "disk-type", "pd-standard", "Type of disk to use on GCP.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 50, "Disk size (in GB) of the cluster.")
	cmd.Flags().IntVar(&o.ScalerMin, "scaler-min", 2, "Minimum autoscale value of the cluster.")
	cmd.Flags().IntVar(&o.ScalerMax, "scaler-max", 3, "Maximum autoscale value of the cluster.")
	cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "One or more arguments provided as the `NAME=VALUE` key-value pairs to configure additional cluster settings. You can use this flag multiple times or enter the key-value pairs as a comma-separated list.")
	cmd.Flags().UintVar(&o.Attempts, "attempts", 3, "Maximum number of attempts to provision the cluster.")

	return cmd
}
