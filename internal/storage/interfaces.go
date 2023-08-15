package storage

import (
	"runtime"
)

type (
	StatsStorage interface {
		Send(url string) error
		Update(memStats runtime.MemStats, count int, rand float64) error
		Put(sType string, name string, value string)
	}

	MetricStorage interface {
		Put(metricType string, metricName string, metricValue string) int
		GetAll() map[string]string
		Get(metricType string, metricName string) (string, int)
	}
)
