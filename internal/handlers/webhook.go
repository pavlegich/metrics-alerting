package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pavlegich/metrics-alerting/internal/storage"
)

var Storage = storage.NewStorage()

// функция webhook обрабатывает HTTP-запрос
func Webhook(w http.ResponseWriter, r *http.Request) {

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
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(update(metricParts[2:]))
		w.Write([]byte(fmt.Sprintf("%v", Storage)))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}
