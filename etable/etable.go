package etable

import (
	"encoding/csv"
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Table style definition.
type TableStyle struct {
	HeaderStyle  lipgloss.Style
	RowStyle     lipgloss.Style
	BorderStyle  lipgloss.Border
	BorderHeader bool
	BorderColumn bool
	BorderTop    bool
	BorderLeft   bool
	BorderBottom bool
	BorderRight  bool
}

// Default TableStyle used by Table. Uses color ANSI termcolor 4 for the heading.
var TableStyleDefault = TableStyle{
	HeaderStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true).Padding(0, 1),
	RowStyle:     lipgloss.NewStyle().Padding(0, 1),
	BorderStyle:  lipgloss.HiddenBorder(),
	BorderHeader: false,
	BorderColumn: false,
	BorderTop:    false,
	BorderLeft:   false,
	BorderBottom: false,
	BorderRight:  false,
}

// TableStyle for markdown formatting of the table
var TableStyleMarkdown = TableStyle{
	HeaderStyle: lipgloss.NewStyle().Bold(true).Padding(0, 1),
	RowStyle:    lipgloss.NewStyle().Padding(0, 1),
	BorderStyle: lipgloss.Border{
		Left:  "|",
		Right: "|",

		Top:      "-",
		TopLeft:  "|",
		TopRight: "|",

		Bottom:      "-",
		BottomLeft:  "|",
		BottomRight: "|",

		Middle:      "|",
		MiddleLeft:  "|",
		MiddleRight: "|",

		MiddleTop:    "|",
		MiddleBottom: "|",
	},
	BorderHeader: true,
	BorderColumn: true,
	BorderTop:    false,
	BorderLeft:   true,
	BorderBottom: false,
	BorderRight:  true,
}

// TableRow is the rapresentation of a row in a Table as a map between
// column keys and the assigned value for the row.
type TableRow = map[string]string

// Alignment of a TableColumn in a Table
//
//	etable.NewTableColumn(key, title).WithAlignment(TableAlignment)
type TableAlignment int

const (
	TableAlignmentLeft TableAlignment = iota
	TableAlignmentRight
	TableAlignmentCenter
)

// TableColumn is a representation of a column in a Table along with
// style and formatting functionalities.
type TableColumn struct {
	key         string
	title       string
	active      bool
	maxWidth    int
	alignment   TableAlignment
	emptyString string
	valueFunc   func(value string) string
	styleFunc   func(style lipgloss.Style, value string) lipgloss.Style
}

// Create a new TableColumn given its key and title.
//
//	c := etable.NewTableColumn("id", "ID")
func NewTableColumn(key string, title string) TableColumn {
	return TableColumn{
		key:         key,
		title:       title,
		active:      true,
		maxWidth:    -1,
		emptyString: "",
		alignment:   TableAlignmentLeft,
		valueFunc: func(value string) string {
			return value
		},
		styleFunc: func(style lipgloss.Style, value string) lipgloss.Style {
			return style
		},
	}
}

// Set a maximum width for the column after which its value will be truncated.
//
//	c := etable.NewTableColumn("id", "ID").WithMaxWidth(30)
func (c TableColumn) WithMaxWidth(w int) TableColumn {
	c.maxWidth = w
	return c
}

// Set the alignment of the column.
//
//	c := etable.NewTableColumn("id", "ID").WithAlignment(etable.TableAlignmentLeft)
func (c TableColumn) WithAlignment(a TableAlignment) TableColumn {
	c.alignment = a
	return c
}

// Show or hide the column.
//
//	c := etable.NewTableColumn("id", "ID").WithActive(false)
func (c TableColumn) WithActive(a bool) TableColumn {
	c.active = a
	return c
}

// Specify a value that will replace empty strings in the column before outputting it.
// Note that this substitution is applied after the valueFunc if provided.
//
//	c := etable.NewTableColumn("id", "ID").WithEmptyString("-")
func (c TableColumn) WithEmptyString(s string) TableColumn {
	c.emptyString = s
	return c
}

// Specify a fuction that will be applied to all the values in the column
// before outputting it.
//
//	c := etable.NewTableColumn("id", "ID").WithValueFunc(func(value string) string {
//		return strings.ToUpper(value)
//	})
func (c TableColumn) WithValueFunc(
	valueFunc func(value string) string,
) TableColumn {
	c.valueFunc = valueFunc
	return c
}

