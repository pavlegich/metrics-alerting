package server

import (
	"fmt"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func MetricsRoutine(wh *handlers.Webhook, store time.Duration, path string) error {
	for {
		if err := storage.Save(path, wh.MemStorage); err != nil {
			return fmt.Errorf("MetricsRoutine: metrics save error %w", err)
		}
		time.Sleep(store)
	}
}
