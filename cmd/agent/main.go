package main

import (
	"context"
	"fmt"

	"github.com/pavlegich/metrics-alerting/internal/agent"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"go.uber.org/zap"
)

var buildVersion string = "N/A"
var buildDate string = "N/A"
var buildCommit = "N/A"

// Пример запуска
// go run -ldflags "-X main.buildVersion=v1.0.1 -X main.buildDate=$(date +'%Y/%m/%d') -X main.buildCommit=1d1wdd1f" main.go
func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	ctx := context.Background()

	// Инициализация логера
	if err := logger.Init(ctx, "Info"); err != nil {
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
	go agent.SendStats(ctx, statsStorage, cfg, c)
	go agent.PollCPUstats(ctx, statsStorage, cfg, c)
	go agent.PollMemStats(ctx, statsStorage, cfg, c)

	for {
		_, ok := <-c
		if !ok {
			logger.Log.Info("routine channel is closed; exit")
			break // exit
		}
	}
}
