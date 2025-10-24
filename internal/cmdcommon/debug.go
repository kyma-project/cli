package cmdcommon

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/spf13/cobra"
)

func AddPersistentDebugFlag(cmd *cobra.Command) {
	// set hidden flag to allow its usage
	_ = cmd.PersistentFlags().Bool("debug", false, "")
	_ = cmd.PersistentFlags().MarkHidden("debug")
}

func SetupOutput(cmd *cobra.Command) {
	out.SetForCmd(cmd)
	if getBoolFlagValue("--debug") {
		out.EnableDebug()
	}
}

// search os.Args manually to find if user pass --<flag_name> path and return its value
func getBoolFlagValue(flag string) bool {
	for i, arg := range os.Args {
		// example: --flag true
		if arg == flag && len(os.Args) > i+1 {
			// ignore error (use default value)
			value, err := strconv.ParseBool(os.Args[i+1])
			if err == nil {
				return value
			}
		}

		// example: --flag
		if arg == flag {
			return true
		}

		// example: --flag=true
		argFields := strings.Split(arg, "=")
		if strings.HasPrefix(arg, fmt.Sprintf("%s=", flag)) && len(argFields) == 2 {
			// ignore error and use default false
			value, err := strconv.ParseBool(argFields[1])
			if err == nil {
				return value
			}
		}
	}

	return false
}
