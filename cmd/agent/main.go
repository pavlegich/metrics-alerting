package main

import (
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/agent"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"github.com/pavlegich/metrics-alerting/internal/logger"
)

func main() {
	if err := logger.Initialize("Info"); err != nil {
		log.Fatalln(err)
	}
	defer logger.Log.Sync()

	cfg, err := agent.ParseFlags()
	if err != nil {
		logger.Log.Info("parse flags error")
	}

	// Интервалы опроса и отправки метрик
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	// Хранилище метрик
	statsStorage := agent.NewStatStorage()

	// Пауза для ожидания запуска сервера
	time.Sleep(time.Duration(2) * time.Second)

	c := make(chan int)
	go metricsRoutine(statsStorage, pollInterval, reportInterval, cfg.Address, c)

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
			if err := st.SendGZIP(addr); err != nil {
				log.Fatal(err)
				close(c)
			}

		}
	}
}
