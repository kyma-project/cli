package module

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type addConfig struct {
	*cmdcommon.KymaConfig
	module      string
	channel     string
	crPath      string
	defaultCR   bool
	autoApprove bool
	community   bool
	version     string
}

func newAddCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := addConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "add <module> [flags]",
		Short: "Add a module",
		Long:  "Use this command to add a module.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkMutuallyExclusive("cr-path", "default-cr"),
				flags.MarkPrerequisites("auto-approve", "community"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runAdd(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.channel, "channel", "c", "", "Name of the Kyma channel to use for the module")
	cmd.Flags().StringVar(&cfg.crPath, "cr-path", "", "Path to the custom resource file")
	cmd.Flags().BoolVar(&cfg.defaultCR, "default-cr", false, "Deploys the module with the default CR")
	cmd.Flags().BoolVar(&cfg.autoApprove, "auto-approve", false, "Automatically approve community module installation")
	cmd.Flags().StringVar(&cfg.version, "version", "", "Specify version of the community module to install")
	cmd.Flags().BoolVar(&cfg.community, "community", false, "Install a community module (no official support, no binding SLA)")

	return cmd
}

func runAdd(cfg *addConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	crs, clierr := loadCustomCRs(cfg.crPath)
	if clierr != nil {
		return clierr
	}

	return addModule(cfg, &client, crs...)
}

func loadCustomCRs(crPath string) ([]unstructured.Unstructured, clierror.Error) {
	if crPath == "" {
		// skip if not set
		return nil, nil
	}

	crs, err := resources.ReadFromFiles(crPath)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to read object from file"))
	}

	return crs, nil
}

func addModule(cfg *addConfig, client *kube.Client, crs ...unstructured.Unstructured) clierror.Error {
	if cfg.community {
		return installCommunityModule(cfg, client, crs...)
	}

	return modules.Enable(cfg.Ctx, *client, cfg.module, cfg.channel, cfg.defaultCR, crs...)
}

func installCommunityModule(cfg *addConfig, client *kube.Client, crs ...unstructured.Unstructured) clierror.Error {
	fmt.Println("Warning:\n  You are about to install a community module.\n" +
		"  Community modules are not officially supported and come with no binding Service Level Agreement (SLA).\n" +
		"  There is no guarantee of support, maintenance, or compatibility.")

	if !cfg.autoApprove {
		proceedPrompt := prompt.NewBool("\nAre you sure you want to proceed with the installation?", true)
		proceedWithInstallation, err := proceedPrompt.Prompt()
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to prompt for the user confirmation", "if error repeats, consider running the command with --auto-approve flag"))
		}
		if !proceedWithInstallation {
			return nil
		}
	}

	versionToInstall, err := selectCommunityModuleVersion(cfg, client)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to prompt for module version", "if error repeats, consider running the command with --version flag"))
	}

	installData := modules.InstallCommunityModuleData{
		ModuleName:            cfg.module,
		Version:               versionToInstall,
		IsDefaultCRApplicable: cfg.defaultCR,
		CustomResources:       crs,
	}

	return modules.Install(cfg.Ctx, *client, installData)
}

func selectCommunityModuleVersion(cfg *addConfig, client *kube.Client) (string, error) {
	if strings.TrimSpace(cfg.version) != "" {
		return cfg.version, nil
	}

	availableVersions, err := modules.ListAvailableVersions(cfg.Ctx, *client, cfg.module, cfg.community)
	if err != nil {
		return "", err
	}

	if len(availableVersions) == 1 {
		return availableVersions[0], nil
	}

	versionPrompt := prompt.NewOneOfStringList("Choose one of the available versions:", "Type the version number: ", availableVersions)
	return versionPrompt.Prompt()
}
