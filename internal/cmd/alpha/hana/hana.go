package hana

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewHanaCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "hana <command> [flags]",
		Short:                 "Manages an SAP HANA instance in the Kyma cluster",
		Long:                  `Use this command to manage an SAP HANA instance in the Kyma cluster.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewMapHanaCMD(kymaConfig))

	return cmd
}
