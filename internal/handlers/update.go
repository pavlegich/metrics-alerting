package handlers

// // функция update проверяет корректность запроса и обновляет хранилище метрик
// func update(metricParts []string) int {
// 	// fmt.Println(metricParts)
// 	// проверка на корректное количество элементов в запросе
// 	if len(metricParts) < 3 {
// 		return http.StatusNotFound
// 	}
// 	metricType, metricName, metricValue := metricParts[0], metricParts[1], metricParts[2]

// 	// проверка на пустое имя метрики
// 	// metric.name = metricParts[2]
// 	if metricName == "" {
// 		return http.StatusNotFound
// 	}

// 	// проверка на корректность типа и значения метрики
// 	// обновление хранлища метрик
// 	switch metricType {
// 	case "gauge":
// 		metric := storage.NewGauge(metricType, metricName, metricValue)
// 		if _, err := strconv.ParseFloat(metric.Value(), 64); err != nil {
// 			return http.StatusBadRequest
// 		}
// 		return Storage.Update(metric)
// 	case "counter":
// 		metric := storage.NewCounter(metricType, metricName, metricValue)
// 		if _, err := strconv.ParseInt(metric.Value(), 10, 64); err != nil {
// 			return http.StatusBadRequest
// 		}
// 		return Storage.Update(metric)
// 	default:
// 		return http.StatusBadRequest
// 	}
// }
