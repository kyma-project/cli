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
		Header: []string{"NAME", "REPOSITORY", "VERSIONS"},
		RowConverter: func(m Module) []string {
			return []string{m.Name, convertRepositories(m.Versions), convertVersions(m.Versions)}
		},
	}
)

func Render(modulesList ModulesList, tableInfo TableInfo, raw bool) {
	render(os.Stdout, modulesList, tableInfo, raw)
}

func render(writer io.Writer, modulesList ModulesList, tableInfo TableInfo, raw bool) {
	renderTable(
		writer,
		convertModuleListToTable(modulesList, tableInfo.RowConverter),
		tableInfo.Header,
		raw,
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

// renderTable renders the table with the provided headers
func renderTable(writer io.Writer, modulesData [][]string, headers []string, raw bool) {
	if raw {
		for _, row := range modulesData {
			println(strings.Join(row, "\t"))
		}
	} else {
		var table [][]string
		table = append(table, modulesData...)

		twTable := setTable(writer, table)
		twTable.SetHeader(headers)
		twTable.Render()
	}
}

// setTable sets the table settings for the tablewriter
func setTable(writer io.Writer, inTable [][]string) *tablewriter.Table {
	table := tablewriter.NewWriter(writer)
	table.AppendBulk(inTable)
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	return table
}

// convert versions to string containing links to repositories without duplicates separated by '\n'
func convertRepositories(versions []ModuleVersion) string {
	values := []string{}
	for _, version := range versions {
		if version.Repository == "" {
			// ignore if repository is empty
			continue
		}

		if !contains(values, version.Repository) {
			values = append(values, version.Repository)
		}
	}

	return strings.Join(values, ", ")
}

func contains(in []string, value string) bool {
	for _, inValue := range in {
		if inValue == value {
			return true
		}
	}

	return false
}

// convert versions to string containing values separated with '\n'
// and in format 'VERSION (CHANNEL)' or 'VERSION' if channel is empty
func convertVersions(versions []ModuleVersion) string {
	values := make([]string, len(versions))
	for i, version := range versions {
		value := version.Version
		if version.Channel != "" {
			value += fmt.Sprintf(" (%s)", version.Channel)
		}

		values[i] = value
	}

	return strings.Join(values, ", ")
}
