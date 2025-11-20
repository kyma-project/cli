package module

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/spf13/cobra"
)

type deleteConfig struct {
	*cmdcommon.KymaConfig
	autoApprove bool
	community   bool

	module     string
	modulePath string
}

func newDeleteCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := deleteConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "delete <module> [flags]",
		Short: "Deletes a module",
		Long:  "Use this command to delete a module.",
		Example: `  # Delete kyma module and auto-approve the deletion
  kyma module delete serverless --auto-approve

  # Delete community module
  kyma module delete my-namespace/my-community-module-1.0.0`,

		Aliases: []string{"del"},
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			clierror.Check(flags.Validate(cmd.Flags(),
				flags.MarkUnsupported("community", "the --community flag is no longer supported - specify community module to delete using argument"),
			))
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg.complete(args)
			clierror.Check(runDelete(&cfg))
		},
	}

	cmd.Flags().BoolVar(&cfg.autoApprove, "auto-approve", false, "Automatically approves module removal")
	cmd.Flags().BoolVar(&cfg.community, "community", false, "Delete the community module (if set, the operation targets a community module instead of a core module)")
	_ = cmd.Flags().MarkHidden("community")

	return cmd
}

func (c *deleteConfig) complete(args []string) {
	if strings.Contains(args[0], "/") {
		// arg is module location in format <namespace>/<module-template-name>
		c.modulePath = args[0]
		return
	}

	// arg is module name
	c.module = args[0]
}

func runDelete(cfg *deleteConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}

	if cfg.modulePath != "" {
		return uninstallCommunityModule(cfg, client)
	}

	return disableModule(cfg, client)
}

func uninstallCommunityModule(cfg *deleteConfig, client kube.Client) clierror.Error {
	repo := repo.NewModuleTemplatesRepo(client)
	namespace, moduleTemplateName, err := validateOrigin(cfg.modulePath)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to identify the community module"))
	}

	communityModuleTemplate, err := modules.FindCommunityModuleTemplate(cfg.Ctx, namespace, moduleTemplateName, repo)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to retrieve the module '%s/%s'", namespace, moduleTemplateName))
	}

	if !cfg.autoApprove {
		runningResources, clierr := modules.GetRunningResourcesOfCommunityModule(cfg.Ctx, repo, *communityModuleTemplate)
		if clierr != nil {
			return clierr
		}
		if len(runningResources) > 0 {
			confirmationPrompt := prompt.NewBool(prepareCommunityPromptMessage(runningResources), false)
			confirmation, err := confirmationPrompt.Prompt()
			if err != nil {
				return clierror.Wrap(err, clierror.New("failed to prompt for user input", "if error repeats, consider running the command with --auto-approve flag"))
			}

			if !confirmation {
				return nil
			}
		}
	}

	return modules.Uninstall(cfg.Ctx, repo, communityModuleTemplate)
}

func disableModule(cfg *deleteConfig, client kube.Client) clierror.Error {
	if !cfg.autoApprove {
		confirmationPrompt := prompt.NewBool(prepareCorePromptMessage(cfg.module), false)
		confirmation, err := confirmationPrompt.Prompt()
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to prompt for user input", "if error repeats, consider running the command with --auto-approve flag"))
		}

		if !confirmation {
			return nil
		}
	}

	return modules.Disable(cfg.Ctx, client, cfg.module)
}

func prepareCommunityPromptMessage(resourcesNames []string) string {
	var buf bytes.Buffer

	fmt.Fprint(&buf, "There are currently associated resources related to this module still running on the cluster:\n")
	for _, name := range resourcesNames {
		fmt.Fprintf(&buf, "  - %s\n", name)
	}
	fmt.Fprint(&buf, "\nDeleting the module may affect these resources.\n")
	fmt.Fprint(&buf, "Are you sure you want to proceed with the deletion?")

	return buf.String()
}

func prepareCorePromptMessage(moduleName string) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Are you sure you want to delete module %s?\n", moduleName)
	fmt.Fprintf(&buf, "Before you delete the %s module, make sure the module resources are no longer needed. This action also permanently removes the namespaces, service instances, and service bindings created by the module.\n", moduleName)
	fmt.Fprintf(&buf, "Are you sure you want to continue?")

	return buf.String()
}
