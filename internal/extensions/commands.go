package extensions

import (
	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/kyma-project/cli.v3/internal/flags"
	"github.com/spf13/cobra"
)

var (
	emptyActionRun       = func(cmd *cobra.Command, _ []string) error { return cmd.Help() }
	unsupportedActionRun = func(_ *cobra.Command, _ []string) {
		clierror.Check(clierror.New("unsupported action",
			"make sure the cli version is compatible with the extension"))
	}
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
		subCmd, err := buildCommand(subExtension, availableActions)
		if err != nil {
			errs = append(errs, err)
		}

		cmd.AddCommand(subCmd)
	}

	return cmd, errors.NewList(errs...)
}

func buildSingleCommand(extension types.Extension, availableActions types.ActionsMap) (*cobra.Command, error) {
	var errs []error

	cmd := &cobra.Command{
		Use:   extension.Metadata.Name,
		Short: extension.Metadata.Description,
		Long:  extension.Metadata.DescriptionLong,
	}

	if extension.Action == "" {
		// no action provided
		// set help command as default run
		cmd.RunE = emptyActionRun
		return cmd, errors.NewList(errs...)
	}

	// set flags
	overwrites := types.ActionConfigOverwrites{
		"flags": map[string]interface{}{},
	}
	values := []parameters.Value{}
	requiredFlags := []string{}
	for _, extensionFlag := range extension.Flags {
		cmdFlag := buildFlag(extensionFlag, overwrites)
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
	cmdArgs := buildArgs(extension.Args, overwrites)
	cmd.Args = cmdArgs.run
	values = append(values, cmdArgs.value)

	// set action runs
	action, ok := availableActions[extension.Action]
	if !ok {
		// action not found
		// set unsupported action run to inform user
		cmd.Run = unsupportedActionRun
		return cmd, errors.NewList(errs...)
	}
	
	cmd.PreRun = func(_ *cobra.Command, _ []string) {
		// check required flags
		clierror.Check(flags.Validate(cmd.Flags(),
			flags.MarkRequired(requiredFlags...),
		))
		// set parameters from flag and args as overwrites
		clierror.Check(parameters.Set(overwrites, values))

		// configure action
		clierror.Check(action.Configure(extension.Config, overwrites))
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		// run action
		clierror.Check(action.Run(cmd, args))
	}

	return cmd, errors.NewList(errs...)
}
