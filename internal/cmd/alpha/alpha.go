package alpha

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/app"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/hana"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/kubeconfig"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/module"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/provision"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/referenceinstance"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/actions"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewAlphaCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "alpha <command> [flags]",
		Short:                 "Groups command prototypes for which the API may still change",
		Long:                  `A set of alpha prototypes that may still change. Use them in automation at your own risk.`,
		DisableFlagsInUseLine: true,
	}

	kymaConfig := cmdcommon.NewKymaConfig(cmd)

	cmd.AddCommand(app.NewAppCMD(kymaConfig))
	cmd.AddCommand(hana.NewHanaCMD(kymaConfig))
	cmd.AddCommand(module.NewModuleCMD(kymaConfig))
	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(referenceinstance.NewReferenceInstanceCMD(kymaConfig))
	cmd.AddCommand(kubeconfig.NewKubeconfigCMD(kymaConfig))

	cmds := kymaConfig.BuildExtensions(&cmdcommon.TemplateCommandsList{
		// list of template commands deffinitions
		Explain: templates.BuildExplainCommand,
		Get:     templates.BuildGetCommand,
		Create:  templates.BuildCreateCommand,
		Delete:  templates.BuildDeleteCommand,
	}, cmdcommon.ActionCommandsMap{
		"function_init":         actions.NewFunctionInit,
		"registry_config":       actions.NewRegistryConfig,
		"registry_image-import": actions.NewRegistryImageImport,
		"resource_create":       actions.NewResourceCreate,
		"resource_get":          actions.NewResourceGet,
		"resource_delete":       actions.NewResourceDelete,
		"resource_explain":      actions.NewResourceExplain,
	},
		cmd)

	kymaConfig.DisplayExtensionsErrors(cmd.ErrOrStderr())

	cmd.AddCommand(cmds...)

	return cmd
}
