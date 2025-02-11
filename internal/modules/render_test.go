package modules

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testCatalogTableView          = "NAME      \tAVAILABLE VERSIONS      \nkeda      \t0.1(regular), 0.2(fast)\t\nserverless\t0.0.1(fast), 0.0.2     \t\n"
	testInstalledModulesTableView = "NAME      \tINSTALLED  \tCR POLICY      \tMANAGED\tSTATUS    \nkeda      \t0.2(fast)  \tCreateAndDelete\tfalse  \tUnmanaged\t\nserverless\t0.0.1(fast)\tIgnore         \ttrue   \tReady    \t\n"
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
