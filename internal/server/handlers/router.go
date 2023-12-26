package handlers

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/server/middlewares"
)

// Webhook содержит локальное хранилище метрик и базу данных для сервера.
type Webhook struct {
	MemStorage interfaces.MetricStorage
	Database   interfaces.Storage
	File       interfaces.Storage
	Config     *config.ServerConfig
}

// NewWebhook создаёт новое хранилище сервера.
func NewWebhook(ctx context.Context, memStorage interfaces.MetricStorage, database interfaces.Storage, file interfaces.Storage, cfg *config.ServerConfig) *Webhook {
	return &Webhook{
		MemStorage: memStorage,
		Database:   database,
		File:       file,
		Config:     cfg,
	}
}

// Route инициализирует обработчики запросов сервера.
func (h *Webhook) Route(ctx context.Context) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewares.WithLogging)
	r.Use(middlewares.WithNetworking(h.Config.Network))
	r.Use(middlewares.WithSign)
	r.Use(middlewares.WithDecryption(h.Config.CryptoKey))
	r.Use(middlewares.WithCompress)

	r.Get("/", h.HandleMain)

	r.Post("/value/", h.HandlePostValue)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetMetric)

	r.Post("/update/", h.HandlePostUpdate)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandlePostMetric)

	r.Get("/ping", h.HandlePing)

	r.Post("/updates/", h.HandlePostUpdates)

	return r
}
