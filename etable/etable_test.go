package etable

import (
	"fmt"
	"testing"
)

func BenchCreateLongTable(b *testing.B) {
	var keys = []string{"a", "b", "c", "d", "e", "f"}

	for b.Loop() {
		cols := make([]TableColumn, 0)
		for _, k := range keys {
			cols = append(cols, NewTableColumn(k, k))
		}

		rows := make([]TableRow, 0)
		for i := 0; i < 1000; i += 1 {
			for _, k := range keys {
				rows = append(rows, TableRow{
					k: fmt.Sprintf("%d", i),
				})
			}
		}

		_ = NewTable(cols).WithRows(rows).Render()
	}
}
