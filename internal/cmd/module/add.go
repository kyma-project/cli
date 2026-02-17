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
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/kyma-project/cli.v3/internal/modulesv2/precheck"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type addConfig struct {
	*cmdcommon.KymaConfig
	module      string
	modulePath  string
	channel     string
	crPath      string
	defaultCR   bool
	autoApprove bool
	community   bool
}

func newAddCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := addConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "add <module> [flags]",
		Short: "Add a module",
		Long:  "Use this command to add a module.",
		Example: `  # Add the Keda module with the default CR
  kyma module add keda --default-config-cr

  # Add the Keda module with a custom CR from a file
  kyma module add keda --config-cr-path ./keda-cr.yaml

  ## Add a community module with a default CR and auto-approve the SLA
  #  passed argument must be in the format <namespace>/<module-template-name>
  #  the module must be pulled from the catalog first using the 'kyma module pull' command
  kyma module add my-namespace/my-module-template-name --default-config-cr --auto-approve`,

		Args: cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkMutuallyExclusive("cr-path", "default-cr", "config-cr-path", "default-config-cr"),
				flags.MarkUnsupported("community", "the --community flag is no longer supported - community modules need to be pulled first using 'kyma module pull' command, then installed"),
				flags.MarkUnsupported("origin", "the --origin flag is no longer supported - use commands argument instead"),
			))
			clierror.Check(precheck.RequireCRD(kymaConfig, precheck.CmdGroupStable))
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg.complete(args)
			clierror.Check(runAdd(&cfg))
		},
	}

	cmd.Flags().StringVarP(&cfg.channel, "channel", "c", "", "Name of the Kyma channel to use for the module")
	cmd.Flags().StringVar(&cfg.crPath, "cr-path", "", "Path to the custom resource file")
	_ = cmd.Flags().MarkHidden("cr-path")
	cmd.Flags().StringVar(&cfg.crPath, "config-cr-path", "", "Path to the manifest file with custom configuration (alias: --cr-path)")
	cmd.Flags().BoolVar(&cfg.defaultCR, "default-cr", false, "Deploys the module with the default CR")
	_ = cmd.Flags().MarkHidden("default-cr")
	cmd.Flags().BoolVar(&cfg.defaultCR, "default-config-cr", false, "Deploys the module with default configuration (alias: --default-cr)")
	cmd.Flags().BoolVar(&cfg.autoApprove, "auto-approve", false, "Automatically approve community module installation")
	cmd.Flags().StringVar(&cfg.modulePath, "origin", "", "Specifies the source of the module (kyma or custom name)")
	_ = cmd.Flags().MarkHidden("origin")
	cmd.Flags().BoolVar(&cfg.community, "community", false, "Install a community module (no official support, no binding SLA)")
	_ = cmd.Flags().MarkHidden("community")

	return cmd
}

func (c *addConfig) complete(args []string) {
	if strings.Contains(args[0], "/") {
		// arg is module location in format <namespace>/<module-template-name>
		c.modulePath = args[0]
		return
	}

	// arg is module name
	c.module = args[0]
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
	moduleTemplatesRepo := repo.NewModuleTemplatesRepo(*client)

	if cfg.modulePath != "" {
		return installCommunityModule(cfg, client, moduleTemplatesRepo, crs...)
	}

	return modules.Enable(cfg.Ctx, *client, moduleTemplatesRepo, cfg.module, cfg.channel, cfg.defaultCR, crs...)
}

func installCommunityModule(cfg *addConfig, client *kube.Client, repo repo.ModuleTemplatesRepository, crs ...unstructured.Unstructured) clierror.Error {
	namespace, moduleTemplateName, err := validateOrigin(cfg.modulePath)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to identify the community module"))
	}

	communityModuleTemplate, err := modules.FindCommunityModuleTemplate(cfg.Ctx, namespace, moduleTemplateName, repo)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to install the community module"))
	}

	out.Msgln("Warning:\n  You are about to install a community module.\n" +
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

	installData := modules.InstallCommunityModuleData{
		CommunityModuleTemplate: communityModuleTemplate,
		IsDefaultCRApplicable:   cfg.defaultCR,
		CustomResources:         crs,
	}

	return modules.Install(cfg.Ctx, *client, repo, installData)
}

func validateOrigin(origin string) (string, string, error) {
	if !strings.Contains(origin, "/") {
		return "", "", fmt.Errorf("invalid origin format - expected <namespace>/<module-template-name>")
	}

	splitOrigin := strings.Split(origin, "/")
	if len(splitOrigin) != 2 {
		return "", "", fmt.Errorf("invalid origin format - expected <namespace>/<module-template-name>")
	}

	namespace := splitOrigin[0]
	moduleTemplateName := splitOrigin[1]

	return namespace, moduleTemplateName, nil
}
