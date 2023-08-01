package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type (
	MetricStorage interface {
		Update(metricType string, metricName string, metricValue string)
	}

	Metric struct {
		Name  string
		Value string
	}

	MemStorage struct {
		Metrics []Metric
	}
)

// метод Update обновляет хранилище данных в зависимости от запроса
func (s MemStorage) Update(metricType string, metricName string, metricValue interface{}) {
	// если нет такой метрики
	// ...
}

// функция update проверяет корректность запроса и обновляет хранилище метрик
func update(metricParts []string, storage *MemStorage) int {

	fmt.Println(metricParts)

	// проверка на корректное количество элементов в запросе
	if len(metricParts) < 3 {
		return http.StatusNotFound
	}

	// проверка на пустое имя метрики
	metricName := metricParts[2]
	if metricName == "" {
		return http.StatusNotFound
	}

	metricType := metricParts[1]
	metricValueStr := metricParts[3]

	// проверка на корректность типа и значения метрики
	// обновление хранлища метрик
	switch metricType {
	case "gauge":
		metricValue, err := strconv.ParseFloat(metricValueStr, 64)
		if err != nil {
			return http.StatusBadRequest
		}
		if _, err := strconv.ParseInt(metricValueStr, 10, 64); err == nil {
			return http.StatusBadRequest
		}
		storage.Update(metricType, metricName, metricValue)
	case "counter":
		metricValue, err := strconv.ParseInt(metricValueStr, 10, 64)
		if err != nil {
			return http.StatusBadRequest
		}
		storage.Update(metricType, metricName, metricValue)
	default:
		return http.StatusBadRequest
	}

	return http.StatusOK
}

// функция webhook обрабатывает HTTP-запрос
func webhook(w http.ResponseWriter, r *http.Request) {
	metricsStorage := &MemStorage{}

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
		w.WriteHeader(update(metricParts[1:], metricsStorage))
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
