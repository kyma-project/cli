package render

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// Table renders the table with the provided headers and data
func Table(writer io.Writer, modulesData [][]string, headers []string) {
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
