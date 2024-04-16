package hana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-project/cli.v3/internal/btp/auth"
	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.KubeClientConfig.Complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runMap(&config)
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
	mapCommands = []func(config *hanaCheckConfig) error{
		checkAndCreateHanaAPIInstance,
		checkAndCreateHanaAPIBinding,
		createHanaInstanceMapping,
	}
)

func runMap(config *hanaCheckConfig) error {
	for _, command := range mapCommands {
		err := command(config)
		if err != nil {
			return err
		}
	}

	fmt.Println("Hana instance was succesfully mapped to the cluster")
	return nil
}

func checkAndCreateHanaAPIInstance(config *hanaCheckConfig) error {
	// check if instance exists, skip API instance creation if it does
	instance, err := kube.GetServiceInstance(config.KubeClient, config.Ctx, config.namespace, hanaBindingAPIName(config.name))
	if err == nil && instance != nil {
		fmt.Printf("Hana API instance already exists (%s/%s)\n", config.namespace, hanaBindingAPIName(config.name))
		return nil
	}
	return createHanaAPIInstance(config)
}

func checkAndCreateHanaAPIBinding(config *hanaCheckConfig) error {
	//check if binding exists, skip API binding creation if it does
	instance, err := kube.GetServiceBinding(config.KubeClient, config.Ctx, config.namespace, hanaBindingAPIName(config.name))
	if err == nil && instance != nil {
		fmt.Printf("Hana API instance already exists (%s/%s)\n", config.namespace, hanaBindingAPIName(config.name))
		return nil
	}

	return createHanaAPIBinding(config)

}

func createHanaAPIInstance(config *hanaCheckConfig) error {
	data, err := hanaAPIInstance(config)
	if err != nil {
		return &clierror.Error{
			Message: "failed to create Hana API instance object",
			Details: err.Error(),
		}
	}
	_, err = config.KubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Create(config.Ctx, data, metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana API instance", config.namespace, hanaBindingAPIName(config.name))
}

func createHanaAPIBinding(config *hanaCheckConfig) error {
	data, err := hanaAPIBinding(config)
	if err != nil {
		return &clierror.Error{
			Message: "failed to create Hana API binding object",
			Details: err.Error(),
		}
	}
	_, err = config.KubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Create(config.Ctx, data, metav1.CreateOptions{})
	return handleProvisionResponse(err, "Hana API binding", config.namespace, hanaBindingAPIName(config.name))
}

func hanaAPIInstance(config *hanaCheckConfig) (*unstructured.Unstructured, error) {
	requestData := kube.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kube.ServicesAPIVersionV1,
			Kind:       kube.KindServiceInstance,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      hanaBindingAPIName(config.name),
			Namespace: config.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": hanaBindingAPIName(config.name),
			},
		},
		Spec: kube.ServiceInstanceSpec{
			Parameters: HanaAPIParameters{
				TechnicalUser: true,
			},
			OfferingName: "hana-cloud",
			PlanName:     "admin-api-access",
		},
	}
	return kube.ToUnstructured(requestData, operator.GVKServiceInstance)
}

func hanaAPIBinding(config *hanaCheckConfig) (*unstructured.Unstructured, error) {
	instanceName := hanaBindingAPIName(config.name)
	requestData := kube.ServiceBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: kube.ServicesAPIVersionV1,
			Kind:       kube.KindServiceBinding,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      instanceName,
			Namespace: config.namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name": hanaBindingAPIName(config.name),
			},
		},
		Spec: kube.ServiceBindingSpec{
			ServiceInstanceName: instanceName,
			SecretName:          instanceName,
		},
	}
	return kube.ToUnstructured(requestData, operator.GVKServiceBinding)
}

func hanaBindingAPIName(name string) string {
	return fmt.Sprintf("%s-api", name)
}

