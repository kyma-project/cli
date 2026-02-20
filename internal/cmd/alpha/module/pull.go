package module

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/modulesv2"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/modulesv2/precheck"
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
		Short: "Pulls a module from a remote repository",
		Long: `Pulls a module from a remote repository to make it available for installation in the cluster.

This command downloads module templates and resources from remote repositories,
making them available locally for subsequent installation. Community modules
must be pulled before they can be installed using the 'kyma module add' command.`,
		Example: `  # Pull a specific community module
  kyma alpha module pull community-module-name

  # Pull the latest version of a module into specific namespace
  kyma alpha module pull community-module-name --namespace module-namespace

  # Pull a module with a specific version into specific namespace
  kyma alpha module pull community-module-name --version v1.0.0 --namespace module-namespace

  # Pull a module from a custom remote repository URL
  kyma alpha module pull community-module-name --remote-url https://example.com/modules.json`,
		Args: cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			clierror.Check(precheck.EnsureCRD(kymaConfig, cfg.force))
		},
		Run: func(_ *cobra.Command, args []string) {
			cfg.moduleName = args[0]
			clierror.Check(pull(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "default", "Destination namespace where the module is stored")
	cmd.Flags().StringVar(&cfg.remote, "remote-url", "", "URL to a file that contains ModuleTemplate CRs (defaults to official community catalog)")
	cmd.Flags().StringVarP(&cfg.version, "version", "v", "", "Specifies the version of the community module to pull")
	cmd.Flags().BoolVar(&cfg.force, "force", false, "Forces application of the module template, overwriting if it already exists")

	return cmd
}

func pull(cfg *pullConfig) clierror.Error {
	moduleOperations := modulesv2.NewModuleOperations(cmdcommon.NewKymaConfig())
	pullOperation, err := moduleOperations.Pull()
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to execute the pull command"))
	}

	pullConfigDto := dtos.NewPullConfig(cfg.moduleName, cfg.namespace, cfg.remote, cfg.version)

	if shouldAbort, clierr := confirmOverwriteIfNeeded(cfg, pullOperation, pullConfigDto); clierr != nil || shouldAbort {
		return clierr
	}

	pulledModule, err := pullOperation.Run(cfg.Ctx, pullConfigDto)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to pull the community module template into the target Kyma environment"))
	}

	out.Msgfln("Module %s pulled successfully into namespace '%s'\n", cfg.moduleName, cfg.namespace)
	out.Msgf("%s", getWarningTextForCommunityModuleUsage(pulledModule))

	return nil
}

// confirmOverwriteIfNeeded checks if the module is already installed and prompts the user for confirmation.
// Returns (shouldAbort, error) - if shouldAbort is true, the caller should return early.
func confirmOverwriteIfNeeded(cfg *pullConfig, pullOperation *modulesv2.PullService, pullConfigDto *dtos.PullConfig) (bool, clierror.Error) {
	if cfg.force {
		return false, nil
	}

	installedModule, err := pullOperation.GetInstalledModuleTemplate(cfg.Ctx, pullConfigDto)
	if err != nil {
		return true, clierror.Wrap(err, clierror.New("failed to check if module template is already installed"))
	}

	if installedModule == nil {
		return false, nil
	}

	confirmed, err := promptForInstallationConfirmation(installedModule.ModuleName, installedModule.Namespace)
	if err != nil {
		return true, clierror.Wrap(err, clierror.New("failed to read user confirmation"))
	}

	return !confirmed, nil
}

func promptForInstallationConfirmation(moduleName, namespace string) (bool, error) {
	out.Msgfln("\nModule template '%s' is already installed in namespace '%s'.", moduleName, namespace)
	out.Msgfln("Proceeding will overwrite the existing module template.\n")
	out.Msgfln("Tip: Use '--force' flag to bypass this confirmation prompt.\n")

	confirmPrompt := prompt.NewBool("Do you want to continue?", false)
	return confirmPrompt.Prompt()
}

func getWarningTextForCommunityModuleUsage(pulledModule *dtos.PullResult) string {
	return "WARNING:\n" +
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
		fmt.Sprintf("  kyma alpha module add %s/%s --default-config-cr\n\n", pulledModule.Namespace, pulledModule.ModuleTemplateName) +
		"For more information about module installation, run: kyma module add --help\n"
}
