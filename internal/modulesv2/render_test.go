package modulesv2

import (
	"bytes"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
)

func TestRenderList_Table(t *testing.T) {
	results := []dtos.ListResult{
		{Name: "api-gateway", Version: "3.5.1", Channel: "regular"},
	}

	var buf bytes.Buffer
	err := RenderList(results, types.DefaultFormat, out.NewToWriter(&buf))

	require.NoError(t, err)
	require.Regexp(t, `MODULE.*VERSION.*CHANNEL`, buf.String())
	require.Regexp(t, `api-gateway.*3\.5\.1.*regular`, buf.String())
}

func TestRenderList_JSON(t *testing.T) {
	results := []dtos.ListResult{
		{Name: "api-gateway", Version: "3.5.1", Channel: "regular"},
	}

	var buf bytes.Buffer
	err := RenderList(results, types.JSONFormat, out.NewToWriter(&buf))

	require.NoError(t, err)
	require.JSONEq(t, `[{"name":"api-gateway","version":"3.5.1","channel":"regular"}]`, buf.String())
}
