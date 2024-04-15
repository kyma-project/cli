package hana

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type hanaCheckConfig struct {
	*cmdcommon.KymaConfig
	cmdcommon.KubeClientConfig

	name      string
	namespace string
	timeout   time.Duration
}

func NewHanaCheckCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaCheckConfig{
		KymaConfig:       kymaConfig,
		KubeClientConfig: cmdcommon.KubeClientConfig{},
	}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if the Hana instance is provisioned.",
		Long:  "Use this command to check if the Hana instance is provisioned on the SAP Kyma platform.",
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.KubeClientConfig.Complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runCheck(&config)
		},
	}

	config.KubeClientConfig.AddFlag(cmd)

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")

	_ = cmd.MarkFlagRequired("name")

	return cmd
}

type somethingWithStatus struct {
	Status Status
}

type Status struct {
	Conditions []metav1.Condition
	Ready      string
	InstanceID string
}

var (
	checkCommands = []func(config *hanaCheckConfig) error{
		checkHanaInstance,
		checkHanaBinding,
		checkHanaBindingUrl,
	}
)

func runCheck(config *hanaCheckConfig) error {
	fmt.Printf("Checking Hana (%s/%s).\n", config.namespace, config.name)

	for _, command := range checkCommands {
		err := command(config)
		if err != nil {
			fmt.Println("Hana is not fully ready.")
			return err
		}
	}
	fmt.Println("Hana is fully ready.")
	return nil
}

func checkHanaInstance(config *hanaCheckConfig) error {
	u, err := getServiceInstance(config, config.name)
	return handleCheckResponse(u, err, "Hana instance", config.namespace, config.name)
}

func checkHanaBinding(config *hanaCheckConfig) error {
	u, err := getServiceBinding(config, config.name)
	return handleCheckResponse(u, err, "Hana binding", config.namespace, config.name)
}

func checkHanaBindingUrl(config *hanaCheckConfig) error {
	urlName := hanaBindingUrlName(config.name)
	u, err := getServiceBinding(config, urlName)
	return handleCheckResponse(u, err, "Hana URL binding", config.namespace, urlName)
}

func handleCheckResponse(u *unstructured.Unstructured, err error, printedName, namespace, name string) error {
	if err != nil {
		return &clierror.Error{
			Message: "failed to get resource data",
			Details: err.Error(),
			Hints: []string{
				"Make sure that Hana was provisioned.",
			},
		}
	}

	ready, _, error := isReady(u)
	if error != nil {
		return &clierror.Error{
			Message: "failed to check readiness of Hana resources",
			Details: error.Error(),
		}
	}
	if !ready {
		fmt.Printf("%s is not ready (%s/%s).\n", printedName, namespace, name)
		return &clierror.Error{
			Message: fmt.Sprintf("%s is not ready", strings.ToLower(printedName[:1])+printedName[1:]),
			Hints: []string{
				"Wait for provisioning of Hana resources.",
				"Check if Hana resources started without errors.",
			},
		}

	}
	fmt.Printf("%s is ready (%s/%s).\n", printedName, namespace, name)
	return nil
}

func getServiceInstance(config *hanaCheckConfig, name string) (*unstructured.Unstructured, error) {
	return config.KubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Get(config.Ctx, config.name, metav1.GetOptions{})
}

func getServiceBinding(config *hanaCheckConfig, name string) (*unstructured.Unstructured, error) {
	return config.KubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Get(config.Ctx, name, metav1.GetOptions{})
}

func getServiceStatus(u *unstructured.Unstructured) (Status, error) {
	instance := somethingWithStatus{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &instance); err != nil {
		return Status{}, &clierror.Error{
			Message: "failed to read resource data",
			Details: err.Error(),
		}
	}

	return instance.Status, nil
}

// isReady returns readiness status, and failed status if at least one of the contitions has failed, or an error was returned
func isReady(u *unstructured.Unstructured) (bool, bool, error) {
	status, err := getServiceStatus(u)
	if err != nil {
		return false, true, err
	}

	failed := (status.Ready == "False") &&
		isConditionTrue(status.Conditions, "Failed")
	if failed {
		return false, true, nil
	}

	ready := (status.Ready == "True") &&
		isConditionTrue(status.Conditions, "Succeeded") &&
		isConditionTrue(status.Conditions, "Ready")
	return ready, false, nil
}

func isConditionTrue(conditions []metav1.Condition, conditionType string) bool {
	condition := meta.FindStatusCondition(conditions, conditionType)
	return condition != nil && condition.Status == metav1.ConditionTrue
}

func getConditionMessage(conditions []metav1.Condition, conditionType string) string {
	condition := meta.FindStatusCondition(conditions, conditionType)
	if condition == nil {
		return ""
	}
	return condition.Message
}
