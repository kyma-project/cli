package cmdcommon

import (
	"context"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/spf13/cobra"
)

// KymaConfig contains data common for all subcommands
type KymaConfig struct {
	*KubeClientConfig

	Ctx        context.Context
	Extensions ExtensionList
}

func NewKymaConfig(cmd *cobra.Command) (*KymaConfig, clierror.Error) {
	ctx := context.Background()

	kubeClient, kubeClientErr := newKubeClientConfig(cmd)
	if kubeClientErr != nil {
		return nil, kubeClientErr
	}

	extensions, err := ListExtensions(ctx, kubeClient.KubeClient.Static())
	if err != nil {
		fmt.Printf("DEBUG ERROR: %s\n", err.Error())
		// TODO: think about handling error later
		// this error should not stop program
		// but I'm not sure what we should do with such information due to it's internal value
	}

	return &KymaConfig{
		Ctx:              ctx,
		KubeClientConfig: kubeClient,
		Extensions:       extensions,
	}, nil
}
