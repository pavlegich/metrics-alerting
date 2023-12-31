package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/entities"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
)

// HandleGetMetric обрабатывает запрос на получение метрики,
// отправляет в ответ полученное значение метрики из хранилища.
func (h *Webhook) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	w.Header().Set("Content-Type", "text/plain")
	value, status := h.MemStorage.Get(ctx, metricType, metricName)
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
	w.Write([]byte(value))
}

// HandlePostValue обрабатывает запрос получения значения метрики.
// Обработчик принимает в JSON формате название и тип метрики,
// в случае успешного получения значения метрики из хранилища,
// формирует и отправляет ответ с метрикой в JSON формате.
func (h *Webhook) HandlePostValue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req entities.Metrics

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

	// проверяем, что пришел запрос понятного типа
	if req.MType != "gauge" && req.MType != "counter" {
		logger.Log.Error("unsupported request type")
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

	// заполняем модель ответа
	metricValue, status := h.MemStorage.Get(ctx, metricType, metricName)

	if status != http.StatusOK {
		logger.Log.Error("metric get error")
		w.WriteHeader(status)
		return
	}

	var resp entities.Metrics
	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
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
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp = entities.Metrics{
			ID:    metricName,
			MType: metricType,
			Delta: &v,
		}
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
