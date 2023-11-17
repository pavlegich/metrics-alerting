package storage

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

// MemStorage хранит данные метрик сервера.
type MemStorage struct {
	Metrics map[string]string
}

// NewMemStorage создаёт новое хранилище метрик сервера.
func NewMemStorage(ctx context.Context) *MemStorage {
	return &MemStorage{make(map[string]string)}
}

// Put обрабатывает данные метрики, в случае успеха сохраняет
// в хранилище сервера.
func (ms *MemStorage) Put(ctx context.Context, metricType string, metricName string, metricValue string) int {
	if metricName == "" {
		return http.StatusNotFound
	}
	switch metricType {
	case "gauge":
		if _, err := strconv.ParseFloat(metricValue, 64); err != nil {
			return http.StatusBadRequest
		}
		ms.Metrics[metricName] = metricValue
	case "counter":
		// проверяем наличие метрики
		if _, ok := ms.Metrics[metricName]; !ok {
			ms.Metrics[metricName] = "0"
		}

		// конвертируем строку в значение float64, проверяем на ошибку
		storageValue, errMetric := strconv.ParseInt(ms.Metrics[metricName], 10, 64)
		if errMetric != nil {
			return http.StatusInternalServerError
		}
		gotValue, errCounter := strconv.ParseInt(metricValue, 10, 64)
		if errCounter != nil {
			return http.StatusBadRequest
		}

		// складываем значения и добавляем в хранилище метрик
		newMetricValue := storageValue + gotValue
		ms.Metrics[metricName] = fmt.Sprintf("%v", newMetricValue)
	default:
		return http.StatusNotImplemented
	}

	return http.StatusOK
}

// Get получает из хранилища значение указанной метрики и возвращает это значение.
func (ms *MemStorage) Get(ctx context.Context, metricType string, metricName string) (string, int) {
	if (metricType != "gauge") && (metricType != "counter") {
		return "", http.StatusNotImplemented
	}
	value, ok := ms.Metrics[metricName]
	if !ok {
		return "", http.StatusNotFound
	}
	return value, http.StatusOK
}

// GetAll возвращает все метрики из хранилища.
func (ms *MemStorage) GetAll(ctx context.Context) (map[string]string, int) {
	return ms.Metrics, http.StatusOK
}
