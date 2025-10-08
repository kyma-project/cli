package modules

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

var testModules = []Module{
	{
		Name: "keda",
		Versions: []ModuleVersion{
			{
				Repository: "url-3",
				Version:    "0.1",
				Channel:    "regular",
			},
			{
				Version: "0.2",
				Channel: "fast",
			},
		},
		CommunityModule: false,
		Origin:          "kyma",
	},
	{
		Name: "serverless",
		Versions: []ModuleVersion{
			{
				Repository: "url-1",
				Version:    "0.0.1",
				Channel:    "fast",
			},
			{
				Repository: "url-2",
				Version:    "0.0.2",
			},
		},
		CommunityModule: false,
		Origin:          "kyma",
	},
	{
		Name: "cluster-ip",
		Versions: []ModuleVersion{
			{
				Repository: "url-1",
				Version:    "0.1.1",
				Channel:    "",
			},
			{
				Repository: "url-2",
				Version:    "0.1.2",
			},
		},
		CommunityModule: true,
		Origin:          "namespace/test",
	},
}

const (
	testCatalogTableView = "NAME         AVAILABLE VERSIONS        ORIGIN           \n" +
		"keda         0.1(regular), 0.2(fast)   kyma             \n" +
		"serverless   0.0.1(fast), 0.0.2        kyma             \n" +
		"cluster-ip   0.1.1, 0.1.2              namespace/test   \n"
	testInstalledModulesTableView = "NAME         VERSION       CR POLICY         MANAGED   MODULE STATUS   INSTALLATION STATUS   \n" +
		"keda         0.2(fast)     CreateAndDelete   false     Unknown         Unmanaged             \n" +
		"serverless   0.0.1(fast)   Ignore            true      Ready           Ready                 \n"

	testCatalogJSONView = `[
  {
    "availableVersions": "0.1(regular), 0.2(fast)",
    "name": "keda",
    "origin": "kyma"
  },
  {
    "availableVersions": "0.0.1(fast), 0.0.2",
    "name": "serverless",
    "origin": "kyma"
  },
  {
    "availableVersions": "0.1.1, 0.1.2",
    "name": "cluster-ip",
    "origin": "namespace/test"
  }
]
`
	testCatalogYAMLView = `- availableVersions: 0.1(regular), 0.2(fast)
  name: keda
  origin: kyma
- availableVersions: 0.0.1(fast), 0.0.2
  name: serverless
  origin: kyma
- availableVersions: 0.1.1, 0.1.2
  name: cluster-ip
  origin: namespace/test

`
)

func TestRender_renderTable(t *testing.T) {
	t.Run("render table from modules catalog", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		err := renderTable(buffer, testModules, CatalogTableInfo)
		require.NoError(t, err)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testCatalogTableView, string(tableViewBytes))
	})

	t.Run("render table from installed modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		err := renderTable(buffer, testInstalledModuleList, ModulesTableInfo)
		require.NoError(t, err)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testInstalledModulesTableView, string(tableViewBytes))
	})
}

func TestRender_renderJSON(t *testing.T) {
	t.Run("render table from modules catalog", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		err := renderJSON(buffer, testModules, CatalogTableInfo)
		require.NoError(t, err)

		jsonViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testCatalogJSONView, string(jsonViewBytes))
	})
}

func TestRender_renderYAML(t *testing.T) {
	t.Run("render table from modules catalog", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		err := renderYAML(buffer, testModules, CatalogTableInfo)
		require.NoError(t, err)

		yamlViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testCatalogYAMLView, string(yamlViewBytes))
	})
}
