package entities

const IndexTemplate = "<html><body><table><tr><th>Название</th><th>Значение</th></tr>" +
	"{{range .Rows}}<tr><td>{{.Name}}</td><td>{{.Value}}</td></tr>{{end}}</table></body></html>"

type (
	Table struct {
		Rows []Row
	}

	Row struct {
		Name  string
		Value string
	}
)

func NewTable() *Table {
	return &Table{make([]Row, 0)}
}

func (d *Table) Put(mName string, mValue string) {
	newRow := Row{Name: mName, Value: mValue}
	d.Rows = append(d.Rows, newRow)
}
