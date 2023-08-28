package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"github.com/pavlegich/metrics-alerting/internal/middlewares"
	"github.com/pavlegich/metrics-alerting/internal/models"
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
	r.Use(middlewares.WithLogging)
	// r.Use(middlewares.GZIP)

	r.Get("/", h.HandleMain)
	r.Post("/value/", h.HandlePostValue)
	r.Post("/update/", h.HandlePostUpdate)
	r.Get("/value/{metricType}/{metricName}", h.HandleGetMetric)
	r.Post("/update/{metricType}/{metricName}/{metricValue}", h.HandlePostMetric)

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

func (h *Webhook) HandleMain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// разрешаем только GET-запросы
		logger.Log.Info("got request with bad method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	metrics, status := h.MemStorage.GetAll()
	if status != http.StatusOK {
		w.WriteHeader(status)
		return
	}
	table := models.NewTable()
	for metric, value := range metrics {
		table.Put(metric, value)
	}
	tmpl, err := template.New("index").Parse(models.IndexTemplate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := tmpl.Execute(w, table); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Webhook) HandleGetMetric(w http.ResponseWriter, r *http.Request) {
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

func (h *Webhook) HandlePostMetric(w http.ResponseWriter, r *http.Request) {
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

func (h *Webhook) HandlePostUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		logger.Log.Info("got request with bad method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req models.Metrics

	// десериализуем запрос в структуру модели
	logger.Log.Info("decoding request")
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

	fmt.Println(metricType, metricName, metricValue)

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

func (h *Webhook) HandlePostValue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только Post-запросы
		logger.Log.Info("got request with bad method")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	fmt.Println(h.MemStorage.GetAll())

	var req models.Metrics

	// десериализуем запрос в структуру модели
	logger.Log.Info("decoding request")
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
		logger.Log.Info("got metric with bad name")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricName := req.ID

	fmt.Println(metricType, metricName)

	// заполняем модель ответа
	metricValue, status := h.MemStorage.Get(metricType, metricName)

	if status != http.StatusOK {
		logger.Log.Info("metric get error")
		w.WriteHeader(status)
		return
	}

	var resp models.Metrics
	switch metricType {
	case "gauge":
		v, err := strconv.ParseFloat(metricValue, 64)
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
		v, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp = models.Metrics{
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
	fmt.Println(h.MemStorage.GetAll())

	// установим правильный заголовок для типа данных
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(respJSON))
}
