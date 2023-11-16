package handlers

import (
	"net/http"
	"text/template"

	"github.com/pavlegich/metrics-alerting/internal/entities"
)

func (h *Webhook) HandleMain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, status := h.MemStorage.GetAll(ctx)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
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
	w.WriteHeader(status)
	if err := tmpl.Execute(w, table); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
