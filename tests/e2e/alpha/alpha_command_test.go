package alpha

import (
	"os"
	"testing"

	"gotest.tools/v3/golden"
	"gotest.tools/v3/icmd"
)

const (
	kymaCommand  = "kyma"
	alphaCommand = "alpha"
)

func TestAlphaCommand(t *testing.T) {
	// Set up the environment if needed
	os.Setenv("MY_CLI_APP_ENV", "test")

	// Execute the CLI command
	result := icmd.RunCmd(icmd.Command(kymaCommand, alphaCommand))

	result.Assert(t, icmd.Success)
	golden.Assert(t, result.Stderr(), "alpha-ok.golden")

}
