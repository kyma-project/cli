package module

import (
	"fmt"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/spf13/cobra"
)

type manageConfig struct {
	*cmdcommon.KymaConfig

	module string
	policy string
}

func newManageCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := manageConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "manage <module>",
		Short: "Manage module.",
		Long:  "Use this command to manage an existing module.",

		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(cfg.validate())
		},
		Run: func(cmd *cobra.Command, args []string) {
			clierror.Check(runManage(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.module, "module", "", "Name of the module to manage")
	cmd.Flags().StringVar(&cfg.policy, "policy", "CreateAndDelete", "Set custom resource policy. (Possible values: CreateAndDelete, Ignore)")
	_ = cmd.MarkFlagRequired("module")
	return cmd
}

func (mc *manageConfig) validate() clierror.Error {
	if mc.policy != "CreateAndDelete" && mc.policy != "Ignore" {
		return clierror.New(fmt.Sprintf("invalid policy %q, only CreateAndDelete and Ignore are allowed", mc.policy))
	}

	return nil
}

func runManage(cfg *manageConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	err := client.Kyma().ManageModule(cfg.Ctx, cfg.module, cfg.policy)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to set module as managed"))
	}
	err = client.Kyma().WaitForModuleState(cfg.Ctx, cfg.module, "Ready", "Warning")
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to check module state"))
	}
	return nil
}
