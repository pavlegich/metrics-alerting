package interfaces

import (
	"runtime"
)

type (
	StatsStorage interface {
		Send(string) error
		Update(runtime.MemStats, int, float64) error
		Put(string, string, string)
	}

	MetricStorage interface {
		Put(string, string, string) int
		GetAll() (map[string]string, int)
		Get(string, string) (string, int)
	}
)
