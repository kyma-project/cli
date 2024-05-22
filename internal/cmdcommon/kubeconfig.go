package cmdcommon

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/spf13/cobra"
)

// KubeClientConfig allows to setup kubeconfig flag and use it to create kube.Client
type KubeClientConfig struct {
	Kubeconfig string
	KubeClient kube.Client
}

func (kcc *KubeClientConfig) AddFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&kcc.Kubeconfig, "kubeconfig", "", "Path to the Kyma kubeconfig file.")
}

func (kcc *KubeClientConfig) Complete() clierror.Error {
	var err clierror.Error
	kcc.KubeClient, err = kube.NewClient(kcc.Kubeconfig)

	return err
}
