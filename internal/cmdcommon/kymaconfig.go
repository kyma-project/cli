package cmdcommon

import (
	"context"
)

// KymaConfig contains data common for all subcommands
type KymaConfig struct {
	*KubeClientConfig
	*KymaExtensionsConfig

	Ctx context.Context
}

func NewKymaConfig() *KymaConfig {
	ctx := context.Background()

	kymaConfig := &KymaConfig{}
	kymaConfig.Ctx = ctx
	kymaConfig.KubeClientConfig = newKubeClientConfig()
	kymaConfig.KymaExtensionsConfig = newExtensionsConfig(kymaConfig)

	return kymaConfig
}
