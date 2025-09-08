package extensions

import (
	"github.com/kyma-project/cli.v3/internal/extensions/errors"
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/cobra"
)

type args struct {
	run   func(*cobra.Command, []string) error
	value parameters.Value
}

func buildArgs(extensionArgs *types.Args, overwrites map[string]interface{}) args {
	if extensionArgs == nil {
		return args{}
	}

	value := parameters.NewTyped(extensionArgs.Type, ".args.value")

	// append args to overwrites
	overwrites["args"] = map[string]interface{}{
		"type":     extensionArgs.Type,
		"optional": extensionArgs.Optional,
		"value":    value.GetValue(),
	}

	return args{
		value: value,
		run: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				for _, sub := range cmd.Commands() {
					if sub.Name() == args[0] {
						return nil
					}
				}
			}

			if extensionArgs.Optional {
				return setOptionalArg(value, args)
			}
			return setRequiredArg(value, args)
		},
	}
}

func setOptionalArg(value parameters.Value, args []string) error {
	if len(args) > 1 {
		return errors.Newf("accepts at most one argument, received %d", len(args))
	}

	if len(args) != 0 {
		return value.Set(args[0])
	}

	return nil
}

func setRequiredArg(value parameters.Value, args []string) error {
	if len(args) != 1 {
		return errors.Newf("requires exactly one argument, received %d", len(args))
	}

	return value.Set(args[0])
}
