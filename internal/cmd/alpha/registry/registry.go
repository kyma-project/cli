package registry

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/registry/config"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewRegistryCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Set of commands for Kyma registry",
		Long:  `Use this command manage resources related to Kyma registry`,
	}

	cmd.AddCommand(config.NewConfigCMD(kymaConfig))

	return cmd
}
