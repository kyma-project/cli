package clierror

import (
	"os"

	"github.com/kyma-project/cli.v3/internal/out"
)

// Check prints error and executes os.Exit(1) if the error is not nil
func Check(err Error) {
	if err != nil {
		out.Errln(err.String())
		os.Exit(1)
	}
}
