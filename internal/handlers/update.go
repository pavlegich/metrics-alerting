package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"github.com/pavlegich/metrics-alerting/internal/models"
)

func (h *Webhook) HandlePostUpdates(w http.ResponseWriter, r *http.Request) {
	req := make([]models.Metrics, 0)

	// десериализуем запрос в структуру модели
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
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

		status := h.MemStorage.Put(metricType, metricName, metricValue)
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

func (h *Webhook) HandlePostMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	w.Header().Set("Content-Type", "text/plain")
	status := h.MemStorage.Put(metricType, metricName, metricValue)
	w.WriteHeader(status)
}

func (h *Webhook) HandlePostUpdate(w http.ResponseWriter, r *http.Request) {
	var req models.Metrics

	// десериализуем запрос в структуру модели
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Log.Error("read body error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		logger.Log.Error("decoding error")
		w.WriteHeader(http.StatusBadRequest)
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
		logger.Log.Error("got metric with bad name")
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
		logger.Log.Error("metric put error")
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

	var resp models.Metrics

	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(newValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp = models.Metrics{
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
		resp = models.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &v,
		}
	default:
		logger.Log.Info("got wrong metric type")
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
