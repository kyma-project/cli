package diagnose

import (
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/spf13/cobra"
)

type diagnoseConfig struct {
	*cmdcommon.KymaConfig
	outputFormat types.Format
	outputPath   string
	verbose      bool
}

func NewDiagnoseCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "diagnose <command> [flags]",
		Short:                 "Run diagnostic commands to troubleshoot your Kyma cluster",
		Long:                  "Use diagnostic commands to collect cluster information, analyze logs, and assess the health of your Kyma installation. Choose from available subcommands to target specific diagnostic areas.",
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(NewDiagnoseClusterCMD(kymaConfig))
	cmd.AddCommand(NewDiagnoseLogsCMD(kymaConfig))
	cmd.AddCommand(NewDiagnoseIstioCMD(kymaConfig))

	return cmd
}
