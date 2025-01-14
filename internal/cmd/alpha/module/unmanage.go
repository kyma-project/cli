package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type unmanageConfig struct {
	*cmdcommon.KymaConfig

	module string
}

func newUnmanageCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := unmanageConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "unmanage <module>",
		Short: "Unmanage module.",
		Long:  "Use this command to unmanage an existing module.",
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(runUnmanage(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.module, "module", "", "Name of the module to unmanage")
	_ = cmd.MarkFlagRequired("module")
	return cmd
}

func runUnmanage(cfg *unmanageConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := client.Kyma().UnmanageModule(cfg.Ctx, cfg.module)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to set module as unmanaged"))
	}

	err = client.Kyma().WaitForModuleState(cfg.Ctx, cfg.module, "Unmanaged")
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to check module state"))
	}
	return nil
}
