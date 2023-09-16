package main

import (
	"context"

	"github.com/pavlegich/metrics-alerting/internal/agent"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	// Инициализация логера
	if err := logger.Initialize(ctx, "Info"); err != nil {
		logger.Log.Error("main: logger initialization error", zap.Error(err))
	}
	defer logger.Log.Sync()

	// Парсинг флагов
	cfg, err := agent.ParseFlags(ctx)
	if err != nil {
		logger.Log.Error("main: parse flags error", zap.Error(err))
	}

	// Хранилище метрик
	statsStorage := agent.NewStatStorage(ctx)

	c := make(chan int)
	// Периодический опрос и отправка метрик
	go agent.StatsRoutine(ctx, statsStorage, cfg, c)

	for {
		_, ok := <-c
		if !ok {
			logger.Log.Info("routine channel is closed; exit")
			break // exit
		}
	}
}
