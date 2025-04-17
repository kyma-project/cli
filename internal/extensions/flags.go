package extensions

import (
	"fmt"
	"strings"

	"github.com/kyma-project/cli.v3/internal/extensions/parameters"
	"github.com/kyma-project/cli.v3/internal/extensions/types"
	"github.com/spf13/pflag"
)

type flag struct {
	pflag   *pflag.Flag
	value   parameters.Value
	warning error
}

func buildFlag(commandFlag types.Flag, overwrites map[string]interface{}) flag {
	flagOverwriteName := strings.ReplaceAll(commandFlag.Name, "-", "")
	valuePath := fmt.Sprintf(".flags.%s.value", flagOverwriteName)
	value := parameters.NewTyped(commandFlag.Type, valuePath)
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

	// append flag to overwrites
	flagsOverwrites := overwrites["flags"].(map[string]interface{})
	flagsOverwrites[flagOverwriteName] = map[string]interface{}{
		"type":        commandFlag.Type,
		"name":        pflag.Name,
		"shorthand":   pflag.Shorthand,
		"description": pflag.Usage,
		"default":     pflag.DefValue,
		"value":       value.GetValue(),
	}
	overwrites["flags"] = flagsOverwrites

	return flag{
		pflag:   pflag,
		value:   value,
		warning: warning,
	}
}
