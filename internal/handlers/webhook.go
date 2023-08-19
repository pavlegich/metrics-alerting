package handlers

import (
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/templates"
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
	r.Handle("/", middlewares.WithLogging(h.HandleMain()))
	r.Route("/value", func(r chi.Router) {
		r.Handle("/", middlewares.WithLogging(h.HandleBadRequest()))
		r.Route("/{metricType}", func(r chi.Router) {
			r.Handle("/", middlewares.WithLogging(h.HandleBadRequest()))
			r.Handle("/{metricName}", middlewares.WithLogging(h.HandleGetMetric()))
		})
	})
	r.Route("/update", func(r chi.Router) {
		r.Handle("/", middlewares.WithLogging(h.HandleNotFound()))
		r.Route("/{metricType}", func(r chi.Router) {
			r.Handle("/", middlewares.WithLogging(h.HandleNotFound()))
			r.Route("/{metricName}", func(r chi.Router) {
				r.Handle("/", middlewares.WithLogging(h.HandleNotFound()))
				r.Handle("/{metricValue}", middlewares.WithLogging(h.HandlePostMetric()))
			})
		})
	})
	return r
}

func (h *Webhook) HandleBadRequest() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
	}
	return http.HandlerFunc(fn)
}

func (h *Webhook) HandleNotFound() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
	}
	return http.HandlerFunc(fn)
}

func (h *Webhook) HandleMain() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// разрешаем только GET-запросы
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		metrics, status := h.MemStorage.GetAll()
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		table := templates.NewTable()
		for metric, value := range metrics {
			table.Put(metric, value)
		}
		tmpl, err := template.New("index").Parse(templates.IndexTemplate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, table); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(status)
	}
	return http.HandlerFunc(fn)
}

func (h *Webhook) HandleGetMetric() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// разрешаем только GET-запросы
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		w.Header().Set("Content-Type", "text/plain")
		value, status := h.MemStorage.Get(metricType, metricName)
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		w.WriteHeader(status)
		w.Write([]byte(value))
	}
	return http.HandlerFunc(fn)
}

func (h *Webhook) HandlePostMetric() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			// разрешаем только GET-запросы
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
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
	return http.HandlerFunc(fn)
}
