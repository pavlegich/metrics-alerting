// Пакет interfaces содержит интерфейсы агента и сервера.
package interfaces

import (
	"context"
	"runtime"
)

type (
	// StatsStorage содержит методы для работы с метриками агента.
	StatsStorage interface {
		SendJSON(ctx context.Context, url string, key string) error
		SendGZIP(ctx context.Context, url string, key string) error
		SendBatch(ctx context.Context, url string, key string) error
		Update(ctx context.Context, memStats runtime.MemStats, count int, rand float64) error
		Put(ctx context.Context, sType string, name string, value string) error
	}

	// MetrciStorage содержит методы для работы с метрики на сервере.
	MetricStorage interface {
		Put(ctx context.Context, metricType string, metricName string, metricValue string) int
		GetAll(ctx context.Context) (map[string]string, int)
		Get(ctx context.Context, metricType string, metricName string) (string, int)
	}
)
