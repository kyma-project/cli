package alpha

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/access"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/add"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/hana"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/imageimport"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/modules"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/oidc"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/provision"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/referenceinstance"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/registry"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/remove"
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
	config, err := cmdcommon.NewKymaConfig(cmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(hana.NewHanaCMD(config))
	cmd.AddCommand(imageimport.NewImportCMD(config))
	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(referenceinstance.NewReferenceInstanceCMD(config))
	cmd.AddCommand(access.NewAccessCMD(config))
	cmd.AddCommand(oidc.NewOIDCCMD(config))
	cmd.AddCommand(modules.NewModulesCMD(config))
	cmd.AddCommand(add.NewAddCMD(config))
	cmd.AddCommand(remove.NewRemoveCMD(config))
	cmd.AddCommand(registry.NewRegistryCMD(config))

	return cmd, nil
}
