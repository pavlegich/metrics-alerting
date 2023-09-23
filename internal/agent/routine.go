package agent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"go.uber.org/zap"
)

func GoutilStats(ctx context.Context, st interfaces.StatsStorage, cfg *Config, c chan int) {
	interval := time.Duration(cfg.PollInterval) * time.Second

	for {
		v, _ := mem.VirtualMemory()

		st.Put(ctx, "gauge", "TotalMemory", fmt.Sprintf("%v", v.Total))
		st.Put(ctx, "gauge", "FreeMemory", fmt.Sprintf("%v", v.Free))
		st.Put(ctx, "gauge", "CPUutilization1", fmt.Sprintf("%v", v.Available))

		time.Sleep(interval)
	}
}

func MemStats(ctx context.Context, st interfaces.StatsStorage, cfg *Config, c chan int) {
	// Runtime метрики
	var memStats runtime.MemStats

	// Дополнительные метрики
	pollCount := 0
	var randomValue float64

	interval := time.Duration(cfg.PollInterval) * time.Second

	for {
		// Обновление метрик
		runtime.ReadMemStats(&memStats)
		pollCount += 1
		randomValue = rand.Float64()

		// Опрос метрик
		if err := st.Update(ctx, memStats, pollCount, randomValue); err != nil {
			logger.Log.Error("StatsRoutine: stats update", zap.Error(err))
		}

		time.Sleep(interval)
	}
}

func SendStats(ctx context.Context, st interfaces.StatsStorage, cfg *Config, c chan int) {
	interval := time.Duration(cfg.ReportInterval) * time.Second

	for {
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

		time.Sleep(interval)
	}
}
