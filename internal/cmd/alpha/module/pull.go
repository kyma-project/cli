package module

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/out"
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
  kyma alpha module pull community-module-name

  # Pull the latest version of a module into specific namespace
  kyma alpha module pull community-module-name --namespace module-namespace

  # Pull a module with a specific version into specific namespace
  kyma alpha module pull community-module-name --version v1.0.0 --namespace module-namespace`,
		Args: cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			cfg.moduleName = args[0]
			clierror.Check(pull(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "default", "Destination namespace there the module is stored")
	cmd.Flags().StringVar(&cfg.remote, "remote", "", "Specifies the community modules repository URL (defaults to official repository)")
	cmd.Flags().StringVarP(&cfg.version, "version", "v", "", "Specify version of the community module to pull")
	cmd.Flags().BoolVar(&cfg.force, "force", false, "Forces application of the module template, overwriting if it already exists")

	return cmd
}

func pull(cfg *pullConfig) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(cmdcommon.NewKymaConfig())
	pullOperation, err := moduleOperations.Pull()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the pull command"))
	}

	pullConfigDto := dtos.NewPullConfig(cfg.moduleName, cfg.namespace, cfg.remote, cfg.version, cfg.force)

	err = pullOperation.Run(cfg.Ctx, pullConfigDto)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to pull the community module template into the target Kyma environment"))
	}

	out.Msgfln("Module %s pulled successfully into namespace '%s'\n", cfg.moduleName, cfg.namespace)
	out.Msgf("%s", getWarningTextForCommunityModuleUsage(pullConfigDto))

	return nil
}

func getWarningTextForCommunityModuleUsage(pullConfigDto *dtos.PullConfig) string {
	return "WARNING: Community Module\n" +
		"This is a community module that is not officially supported by the Kyma team.\n" +
		"Community modules:\n" +
		"  • Do not guarantee compatibility or stability\n" +
		"  • Are not covered by any Service Level Agreement (SLA)\n" +
		"  • May contain security vulnerabilities or bugs\n" +
		"  • Are maintained by the community, not by SAP\n\n" +
		"Use community modules at your own risk in production environments.\n\n" +
		"Next Steps:\n" +
		"To install this module on your cluster, you can use the sample command:\n\n" +
		"  # Install with default configuration:\n" +
		fmt.Sprintf("  kyma module add %s/%s --default-cr\n\n", pullConfigDto.Namespace, pullConfigDto.ModuleName) +
		"For more information about module installation, run: kyma module add --help\n"
}
