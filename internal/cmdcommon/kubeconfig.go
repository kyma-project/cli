package cmdcommon

import (
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
	kubeconfigPath := getKubeconfigPath()

	kcc.kubeClient, kcc.kubeClientErr = kube.NewClient(kubeconfigPath)
}

// search os.Args manually to find if user pass --kubeconfig path and return its value
func getKubeconfigPath() string {
	path := ""
	for i, arg := range os.Args {
		// example: --kubeconfig /path/to/file
		if arg == "--kubeconfig" && len(os.Args) > i+1 {
			path = os.Args[i+1]
		}

		// example: --kubeconfig=/path/to/file
		argFields := strings.Split(arg, "=")
		if strings.HasPrefix(arg, "--kubeconfig=") && len(argFields) == 2 {
			path = argFields[1]
		}
	}

	return path
}
