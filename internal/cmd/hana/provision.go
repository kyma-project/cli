package hana

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type hanaProvisionConfig struct {
	ctx        context.Context
	kubeClient kube.Client

	kubeconfig  string
	name        string
	namespace   string
	planName    string
	memory      int
	cpu         int
	whitelistIP []string
}

func NewHanaProvisionCMD() *cobra.Command {
	config := hanaProvisionConfig{}

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
	cmd.Flags().IntVar(&config.memory, "memory", 30, "Memory size for Hana.")                                        //TODO: fulfill proper usage
	cmd.Flags().IntVar(&config.cpu, "cpu", 2, "Number of CPUs for Hana.")                                            //TODO: fulfill proper usage
	cmd.Flags().StringSliceVar(&config.whitelistIP, "whitelist-ip", []string{"0.0.0.0/0"}, "IP whitelist for Hana.") //TODO: fulfill proper usage

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (pc *hanaProvisionConfig) complete() error {
	// TODO: think about itmeout and moving context to persistent `kyma` command configuration
	pc.ctx = context.Background()

	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

func runProvision(config *hanaProvisionConfig) error {
	fmt.Printf("Provisioning Hana %s/%s.\n", config.namespace, config.name)

	if _, err := createHanaInstance(config); err != nil {
		return err
	}
	fmt.Println("Created Hana.")

	if _, err := createHanaBinding(config); err != nil {
		return err
	}
	fmt.Println("Created Hana binding.")

	if _, err := createHanaBindingUrl(config); err != nil {
		return err
	}
	fmt.Println("Created Hana URL binding.")
	fmt.Println("Operation completed.")

	return nil
}

func createHanaInstance(config *hanaProvisionConfig) (*unstructured.Unstructured, error) {
	return config.kubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Create(config.ctx, hanaInstance(config), metav1.CreateOptions{})
}

func createHanaBinding(config *hanaProvisionConfig) (*unstructured.Unstructured, error) {
	return config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.ctx, hanaBinding(config), metav1.CreateOptions{})
}

func createHanaBindingUrl(config *hanaProvisionConfig) (*unstructured.Unstructured, error) {
	return config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.ctx, hanaBindingUrl(config), metav1.CreateOptions{})
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
	urlName := fmt.Sprintf("%s-url", config.name)
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
