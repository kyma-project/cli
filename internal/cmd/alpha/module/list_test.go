package module

import (
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/stretchr/testify/require"
)

func TestListCmd_Exists(t *testing.T) {
	cmd := NewListV2CMD(&cmdcommon.KymaConfig{})
	require.NotNil(t, cmd)
	require.Equal(t, "list [flags]", cmd.Use)
}
