package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func main() {
	// Считывание флагов
	// addr := models.NewAddress()
	// _ = flag.Value(addr)
	// flag.Var(addr, "a", "HTTP-server endpoint address host:port")

	addr := flag.String("a", "localhost:8080", "address")
	report := flag.Int("r", 10, "Frequency of sending metrics to HTTP-server")
	poll := flag.Int("p", 2, "Frequency of metrics polling from the runtime package")
	flag.Parse()

	// if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
	// 	addr.Set(envAddr)
	// }
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		*addr = envAddr
	}
	if envReport := os.Getenv("REPORT_INTERVAL"); envReport != "" {
		var err error
		*report, err = strconv.Atoi(envReport)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	if envPoll := os.Getenv("POLL_INTERVAL"); envPoll != "" {
		var err error
		*poll, err = strconv.Atoi(envPoll)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	// Интервалы опроса и отправки метрик
	pollInterval := time.Duration(*poll) * time.Second
	reportInterval := time.Duration(*report) * time.Second

	// Хранилище метрик
	statsStorage := storage.NewStatStorage()

	// Пауза для ожидания запуска сервера
	time.Sleep(time.Duration(1) * time.Second)

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
func metricsRoutine(st interfaces.StatsStorage, poll time.Duration, report time.Duration, addr string, c chan int) {
	tickerPoll := time.NewTicker(poll)
	tickerReport := time.NewTicker(report)
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
			if err := st.Send(addr); err != nil {
				log.Fatal(err)
				close(c)
			}

		}
	}
}
