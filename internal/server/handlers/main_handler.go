package handlers

import (
	"net/http"
	"text/template"

	"github.com/pavlegich/metrics-alerting/internal/entities"
)

// HandleMain обрабатывает запрос получения корневой веб-страницы,
// формирумя страницу, содержащую таблицу с информацией о текущих
// значениях метрик.
func (h *Webhook) HandleMain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metrics := h.MemStorage.GetAll(ctx)
	table := entities.NewTable()
	for metric, value := range metrics {
		table.Put(metric, value)
	}
	tmpl, err := template.New("index").Parse(entities.IndexTemplate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := tmpl.Execute(w, table); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
