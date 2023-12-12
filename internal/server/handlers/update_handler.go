package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
)

// HandlePostUpdates обрабатывает и сохраняет полученные метрики.
func (h *Webhook) HandlePostUpdates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := make([]entities.Metrics, 0)

	// десериализуем запрос в структуру модели
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.Log.Error("HandlePostUpdates: read body error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Log.Error("HandlePostUpdates: decoding error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, metric := range req {
		// проверяем, то пришел запрос понятного типа
		if metric.MType != "gauge" && metric.MType != "counter" {
			logger.Log.Error("HandlePostUpdates: unsupported request type")
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		metricType := metric.MType

		// при правильном имени метрики, помещаем метрику в хранилище
		if metric.ID == "" {
			logger.Log.Error("HandlePostUpdates: got metric with bad name")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		metricName := metric.ID

		var metricValue string
		switch metric.MType {
		case "gauge":
			metricValue = fmt.Sprintf("%v", *metric.Value)
		case "counter":
			metricValue = fmt.Sprintf("%v", *metric.Delta)
		}

		status := h.MemStorage.Put(ctx, metricType, metricName, metricValue)
		if status != http.StatusOK {
			logger.Log.Error("HandlePostUpdates: metric put error")
			w.WriteHeader(status)
			return
		}
	}

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// HandlePostMetric обрабатывает и сохраняет полученную метрику.
func (h *Webhook) HandlePostMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	status := h.MemStorage.Put(ctx, metricType, metricName, metricValue)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
}

// HandlePostUpdate обрабатывает и сохраняет полученную в JSON формате метрику.
// В случае успешного сохранения обработчик получает новое значение метрики
// из хранилища и отправляет в ответ метрику в JSON формате.
func (h *Webhook) HandlePostUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entities.Metrics

	// десериализуем запрос в структуру модели
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.Log.Error("HandlePostUpdate: read body error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Log.Error("HandlePostUpdate: decoding error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// проверяем, то пришел запрос понятного типа
	if req.MType != "gauge" && req.MType != "counter" {
		logger.Log.Error("unsupported request type")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	metricType := req.MType

	// при правильном имени метрики, помещаем метрику в хранилище
	if req.ID == "" {
		logger.Log.Error("HandlePostUpdate: got metric with bad name")
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

	status := h.MemStorage.Put(ctx, metricType, metricName, metricValue)
	if status != http.StatusOK {
		logger.Log.Error("HandlePostUpdate: metric put error")
		w.WriteHeader(status)
		return
	}

	// заполняем модель ответа
	newValue, status := h.MemStorage.Get(ctx, metricType, metricName)
	if status != http.StatusOK {
		logger.Log.Error("HandlePostUpdate: metric get error")
		w.WriteHeader(status)
		return
	}

	var resp entities.Metrics

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(newValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp = entities.Metrics{
			ID:    metricName,
			MType: metricType,
			Value: &v,
		}
	case "counter":
		v, err := strconv.ParseInt(newValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp = entities.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &v,
		}
	default:
		logger.Log.Error("HandlePostUpdate: got wrong metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// сериализуем ответ сервера
	respJSON, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respJSON))
}
