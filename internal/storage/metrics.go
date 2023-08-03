package storage

import (
	"fmt"
	"net/http"
	"strconv"
)

type (
	Metric interface {
		Type() string
		Name() string
		Value() string
	}

	metric struct {
		mtype string
		name  string
		value string
	}

	MetricStorage interface {
		Update(metricType string, metricName string, metricValue string)
	}

	MemStorage struct {
		Metrics map[string]string
	}
)

func (m metric) Type() string {
	return m.mtype
}

func (m metric) Name() string {
	return m.name
}

func (m metric) Value() string {
	return m.value
}

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
		ms.Metrics[metric.Name()] = metric.Value()
	case "counter":
		// проверяем наличие метрики
		if _, ok := ms.Metrics[metric.Name()]; !ok {
			ms.Metrics[metric.Name()] = "0"
		}

		// конвертируем строку в значение float64, проверяем на ошибку
		metricValue, errMetric := strconv.ParseFloat(ms.Metrics[metric.Name()], 64)
		if errMetric != nil {
			panic("metric value from storage cannot be converted")
		}
		metricCounter, errCounter := strconv.ParseFloat(metric.Value(), 64)
		if errCounter != nil {
			return http.StatusBadRequest
		}

		// складываем значения и добавляем в хранилище метрик
		newMetricValue := metricValue + metricCounter
		ms.Metrics[metric.Name()] = fmt.Sprintf("%v", newMetricValue)
	}

	return http.StatusOK
}

func NewStorage() *MemStorage {
	return &MemStorage{make(map[string]string)}
}

func NewMetric(metricType string, metricName string, metricValue string) Metric {
	return &metric{mtype: metricType, name: metricName, value: metricValue}
}
