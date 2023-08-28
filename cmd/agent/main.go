package main

import (
	"log"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/agent"
	"github.com/pavlegich/metrics-alerting/internal/logger"
)

func main() {
	// Инициализация логера
	if err := logger.Initialize("Info"); err != nil {
		log.Fatalln(err)
	}
	defer logger.Log.Sync()

	// Парсинг флагов
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
	// time.Sleep(time.Duration(2) * time.Second)

	c := make(chan int)
	// Периодический опрос и отправка метрик
	go agent.StatsRoutine(statsStorage, pollInterval, reportInterval, cfg.Address, c)

	for {
		_, ok := <-c
		if !ok {
			break // exit
		}
	}
}
