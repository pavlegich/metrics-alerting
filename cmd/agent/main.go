package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/agent"
	"github.com/pavlegich/metrics-alerting/internal/infra/config"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	wg := &sync.WaitGroup{}

	// Инициализация логера
	if err := logger.Init(ctx, "Info"); err != nil {
		logger.Log.Error("main: logger initialization error", zap.Error(err))
	}
	defer logger.Log.Sync()

	// Парсинг флагов
	cfg, err := config.AgentParseFlags(ctx)
	if err != nil {
		logger.Log.Error("main: parse flags error", zap.Error(err))
	}

	// Хранилище метрик
	statsStorage := agent.NewStatStorage(ctx)

	// Периодический опрос и отправка метрик
	wg.Add(1)
	go func() {
		agent.SendStats(ctx, statsStorage, cfg)
		wg.Done()
	}()

	go func() {
		agent.PollCPUstats(ctx, statsStorage, cfg)
	}()

	go func() {
		agent.PollMemStats(ctx, statsStorage, cfg)
	}()

	<-ctx.Done()
	if ctx.Err() != nil {
		logger.Log.Info("shutting down gracefully...",
			zap.Error(ctx.Err()))

		// Ожидание завершения процессов
		connsClosed := make(chan struct{})
		go func() {
			wg.Wait()
			close(connsClosed)
		}()

		// Обработка завершения программы
		select {
		case <-connsClosed:
		case <-time.After(15 * time.Second):
			panic("shutdown timeout")
		}

		logger.Log.Info("quit")
	}
}
