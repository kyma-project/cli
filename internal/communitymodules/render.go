package communitymodules

import (
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func RenderTableForCollective(raw bool, moduleMap moduleMap) {
	renderTable(raw,
		convertRowToCollective(moduleMap),
		[]string{"NAME", "REPOSITORY", "VERSION INSTALLED", "CONTROL-PLANE"})
}

func convertRowToCollective(moduleMap moduleMap) [][]string {
	var result [][]string
	for _, row := range moduleMap {
		result = append(result, []string{row.Name, row.Repository, row.Version, row.Managed})
	}
	return result
}

func RenderTableForInstalled(raw bool, moduleMap moduleMap) {
	renderTable(raw,
		convertRowToInstalled(moduleMap),
		[]string{"NAME", "VERSION"})
}

func convertRowToInstalled(moduleMap moduleMap) [][]string {
	var result [][]string
	for _, row := range moduleMap {
		result = append(result, []string{row.Name, row.Version})
	}
	return result
}

func RenderTableForManaged(raw bool, moduleMap moduleMap) {
	renderTable(raw,
		convertRowToManaged(moduleMap),
		[]string{"NAME"})
}

func convertRowToManaged(moduleMap moduleMap) [][]string {
	var result [][]string
	for _, row := range moduleMap {
		result = append(result, []string{row.Name})
	}
	return result
}
func RenderTableForCatalog(raw bool, moduleMap moduleMap) {
	renderTable(raw,
		convertRowToCatalog(moduleMap),
		[]string{"NAME", "REPOSITORY", "LATEST VERSION"})
}

func convertRowToCatalog(moduleMap moduleMap) [][]string {
	var result [][]string
	for _, row := range moduleMap {
		result = append(result, []string{row.Name, row.Repository, row.LatestVersion})
	}
	return result
}

// renderTable renders the table with the provided headers
func renderTable(raw bool, modulesData [][]string, headers []string) {
	if raw {
		for _, row := range modulesData {
			println(strings.Join(row, "\t"))
		}
	} else {

		var table [][]string
		table = append(table, modulesData...)

		twTable := setTable(table)
		twTable.SetHeader(headers)
		twTable.Render()
	}
}

// setTable sets the table settings for the tablewriter
func setTable(inTable [][]string) *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.AppendBulk(inTable)
	table.SetRowLine(true)
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetBorder(false)
	return table
}
