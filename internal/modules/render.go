package modules

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type RowConverter func(Module) []string

type TableInfo struct {
	Header       []string
	RowConverter RowConverter
}

var (
	ModulesTableInfo = TableInfo{
		Header: []string{"NAME", "AVAILABLE VERSIONS", "INSTALLED", "MANAGED"},
		RowConverter: func(m Module) []string {
			return []string{
				m.Name,
				convertVersions(m.Versions),
				convertInstall(m.InstallDetails),
				string(m.InstallDetails.Managed),
			}
		},
	}
)

func Render(modulesList ModulesList, tableInfo TableInfo) {
	render(os.Stdout, modulesList, tableInfo)
}

func render(writer io.Writer, modulesList ModulesList, tableInfo TableInfo) {
	renderTable(
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

// renderTable renders the table with the provided headers and data
func renderTable(writer io.Writer, modulesData [][]string, headers []string) {
	twTable := setTable(writer)
	twTable.AppendBulk(modulesData)
	twTable.SetHeader(headers)
	twTable.Render()
}

// setTable sets the table settings for the tablewriter
func setTable(writer io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	return table
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
