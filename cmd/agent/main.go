package main

import (
	"flag"
	"log"
	"math/rand"
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
	pollInterval := time.Duration(*poll)
	reportInterval := time.Duration(*report)

	// Хранилище метрик
	statsStorage := storage.NewStatsStorage()

	// Пауза для ожидания запуска сервера
	time.Sleep(time.Duration(2) * time.Second)

	c := make(chan int)
	go metricsRoutine(statsStorage, pollInterval, reportInterval, *addr, c)

	for {
		_, ok := <-c
		if !ok {
			break // exit
		}
	}
}

// Периодический опрос и отправка метрик
func metricsRoutine(st storage.StatsStorage, poll time.Duration, report time.Duration, addr storage.Address, c chan int) {
	tickerPoll := time.NewTicker(poll * time.Second)
	tickerReport := time.NewTicker(report * time.Second)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()

	// Runtime метрики
	var memStats runtime.MemStats

	// Дополнительные метрики
	pollCount := 0
	var randomValue float64

	for {
		select {
		case <-tickerPoll.C:
			// Обновление метрик
			runtime.ReadMemStats(&memStats)
			pollCount += 1
			randomValue = rand.Float64()

			// Опрос метрик
			if err := st.Update(memStats, pollCount, randomValue); err != nil {
				log.Fatal(err)
				close(c)
			}
		case <-tickerReport.C:
			if err := st.Send(addr.String()); err != nil {
				log.Fatal(err)
				close(c)
			}

		}
	}
}
