package storage

import (
	"fmt"
	"net/http"
	"strconv"
)

type (
	MemStorage struct {
		Metrics map[string]string `json:"metrics"`
	}
)

// метод Update обновляет хранилище данных в зависимости от запроса
func (ms *MemStorage) Put(metricType string, metricName string, metricValue string) int {
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

func NewMemStorage() *MemStorage {
	return &MemStorage{make(map[string]string)}
}

func (ms *MemStorage) GetAll() (map[string]string, int) {
	return ms.Metrics, http.StatusOK
}

func (ms *MemStorage) Get(metricType string, metricName string) (string, int) {
	if (metricType != "gauge") && (metricType != "counter") {
		return "", http.StatusNotImplemented
	}
	value, ok := ms.Metrics[metricName]
	if !ok {
		return "", http.StatusNotFound
	}
	return value, http.StatusOK
}
