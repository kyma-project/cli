package extensions

import (
	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/pflag"
)

type flag struct {
	pflag   *pflag.Flag
	value   parameters.Value
	warning error
}

func buildFlag(commandFlag types.Flag) flag {
	value := parameters.NewTyped(commandFlag.Type, commandFlag.ConfigPath)
	warning := value.SetValue(commandFlag.DefaultValue)

	pflag := &pflag.Flag{
		Name:      commandFlag.Name,
		Shorthand: commandFlag.Shorthand,
		Usage:     commandFlag.Description,
		Value:     value,
		DefValue:  value.String(),
	}

	if commandFlag.Type == parameters.BoolCustomType {
		// set default value for bool flag used without value (for example "--flag" instead of "--flag value")
		pflag.NoOptDefVal = "true"
	}

	return flag{
		pflag:   pflag,
		value:   value,
		warning: warning,
	}
}