// Specify a style that will be applied to all the cells in the column.
// Note that this is applied after the valueFunc setted with WithValueFunc.
//
//		c := etable.NewTableColumn("id", "ID").WithStyleFunc(func(style, value) lipgloss.Style {
//			if value == "OK" {
//				return style.Bold(true)
//			}
//			else return style
//	})
func (c TableColumn) WithStyleFunc(
	styleFunc func(style lipgloss.Style, value string) lipgloss.Style,
) TableColumn {
	c.styleFunc = styleFunc
	return c
}

// A rapresentation of a Table.
type Table struct {
	columns []TableColumn
	rows    []TableRow
	style   TableStyle
}

// Create a new Table given its columns as TableColumn.
//
//	columns := []etable.TableColumn{
//		etable.NewTableColumn(key, title)
//	}
//	t := etable.NewTable(columns)
func NewTable(columns []TableColumn) Table {
	return Table{
		columns: columns,
		rows:    []TableRow{},
		style:   TableStyleDefault,
	}
}

// Specify the style of the Table.
//
//	t := etable.NewTable(columns).WithStyle(etable.TableStyleMarkdown)
func (t Table) WithStyle(s TableStyle) Table {
	t.style = s
	return t
}

// Adds a slice of TableRow to the Table
//
//	t := etable.NewTable(columns)
//	rows := make([]etable.TableRow, 0)
//	// fill rows
//	t.WithRows(rows)
func (t Table) WithRows(rows []TableRow) Table {
	t.rows = rows
	return t
}

func (t *Table) getRowMatrix() [][]string {
	rows := make([][]string, 0)
	for _, rowEntry := range t.rows {
		row := []string{}
		for _, col := range t.columns {
			if !col.active {
				continue
			}

			value := col.valueFunc(rowEntry[col.key])
			if value == "" {
				value = col.emptyString
			}
			if col.maxWidth > 0 && col.maxWidth < len(value) {
				value = fmt.Sprintf("%.*s...", col.maxWidth-3, value)
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}
	return rows
}

// Render the Table.
//
//	t := etable.NewTable(...).WithRows(...)
//	fmt.Println(t.Render())
func (t *Table) Render() string {
	headers := make([]string, 0)

	columnOffset := 0
	columnOffsets := make([]int, 0)
	for _, col := range t.columns {
		if !col.active {
			columnOffset += 1
			continue
		}

		columnOffsets = append(columnOffsets, columnOffset)
		headers = append(headers, col.title)
	}

	rows := t.getRowMatrix()

	lt := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(t.style.BorderStyle).
		BorderLeft(t.style.BorderLeft).BorderRight(t.style.BorderRight).
		BorderTop(t.style.BorderTop).BorderBottom(t.style.BorderBottom).
		BorderHeader(t.style.BorderHeader).BorderColumn(t.style.BorderColumn).
		StyleFunc(func(row int, col int) lipgloss.Style {
			var sty lipgloss.Style
			column := t.columns[col+columnOffsets[col]]

			if row == table.HeaderRow {
				sty = t.style.HeaderStyle
			} else {
				sty = column.styleFunc(t.style.RowStyle, rows[row][col])
			}

			switch column.alignment {
			case TableAlignmentLeft:
				sty = sty.Align(lipgloss.Left)
			case TableAlignmentCenter:
				sty = sty.Align(lipgloss.Center)
			case TableAlignmentRight:
				sty = sty.Align(lipgloss.Right)
			}

			return sty
		})

	return lt.Render()
}

// Export the table as a .csv file.
//
// t := t.NewTable(...).WithRows(...)
// fd, _ := os.Create("path_to_file.csv")
// t.ExportCSV(fd)
func (t *Table) ExportCSV(w io.Writer) error {
	csvWriter := csv.NewWriter(w)

	header := make([]string, 0)
	for _, col := range t.columns {
		if col.active {
			header = append(header, col.title)
		}
	}

	err := csvWriter.Write(header)
	if err != nil {
		return err
	}
	err = csvWriter.WriteAll(t.getRowMatrix())
	if err != nil {
		return err
	}

	return nil
}
