package templates

import (
	"fmt"

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
