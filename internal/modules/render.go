package modules

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/kyma-project/cli.v3/internal/render"
)

type RowConverter func(Module) []string

type TableInfo struct {
	Header       []string
	RowConverter RowConverter
}

var (
	ModulesTableInfo = TableInfo{
		Header: []string{"NAME", "AVAILABLE VERSIONS", "INSTALLED", "MANAGED", "STATUS"},
		RowConverter: func(m Module) []string {
			return []string{
				m.Name,
				convertVersions(m.Versions),
				convertInstall(m.InstallDetails),
				string(m.InstallDetails.Managed),
				m.InstallDetails.State,
			}
		},
	}
)

// Renders uses standard output to print ModuleList in table view
// TODO: support other formats like YAML or JSON
func Render(modulesList ModulesList, tableInfo TableInfo) {
	renderTable(os.Stdout, modulesList, tableInfo)
}

func renderTable(writer io.Writer, modulesList ModulesList, tableInfo TableInfo) {
	render.Table(
		writer,
		convertModuleListToTable(modulesList, tableInfo.RowConverter),
		tableInfo.Header,
	)
}

func convertModuleListToTable(modulesList ModulesList, rowConverter RowConverter) [][]string {
	slices.SortFunc(modulesList, func(a, b Module) int {
		return cmp.Compare(a.Name, b.Name)
	})

	var result [][]string
	for _, module := range modulesList {
		result = append(result, rowConverter(module))
	}
	return result
}

// convert version and channel into field in format 'version (channel)' for core modules and 'version' for community ones
func convertInstall(details ModuleInstallDetails) string {
	if details.Channel != "" {
		return fmt.Sprintf("%s(%s)", details.Version, details.Channel)
	}

	return details.Version
}

// convert versions to string containing values separated with '\n'
// and in format 'VERSION (CHANNEL)' or 'VERSION' if channel is empty
func convertVersions(versions []ModuleVersion) string {
	values := make([]string, len(versions))
	for i, version := range versions {
		value := version.Version
		if version.Channel != "" {
			value += fmt.Sprintf("(%s)", version.Channel)
		}

		values[i] = value
	}

	return strings.Join(values, ", ")
}
