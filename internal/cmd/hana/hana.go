package hana

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewHanaCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "hana",
		Short:                 "Manage a Hana instance on the Kyma platform.",
		Long:                  `Use this command to manage a Hana instance on the SAP Kyma platform.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewHanaProvisionCMD(kymaConfig))
	cmd.AddCommand(NewHanaCheckCMD(kymaConfig))
	cmd.AddCommand(NewHanaDeleteCMD(kymaConfig))
	cmd.AddCommand(NewHanaCredentialsCMD(kymaConfig))
	cmd.AddCommand(NewMapHanaCMD(kymaConfig))

	return cmd
}
