package module

import (
	"errors"
	"fmt"
	"maps"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/cmdcommon/prompt"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/kyma"
	"github.com/kyma-project/cli.v3/internal/modules"
	"github.com/kyma-project/cli.v3/internal/modules/repo"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
		Use:   "manage <module> [flags]",
		Short: "Sets the module to the managed state",
		Long:  "Use this command to set an existing module to the managed state.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(_ *cobra.Command, args []string) {
			clierror.Check(cfg.validate())
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg.module = args[0]
			clierror.Check(runManage(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.policy, "policy", "CreateAndDelete", "Sets a custom resource policy (Possible values: CreateAndDelete, Ignore)")

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

	exists, err := modules.ModuleExistsInKymaCR(cfg.Ctx, client, cfg.module)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to check if module exists in the target Kyma environment"))
	}

	if exists {
		return manageModuleInKyma(cfg, client)
	}

	if clierr = manageModuleMissingInKyma(cfg, client); clierr != nil {
		return clierr
	}

	return nil
}

func manageModuleInKyma(cfg *manageConfig, client kube.Client) clierror.Error {
	err := modules.ManageModuleInKymaCR(cfg.Ctx, client, cfg.module, cfg.policy)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to manage module in the target Kyma environment"))
	}
	fmt.Printf("Module %s set to managed\n", cfg.module)

	return nil
}

func manageModuleMissingInKyma(cfg *manageConfig, client kube.Client) clierror.Error {
	moduleTemplatesRepo := repo.NewModuleTemplatesRepo(client)

	err := modules.ManageModuleMissingInKyma(cfg.Ctx, client, moduleTemplatesRepo, cfg.module, cfg.policy)
	if err == nil {
		fmt.Printf("Module %s set to managed\n", cfg.module)
		return nil
	}
	if !errors.Is(err, modules.ErrModuleInstalledVersionNotInKymaChannel) {
		return clierror.Wrap(err, clierror.New("failed to set module as managed"))
	}

	// If not found, prompt for alternative channel
	channelsAndVersions, err := modules.GetAvailableChannelsAndVersions(cfg.Ctx, client, moduleTemplatesRepo, cfg.module)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to get available channels and versions"))
	}
	if len(channelsAndVersions) == 0 {
		return clierror.New("no available channels found for the module")
	}

	selectedChannel, clierr := promptForAlternativeChannel(channelsAndVersions)
	if clierr != nil {
		return clierr
	}

	defaultCrFlag := cfg.policy == kyma.CustomResourcePolicyCreateAndDelete

	clierr = modules.Enable(cfg.Ctx, client, moduleTemplatesRepo, cfg.module, selectedChannel, defaultCrFlag, []unstructured.Unstructured{}...)
	if clierr != nil {
		return clierr
	}

	fmt.Printf("Module %s set to managed (channel: %s)\n", cfg.module, selectedChannel)
	return nil
}

func promptForAlternativeChannel(channelsAndVersions map[string]string) (string, clierror.Error) {
	channelsIterator := maps.Keys(channelsAndVersions)
	var channelOpts []prompt.EnumValueWithDescription
	for channel := range channelsIterator {
		valWithDesc := prompt.NewEnumValWithDesc(channel, channelsAndVersions[channel])
		channelOpts = append(channelOpts, *valWithDesc)
	}

	fmt.Println("The version of the module you have installed is not available in the default Kyma channel.")
	fmt.Println("To proceed, please select one of the available channels below to manage the module with the desired version.")

	channelPrompt := prompt.NewOneOfEnumList("Available versions:\n", "Type the option number: ", channelOpts)
	selectedChannel, err := channelPrompt.Prompt()
	if err != nil {
		return "", clierror.Wrap(err, clierror.New("failed to prompt for channel"))
	}

	return selectedChannel, nil
}
