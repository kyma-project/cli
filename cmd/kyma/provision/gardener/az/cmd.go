package az

import (
	"github.com/spf13/cobra"
)

//NewCmd creates a new az command
func NewCmd(o *Options) *cobra.Command {
	c := newAzCmd(o)

	cmd := &cobra.Command{
		Use:   "az",
		Short: "Provisions a Kubernetes cluster using Gardener on Azure.",
		Long: `Use this command to provision Kubernetes clusters with Gardener on Azure for Kyma installation. 
To successfully provision a cluster on Azure, you must first create a service account to pass its details as one of the command parameters. 
Create a service account with the ` + "`contributor`" + ` role. Use service account details to create a Secret and import it in Gardener.`,

		RunE: func(_ *cobra.Command, _ []string) error { return c.Run() },
	}

	cmd.Flags().StringVarP(&o.Name, "name", "n", "", "Name of the cluster to provision. (required)")
	cmd.Flags().StringVarP(&o.Project, "project", "p", "", "Name of the Gardener project where you provision the cluster. (required)")
	cmd.Flags().StringVarP(&o.CredentialsFile, "credentials", "c", "", "Path to the kubeconfig file of the Gardener service account for Azure. (required)")
	cmd.Flags().StringVarP(&o.Secret, "secret", "s", "", "Name of the Gardener secret used to access Azure. (required)")
	cmd.Flags().StringVarP(&o.KubernetesVersion, "kube-version", "k", "1.20", "Kubernetes version of the cluster.")
	cmd.Flags().StringVarP(&o.Region, "region", "r", "westeurope", "Region of the cluster.")
	cmd.Flags().StringSliceVarP(&o.Zones, "zones", "z", []string{"1"}, "Zones specify availability zones that are used to evenly distribute the worker pool. eg. --zones=\"europe-west3-a,europe-west3-b\"")
	cmd.Flags().StringVarP(&o.MachineType, "type", "t", "Standard_D4_v3", "Machine type used for the cluster.")
	cmd.Flags().StringVar(&o.DiskType, "disk-type", "Standard_LRS", "Type of disk to use on Azure.")
	cmd.Flags().IntVar(&o.DiskSizeGB, "disk-size", 50, "Disk size (in GB) of the cluster.")
	cmd.Flags().IntVar(&o.ScalerMin, "scaler-min", 2, "Minimum autoscale value of the cluster.")
	cmd.Flags().IntVar(&o.ScalerMax, "scaler-max", 3, "Maximum autoscale value of the cluster.")
	cmd.Flags().StringSliceVarP(&o.Extra, "extra", "e", nil, "One or more arguments provided as the `NAME=VALUE` key-value pairs to configure additional cluster settings. You can use this flag multiple times or enter the key-value pairs as a comma-separated list.")
	cmd.Flags().UintVar(&o.Attempts, "attempts", 3, "Maximum number of attempts to provision the cluster.")

	return cmd
}
