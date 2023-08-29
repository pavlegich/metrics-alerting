package main

import (
	"time"

	"github.com/pavlegich/metrics-alerting/internal/agent"
	"github.com/pavlegich/metrics-alerting/internal/logger"
	"go.uber.org/zap"
)

func main() {
	// Инициализация логера
	if err := logger.Initialize("Info"); err != nil {
		logger.Log.Error("main: logger initialization error", zap.Error(err))
	}
	defer logger.Log.Sync()

	// Парсинг флагов
	cfg, err := agent.ParseFlags()
	if err != nil {
		logger.Log.Error("main: parse flags error", zap.Error(err))
	}

	// Интервалы опроса и отправки метрик
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	// Хранилище метрик
	statsStorage := agent.NewStatStorage()

	// Пауза для ожидания запуска сервера

	c := make(chan int)
	// Периодический опрос и отправка метрик
	go agent.StatsRoutine(statsStorage, pollInterval, reportInterval, cfg.Address, c)

	for {
		_, ok := <-c
		if !ok {
			logger.Log.Info("routine channel is closed; exit")
			break // exit
		}
	}
}
