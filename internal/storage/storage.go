package storage

import (
	"fmt"
	"net/http"
	"strconv"
)

type (
	MetricStorage interface {
		Update(metricType string, metricName string, metricValue string)
	}

	MemStorage struct {
		metrics map[string]string
	}
)

// метод Update обновляет хранилище данных в зависимости от запроса
func (ms MemStorage) Update(metric Metric) int {

	// в случае паники возвращаем ее значение
	defer func() {
		if p := recover(); p != nil {
			fmt.Println(`Возникла паника: `, p)
		}
	}()

	switch metric.Type() {
	case "gauge":
		ms.metrics[metric.Name()] = metric.Value()
	case "counter":
		// проверяем наличие метрики
		if _, ok := ms.metrics[metric.Name()]; !ok {
			ms.metrics[metric.Name()] = "0"
		}

		// конвертируем строку в значение float64, проверяем на ошибку
		metricValue, errMetric := strconv.ParseInt(ms.metrics[metric.Name()], 10, 64)
		if errMetric != nil {
			panic("metric value from storage cannot be converted")
		}
		metricCounter, errCounter := strconv.ParseInt(metric.Value(), 10, 64)
		if errCounter != nil {
			return http.StatusBadRequest
		}

		// складываем значения и добавляем в хранилище метрик
		newMetricValue := metricValue + metricCounter
		ms.metrics[metric.Name()] = fmt.Sprintf("%v", newMetricValue)
	}

	return http.StatusOK
}

func NewStorage() *MemStorage {
	return &MemStorage{make(map[string]string)}
}
