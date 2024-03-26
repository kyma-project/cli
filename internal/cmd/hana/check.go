package hana

import (
	"context"
	"fmt"
	"github.com/kyma-project/cli.v3/internal/btp/operator"
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
}

var (
	checks = []func(config *hanaCheckConfig) error{
		checkHanaInstance,
		checkHanaBinding,
		checkHanaBindingUrl,
	}
)

func runCheck(config *hanaCheckConfig) error {
	fmt.Printf("Checkinging Hana %s/%s.\n", config.namespace, config.name)

	for _, check := range checks {
		err := check(config)
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
	if err != nil {
		return err
	}

	ready, err := isReady(u)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("hana instance is not ready")
	}
	fmt.Println("Hana instance is ready.")
	return nil
}

func checkHanaBinding(config *hanaCheckConfig) error {
	u, err := getServiceBinding(config, config.name)
	if err != nil {
		return err
	}

	ready, err := isReady(u)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("hana binding is not ready")
	}
	fmt.Println("Hana binding is ready.")
	return nil
}

func checkHanaBindingUrl(config *hanaCheckConfig) error {
	urlName := fmt.Sprintf("%s-url", config.name)
	u, err := getServiceBinding(config, urlName)
	if err != nil {
		return err
	}

	ready, err := isReady(u)
	if err != nil {
		return err
	}
	if !ready {
		return fmt.Errorf("hana binding url is not ready")
	}
	fmt.Println("Hana binding url is ready.")
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
		return false, err
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
