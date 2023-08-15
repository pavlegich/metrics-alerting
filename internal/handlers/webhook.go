package handlers

import (
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/templates"
	log "github.com/sirupsen/logrus"
)

type Webhook struct {
	MemStorage interfaces.MetricStorage
}

func NewWebhook(memStorage interfaces.MetricStorage) *Webhook {
	return &Webhook{
		MemStorage: memStorage,
	}
}

func (h *Webhook) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", h.HandleMain)
	r.Route("/value", func(r chi.Router) {
		r.Get("/", h.HandleBadRequest)
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/", h.HandleBadRequest)
			r.Get("/{metricName}", h.HandleGetMetric)
		})
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.HandleNotFound)
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", h.HandleNotFound)
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", h.HandleNotFound)
				r.Post("/{metricValue}", h.HandlePostMetric)
			})
		})
	})
	return r
}

func (h *Webhook) HandleBadRequest(w http.ResponseWriter, r *http.Request) {
	log.Info("bad request")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
}

func (h *Webhook) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	log.Info("not found")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
}

func (h *Webhook) HandleMain(w http.ResponseWriter, r *http.Request) {
	log.Info("main")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	metrics := h.MemStorage.GetAll()
	table := templates.NewTable()
	for metric, value := range metrics {
		table.Put(metric, value)
	}
	tmpl, err := template.New("index").Parse(templates.IndexTemplate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if err := tmpl.Execute(w, table); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Webhook) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
	log.Info("get metric")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	w.Header().Set("Content-Type", "text/plain")
	value, status := h.MemStorage.Get(metricType, metricName)
	if status == http.StatusOK {
		w.Write([]byte(value))
	}
	w.WriteHeader(status)
}

func (h *Webhook) HandlePostMetric(w http.ResponseWriter, r *http.Request) {
	log.Info("post metric")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	w.Header().Set("Content-Type", "text/plain")
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	status := h.MemStorage.Put(metricType, metricName, metricValue)
	w.WriteHeader(status)
}
