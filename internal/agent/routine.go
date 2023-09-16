package agent

import (
	"context"
	"errors"
	"math/rand"
	"runtime"
	"syscall"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"go.uber.org/zap"
)

func StatsRoutine(ctx context.Context, st interfaces.StatsStorage, cfg *Config, c chan int) {
	// Интервалы опроса и отправки метрик
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	reportInterval := time.Duration(cfg.ReportInterval) * time.Second

	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)
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
			if err := st.Update(ctx, memStats, pollCount, randomValue); err != nil {
				logger.Log.Error("StatsRoutine: stats update", zap.Error(err))
			}
		case <-tickerReport.C:
			if err := st.SendBatch(ctx, cfg.Address, cfg.Key); err != nil {
				if errors.Is(err, syscall.ECONNREFUSED) {
					intervals := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}
					for _, interval := range intervals {
						time.Sleep(interval)
						if err := st.SendBatch(ctx, cfg.Address, cfg.Key); !errors.Is(err, syscall.ECONNREFUSED) {
							break
						}
					}
					logger.Log.Error("StatsRoutine: retriable error connection refused", zap.Error(err))
				} else {
					logger.Log.Error("StatsRoutine: send stats failed", zap.Error(err))
				}
			}
		}
	}
}
