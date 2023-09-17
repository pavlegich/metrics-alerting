package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
)

type Webhook struct {
	MemStorage interfaces.MetricStorage
	Database   *sql.DB
}

func NewWebhook(ctx context.Context, memStorage interfaces.MetricStorage, db *sql.DB) *Webhook {
	return &Webhook{
		MemStorage: memStorage,
		Database:   db,
	}
}

func (h *Webhook) Route(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)
	r.Use(middlewares.GZIP)

	r.Get("/", h.HandleMain)

	r.Post("/value/", h.HandlePostValue)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetMetric)

	r.Post("/update/", h.HandlePostUpdate)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandlePostMetric)

	r.Get("/ping", h.HandlePing)

	r.Post("/updates/", h.HandlePostUpdates)

	r.HandleFunc("/value/{metricType}/", h.HandleBadRequest)
	r.HandleFunc("/update/{metricType}/", h.HandleNotFound)
	r.HandleFunc("/update/{metricType}/{metricName}/", h.HandleNotFound)

	return r
}

func (h *Webhook) HandleBadRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
}

func (h *Webhook) HandleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
}
