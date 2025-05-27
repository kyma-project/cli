package alpha

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/app"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/hana"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/kubeconfig"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/provision"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/referenceinstance"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions"
	"github.com/kyma-project/cli.v3/internal/extensions/actions"
	extensionstypes "github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

func NewAlphaCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "alpha <command> [flags]",
		Short:                 "Groups command prototypes for which the API may still change",
		Long:                  `A set of alpha prototypes that may still change. Use them in automation at your own risk.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(app.NewAppCMD(kymaConfig))
	cmd.AddCommand(hana.NewHanaCMD(kymaConfig))
	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(referenceinstance.NewReferenceInstanceCMD(kymaConfig))
	cmd.AddCommand(kubeconfig.NewKubeconfigCMD(kymaConfig))

	builder := extensions.NewBuilder(kymaConfig)
	builder.Build(cmd, extensionstypes.ActionsMap{
		"function_init":         actions.NewFunctionInit(kymaConfig),
		"registry_config":       actions.NewRegistryConfig(kymaConfig),
		"registry_image_import": actions.NewRegistryImageImport(kymaConfig),
		"resource_create":       actions.NewResourceCreate(kymaConfig),
		"resource_get":          actions.NewResourceGet(kymaConfig),
		"resource_delete":       actions.NewResourceDelete(kymaConfig),
		"resource_explain":      actions.NewResourceExplain(),
	})
	builder.DisplayWarnings(cmd.ErrOrStderr())

	return cmd
}
