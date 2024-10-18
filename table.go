package essay

import (
	"math"

	"github.com/jmacd/essay/internal/recovery"
)

type (
	Table struct {
		Cells   [][]interface{}
		TopRow  []interface{}
		LeftCol []interface{}
	}
)

func (t Table) Render(builtin Builtin) (interface{}, error) {
	defer recovery.Here()()
	return builtin.RenderTable(t)
}

func (e *Essay) RenderTable(t Table) (interface{}, error) {
	defer recovery.Here()()
	return e.execute("table.html", t)
}

func RowTable(cells ...interface{}) Table {
	return Table{
		Cells: [][]interface{}{cells},
	}
}

func SquareTable(cells ...interface{}) Table {
	var out [][]interface{}

	n := int(math.Sqrt(float64(len(cells))) + 0.5)

	for len(cells) != 0 {
		row := make([]interface{}, 0, n)
		for i := 0; i < n; i++ {
			row = append(row, cells[0])
			cells = cells[1:]
		}
		out = append(out, row)
	}
	return Table{Cells: out}
}

func HeadRowTable(cells ...interface{}) Table {
	var top []interface{}
	var row []interface{}
	for i := 0; i < len(cells); i += 2 {
		top = append(top, cells[i])
		row = append(row, cells[i+1])
	}

	return Table{
		TopRow: top,
		Cells:  [][]interface{}{row},
	}
}
