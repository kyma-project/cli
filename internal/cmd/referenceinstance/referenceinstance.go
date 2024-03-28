package referenceinstance

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		Short: "Add an instance reference to a shared service instance.",
		Long: `Use this command to add an instance reference to a shared service instance on the SAP Kyma platform.
`,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return config.complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runReferenceInstance(config)
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

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("offering-name")
	_ = cmd.MarkFlagRequired("reference-name")

	return cmd
}

func (pc *referenceInstanceConfig) complete() error {
	// TODO: think about timeout and moving context to persistent `kyma` command configuration
	pc.ctx = context.Background()

	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

func runReferenceInstance(config referenceInstanceConfig) error {
	requestData := fillRequestData(config)
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&requestData)
	if err != nil {
		return err
	}

	unstructuredObj := &unstructured.Unstructured{
		Object: u,
	}

	unstructuredObj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "services.cloud.sap.com",
		Version: "v1",
		Kind:    "ServiceInstance",
	})

	resourceSchema := schema.GroupVersionResource{
		Group:    "services.cloud.sap.com",
		Version:  "v1",
		Resource: "serviceinstances",
	}
	_, err = config.kubeClient.Dynamic().Resource(resourceSchema).Namespace("default").Create(config.ctx, unstructuredObj, metav1.CreateOptions{})
	return err
}

func fillRequestData(config referenceInstanceConfig) KubernetesResource {
	requestData := KubernetesResource{
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
		Spec: KubernetesResourceSpec{
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
