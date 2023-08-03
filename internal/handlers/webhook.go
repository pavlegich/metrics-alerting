package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/pavlegich/metrics-alerting/internal/storage"
)

var Storage = storage.NewStorage()

// функция webhook обрабатывает HTTP-запрос
func Webhook(w http.ResponseWriter, r *http.Request) {
	// var Storage = &MemStorage{make(map[string]string)}

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
		w.WriteHeader(update(metricParts[2:]))
		w.Write([]byte(fmt.Sprintf("%v", Storage)))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// функция update проверяет корректность запроса и обновляет хранилище метрик
func update(metricParts []string) int {

	fmt.Println(metricParts)
	// проверка на корректное количество элементов в запросе
	if len(metricParts) < 3 {
		return http.StatusNotFound
	}

	metric := storage.NewMetric(metricParts[0], metricParts[1], metricParts[2])

	// проверка на пустое имя метрики
	// metric.name = metricParts[2]
	if metric.Name() == "" {
		return http.StatusNotFound
	}

	// проверка на корректность типа и значения метрики
	// обновление хранлища метрик
	switch metric.Type() {
	case "gauge":
		if _, err := strconv.ParseFloat(metric.Value(), 64); err != nil {
			return http.StatusBadRequest
		}
		return Storage.Update(metric)
	case "counter":
		if _, err := strconv.ParseInt(metric.Value(), 10, 64); err != nil {
			return http.StatusBadRequest
		}
		return Storage.Update(metric)
	default:
		return http.StatusBadRequest
	}
}
