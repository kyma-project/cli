package diagnose

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewDiagnoseCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "diagnose <command> [flags]",
		Short:                 "Runs diagnostic commands to troubleshoot your Kyma cluster",
		Long:                  "Use diagnostic commands to collect cluster information, analyze logs, and assess the health of your Kyma installation. Choose from available subcommands to target specific diagnostic areas.",
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewDiagnoseClusterCMD(kymaConfig))
	cmd.AddCommand(NewDiagnoseLogsCMD(kymaConfig))
	cmd.AddCommand(NewDiagnoseIstioCMD(kymaConfig))

	return cmd
}
