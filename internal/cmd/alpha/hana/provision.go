package hana

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		createHanaBindingURL,
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
	instance := hanaInstance(config)

	err := config.KubeClient.Btp().CreateServiceInstance(config.Ctx, instance)
	return handleProvisionResponse(err, "Hana instance", config.namespace, config.name)
}

func createHanaBinding(config *hanaProvisionConfig) clierror.Error {
	binding := hanaBinding(config)

	err := config.KubeClient.Btp().CreateServiceBinding(config.Ctx, binding)
	return handleProvisionResponse(err, "Hana binding", config.namespace, config.name)
}

func createHanaBindingURL(config *hanaProvisionConfig) clierror.Error {
	bindingURL := hanaBindingURL(config)

	err := config.KubeClient.Btp().CreateServiceBinding(config.Ctx, bindingURL)
	return handleProvisionResponse(err, "Hana URL binding", config.namespace, hanaBindingURLName(config.name))
}

func handleProvisionResponse(err error, printedName, namespace, name string) clierror.Error {
	if err == nil {
		fmt.Printf("Created %s (%s/%s).\n", printedName, namespace, name)
		return nil
	}
	return clierror.Wrap(err, clierror.New("failed to provision Hana resource"))
}

func hanaInstance(config *hanaProvisionConfig) *btp.ServiceInstance {
	return &btp.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: btp.ServicesAPIVersionV1,
			Kind:       btp.KindServiceInstance,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.name,
		},
		Spec: btp.ServiceInstanceSpec{
			ServiceOfferingName: "hana-cloud", // fixed
			ServicePlanName:     config.planName,
			ExternalName:        config.name,
			Parameters: HanaInstanceParameters{
				Data: HanaInstanceParametersData{
					Memory:                 config.memory,
					Vcpu:                   config.cpu,
					WhitelistIPs:           config.whitelistIP,
					GenerateSystemPassword: true,    // TODO: manage it later
					Edition:                "cloud", // TODO: is it necessary?
				},
			},
		},
	}
}

func hanaBinding(config *hanaProvisionConfig) *btp.ServiceBinding {
	return &btp.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: btp.ServicesAPIVersionV1,
			Kind:       btp.KindServiceBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: config.name,
		},
		Spec: btp.ServiceBindingSpec{
			ServiceInstanceName: config.name,
			ExternalName:        config.name,
			SecretName:          config.name,
			Parameters: HanaBindingParameters{
				Scope:           "administration",     // fixed
				CredentialsType: "PASSWORD_GENERATED", // fixed
			},
		},
	}
}

func hanaBindingURL(config *hanaProvisionConfig) *btp.ServiceBinding {
	urlName := hanaBindingURLName(config.name)
	return &btp.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: btp.ServicesAPIVersionV1,
			Kind:       btp.KindServiceBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: urlName,
		},
		Spec: btp.ServiceBindingSpec{
			ServiceInstanceName: config.name,
			ExternalName:        urlName,
			SecretName:          urlName,
		},
	}
}

func hanaBindingURLName(name string) string {
	return fmt.Sprintf("%s-url", name)
}
