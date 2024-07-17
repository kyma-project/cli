package referenceinstance

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type referenceInstanceConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	namespace     string
	offeringName  string
	referenceName string
	instanceID    string
	labelSelector []string
	nameSelector  string
	planSelector  string
	btpSecretName string
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
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runReferenceInstance(config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace of the reference instance.")
	cmd.Flags().StringVar(&config.offeringName, "offering-name", "", "Offering name.")
	cmd.Flags().StringVar(&config.referenceName, "reference-name", "", "Name of the reference.")
	cmd.Flags().StringVar(&config.instanceID, "instance-id", "", "ID of the instance.")
	cmd.Flags().StringSliceVar(&config.labelSelector, "label-selector", nil, "Label selector for filtering instances.")
	cmd.Flags().StringVar(&config.nameSelector, "name-selector", "", "Instance name selector for filtering instances.")
	cmd.Flags().StringVar(&config.planSelector, "plan-selector", "", "Plan name selector for filtering instances.")
	cmd.Flags().StringVar(&config.btpSecretName, "btp-secret-name", "", "name of the BTP secret containing credentials to another subaccount Service Manager:\nhttps://github.com/SAP/sap-btp-service-operator/blob/main/README.md#working-with-multiple-subaccounts")

	// either instance id or selectors can be used
	cmd.MarkFlagsOneRequired("instance-id", "label-selector", "name-selector", "plan-selector")
	cmd.MarkFlagsMutuallyExclusive("instance-id", "label-selector")
	cmd.MarkFlagsMutuallyExclusive("instance-id", "name-selector")
	cmd.MarkFlagsMutuallyExclusive("instance-id", "plan-selector")

	_ = cmd.MarkFlagRequired("offering-name")
	_ = cmd.MarkFlagRequired("reference-name")

	return cmd
}

func runReferenceInstance(config referenceInstanceConfig) error {
	requestData := fillRequestData(config)

	return config.KubeClient.Btp().CreateServiceInstance(config.Ctx, &requestData)
}

func fillRequestData(config referenceInstanceConfig) btp.ServiceInstance {
	requestData := btp.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "services.cloud.sap.com/v1",
			Kind:       "ServiceInstance",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.referenceName,
			Namespace: config.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": config.referenceName,
			},
		},
		Spec: btp.ServiceInstanceSpec{
			Parameters: KubernetesResourceSpecParameters{
				InstanceID: config.instanceID,
				Selectors: SpecSelectors{
					InstanceLabelSelector: config.labelSelector,
					InstanceNameSelector:  config.nameSelector,
					PlanNameSelector:      config.planSelector,
				},
			},
			ServiceOfferingName:        config.offeringName,
			ServicePlanName:            "reference-instance",
			BTPAccessCredentialsSecret: config.btpSecretName,
		},
	}
	return requestData
}
