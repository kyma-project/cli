package hana

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	u, err := kube.GetServiceInstance(config.kubeClient, config.ctx, config.namespace, config.name)
	return handleCheckResponse(u, err, "Hana instance", config.namespace, config.name)
}

func checkHanaBinding(config *hanaCheckConfig) error {
	u, err := kube.GetServiceBinding(config.kubeClient, config.ctx, config.namespace, config.name)
	return handleCheckResponse(u, err, "Hana binding", config.namespace, config.name)
}

func checkHanaBindingUrl(config *hanaCheckConfig) error {
	urlName := hanaBindingUrlName(config.name)
	u, err := kube.GetServiceBinding(config.kubeClient, config.ctx, config.namespace, urlName)
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

	ready, error := kube.IsReady(u)
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
