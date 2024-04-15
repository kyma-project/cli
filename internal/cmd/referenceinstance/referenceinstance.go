package referenceinstance

import (
	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type referenceInstanceConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	offeringName  string
	referenceName string
	instanceID    string
	labelSelector []string
	nameSelector  string
	planSelector  string
}

func NewReferenceInstanceCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := referenceInstanceConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "reference-instance",
		Short: "Add an instance reference to a shared service instance.",
		Long: `Use this command to add an instance reference to a shared service instance on the SAP Kyma platform.
`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return config.KubeClientConfig.Complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runReferenceInstance(config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

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

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("offering-name")
	_ = cmd.MarkFlagRequired("reference-name")

	return cmd
}

func runReferenceInstance(config referenceInstanceConfig) error {
	requestData := fillRequestData(config)
	unstructuredObj, err := kube.ToUnstructured(requestData, operator.GVKServiceInstance)
	if err != nil {
		return err
	}

	_, err = config.KubeClient.Dynamic().Resource(operator.GVRServiceInstance).Namespace("default").Create(config.Ctx, unstructuredObj, metav1.CreateOptions{})
	return err
}

func fillRequestData(config referenceInstanceConfig) kube.ServiceInstance {
	requestData := kube.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "services.cloud.sap.com/v1",
			Kind:       "ServiceInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.referenceName,
			Namespace: "default",
			Labels: map[string]string{
				"app.kubernetes.io/name": config.referenceName,
			},
		},
		Spec: kube.ServiceInstanceSpec{
			Parameters: KubernetesResourceSpecParameters{
				InstanceID: config.instanceID,
				Selectors: SpecSelectors{
					InstanceLabelSelector: config.labelSelector,
					InstanceNameSelector:  config.nameSelector,
					PlanNameSelector:      config.planSelector,
				},
			},
			OfferingName: config.offeringName,
			PlanName:     "reference-instance",
		},
	}
	return requestData
}
