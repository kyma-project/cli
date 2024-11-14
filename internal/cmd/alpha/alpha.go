package alpha

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/access"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/add"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/app"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/hana"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/modules"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/oidc"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/provision"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/referenceinstance"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/registry/config"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/registry/imageimport"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/remove"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewAlphaCMD() (*cobra.Command, clierror.Error) {
	cmd := &cobra.Command{
		Use:                   "alpha",
		Short:                 "Groups command prototypes the API for which may still change.",
		Long:                  `A set of alpha prototypes that may still change. Use in automations on your own risk.`,
		DisableFlagsInUseLine: true,
	}

	kymaConfig, err := cmdcommon.NewKymaConfig(cmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(access.NewAccessCMD(kymaConfig))
	cmd.AddCommand(add.NewAddCMD(kymaConfig))
	cmd.AddCommand(app.NewAppCMD(kymaConfig))
	cmd.AddCommand(hana.NewHanaCMD(kymaConfig))
	cmd.AddCommand(modules.NewModulesCMD(kymaConfig))
	cmd.AddCommand(oidc.NewOIDCCMD(kymaConfig))
	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(referenceinstance.NewReferenceInstanceCMD(kymaConfig))
	cmd.AddCommand(remove.NewRemoveCMD(kymaConfig))

	cmds := kymaConfig.BuildExtensions(&cmdcommon.TemplateCommandsList{
		// list of template commands deffinitions
		Explain: templates.BuildExplainCommand,
	}, cmdcommon.CoreCommandsMap{
		// map of available core commands
		"registry_config":       config.NewConfigCMD,
		"registry_image-import": imageimport.NewImportCMD,
	})
	cmd.AddCommand(cmds...)

	return cmd, nil
}
