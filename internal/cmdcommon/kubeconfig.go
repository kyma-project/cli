package cmdcommon

import (
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
)

type KubeClientConfig interface {
	GetKubeClient() (kube.Client, error)
	GetKubeClientWithClierr() (kube.Client, clierror.Error)
}

// KubeClientConfig allows to setup kubeconfig flag and use it to create kube.Client
type kubeClientConfig struct {
	kubeClient    kube.Client
	kubeClientErr error
}

func newKubeClientConfig() KubeClientConfig {
	cfg := &kubeClientConfig{}
	cfg.complete()

	return cfg
}

func AddCmdPersistentKubeconfigFlag(cmd *cobra.Command) {
	// this flag is not operational. it's only to print help description and help cobra with validation
	_ = cmd.PersistentFlags().String("kubeconfig", "", "Path to the Kyma kubeconfig file")
	_ = cmd.PersistentFlags().String("context", "", "The name of the kubeconfig context to use")
}

func (kcc *kubeClientConfig) GetKubeClient() (kube.Client, error) {
	return kcc.kubeClient, kcc.kubeClientErr
}

func (kcc *kubeClientConfig) GetKubeClientWithClierr() (kube.Client, clierror.Error) {
	if kcc.kubeClientErr != nil {
		return nil, clierror.Wrap(kcc.kubeClientErr,
			clierror.New("failed to create connection with the target Kyma environment", "Make sure that kubeconfig is proper."),
		)
	}

	return kcc.kubeClient, nil
}

func (kcc *kubeClientConfig) complete() {
	kubeconfigPath := getStringFlagValue("--kubeconfig")
	context := getStringFlagValue("--context")

	kcc.kubeClient, kcc.kubeClientErr = kube.NewClient(kubeconfigPath, context)
}

// search os.Args manually to find if user pass --<flag_name> path and return its value
func getStringFlagValue(flag string) string {
	value := ""
	for i, arg := range os.Args {
		// example: --kubeconfig /path/to/file
		if arg == flag && len(os.Args) > i+1 {
			value = os.Args[i+1]
		}

		// example: --kubeconfig=/path/to/file
		argFields := strings.Split(arg, "=")
		if strings.HasPrefix(arg, fmt.Sprintf("%s=", flag)) && len(argFields) == 2 {
			value = argFields[1]
		}
	}

	return value
}
