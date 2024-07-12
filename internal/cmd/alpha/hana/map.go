package hana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/btp"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

// this command uses the same config as check command
func NewMapHanaCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaCheckConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "map",
		Short: "Map the Hana instance to the Kyma cluster.",
		Long:  "Use this command to map the Hana instance to the Kyma cluster.",
		PreRun: func(_ *cobra.Command, _ []string) {
			clierror.Check(config.KubeClientConfig.Complete())
		},
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runMap(&config))
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")
	cmd.Flags().DurationVar(&config.timeout, "timeout", 7*time.Minute, "Timeout for the command")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

var (
	mapCommands = []func(config *hanaCheckConfig) clierror.Error{
		createHanaAPIInstanceIfNeeded,
		createHanaAPIBindingIfNeeded,
		createHanaInstanceMapping,
	}
)

func runMap(config *hanaCheckConfig) clierror.Error {
	for _, command := range mapCommands {
		err := command(config)
		if err != nil {
			return err
		}
	}

	fmt.Println("Hana instance was successfully mapped to the cluster")
	fmt.Println("You may want to create a Hana HDI container: see how to do it under https://help.sap.com/docs/hana-cloud/sap-hana-cloud-getting-started-guide/set-up-hdi-container-kyma")
	return nil
}

func createHanaAPIInstanceIfNeeded(config *hanaCheckConfig) clierror.Error {
	// check if instance exists, skip API instance creation if it does
	instance, err := config.KubeClient.Btp().GetServiceInstance(config.Ctx, config.namespace, hanaBindingAPIName(config.name))
	if err == nil && instance != nil {
		fmt.Printf("Hana API instance already exists (%s/%s)\n", config.namespace, hanaBindingAPIName(config.name))
		return nil
	}
	return createHanaAPIInstance(config)
}

func createHanaAPIBindingIfNeeded(config *hanaCheckConfig) clierror.Error {
	//check if binding exists, skip API binding creation if it does
	instance, err := config.KubeClient.Btp().GetServiceBinding(config.Ctx, config.namespace, hanaBindingAPIName(config.name))
	if err == nil && instance != nil {
		fmt.Printf("Hana API instance already exists (%s/%s)\n", config.namespace, hanaBindingAPIName(config.name))
		return nil
	}

	return createHanaAPIBinding(config)

}

func createHanaAPIInstance(config *hanaCheckConfig) clierror.Error {
	instance := hanaAPIInstance(config)

	err := config.KubeClient.Btp().CreateServiceInstance(config.Ctx, instance)
	return handleProvisionResponse(err, "Hana API instance", config.namespace, hanaBindingAPIName(config.name))
}

func createHanaAPIBinding(config *hanaCheckConfig) clierror.Error {
	binding := hanaAPIBinding(config)

	err := config.KubeClient.Btp().CreateServiceBinding(config.Ctx, binding)
	return handleProvisionResponse(err, "Hana API binding", config.namespace, hanaBindingAPIName(config.name))
}

func hanaAPIInstance(config *hanaCheckConfig) *btp.ServiceInstance {
	return &btp.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: btp.ServicesAPIVersionV1,
			Kind:       btp.KindServiceInstance,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hanaBindingAPIName(config.name),
			Namespace: config.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": hanaBindingAPIName(config.name),
			},
		},
		Spec: btp.ServiceInstanceSpec{
			Parameters: HanaAPIParameters{
				TechnicalUser: true,
			},
			ServiceOfferingName: "hana-cloud",
			ServicePlanName:     "admin-api-access",
		},
	}
}

func hanaAPIBinding(config *hanaCheckConfig) *btp.ServiceBinding {
	instanceName := hanaBindingAPIName(config.name)
	return &btp.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: btp.ServicesAPIVersionV1,
			Kind:       btp.KindServiceBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: config.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": hanaBindingAPIName(config.name),
			},
		},
		Spec: btp.ServiceBindingSpec{
			ServiceInstanceName: instanceName,
			SecretName:          instanceName,
		},
	}
}

func hanaBindingAPIName(name string) string {
	return fmt.Sprintf("%s-api", name)
}

