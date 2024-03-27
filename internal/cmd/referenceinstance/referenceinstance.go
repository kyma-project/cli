package referenceinstance

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
)

type referenceInstanceConfig struct {
	ctx        context.Context
	kubeClient kube.Client

	kubeconfig    string
	offeringName  string
	referenceName string
	instanceID    string
	labelSelector []string
	nameSelector  string
	planSelector  string
}

func NewReferenceInstanceCMD() *cobra.Command {
	config := referenceInstanceConfig{}

	cmd := &cobra.Command{
		Use:   "reference-instance",
		Short: "Manage reference instances.",
		Long: `Use this command to manage reference instances on the SAP BTP platform.
`,
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	cmd.Flags().StringVar(&config.kubeconfig, "kubeconfig", "", "Path to the Kyma kubecongig file.")
	cmd.Flags().StringVar(&config.offeringName, "offering-name", "", "Offering name.")
	cmd.Flags().StringVar(&config.referenceName, "reference-name", "", "Name of the reference.")
	cmd.Flags().StringVar(&config.instanceID, "instance-id", "", "ID of the instance.")
	cmd.Flags().StringSliceVar(&config.labelSelector, "label-selector", nil, "Label selector for filtering instances.")
	cmd.Flags().StringVar(&config.nameSelector, "name-selector", "", "Instance name selector for filtering instances.")
	cmd.Flags().StringVar(&config.planSelector, "plan-selector", "", "Plan name selector for filtering instances.")

	// either instance id or selectors can be used
	cmd.MarkFlagsOneRequired("instance-id", "label-selector", "name-selector", "plan-selector")
	cmd.MarkFlagsMutuallyExclusive("instance-id", "label-selector")
	cmd.MarkFlagsMutuallyExclusive("instance-id", "name-selector")
	cmd.MarkFlagsMutuallyExclusive("instance-id", "plan-selector")

	cmd.MarkFlagRequired("offering-name")
	cmd.MarkFlagRequired("reference-name")

	return cmd
}

func (pc *referenceInstanceConfig) complete() error {
	// TODO: think about timeout and moving context to persistent `kyma` command configuration
	pc.ctx = context.Background()

	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}
