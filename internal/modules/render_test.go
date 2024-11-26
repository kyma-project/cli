package modules

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testModulesTableView = "NAME      \tREPOSITORY  \tVERSIONS                  \nkeda      \turl-3       \t0.1 (regular), 0.2 (fast)\t\nserverless\turl-1, url-2\t0.0.1 (fast), 0.0.2      \t\n"
)

func TestRender(t *testing.T) {
	t.Run("render table from modules", func(t *testing.T) {
		buffer := bytes.NewBuffer([]byte{})

		render(buffer, fixModuleList(), ModulesTableInfo, false)

		tableViewBytes, err := io.ReadAll(buffer)
		require.NoError(t, err)
		require.Equal(t, testModulesTableView, string(tableViewBytes))
	})
}
