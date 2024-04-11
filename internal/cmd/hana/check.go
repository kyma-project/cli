package hana

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type hanaCheckConfig struct {
	ctx        context.Context
	kubeClient kube.Client

	kubeconfig string
	name       string
	namespace  string
	timeout    time.Duration
}

func NewHanaCheckCMD() *cobra.Command {
	config := hanaCheckConfig{}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if the Hana instance is provisioned.",
		Long:  "Use this command to check if the Hana instance is provisioned on the SAP Kyma platform.",
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runCheck(&config)
		},
	}

	cmd.Flags().StringVar(&config.kubeconfig, "kubeconfig", "", "Path to the Kyma kubecongig file.")

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (pc *hanaCheckConfig) complete() error {
	// TODO: think about timeout and moving context to persistent `kyma` command configuration
	pc.ctx = context.Background()

	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

type somethingWithStatus struct {
	Status status
}

type status struct {
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
	u, err := getServiceInstance(config)
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

	ready, error := isReady(u)
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

func getServiceInstance(config *hanaCheckConfig) (*unstructured.Unstructured, error) {
	return config.kubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Get(config.ctx, config.name, metav1.GetOptions{})
}

func getServiceBinding(config *hanaCheckConfig, name string) (*unstructured.Unstructured, error) {
	return config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Get(config.ctx, name, metav1.GetOptions{})
}

func isReady(u *unstructured.Unstructured) (bool, error) {
	instance := somethingWithStatus{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &instance); err != nil {
		return false, &clierror.Error{
			Message: "failed to read resource data",
			Details: err.Error(),
		}
	}
	status := instance.Status
	ready := (status.Ready == "True") &&
		isConditionTrue(status.Conditions, "Succeeded") &&
		isConditionTrue(status.Conditions, "Ready")
	if !ready {
		return false, nil
	}
	return true, nil
}

func isConditionTrue(conditions []metav1.Condition, conditionType string) bool {
	condition := meta.FindStatusCondition(conditions, conditionType)
	return condition != nil && condition.Status == metav1.ConditionTrue
}