func createHanaInstanceMapping(config *hanaCheckConfig) error {
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

func getClusterID(config *hanaCheckConfig) (string, error) {
	cm, err := config.KubeClient.Static().CoreV1().ConfigMaps("kyma-system").Get(config.Ctx, "sap-btp-operator-config", metav1.GetOptions{})
	if err != nil {
		return "", &clierror.Error{
			Message: "failed to get cluster ID",
			Details: err.Error(),
		}
	}
	return cm.Data["CLUSTER_ID"], nil
}

func getHanaID(config *hanaCheckConfig) (string, error) {
	// wait for until Hana instance is ready, for default setting it should take 5 minutes
	fmt.Print("waiting for Hana instance to be ready... ")
	err := wait.PollUntilContextTimeout(config.Ctx, 10*time.Second, config.timeout, true, kube.IsInstanceReady(config.KubeClient, config.Ctx, config.namespace, config.name))
	if err != nil {
		return "", clierror.Wrap(err, &clierror.Error{
			Message: "timeout while waiting for Hana instance to be ready",
			Hints:   []string{"make sure the hana-cloud hana entitlement is enabled"},
		})
	}
	fmt.Println("done")

	u, err := config.KubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Get(config.Ctx, config.name, metav1.GetOptions{})
	if err != nil {
		return "", &clierror.Error{
			Message: "failed to get Hana instance",
			Details: err.Error(),
		}
	}
	status, err := kube.GetServiceStatus(u)
	if err != nil {
		return "", &clierror.Error{
			Message: "failed to read resource data",
			Details: err.Error(),
		}
	}

	return status.InstanceID, nil
}

func readHanaAPISecret(config *hanaCheckConfig) (string, *auth.UAA, error) {
	fmt.Print("waiting for Hana API instance to be ready... ")
	err := wait.PollUntilContextTimeout(config.Ctx, 5*time.Second, 2*time.Minute, true, kube.IsInstanceReady(config.KubeClient, config.Ctx, config.namespace, hanaBindingAPIName(config.name)))
	if err != nil {
		return "", nil, clierror.Wrap(err, &clierror.Error{
			Message: "timeout while waiting for Hana API instance",
			Hints:   []string{"make sure the hana-cloud admin-api-access entitlement is enabled"},
		})
	}
	fmt.Println("done")

	fmt.Print("waiting for Hana API binding to be ready... ")
	err = wait.PollUntilContextTimeout(config.Ctx, 5*time.Second, 2*time.Minute, true, kube.IsBindingReady(config.KubeClient, config.Ctx, config.namespace, hanaBindingAPIName(config.name)))
	if err != nil {
		return "", nil, clierror.Wrap(err, &clierror.Error{
			Message: "timeout while waiting for Hana API binding",
			Details: err.Error(),
		})
	}
	fmt.Println("done")
	secret, err := config.KubeClient.Static().CoreV1().Secrets(config.namespace).Get(config.Ctx, hanaBindingAPIName(config.name), metav1.GetOptions{})
	if err != nil {
		return "", nil, &clierror.Error{
			Message: "failed to get secret",
			Details: err.Error(),
		}
	}
	baseURL := secret.Data["baseurl"]
	uaaData := secret.Data["uaa"]

	uaa := &auth.UAA{}
	err = json.Unmarshal(uaaData, uaa)
	if err != nil {
		return "", nil, &clierror.Error{
			Message: "failed to decode UAA data",
			Details: err.Error(),
		}
	}
	return string(baseURL), uaa, nil
}

func hanaInstanceMapping(baseURL, clusterID, hanaID, token string) error {
	client := &http.Client{}

	requestData := HanaMapping{
		Platform:  "kubernetes",
		PrimaryID: clusterID,
	}

	requestString, err := json.Marshal(requestData)
	if err != nil {
		return &clierror.Error{
			Message: "failed to create mapping request",
			Details: err.Error(),
		}
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("https://%s/inventory/v2/serviceInstances/%s/instanceMappings", baseURL, hanaID), bytes.NewBuffer(requestString))
	if err != nil {
		return &clierror.Error{
			Message: "failed to create mapping request",
			Details: err.Error(),
		}
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(request)
	if err != nil {
		return &clierror.Error{
			Message: "failed to create mapping",
			Details: err.Error(),
		}
	}

	// server sends status Created when mapping is created, and 200 if it already exists
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return &clierror.Error{
			Message: "failed to create mapping",
			Details: fmt.Sprintf("status code: %d", resp.StatusCode),
		}
	}

	return nil
}
