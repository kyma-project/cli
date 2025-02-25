package templates

import (
	"bytes"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmd/alpha/templates/types"
	"github.com/stretchr/testify/require"
)

func Test_explain(t *testing.T) {
	t.Run("print output string", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte{})
		cmd := buildExplainCommand(buf, &ExplainOptions{
			ExplainCommand: types.ExplainCommand{
				Description:     "test explain command",
				DescriptionLong: "this is test explain command",
				Output:          "test output",
			},
		})

		require.Equal(t, "explain [flags]", cmd.Use)
		require.Equal(t, "test explain command", cmd.Short)
		require.Equal(t, "this is test explain command", cmd.Long)

		err := cmd.Execute()
		require.NoError(t, err)

		require.Equal(t, "test output\n", buf.String())
	})
}
