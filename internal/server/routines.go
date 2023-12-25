package server

import (
	"context"
	"fmt"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/server/handlers"
)

// SaveToFileRoutine сохраняет метрики в файл с указанным интервалом времени.
func SaveToFileRoutine(ctx context.Context, wh *handlers.Webhook, store time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			if err := wh.File.Save(context.Background(), wh.MemStorage); err != nil {
				return fmt.Errorf("SaveToFileRoutine: metrics save error %w", err)
			}
			return nil
		default:
			if err := wh.File.Save(ctx, wh.MemStorage); err != nil {
				return fmt.Errorf("SaveToFileRoutine: metrics save error %w", err)
			}
			time.Sleep(store)
		}
	}
}

// SaveToDBRoutine сохраняет метрики в базу данных с указанным интервалом времени.
func SaveToDBRoutine(ctx context.Context, wh *handlers.Webhook, store time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			if err := wh.Database.Save(context.Background(), wh.MemStorage); err != nil {
				return fmt.Errorf("SaveToDBRoutine: metrics save error %w", err)
			}
			return nil
		default:
			if err := wh.Database.Save(ctx, wh.MemStorage); err != nil {
				return fmt.Errorf("SaveToDBRoutine: metrics save error %w", err)
			}
			time.Sleep(store)
		}
	}
}
