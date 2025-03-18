package cmdcommon

import (
	"context"

	"github.com/spf13/cobra"
)

// KymaConfig contains data common for all subcommands
type KymaConfig struct {
	*KubeClientConfig
	*KymaExtensionsConfig

	Ctx context.Context
}

func NewKymaConfig(cmd *cobra.Command) *KymaConfig {
	ctx := context.Background()

	kymaConfig := &KymaConfig{}
	kymaConfig.Ctx = ctx
	kymaConfig.KubeClientConfig = newKubeClientConfig(cmd)
	kymaConfig.KymaExtensionsConfig = newExtensionsConfig(kymaConfig)

	return kymaConfig
}
