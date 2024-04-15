package hana

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/btp/operator"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type hanaDeleteConfig struct {
	*cmdcommon.KymaConfig
	kubeClient kube.Client

	kubeconfig string
	name       string
	namespace  string
}

func NewHanaDeleteCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	config := hanaDeleteConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Hana instance on the Kyma.",
		Long:  "Use this command to delete a Hana instance on the SAP Kyma platform.",
		PreRunE: func(_ *cobra.Command, args []string) error {
			return config.complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDelete(&config)
		},
	}

	cmd.Flags().StringVar(&config.kubeconfig, "kubeconfig", "", "Path to the Kyma kubecongig file.")

	cmd.Flags().StringVar(&config.name, "name", "", "Name of Hana instance.")
	cmd.Flags().StringVar(&config.namespace, "namespace", "default", "Namespace for Hana instance.")

	_ = cmd.MarkFlagRequired("kubeconfig")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func (pc *hanaDeleteConfig) complete() error {
	var err error
	pc.kubeClient, err = kube.NewClient(pc.kubeconfig)

	return err
}

var (
	deleteCommands = []func(*hanaDeleteConfig) error{
		deleteHanaBinding,
		deleteHanaBindingUrl,
		deleteHanaInstance,
	}
)

func runDelete(config *hanaDeleteConfig) error {
	fmt.Printf("Deleting Hana (%s/%s).\n", config.namespace, config.name)

	for _, command := range deleteCommands {
		err := command(config)
		if err != nil {
			return err
		}
	}
	fmt.Println("Operation completed.")
	return nil
}

func deleteHanaInstance(config *hanaDeleteConfig) error {
	err := config.kubeClient.Dynamic().Resource(operator.GVRServiceInstance).
		Namespace(config.namespace).
		Delete(config.Ctx, config.name, metav1.DeleteOptions{})
	return handleDeleteResponse(err, "Hana instance", config.namespace, config.name)
}

func deleteHanaBinding(config *hanaDeleteConfig) error {
	err := config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Delete(config.Ctx, config.name, metav1.DeleteOptions{})
	return handleDeleteResponse(err, "Hana binding", config.namespace, config.name)
}

func deleteHanaBindingUrl(config *hanaDeleteConfig) error {
	urlName := hanaBindingUrlName(config.name)
	err := config.kubeClient.Dynamic().Resource(operator.GVRServiceBinding).
		Namespace(config.namespace).
		Delete(config.Ctx, urlName, metav1.DeleteOptions{})
	return handleDeleteResponse(err, "Hana URL binding", config.namespace, urlName)
}

func handleDeleteResponse(err error, printedName, namespace, name string) error {
	if err == nil {
		fmt.Printf("Deleted %s (%s/%s).\n", printedName, namespace, name)
		return nil
	}
	if errors.IsNotFound(err) {
		fmt.Printf("%s (%s/%s) not found.\n", printedName, namespace, name)
		return nil
	}
	return &clierror.Error{
		Message: "failed to delete Hana resource.",
		Details: err.Error(),
	}
}
