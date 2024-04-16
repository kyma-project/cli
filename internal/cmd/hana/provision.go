package hana

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type hanaProvisionConfig struct {
	*cmdcommon.KymaConfig
	kubeClient kube.Client

	kubeconfig  string
	name        string
	namespace   string
	planName    string
	memory      int
	cpu         int
	whitelistIP []string
}

func NewHanaProvisionCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaProvisionConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a Hana instance on the Kyma.",
		Long:  "Use this command to provision a Hana instance on the SAP Kyma platform.",
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runProvision(&config)
		},
	}

	cmd.Flags().StringVar(&config.kubeconfig, "kubeconfig", "", "Path to the Kyma kubecongig file.")

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")
	cmd.Flags().StringVar(&config.planName, "plan", "hana", "Name of the service plan.")
	cmd.Flags().IntVar(&config.memory, "memory", 32, "Memory size for Hana.")                                        //TODO: fulfill proper usage
	cmd.Flags().IntVar(&config.cpu, "cpu", 2, "Number of CPUs for Hana.")                                            //TODO: fulfill proper usage
	cmd.Flags().StringSliceVar(&config.whitelistIP, "whitelist-ip", []string{"0.0.0.0/0"}, "IP whitelist for Hana.") //TODO: fulfill proper usage

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (pc *hanaProvisionConfig) complete() error {
	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

var (
	provisionCommands = []func(*hanaProvisionConfig) error{
		createHanaInstance,
		createHanaBinding,
		createHanaBindingUrl,
	}
)

func runProvision(config *hanaProvisionConfig) error {
	fmt.Printf("Provisioning Hana (%s/%s).\n", config.namespace, config.name)

	for _, command := range provisionCommands {
		err := command(config)
		if err != nil {
			return err
		}
	}
	fmt.Println("Operation completed.")
	return nil
}

func createHanaInstance(config *hanaProvisionConfig) error {
	_, err := config.kubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Create(config.Ctx, hanaInstance(config), metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana instance", config.namespace, config.name)
}

func createHanaBinding(config *hanaProvisionConfig) error {
	_, err := config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.Ctx, hanaBinding(config), metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana binding", config.namespace, config.name)
}

func createHanaBindingUrl(config *hanaProvisionConfig) error {
	_, err := config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.Ctx, hanaBindingUrl(config), metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana URL binding", config.namespace, hanaBindingUrlName(config.name))
}

func handleProvisionResponse(err error, printedName, namespace, name string) error {
	if err == nil {
		fmt.Printf("Created %s (%s/%s).\n", printedName, namespace, name)
		return nil
	}
	return &clierror.Error{
		Message: "failed to provision Hana resource",
		Details: err.Error(),
	}
}

func hanaInstance(config *hanaProvisionConfig) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "services.cloud.sap.com/v1",
			"kind":       "ServiceInstance",
			"metadata": map[string]interface{}{
				"name": config.name,
			},
			"spec": map[string]interface{}{
				"serviceOfferingName": "hana-cloud", // fixed
				"servicePlanName":     config.planName,
				"externalName":        config.name,
				"parameters": map[string]interface{}{
					"data": map[string]interface{}{
						"memory":                 config.memory,
						"vcpu":                   config.cpu,
						"whitelistIPs":           config.whitelistIP,
						"generateSystemPassword": true,    // TODO: manage it later
						"edition":                "cloud", // TODO: is it necessary?
					},
				},
			},
		},
	}
}

func hanaBinding(config *hanaProvisionConfig) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "services.cloud.sap.com/v1",
			"kind":       "ServiceBinding",
			"metadata": map[string]interface{}{
				"name": config.name,
			},
			"spec": map[string]interface{}{
				"serviceInstanceName": config.name,
				"externalName":        config.name,
				"secretName":          config.name,
				"parameters": map[string]interface{}{
					"scope":           "administration",     // fixed
					"credential-type": "PASSWORD_GENERATED", // fixed
				},
			},
		},
	}
}

func hanaBindingUrl(config *hanaProvisionConfig) *unstructured.Unstructured {
	urlName := hanaBindingUrlName(config.name)
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "services.cloud.sap.com/v1",
			"kind":       "ServiceBinding",
			"metadata": map[string]interface{}{
				"name": urlName,
			},
			"spec": map[string]interface{}{
				"serviceInstanceName": config.name,
				"externalName":        urlName,
				"secretName":          urlName,
			},
		},
	}
}

func hanaBindingUrlName(name string) string {
	return fmt.Sprintf("%s-url", name)
}
