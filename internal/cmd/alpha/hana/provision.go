package hana

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type hanaProvisionConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	name        string
	namespace   string
	planName    string
	memory      int
	cpu         int
	whitelistIP []string
}

func NewHanaProvisionCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaProvisionConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "provision",
		Short: "Provisions a Hana instance on the Kyma.",
		Long:  "Use this command to provision a Hana instance on the SAP Kyma platform.",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runProvision(&config))
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")
	cmd.Flags().StringVar(&config.planName, "plan", "hana", "Name of the service plan.")
	cmd.Flags().IntVar(&config.memory, "memory", 32, "Memory size for Hana.")                                        //TODO: fulfill proper usage
	cmd.Flags().IntVar(&config.cpu, "cpu", 2, "Number of CPUs for Hana.")                                            //TODO: fulfill proper usage
	cmd.Flags().StringSliceVar(&config.whitelistIP, "whitelist-ip", []string{"0.0.0.0/0"}, "IP whitelist for Hana.") //TODO: fulfill proper usage

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

var (
	provisionCommands = []func(*hanaProvisionConfig) clierror.Error{
		createHanaInstance,
		createHanaBinding,
		createHanaBindingUrl,
	}
)

func runProvision(config *hanaProvisionConfig) clierror.Error {
	fmt.Printf("Provisioning Hana (%s/%s).\n", config.namespace, config.name)

	for _, command := range provisionCommands {
		err := command(config)
		if err != nil {
			return err
		}
	}
	fmt.Println("Operation completed.")
	fmt.Println("You may want to map the Hana instance to use it inside the cluster: see the 'kyma hana map' command.")
	return nil
}

func createHanaInstance(config *hanaProvisionConfig) clierror.Error {
	_, err := config.KubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Create(config.Ctx, hanaInstance(config), metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana instance", config.namespace, config.name)
}

func createHanaBinding(config *hanaProvisionConfig) clierror.Error {
	_, err := config.KubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.Ctx, hanaBinding(config), metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana binding", config.namespace, config.name)
}

func createHanaBindingUrl(config *hanaProvisionConfig) clierror.Error {
	_, err := config.KubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.Ctx, hanaBindingUrl(config), metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana URL binding", config.namespace, hanaBindingURLName(config.name))
}

func handleProvisionResponse(err error, printedName, namespace, name string) clierror.Error {
	if err == nil {
		fmt.Printf("Created %s (%s/%s).\n", printedName, namespace, name)
		return nil
	}
	return clierror.Wrap(err, clierror.New("failed to provision Hana resource"))
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
	urlName := hanaBindingURLName(config.name)
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

func hanaBindingURLName(name string) string {
	return fmt.Sprintf("%s-url", name)
}
