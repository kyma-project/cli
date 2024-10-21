package cmdcommon

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/spf13/cobra"
)

// KymaConfig contains data common for all subcommands
type KymaConfig struct {
	*KubeClientConfig

	Ctx context.Context
}

func NewKymaConfig(cmd *cobra.Command) (*KymaConfig, clierror.Error) {
	kubeClient, err := newKubeClientConfig(cmd)
	if err != nil {
		return nil, err
	}

	return &KymaConfig{
		Ctx:              context.Background(),
		KubeClientConfig: kubeClient,
	}, nil
}
