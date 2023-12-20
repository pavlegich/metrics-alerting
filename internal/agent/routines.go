package agent

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"go.uber.org/zap"
)

// PollCPUstats считывает информацию о занимаемой памяти с указанным интервалом времени
// и обновляет данные в хранилище.
func PollCPUstats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
	interval := time.Duration(cfg.PollInterval) * time.Second

	for {
		v, err := mem.VirtualMemory()
		if err != nil {
			logger.Log.Error("PollGoutilStats: get virtual memory stats failed", zap.Error(err))
		}
		c, err := cpu.PercentWithContext(ctx, 0, false)
		if err != nil {
			logger.Log.Error("PollGoutilStats: get cpu stats failed", zap.Error(err))
		}

		st.Put(ctx, "gauge", "TotalMemory", fmt.Sprintf("%v", v.Total))
		st.Put(ctx, "gauge", "FreeMemory", fmt.Sprintf("%v", v.Free))
		st.Put(ctx, "gauge", "CPUutilization1", fmt.Sprintf("%v", c))

		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(interval)
		}
	}
}

// PollMemStats считывает метрики с указанным интервалом времени
// и обновляет данные в хранилище.
func PollMemStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
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

		// Обновление метрик
		if err := st.Update(ctx, memStats, pollCount, randomValue); err != nil {
			logger.Log.Error("PollMemStats: stats update", zap.Error(err))
		}

		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(interval)
		}
	}
}

// SendStats создаёт worker-ов и отправляет данные из хранилища в работу worker-ам
// через канал с указанным интервалом.
func SendStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
	interval := time.Duration(cfg.ReportInterval) * time.Second
	for {
		select {
		case <-ctx.Done():
			return
		default:
			jobs := make(chan interfaces.StatsStorage)
			g := new(errgroup.Group)
			for w := 1; w <= cfg.RateLimit; w++ {
				g.Go(func() error {
					return sendWorker(ctx, cfg, jobs)
				})
			}
			jobs <- st
			close(jobs)
			if err := g.Wait(); err != nil {
				logger.Log.Error("SendStats: sendWorker run failed",
					zap.Error(err))
			}
		}
		time.Sleep(interval)
	}
}

// sendWorker принимает метрики из канала и отправляет их по указанному адресу.
// Если соединение с сервером получить не удаётся, прерывает отправку метрик.
func sendWorker(ctx context.Context, cfg *config.AgentConfig, jobs <-chan interfaces.StatsStorage) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case job, ok := <-jobs:
			if !ok {
				return nil
			}
			var err error = nil
			intervals := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}
			for _, interval := range intervals {
				time.Sleep(interval)
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
				err = job.SendBatch(ctx, cfg)
				cancel()
				if !errors.Is(err, syscall.ECONNREFUSED) {
					break
				}
			}
			if err != nil {
				return fmt.Errorf("sendWorker: send stats failed %w", err)
				// logger.Log.Error("sendWorker: send stats failed",
				// 	zap.Error(err))
			}
		}
	}
}
