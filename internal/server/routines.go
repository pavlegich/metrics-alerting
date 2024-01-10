// Пакет server содержит горутины для сохранения метрик в файловое хранилище или базу данных.
package server

import (
	"context"
	"fmt"
	"time"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

// SaveToFileRoutine сохраняет метрики в файл с указанным интервалом времени.
func SaveToFileRoutine(ctx context.Context, ms interfaces.MetricStorage, db interfaces.Storage,
	f interfaces.Storage, store time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			if err := f.Save(context.Background(), ms); err != nil {
				return fmt.Errorf("SaveToFileRoutine: metrics save error %w", err)
			}
			return nil
		default:
			if err := f.Save(ctx, ms); err != nil {
				return fmt.Errorf("SaveToFileRoutine: metrics save error %w", err)
			}
			time.Sleep(store)
		}
	}
}

// SaveToDBRoutine сохраняет метрики в базу данных с указанным интервалом времени.
func SaveToDBRoutine(ctx context.Context, ms interfaces.MetricStorage, db interfaces.Storage,
	f interfaces.Storage, store time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			if err := db.Save(context.Background(), ms); err != nil {
				return fmt.Errorf("SaveToDBRoutine: metrics save error %w", err)
			}
			return nil
		default:
			if err := db.Save(ctx, ms); err != nil {
				return fmt.Errorf("SaveToDBRoutine: metrics save error %w", err)
			}
			time.Sleep(store)
		}
	}
}
