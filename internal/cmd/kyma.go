package cmd

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha"
	"github.com/kyma-project/cli.v3/internal/cmd/app"
	"github.com/kyma-project/cli.v3/internal/cmd/module"
	"github.com/kyma-project/cli.v3/internal/cmd/version"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions"
	"github.com/kyma-project/cli.v3/internal/extensions/actions"
	extensionstypes "github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

func NewKymaCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kyma <command> [flags]",
		Short: "A simple set of commands to manage a Kyma cluster",
		Long:  "Use this command to manage Kyma modules and resources on a cluster.",
		// Affects children as well
		// by default Cobra adds `Error:` to the front of the error message, we want to suppress it
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	extensions.AddCmdPersistentFlags(cmd)
	cmdcommon.AddCmdPersistentKubeconfigFlag(cmd)
	cmd.PersistentFlags().BoolP("help", "h", false, "Help for the command")

	kymaConfig := cmdcommon.NewKymaConfig()

	alpha := alpha.NewAlphaCMD(kymaConfig)

	cmd.AddCommand(alpha)
	cmd.AddCommand(version.NewCmd())
	cmd.AddCommand(module.NewModuleCMD(kymaConfig))
	cmd.AddCommand(app.NewAppCMD(kymaConfig))

	builder := extensions.NewBuilder(kymaConfig)
	builder.Build(cmd, extensionstypes.ActionsMap{
		"function_init":         actions.NewFunctionInit(kymaConfig),
		"registry_config":       actions.NewRegistryConfig(kymaConfig),
		"registry_image_import": actions.NewRegistryImageImport(kymaConfig),
		"resource_create":       actions.NewResourceCreate(kymaConfig),
		"resource_get":          actions.NewResourceGet(kymaConfig),
		"resource_delete":       actions.NewResourceDelete(kymaConfig),
		"resource_explain":      actions.NewResourceExplain(),
		"call_files_to_save":    actions.NewCallFilesToSaveAction(kymaConfig),
	})
	builder.DisplayWarnings(cmd.ErrOrStderr())

	return cmd
}
