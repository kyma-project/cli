package module

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/spf13/cobra"
)

type pullConfig struct {
	*cmdcommon.KymaConfig

	moduleName string
	namespace  string
	remote     string
	version    string
	force      bool
}

func NewPullV2CMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := pullConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "pull <module-name> [flags]",
		Short: "Pull a module from a remote repository",
		Long: `Pull a module from a remote repository to make it available for installation on the cluster.

This command downloads module templates and resources from remote repositories,
making them available locally for subsequent installation. Community modules
must be pulled before they can be installed using the 'kyma module add' command.`,
		Example: `  # Pull a specific community module
  kyma module pull community-module-name

  # Pull the latest version of a module into specific namespace
  kyma module pull community-module-name --namespace module-namespace

  # Pull a module with a specific version into specific namespace
  kyma module pull community-module-name --version v1.0.0 --namespace module-namespace`,
		Args: cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			cfg.moduleName = args[0]
			clierror.Check(pull(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "default", "Destination namespace there the module is stored")
	cmd.Flags().StringVar(&cfg.remote, "remote", "", "Specifies the community modules repository URL (defaults to official repository)")
	cmd.Flags().StringVarP(&cfg.version, "version", "v", "", "Specify version of the community module to pull")
	cmd.Flags().BoolVar(&cfg.force, "force", false, "Forces application of the binding, overwriting if it already exists")

	return cmd
}

func pull(cfg *pullConfig) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(cmdcommon.NewKymaConfig())
	pullOperation, err := moduleOperations.Pull()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the pull command"))
	}

	err = pullOperation.Run(cfg.Ctx)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to pull the community module template into the target Kyma environment"))
	}

	return nil
}
