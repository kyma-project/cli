package schema

import (
	"os"
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/kyma-project/hydroform/function/pkg/workspace"
	"github.com/stretchr/testify/require"
)

// TestFunctionFlags ensures that the provided command flags are stored in the options.
func TestFunctionFlags(t *testing.T) {
	t.Parallel()
	o := NewOptions(&cli.Options{}, os.Stdout, map[string]func() ([]byte, error){
		"serverless": workspace.ReflectSchema,
	})
	c := NewCmd(o)

	// test passing flags
	err := c.ParseFlags([]string{
		"serverless",
	})
	require.NoError(t, err, "Parsing flags should not return an error")
}
