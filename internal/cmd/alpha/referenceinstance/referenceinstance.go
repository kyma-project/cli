package referenceinstance

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type referenceInstanceConfig struct {
	*cmdcommon.KymaConfig

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
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "reference-instance [flags]",
		Short: "Adds an instance reference to a shared service instance",
		Long:  `Use this command to add an instance reference to a shared service instance in the Kyma cluster.`,
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkRequired("offering-name", "reference-name"),
				// either instance id or selectors can be used
				flags.MarkOneRequired("instance-id", "label-selector", "name-selector", "plan-selector"),
				flags.MarkExclusive("instance-id", "label-selector", "name-selector", "plan-selector"),
			))
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runReferenceInstance(config)
		},
	}

	cmd.Flags().StringVarP(&config.namespace, "namespace", "n", "default", "Namespace of the reference instance")
	cmd.Flags().StringVar(&config.offeringName, "offering-name", "", "Offering name")
	cmd.Flags().StringVar(&config.referenceName, "reference-name", "", "Name of the reference")
	cmd.Flags().StringVar(&config.instanceID, "instance-id", "", "ID of the instance")
	cmd.Flags().StringSliceVar(&config.labelSelector, "label-selector", nil, "Label selector for filtering instances")
	cmd.Flags().StringVar(&config.nameSelector, "name-selector", "", "Instance name selector for filtering instances")
	cmd.Flags().StringVar(&config.planSelector, "plan-selector", "", "Plan name selector for filtering instances")
	cmd.Flags().StringVar(&config.btpSecretName, "btp-secret-name", "", "name of the BTP secret containing credentials to another subaccount Service Manager:\nhttps://github.com/SAP/sap-btp-service-operator/blob/main/README.md#working-with-multiple-subaccounts")

	return cmd
}

func runReferenceInstance(config referenceInstanceConfig) error {
	requestData := fillRequestData(config)

	client, err := config.GetKubeClient()
	if err != nil {
		return err
	}

	return client.Btp().CreateServiceInstance(config.Ctx, &requestData)
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
