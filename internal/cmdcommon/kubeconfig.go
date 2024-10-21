package cmdcommon

import (
	"os"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
)

// KubeClientConfig allows to setup kubeconfig flag and use it to create kube.Client
type KubeClientConfig struct {
	KubeClient kube.Client
}

func newKubeClientConfig(cmd *cobra.Command) (*KubeClientConfig, clierror.Error) {
	cfg := &KubeClientConfig{}
	cfg.addFlag(cmd)

	return cfg, cfg.complete()
}

func (kcc *KubeClientConfig) addFlag(cmd *cobra.Command) {
	// this flag is not operational. it's only to print help description and help cobra with validation
	_ = cmd.PersistentFlags().String("kubeconfig", "", "Path to the Kyma kubeconfig file.")
}

func (kcc *KubeClientConfig) complete() clierror.Error {
	kubeconfigPath := getKubeconfigPath()

	var err clierror.Error
	kcc.KubeClient, err = kube.NewClient(kubeconfigPath)

	return err
}

// search os.Args manually to find if user pass --kubeconfig path and return its value
func getKubeconfigPath() string {
	path := ""
	for i, arg := range os.Args {
		if arg == "--kubeconfig" && len(os.Args) > i+1 {
			path = os.Args[i+1]
		}
	}

	return path
}
