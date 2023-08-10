package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
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

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr.Set(envAddr)
	}
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		*report, _ = strconv.Atoi(envReport)
	}
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		*poll, _ = strconv.Atoi(envPoll)
	}

	// Интервалы опроса и отправки метрик
	pollInterval := time.Duration(*poll) * time.Second
	reportInterval := *report

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
	if status := StatsStorage.Send(addr.String()); status != http.StatusOK {
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
			if status := StatsStorage.Send(addr.String()); status != http.StatusOK {
				log.Fatal(status)
			}
		}
	}
}
