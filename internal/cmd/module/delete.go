package module

import (
	"bytes"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/spf13/cobra"
)

type deleteConfig struct {
	*cmdcommon.KymaConfig
	autoApprove bool
	community   bool

	module string
}

func newDeleteCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := deleteConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:     "delete <module> [flags]",
		Short:   "Deletes a module",
		Long:    "Use this command to delete a module.",
		Aliases: []string{"del"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runDelete(&cfg))
		},
	}

	cmd.Flags().BoolVar(&cfg.autoApprove, "auto-approve", false, "Automatically approves module removal")
	cmd.Flags().BoolVar(&cfg.community, "community", false, "Delete the community module (if set, the operation targets a community module instead of a core module)")

	return cmd
}

func runDelete(cfg *deleteConfig) clierror.Error {
	client, clierr := cfg.GetKubeClientWithClierr()
	if clierr != nil {
		return clierr
	}
	moduleTemplatesRepo := repo.NewModuleTemplatesRepo(client)

	if cfg.community && !cfg.autoApprove {
		runningResources, clierr := modules.GetRunningResourcesOfCommunityModule(cfg.Ctx, moduleTemplatesRepo, cfg.module)
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

	if !cfg.community && !cfg.autoApprove {
		confirmationPrompt := prompt.NewBool(prepareCorePromptMessage(cfg.module), false)
		confirmation, err := confirmationPrompt.Prompt()
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to prompt for user input", "if error repeats, consider running the command with --auto-approve flag"))
		}

		if !confirmation {
			return nil
		}
	}

	return modules.Disable(cfg.Ctx, client, moduleTemplatesRepo, cfg.module, cfg.community)
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
