package extensions

import (
	"errors"
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/spf13/cobra"
)

func buildCommand(kymaConfig *cmdcommon.KymaConfig, extension types.Extension, availableActions types.ActionsMap) (*cobra.Command, error) {
	var buildError error

	cmd := &cobra.Command{
		Use:   extension.Metadata.Name,
		Short: extension.Metadata.Description,
		Long:  extension.Metadata.DescriptionLong,
	}

	// build sub-commands
	for _, subCommand := range extension.SubCommands {
		subCmd, subErr := buildSubCommand(kymaConfig, subCommand, availableActions, extension.Config)
		if subErr != nil {
			buildError = errors.Join(buildError,
				fmt.Errorf("failed to build sub-command '%s': %s", subCommand.Metadata.Name, subErr.Error()))
		}

		cmd.AddCommand(subCmd)
	}

	if extension.Action == "" {
		return cmd, buildError
	}

	// set flags
	values := []parameters.Value{}
	requiredFlags := []string{}
	for _, extensionFlag := range extension.Flags {
		flag := buildFlag(extensionFlag)
		if flag.warning != nil {
			buildError = errors.Join(buildError,
				fmt.Errorf("failed to build flag '%s' for '%s' command: %s", extensionFlag.Name, extension.Metadata.Name, flag.warning.Error()))
		}

		if extensionFlag.Required {
			requiredFlags = append(requiredFlags, extensionFlag.Name)
		}

		cmd.Flags().AddFlag(flag.pflag)
		values = append(values, flag.value)
	}

	// set args
	args := buildArgs(extension.Args)
	cmd.Args = args.run
	values = append(values, args.value)

	// set action runs
	action, ok := availableActions[extension.Action]
	if !ok {
		// unexpected behavior because actions list is validated and should be filled only with available data
		return cmd, buildError
	}

	cmd.PreRun = func(_ *cobra.Command, _ []string) {
		// check required flags
		clierror.Check(flags.Validate(cmd.Flags(),
			flags.MarkRequired(requiredFlags...),
		))
		// set parameters from flag
		clierror.Check(parameters.Set(extension.Config, values))

		// configure action
		clierror.Check(action.Configure(extension.Config))
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// run action
		clierror.Check(action.Run(cmd, args))
	}

	return cmd, buildError
}

func buildSubCommand(kymaConfig *cmdcommon.KymaConfig, subCommand types.Extension, availableActions types.ActionsMap, parentConfig types.ActionConfig) (*cobra.Command, error) {
	err := parameters.MergeMaps(parentConfig, subCommand.Config)
	if err != nil {
		return nil, err
	}

	return buildCommand(kymaConfig, subCommand, availableActions)
}
