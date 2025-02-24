package cmdcommon

import (
	"os"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
)

// KubeClientConfig allows to setup kubeconfig flag and use it to create kube.Client
type KubeClientConfig struct {
	KubeClient    kube.Client
	KubeClientErr error
}

func newKubeClientConfig(cmd *cobra.Command) *KubeClientConfig {
	cfg := &KubeClientConfig{}
	cfg.addFlag(cmd)
	cfg.complete()

	return cfg
}

func (kcc *KubeClientConfig) GetKubeClient() (kube.Client, error) {
	return kcc.KubeClient, kcc.KubeClientErr
}

func (kcc *KubeClientConfig) GetKubeClientWithClierr() (kube.Client, clierror.Error) {
	if kcc.KubeClientErr != nil {
		return nil, clierror.Wrap(kcc.KubeClientErr,
			clierror.New("failed to create cluster connection", "Make sure that kubeconfig is proper."),
		)
	}

	return kcc.KubeClient, nil
}

func (kcc *KubeClientConfig) addFlag(cmd *cobra.Command) {
	// this flag is not operational. it's only to print help description and help cobra with validation
	_ = cmd.PersistentFlags().String("kubeconfig", "", "Path to the Kyma kubeconfig file")
}

func (kcc *KubeClientConfig) complete() {
	kubeconfigPath := getKubeconfigPath()

	kcc.KubeClient, kcc.KubeClientErr = kube.NewClient(kubeconfigPath)
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
