package interfaces

import (
	"context"
	"runtime"
)

type (
	StatsStorage interface {
		SendJSON(ctx context.Context, url string) error
		SendGZIP(ctx context.Context, url string) error
		SendBatch(ctx context.Context, url string) error
		Update(ctx context.Context, memStats runtime.MemStats, count int, rand float64) error
		Put(ctx context.Context, sType string, name string, value string) error
	}

	MetricStorage interface {
		Put(ctx context.Context, metricType string, metricName string, metricValue string) int
		GetAll(ctx context.Context) (map[string]string, int)
		Get(ctx context.Context, metricType string, metricName string) (string, int)
	}
)
