package extensions

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/spf13/cobra"
)

func buildCommand(extension types.Extension, availableActions types.ActionsMap) (*cobra.Command, error) {
	var errs []error

	// build command
	cmd, err := buildSingleCommand(extension, availableActions)
	if err != nil {
		errs = append(errs, errors.Wrapf(err, "failed to build command '%s'", extension.Metadata.Name))
	}

	// build sub-commands
	for _, subExtension := range extension.SubCommands {
		subCmd, err := buildSubCommand(subExtension, availableActions, extension.Config)
		if err != nil {
			errs = append(errs, err)
		}

		cmd.AddCommand(subCmd)
	}

	return cmd, errors.NewList(errs...)
}

func buildSubCommand(subCommand types.Extension, availableActions types.ActionsMap, parentConfig types.ActionConfig) (*cobra.Command, error) {
	err := parameters.MergeMaps(parentConfig, subCommand.Config)
	if err != nil {
		return nil, err
	}

	return buildCommand(subCommand, availableActions)
}

func buildSingleCommand(extension types.Extension, availableActions types.ActionsMap) (*cobra.Command, error) {
	var errs []error

	cmd := &cobra.Command{
		Use:   extension.Metadata.Name,
		Short: extension.Metadata.Description,
		Long:  extension.Metadata.DescriptionLong,
	}

	if extension.Action == "" {
		return cmd, errors.NewList(errs...)
	}

	// set flags
	values := []parameters.Value{}
	requiredFlags := []string{}
	for _, extensionFlag := range extension.Flags {
		cmdFlag := buildFlag(extensionFlag)
		if cmdFlag.warning != nil {
			errs = append(errs, errors.Newf("flag '%s' error: %s", extensionFlag.Name, cmdFlag.warning.Error()))
		}

		if extensionFlag.Required {
			requiredFlags = append(requiredFlags, extensionFlag.Name)
		}

		cmd.Flags().AddFlag(cmdFlag.pflag)
		values = append(values, cmdFlag.value)
	}

	// set args
	args := buildArgs(extension.Args)
	cmd.Args = args.run
	values = append(values, args.value)

	// set action runs
	action, ok := availableActions[extension.Action]
	if !ok {
		// unexpected behavior because actions list is validated and should be filled only with available data
		errs = append(errs, errors.Newf("action '%s' not found", extension.Action))
		return cmd, errors.NewList(errs...)
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

	return cmd, errors.NewList(errs...)
}
