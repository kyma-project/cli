package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modulesv2/precheck"
	"github.com/kyma-project/cli.v3/internal/out"
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
		Use:   "unmanage <module> [flags]",
		Short: "Sets a module to the unmanaged state",
		Long:  "Use this command to set an existing module to the unmanaged state.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(precheck.RequireKLMManaged(kymaConfig, precheck.CmdGroupStable))
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runUnmanage(&cfg))
		},
	}

	return cmd
}

func runUnmanage(cfg *unmanageConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := client.Kyma().UnmanageModule(cfg.Ctx, cfg.module)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to set the module as unmanaged"))
	}

	err = client.Kyma().WaitForModuleState(cfg.Ctx, cfg.module, "Unmanaged")
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to check the module state"))
	}

	out.Msgfln("Module %s set to unmanaged", cfg.module)

	return nil
}
