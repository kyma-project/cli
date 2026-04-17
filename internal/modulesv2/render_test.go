package modulesv2

import (
	"bytes"
	"testing"

	"github.com/kyma-project/cli.v3/internal/cmdcommon/types"
	"github.com/kyma-project/cli.v3/internal/modulesv2/dtos"
	"github.com/kyma-project/cli.v3/internal/out"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRenderList_Table(t *testing.T) {
	results := []dtos.ListResult{
		{Name: "api-gateway", Version: "3.5.1", Channel: "regular", State: "Ready", Managed: true, CustomResourcePolicy: "CreateAndDelete", InstallationState: "Ready"},
	}

	var buf bytes.Buffer
	err := RenderList(results, types.DefaultFormat, out.NewToWriter(&buf))

	require.NoError(t, err)
	require.Regexp(t, `MODULE.*VERSION.*CR POLICY.*MANAGED.*MODULE STATUS.*INSTALLATION STATUS`, buf.String())
	require.Regexp(t, `api-gateway.*3\.5\.1\(regular\).*CreateAndDelete.*true.*Ready.*Ready`, buf.String())
}

func TestRenderList_JSON(t *testing.T) {
	results := []dtos.ListResult{
		{Name: "api-gateway", Version: "3.5.1", Channel: "regular", State: "Ready", Managed: true, CustomResourcePolicy: "CreateAndDelete", InstallationState: "Ready"},
	}

	var buf bytes.Buffer
	err := RenderList(results, types.JSONFormat, out.NewToWriter(&buf))

	require.NoError(t, err)
	require.JSONEq(t, `[{"name":"api-gateway","version":"3.5.1","channel":"regular","state":"Ready","managed":true,"customResourcePolicy":"CreateAndDelete","installationState":"Ready"}]`, buf.String())
}

func TestRenderList_Table_SortedByName(t *testing.T) {
	results := []dtos.ListResult{
		{Name: "istio"},
		{Name: "api-gateway"},
	}

	var buf bytes.Buffer
	err := RenderList(results, types.DefaultFormat, out.NewToWriter(&buf))

	require.NoError(t, err)
	require.Regexp(t, `(?s)api-gateway.*istio`, buf.String())
}

func TestRenderList_YAML(t *testing.T) {
	results := []dtos.ListResult{
		{Name: "api-gateway", Version: "3.5.1", Channel: "regular", State: "Ready", Managed: true, CustomResourcePolicy: "CreateAndDelete", InstallationState: "Ready"},
	}

	var buf bytes.Buffer
	err := RenderList(results, types.YAMLFormat, out.NewToWriter(&buf))

	require.NoError(t, err)
	var parsed []map[string]interface{}
	require.NoError(t, yaml.Unmarshal(buf.Bytes(), &parsed))
	require.Len(t, parsed, 1)
	module := parsed[0]
	require.Equal(t, "api-gateway", module["name"])
	require.Equal(t, "3.5.1", module["version"])
	require.Equal(t, "regular", module["channel"])
	require.Equal(t, "Ready", module["state"])
	require.Equal(t, true, module["managed"])
	require.Equal(t, "CreateAndDelete", module["customResourcePolicy"])
	require.Equal(t, "Ready", module["installationState"])
}