func createHanaInstanceMapping(config *hanaCheckConfig) clierror.Error {
	clusterID, err := getClusterID(config)
	if err != nil {
		return err
	}

	hanaID, err := getHanaID(config)
	if err != nil {
		return err
	}

	// read secret
	baseurl, uaa, err := readHanaAPISecret(config)
	if err != nil {
		return err
	}

	// authenticate
	token, err := auth.GetOAuthToken("client_credentials", uaa.URL, uaa.ClientID, uaa.ClientSecret)
	if err != nil {
		return err
	}
	// create mapping
	return hanaInstanceMapping(baseurl, clusterID, hanaID, token.AccessToken)
}

func getClusterID(config *hanaCheckConfig) (string, clierror.Error) {
	cm, err := config.KubeClient.Static().CoreV1().ConfigMaps("kyma-system").Get(config.Ctx, "sap-btp-operator-config", metav1.GetOptions{})
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to get cluster ID"))
	}
	return cm.Data["CLUSTER_ID"], nil
}

func getHanaID(config *hanaCheckConfig) (string, clierror.Error) {
	// wait for until Hana instance is ready, for default setting it should take 5 minutes
	fmt.Print("waiting for Hana instance to be ready... ")
	instanceReadyCheck := config.KubeClient.Btp().IsInstanceReady(config.Ctx, config.namespace, config.name)
	err := wait.PollUntilContextTimeout(config.Ctx, 10*time.Second, config.timeout, true, instanceReadyCheck)
	if err != nil {
		fmt.Println("Failed")
		return "", clierror.Wrap(err,
			clierror.New("timeout while waiting for Hana instance to be ready", "make sure the hana-cloud hana entitlement is enabled"),
		)
	}
	fmt.Println("done")

	instance, err := config.KubeClient.Btp().GetServiceInstance(config.Ctx, config.namespace, config.name)
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to get Hana instance"))
	}

	return instance.Status.InstanceID, nil
}

func readHanaAPISecret(config *hanaCheckConfig) (string, *auth.UAA, clierror.Error) {
	fmt.Print("waiting for Hana API instance to be ready... ")
	instanceReadyCheck := config.KubeClient.Btp().IsInstanceReady(config.Ctx, config.namespace, hanaBindingAPIName(config.name))
	err := wait.PollUntilContextTimeout(config.Ctx, 5*time.Second, 2*time.Minute, true, instanceReadyCheck)
	if err != nil {
		fmt.Println("Failed")
		return "", nil, clierror.Wrap(err,
			clierror.New("timeout while waiting for Hana API instance", "make sure the hana-cloud admin-api-access entitlement is enabled"),
		)
	}
	fmt.Println("done")

	fmt.Print("waiting for Hana API binding to be ready... ")
	bindingReadyCheck := config.KubeClient.Btp().IsBindingReady(config.Ctx, config.namespace, hanaBindingAPIName(config.name))
	err = wait.PollUntilContextTimeout(config.Ctx, 5*time.Second, 2*time.Minute, true, bindingReadyCheck)
	if err != nil {
		fmt.Println("Failed")
		return "", nil, clierror.Wrap(err, clierror.New("timeout while waiting for Hana API binding"))
	}
	fmt.Println("done")
	secret, err := config.KubeClient.Static().CoreV1().Secrets(config.namespace).Get(config.Ctx, hanaBindingAPIName(config.name), metav1.GetOptions{})
	if err != nil {
		return "", nil, clierror.Wrap(err, clierror.New("failed to get secret"))
	}
	baseURL := secret.Data["baseurl"]
	uaaData := secret.Data["uaa"]

	uaa := &auth.UAA{}
	err = json.Unmarshal(uaaData, uaa)
	if err != nil {
		return "", nil, clierror.Wrap(err, clierror.New("failed to decode UAA data"))
	}
	return string(baseURL), uaa, nil
}

func hanaInstanceMapping(baseURL, clusterID, hanaID, token string) clierror.Error {
	client := &http.Client{}

	requestData := HanaMapping{
		Platform:  "kubernetes",
		PrimaryID: clusterID,
	}

	requestString, err := json.Marshal(requestData)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create mapping request"))
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("https://%s/inventory/v2/serviceInstances/%s/instanceMappings", baseURL, hanaID), bytes.NewBuffer(requestString))
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create mapping request"))
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(request)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create mapping"))
	}

	// server sends status Created when mapping is created, and 200 if it already exists
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return clierror.Wrap(fmt.Errorf("status code: %d", resp.StatusCode), clierror.New("failed to create mapping"))
	}

	return nil
}
