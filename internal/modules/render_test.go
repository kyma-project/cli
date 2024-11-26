package modules

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testModulesTableView        = "NAME      \tREPOSITORY  \tVERSIONS                 \tINSTALLED\tMANAGED \n   keda   \t   url-3    \t0.1 (regular), 0.2 (fast)\t  false  \t false \t\nserverless\turl-1, url-2\t   0.0.1 (fast), 0.0.2   \t  false  \t false \t\n"
	testManagedModulesTableView = "NAME      \tREPOSITORY  \tVERSIONS                 \tINSTALLED\tMANAGED \n   keda   \t   url-3    \t0.1 (regular), 0.2 (fast)\t  true   \t true  \t\nserverless\turl-1, url-2\t   0.0.1 (fast), 0.0.2   \t  true   \t true  \t\n"
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
