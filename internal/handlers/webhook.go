package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

type Logger interface {
	Info(args ...interface{})
}

type Webhook struct {
	logger     Logger
	memStorage storage.MemStorage
}

func NewWebhook(logger Logger, memStorage *storage.MemStorage) *Webhook {
	return &Webhook{
		logger:     logger,
		memStorage: *memStorage,
	}
}

func (h *Webhook) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Get("/", h.handleMain)
	r.Route("/value", func(r chi.Router) {
		r.Get("/", h.handleBadRequest)
		r.Route("/{metricType}", func(r chi.Router) {
			r.Get("/", h.handleBadRequest)
			r.Get("/{metricName}", h.handleGetMetric)
		})
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.handleNotFound)
		r.Route("/{metricType}", func(r chi.Router) {
			r.Post("/", h.handleNotFound)
			r.Route("/{metricName}", func(r chi.Router) {
				r.Post("/", h.handleNotFound)
				r.Post("/{metricValue}", h.handlePostMetric)
			})
		})
	})
	return r
}

func (h *Webhook) handleBadRequest(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("bad request")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusBadRequest)
}

func (h *Webhook) handleNotFound(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("not found")
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
}

func (h *Webhook) handleMain(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("main")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(h.memStorage.HTML()))
}

func (h *Webhook) handleGetMetric(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get metric")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	w.Header().Set("Content-Type", "text/plain")
	value, status := h.memStorage.Get(metricType, metricName)
	if status == http.StatusOK {
		w.Write([]byte(value))
	}
	w.WriteHeader(status)
}

func (h *Webhook) handlePostMetric(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("post metric")
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	w.Header().Set("Content-Type", "text/plain")
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	status := h.memStorage.Put(metricType, metricName, metricValue)
	w.WriteHeader(status)
}
