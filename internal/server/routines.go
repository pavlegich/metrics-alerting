package server

import (
	"context"
	"fmt"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/server/handlers"
	"github.com/pavlegich/metrics-alerting/internal/storage"
)

// SaveToFileRoutine сохраняет метрики в файл с указанным интервалом времени.
func SaveToFileRoutine(ctx context.Context, wh *handlers.Webhook, store time.Duration, path string) error {
	for {
		if err := storage.SaveToFile(ctx, path, wh.MemStorage); err != nil {
			return fmt.Errorf("SaveToFileRoutine: metrics save error %w", err)
		}
		time.Sleep(store)
	}
}

// SaveToFileRoutine сохраняет метрики в базу данных с указанным интервалом времени.
func SaveToDBRoutine(ctx context.Context, wh *handlers.Webhook, store time.Duration) error {
	for {
		if err := storage.SaveToDB(ctx, wh.Database, wh.MemStorage); err != nil {
			return fmt.Errorf("SaveToDBRoutine: metrics save error %w", err)
		}
		time.Sleep(store)
	}
}
