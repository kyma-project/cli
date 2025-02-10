package modules

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testCatalogTableView = "NAME         AVAILABLE VERSIONS        \n" +
		"keda         0.1(regular), 0.2(fast)   \n" +
		"serverless   0.0.1(fast), 0.0.2        \n"
	testInstalledModulesTableView = "NAME         VERSION       CR POLICY         MANAGED   STATUS      \n" +
		"keda         0.2(fast)     CreateAndDelete   false     Unmanaged   \n" +
		"serverless   0.0.1(fast)   Ignore            true      Ready       \n"
)

func TestRender(t *testing.T) {
	t.Run("render table from modules catalog", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		renderTable(buffer, testModuleList, CatalogTableInfo)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testCatalogTableView, string(tableViewBytes))
	})

	t.Run("render table from installed modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		renderTable(buffer, testInstalledModuleList, ModulesTableInfo)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testInstalledModulesTableView, string(tableViewBytes))
	})
}
