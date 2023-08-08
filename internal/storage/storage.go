package storage

import (
	"fmt"
	"net/http"
	"strconv"
)

type (
	MetricStorage interface {
		Put(metricType string, metricName string, metricValue string)
		HTML()
		Get(metricName string)
	}

	MemStorage struct {
		Metrics map[string]string
	}
)

// метод Update обновляет хранилище данных в зависимости от запроса
func (ms *MemStorage) Put(metricType string, metricName string, metricValue string) int {

	// в случае паники возвращаем ее значение
	defer func() {
		if p := recover(); p != nil {
			fmt.Println(`Возникла паника: `, p)
		}
	}()

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
			panic("metric value from storage cannot be converted")
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

func (ms *MemStorage) HTML() string {
	page := `<html>
	<head>
		<title>Список известных метрик</title>
	</head>
	<body>
		<table>
			<tr>
				<th>Название</th>
				<th>Значение</th>
			</tr>`
	for metric, value := range ms.Metrics {
		page += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%s</td>
			</tr>`, metric, value)
	}
	page += `
		</table>
	</body>
</html>`
	return page
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
