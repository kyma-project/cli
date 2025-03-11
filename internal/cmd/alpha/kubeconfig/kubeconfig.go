package kubeconfig

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewKubeconfigCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig <command> [flags]",
		Short: "Manages access to the Kyma cluster",
		Long:  "Use this command to manage access to the Kyma cluster",
	}

	cmd.AddCommand(newGenerateCMD(kymaConfig))

	return cmd
}
