package hana

import (
	"github.com/spf13/cobra"
)

func NewHanaCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "hana",
		Short:                 "Manage a Hana instance on the Kyma platform.",
		Long:                  `Use this command to manage a Hana instance on the SAP Kyma platform.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewHanaProvisionCMD())
	cmd.AddCommand(NewHanaCheckCMD())
	cmd.AddCommand(NewHanaDeleteCMD())
	cmd.AddCommand(NewHanaCredentialsCMD())

	return cmd
}
