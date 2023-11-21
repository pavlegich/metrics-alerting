package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

// FileMetrics содержит метрики для хранения в файле.
type FileMetrics struct {
	Metrics map[string]string `json:"metrics"`
}

// NewFileMetrics создаёт новое хранилище метрик для файла.
func NewFileMetrics(ctx context.Context) *FileMetrics {
	return &FileMetrics{Metrics: make(map[string]string)}
}

// SaveToFile получает все текущие метрики из хранилища сервера,
// преобразует их в JSON формат и сохраняет в файл.
func SaveToFile(ctx context.Context, path string, ms interfaces.MetricStorage) error {
	// сериализуем структуру в JSON формат
	metrics := ms.GetAll(ctx)
	storage := NewFileMetrics(ctx)
	for m, v := range metrics {
		storage.Metrics[m] = v
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("SaveToFile: data marshal %w", err)
	}
	// сохраняем данные в файл
	if err := os.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("SaveToFile: write file error %w", err)
	}
	return nil
}

// LoadFromFile получает и конвертирует метрики из JSON формата,
// сохраняет в хранилище сервера.
func LoadFromFile(ctx context.Context, path string, ms interfaces.MetricStorage) error {
	if _, err := os.Stat(path); err != nil {
		if _, err := os.Create(path); err != nil {
			return fmt.Errorf("LoadFromFile: file create %w", err)
		}
		if err := SaveToFile(ctx, path, ms); err != nil {
			return fmt.Errorf("LoadFromFile: data save %w", err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("LoadFromFile: read file error %w", err)
	}

	storage := NewFileMetrics(ctx)

	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("LoadFromFile: data unmarshal %w", err)
	}

	for m, v := range storage.Metrics {
		// Все метрики с типом gauge, чтобы не было проблем с конвертацией
		if status := ms.Put(ctx, "gauge", m, v); status != http.StatusOK {
			return fmt.Errorf("LoadFromFile: put metric status %v", status)
		}
	}

	return nil
}
