package handlers

import (
	"context"
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/server/middlewares"
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
	r.Use(middlewares.WithSign)
	r.Use(middlewares.GZIP)

	r.Get("/", h.HandleMain)

	r.Post("/value/", h.HandlePostValue)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetMetric)

	r.Post("/update/", h.HandlePostUpdate)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandlePostMetric)

	r.Get("/ping", h.HandlePing)

	r.Post("/updates/", h.HandlePostUpdates)

	return r
}