// Пакет interfaces содержит интерфейсы агента и сервера.
package interfaces

import (
	"context"
	"runtime"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
)

type (
	// StatsStorage содержит методы для работы с метриками агента.
	StatsStorage interface {
		SendJSON(ctx context.Context, cfg *config.AgentConfig) error
		SendGZIP(ctx context.Context, cfg *config.AgentConfig) error
		SendBatch(ctx context.Context, cfg *config.AgentConfig) error
		Update(ctx context.Context, memStats runtime.MemStats, count int, rand float64) error
		Put(ctx context.Context, sType string, name string, value string) error
	}

	// MetricStorage содержит методы для работы с метрики на сервере.
	MetricStorage interface {
		Put(ctx context.Context, metricType string, metricName string, metricValue string) int
		GetAll(ctx context.Context) map[string]string
		Get(ctx context.Context, metricType string, metricName string) (string, int)
	}

	Storage interface {
		Save(ctx context.Context, ms MetricStorage) error
		Load(ctx context.Context, ms MetricStorage) error
		Ping(ctx context.Context) error
	}
)
