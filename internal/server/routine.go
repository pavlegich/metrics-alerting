package server

import (
	"time"

	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func MetricsRoutine(wh *handlers.Webhook, store time.Duration, path string) error {
	for {
		if err := storage.Save(path, &wh.MemStorage); err != nil {
			return err
		}
		time.Sleep(store)
	}
}
