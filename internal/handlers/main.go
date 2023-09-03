package handlers

import (
	"net/http"
	"text/template"

	"github.com/pavlegich/metrics-alerting/internal/models"
)

func (h *Webhook) HandleMain(w http.ResponseWriter, r *http.Request) {
	metrics, status := h.MemStorage.GetAll()
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	table := models.NewTable()
	for metric, value := range metrics {
		table.Put(metric, value)
	}
	tmpl, err := template.New("index").Parse(models.IndexTemplate)
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
