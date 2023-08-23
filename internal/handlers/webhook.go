package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/models"
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
		r.Handle("/", middlewares.WithLogging(h.HandleGetJSONMetric()))
		r.Route("/{metricType}", func(r chi.Router) {
			r.Handle("/", middlewares.WithLogging(h.HandleBadRequest()))
			r.Handle("/{metricName}", middlewares.WithLogging(h.HandleGetMetric()))
		})
	})
	r.Route("/update", func(r chi.Router) {
		r.Handle("/", middlewares.WithLogging(h.HandlePostJSONMetric()))
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
			logger.Log.Info("got request with bad method")
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
			logger.Log.Info("got request with bad method")
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
			// разрешаем только POST-запросы
			logger.Log.Info("got request with bad method")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")
		w.Header().Set("Content-Type", "text/plain")
		status := h.MemStorage.Put(metricType, metricName, metricValue)
		w.WriteHeader(status)
	}
	return http.HandlerFunc(fn)
}

func (h *Webhook) HandlePostJSONMetric() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			// разрешаем только POST-запросы
			logger.Log.Info("got request with bad method")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// десериализуем запрос в структуру модели
		logger.Log.Info("decoding request")
		var req models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// проверяем, то пришел запрос понятного типа
		if req.MType != "gauge" && req.MType != "counter" {
			logger.Log.Info("unsupported request type")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		metricType := req.MType

		// при правильном имени метрики, помещаем метрику в хранилище
		if req.ID == "" {
			logger.Log.Info("got metric with bad name")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		metricName := req.ID

		var metricValue string
		switch req.MType {
		case "gauge":
			metricValue = fmt.Sprintf("%v", *req.Value)
		case "counter":
			metricValue = fmt.Sprintf("%v", *req.Delta)
		}

		status := h.MemStorage.Put(metricType, metricName, metricValue)
		if status != http.StatusOK {
			logger.Log.Info("metric put error")
			w.WriteHeader(status)
			return
		}

		// заполняем модель ответа
		newValue, status := h.MemStorage.Get(metricType, metricName)
		if status != http.StatusOK {
			logger.Log.Info("metric get error")
			w.WriteHeader(status)
			return
		}
		resp := models.Metrics{
			ID:    metricName,
			MType: metricType,
		}
		switch metricType {
		case "gauge":
			v, err := strconv.ParseFloat(newValue, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.Value = &v
		case "counter":
			v, err := strconv.ParseInt(newValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.Delta = &v
		}

		// установим правильный заголовок для типа данных
		w.Header().Set("Content-Type", "application/json")

		// сериализуем ответ сервера
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}

func (h *Webhook) HandleGetJSONMetric() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			// разрешаем только GET-запросы
			logger.Log.Info("got request with bad method")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// десериализуем запрос в структуру модели
		logger.Log.Info("decoding request")
		var req models.Metrics
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// проверяем, то пришел запрос понятного типа
		if req.MType != "gauge" && req.MType != "counter" {
			logger.Log.Info("unsupported request type")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		metricType := req.MType

		// при правильном имени метрики, помещаем метрику в хранилище
		if req.ID == "" {
			logger.Log.Info("got metric with bad name")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		metricName := req.ID

		// заполняем модель ответа
		metricValue, status := h.MemStorage.Get(metricType, metricName)
		if status != http.StatusOK {
			logger.Log.Info("metric get error")
			w.WriteHeader(status)
			return
		}
		resp := models.Metrics{
			ID:    metricName,
			MType: metricType,
		}
		switch metricType {
		case "gauge":
			v, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.Value = &v
		case "counter":
			v, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			resp.Delta = &v
		}

		// установим правильный заголовок для типа данных
		w.Header().Set("Content-Type", "application/json")

		// сериализуем ответ сервера
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
	return http.HandlerFunc(fn)
}
