package server

import (
	"fmt"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

func SaveToFileRoutine(wh *handlers.Webhook, store time.Duration, path string) error {
	for {
		if err := storage.SaveToFile(path, wh.MemStorage); err != nil {
			return fmt.Errorf("SaveToFileRoutine: metrics save error %w", err)
		}
		time.Sleep(store)
	}
}

func SaveToDBRoutine(wh *handlers.Webhook, store time.Duration) error {
	for {
		if err := storage.SaveToDB(wh.Database, wh.MemStorage); err != nil {
			return fmt.Errorf("SaveToDBRoutine: metrics save error %w", err)
		}
		time.Sleep(store)
	}
}
