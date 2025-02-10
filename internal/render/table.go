package render

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
)

const space = " "

var (
	tableStyle = table.Style{
		Name: "StyleDefault",
		Box: table.BoxStyle{
			EmptySeparator: space,
			PaddingLeft:    "", // empty to avoid white symbol at the beginning of every line
			PaddingRight:   space + space + space,
		},
		Color:   table.ColorOptionsDefault,
		Format:  table.FormatOptionsDefault,
		HTML:    table.DefaultHTMLOptions,
		Options: table.OptionsNoBordersAndSeparators,
		Size:    table.SizeOptionsDefault,
		Title:   table.TitleOptionsDefault,
	}
)

// Table renders the table with the provided headers and data
func Table(writer io.Writer, headers []interface{}, rows [][]interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(writer)
	t.AppendHeader(headers)
	t.AppendRows(convertRows(rows))
	t.SetStyle(tableStyle)
	t.Render()
}

// this func help converting type [][]interface{} to []table.Row
func convertRows(rows [][]interface{}) []table.Row {
	converted := make([]table.Row, len(rows))
	for i := range rows {
		converted[i] = rows[i]
	}

	return converted
}
