package entities

// IndexTemplate содержит шаблон HTML разметки страницы.
const IndexTemplate = "<html><body><table><tr><th>Название</th><th>Значение</th></tr>" +
	"{{range .Rows}}<tr><td>{{.Name}}</td><td>{{.Value}}</td></tr>{{end}}</table></body></html>"

type (
	// Table содержит строки с данными метрик.
	Table struct {
		Rows []Row
	}

	// Row содержит имя и значение метрики.
	Row struct {
		Name  string
		Value string
	}
)

// NewTable создаёт таблицу.
func NewTable() *Table {
	return &Table{make([]Row, 0)}
}

// Put добавляет новую строку с данными метрики в таблицу.
func (d *Table) Put(mName string, mValue string) {
	newRow := Row{Name: mName, Value: mValue}
	d.Rows = append(d.Rows, newRow)
}
