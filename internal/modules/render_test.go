package modules

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testModulesTableView        = "NAME      \tAVAILABLE VERSIONS     \tINSTALLED\tMANAGED\tSTATUS \nkeda      \t0.1(regular), 0.2(fast)\t         \t       \t      \t\nserverless\t0.0.1(fast), 0.0.2     \t         \t       \t      \t\n"
	testManagedModulesTableView = "NAME      \tAVAILABLE VERSIONS     \tINSTALLED  \tMANAGED\tSTATUS \nkeda      \t0.1(regular), 0.2(fast)\t0.2(fast)  \ttrue   \t      \t\nserverless\t0.0.1(fast), 0.0.2     \t0.0.1(fast)\tfalse  \tReady \t\n"
)

func TestRender(t *testing.T) {
	t.Run("render table from modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		render(buffer, testModuleList, ModulesTableInfo)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testModulesTableView, string(tableViewBytes))
	})

	t.Run("render table from managed modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		render(buffer, testManagedModuleList, ModulesTableInfo)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testManagedModulesTableView, string(tableViewBytes))
	})
}
