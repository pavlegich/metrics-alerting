// Пакет httpagent содержит объекты и методы для работы с http-агентом.
package httpagent

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
	"github.com/pavlegich/metrics-alerting/internal/infra/logger"
	"github.com/pavlegich/metrics-alerting/internal/interfaces"
	"go.uber.org/zap"
)

type Agent struct {
}

func NewAgent(ctx context.Context) *Agent {
	return &Agent{}
}

// SendStats создаёт worker-ов и отправляет данные из хранилища в работу worker-ам
// через канал с указанным интервалом.
func (a *Agent) SendStats(ctx context.Context, st interfaces.StatsStorage, cfg *config.AgentConfig) {
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
