package alpha

import (
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/authorize"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/diagnose"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/hana"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/kubeconfig"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/module"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/provision"
	"github.com/kyma-project/cli.v3/internal/cmd/alpha/referenceinstance"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

func NewAlphaCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "alpha <command> [flags]",
		Short:                 "Groups command prototypes for which the API may still change",
		Long:                  `A set of alpha prototypes that may still change. Use them in automation at your own risk.`,
		DisableFlagsInUseLine: true,
	}

	cmd.AddCommand(authorize.NewAuthorizeCMD(kymaConfig))
	cmd.AddCommand(hana.NewHanaCMD(kymaConfig))
	cmd.AddCommand(provision.NewProvisionCMD())
	cmd.AddCommand(referenceinstance.NewReferenceInstanceCMD(kymaConfig))
	cmd.AddCommand(kubeconfig.NewKubeconfigCMD(kymaConfig))
	cmd.AddCommand(diagnose.NewDiagnoseCMD(kymaConfig))
	cmd.AddCommand(module.NewModuleCMD(kymaConfig))

	return cmd
}
