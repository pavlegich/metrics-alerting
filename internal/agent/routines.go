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

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"go.uber.org/zap"
)

// PollCPUstats считывает информацию о занимаемой памяти с указанным интервалом времени
// и обновляет данные в хранилище.
func PollCPUstats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig, c chan int) {
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

		time.Sleep(interval)
	}
}

// PollMemStats считывает метрики с указанным интервалом времени
// и обновляет данные в хранилище.
func PollMemStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig, c chan int) {
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

		time.Sleep(interval)
	}
}

// SendStats создаёт worker-ов и отправляет данные из хранилища в работу worker-ам
// через канал с указанным интервалом.
func SendStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig, c chan int) {
	interval := time.Duration(cfg.ReportInterval) * time.Second
	jobs := make(chan interfaces.StatsStorage)
	for w := 1; w <= cfg.RateLimit; w++ {
		go sendWorker(ctx, cfg, jobs)
	}
	for {
		jobs <- st
		time.Sleep(interval)
	}
}

// sendWorker принимает метрики из канала и отправляет их по указанному адресу.
// Если соединение с сервером получить не удаётся, прерывает отправку метрик.
func sendWorker(ctx context.Context, cfg *config.AgentConfig, jobs <-chan interfaces.StatsStorage) {
	for j := range jobs {
		var err error = nil
		intervals := []time.Duration{0, time.Second, 3 * time.Second, 5 * time.Second}
		for _, interval := range intervals {
			time.Sleep(interval)
			err = j.SendBatch(ctx, cfg)
			if !errors.Is(err, syscall.ECONNREFUSED) {
				break
			}
		}
		if err != nil {
			logger.Log.Error("sendWorker: send stats failed", zap.Error(err))
		}
	}
}
