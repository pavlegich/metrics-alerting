package interfaces

import (
	"runtime"
)

type (
	StatsStorage interface {
		SendJSON(url string) error
		SendGZIP(url string) error
		SendBatch(url string) error
		Update(memStats runtime.MemStats, count int, rand float64) error
		Put(sType string, name string, value string) error
	}

	MetricStorage interface {
		Put(metricType string, metricName string, metricValue string) int
		GetAll() (map[string]string, int)
		Get(metricType string, metricName string) (string, int)
	}
)
