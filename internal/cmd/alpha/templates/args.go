package templates

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/parameters"
	"github.com/spf13/cobra"
)

func AssignOptionalNameArg(name *string) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most one resource name as argument, received %d", len(args))
		}
		if len(args) == 1 {
			*name = args[0]
		}

		return nil
	}
}

func AssignRequiredNameArg(name parameters.Value) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("requires exactly one resource name as an argument, received %d", len(args))
		}

		return name.Set(args[0])
	}
}
