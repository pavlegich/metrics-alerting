package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type (
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

var Storage = &MemStorage{make(map[string]string)}

// метод Update обновляет хранилище данных в зависимости от запроса
func (ms MemStorage) Update(metric *metric) int {

	// в случае паники возвращаем ее значение
	defer func() {
		if p := recover(); p != nil {
			fmt.Println(`Возникла паника: `, p)
		}
	}()

	switch metric.mtype {
	case "gauge":
		ms.Metrics[metric.name] = metric.value
	case "counter":
		// проверяем наличие метрики
		if _, ok := ms.Metrics[metric.name]; !ok {
			ms.Metrics[metric.name] = "0"
		}

		// конвертируем строку в значение float64, проверяем на ошибку
		metricValue, errMetric := strconv.ParseFloat(ms.Metrics[metric.name], 64)
		if errMetric != nil {
			panic("metric value from storage cannot be converted")
		}
		metricCounter, errCounter := strconv.ParseFloat(metric.value, 64)
		if errCounter != nil {
			return http.StatusBadRequest
		}

		// складываем значения и добавляем в хранилище метрик
		newMetricValue := metricValue + metricCounter
		ms.Metrics[metric.name] = fmt.Sprintf("%v", newMetricValue)
	}

	return http.StatusOK
}

// функция update проверяет корректность запроса и обновляет хранилище метрик
func update(metricParts []string) int {

	// проверка на корректное количество элементов в запросе
	if len(metricParts) < 3 {
		return http.StatusNotFound
	}

	metric := &metric{}
	// проверка на пустое имя метрики
	metric.name = metricParts[2]
	if metric.name == "" {
		return http.StatusNotFound
	}

	metric.mtype = metricParts[1]
	metric.value = metricParts[3]

	// проверка на корректность типа и значения метрики
	// обновление хранлища метрик
	switch metric.mtype {
	case "gauge":
		if _, err := strconv.ParseFloat(metric.value, 64); err != nil {
			return http.StatusBadRequest
		}
		return Storage.Update(metric)
	case "counter":
		if _, err := strconv.ParseInt(metric.value, 10, 64); err != nil {
			return http.StatusBadRequest
		}
		return Storage.Update(metric)
	default:
		return http.StatusBadRequest
	}
}

// функция webhook обрабатывает HTTP-запрос
func webhook(w http.ResponseWriter, r *http.Request) {

	// делим URL на части
	path := r.URL.Path
	metricParts := strings.Split(path, "/")

	// первый элемент - пустой
	metricAction := metricParts[1]

	// проверяется и используется метод из запроса
	switch metricAction {
	case "update":
		if r.Method != http.MethodPost {
			// разрешаем только POST-запросы
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// отправляем в функцию update без названия метода update
		w.WriteHeader(update(metricParts[1:]))
		// w.Write([]byte(fmt.Sprintf("%v", Storage)))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// функция run запускает сервер
func run() error {
	return http.ListenAndServe(`:8080`, http.HandlerFunc(webhook))
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
