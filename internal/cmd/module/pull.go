package module

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/spf13/cobra"
)

type pullConfig struct {
	*cmdcommon.KymaConfig

	moduleName string
	namespace  string
	version    string
}

func newPullCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := pullConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "pull <module-name> [flags]",
		Short: "Pull a module from remote repository",
		Long: `Pull a module from a remote repository to make it available for installation on the cluster.

This command downloads module templates and resources from remote repositories,
making them available locally for subsequent installation. Community modules
must be pulled before they can be installed using the 'kyma module add' command.

Examples:
  # Pull a specific community module
  kyma module pull my-community-module

  # Pull a module with a specific version into specific namespace
  kyma module pull my-module --version v1.0.0 --namespace my-namespace`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.moduleName = args[0]
			clierror.Check(pullModule(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.namespace, "namespace", "n", "default", "Destination namespace where module will be stored")
	cmd.Flags().StringVarP(&cfg.version, "version", "v", "", "Specify version of the community module to pull")

	return cmd
}

func pullModule(cfg *pullConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	if cfg.namespace == "kyma-system" {
		return clierror.New(getErrorTextForInvalidNamespace(cfg.moduleName))
	}

	moduleTemplatesRepo := repo.NewModuleTemplatesRepo(client)
	moduleTemplate, err := modules.GetModuleTemplateFromRemote(cfg.Ctx, moduleTemplatesRepo, cfg.moduleName, cfg.version)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to pull image from the community modules repository"))
	}

	err = modules.PersistModuleTemplateInNamespace(cfg.Ctx, client, moduleTemplate, cfg.namespace)

	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to store module in the provided namespace"))
	}

	fmt.Printf("Module %s pulled successfully into namespace '%s'\n\n", cfg.moduleName, cfg.namespace)
	fmt.Printf("%s", getWarningTextForCommunityModuleUsage(moduleTemplate))

	return nil
}

func getErrorTextForInvalidNamespace(moduleName string) string {
	return "Cannot pull modules into the 'kyma-system' namespace.\n\n" +
		"The 'kyma-system' namespace is reserved for core Kyma components and system modules. " +
		"Pulling community modules into this namespace could cause conflicts with system resources " +
		"and may impact cluster stability.\n\n" +
		"Please specify a different namespace using the --namespace flag, or use the default namespace by omitting the flag.\n\n" +
		"Example: kyma module pull " + moduleName + " --namespace my-modules"
}

func getWarningTextForCommunityModuleUsage(moduleTemplate *kyma.ModuleTemplate) string {
	return "WARNING: Community Module\n" +
		"This is a community module that is not officially supported by the Kyma team.\n" +
		"Community modules:\n" +
		"  • Do not guarantee compatibility or stability\n" +
		"  • Are not covered by any Service Level Agreement (SLA)\n" +
		"  • May contain security vulnerabilities or bugs\n" +
		"  • Are maintained by the community, not by SAP\n\n" +
		"Use community modules at your own risk in production environments.\n\n" +
		"Next Steps:\n" +
		"To install this module on your cluster, you can use the sample command:\n" +
		"  # Install with default configuration:\n" +
		fmt.Sprintf("  kyma module add %s --origin %s/%s --default-cr\n\n", moduleTemplate.Spec.ModuleName, moduleTemplate.Namespace, moduleTemplate.Name) +
		"For more information about module installation, run: kyma module add --help\n"
}
