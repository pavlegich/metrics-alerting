package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func main() {
	// Считывание флагов
	addr := storage.NewAddress()
	_ = flag.Value(addr)
	flag.Var(addr, "a", "HTTP-server endpoint address host:port")
	report := flag.Int("r", 10, "Frequency of sending metrics to HTTP-server")
	poll := flag.Int("p", 2, "Frequency of metrics polling from the runtime package")
	flag.Parse()

	// Интервалы опроса и отправки метрик
	reportInterval := *report
	pollInterval := time.Duration(*poll) * time.Second

	// Хранилище метрик
	StatsStorage := storage.NewStatsStorage()

	// Runtime метрики
	var memStats runtime.MemStats

	// Дполнительные метрики
	pollCount := 0
	randomValue := rand.Float64()

	// Начальный опрос метрик
	if err := StatsStorage.Update(memStats, pollCount, randomValue); err != nil {
		log.Fatal(err)
	}

	// Пауза для ожидания запуска сервера
	time.Sleep(time.Duration(2) * time.Second)

	// Начальная отправка метрик
	if status := StatsStorage.Send("http://localhost:8080/update"); status != http.StatusOK {
		log.Fatal(status)
	}

	// Периодический опрос и отправка метрик
	for {
		time.Sleep(pollInterval)
		runtime.ReadMemStats(&memStats)
		pollCount += 1
		randomValue = rand.Float64()
		if err := StatsStorage.Update(memStats, pollCount, randomValue); err != nil {
			log.Fatal(err)
		}
		if (pollCount*2)%reportInterval == 0 {
			if status := StatsStorage.Send("http://localhost:8080/update"); status != http.StatusOK {
				log.Fatal(status)
			}
		}
	}
}
