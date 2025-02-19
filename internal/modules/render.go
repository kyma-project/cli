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

type RowConverter func(Module) []interface{}

type TableInfo struct {
	Header       []interface{}
	RowConverter RowConverter
}

var (
	ModulesTableInfo = TableInfo{
		Header: []interface{}{"NAME", "VERSION", "CR POLICY", "MANAGED", "STATUS"},
		RowConverter: func(m Module) []interface{} {
			return []interface{}{
				m.Name,
				convertInstall(m.InstallDetails),
				string(m.InstallDetails.CustomResourcePolicy),
				string(m.InstallDetails.Managed),
				m.InstallDetails.State,
			}
		},
	}

	CatalogTableInfo = TableInfo{
		Header: []interface{}{"NAME", "AVAILABLE VERSIONS"},
		RowConverter: func(m Module) []interface{} {
			return []interface{}{
				m.Name,
				convertVersions(m.Versions),
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
		tableInfo.Header,
		convertModuleListToTable(modulesList, tableInfo.RowConverter),
	)
}

func convertModuleListToTable(modulesList ModulesList, rowConverter RowConverter) [][]interface{} {
	slices.SortFunc(modulesList, func(a, b Module) int {
		return cmp.Compare(a.Name, b.Name)
	})

	var result [][]interface{}
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
