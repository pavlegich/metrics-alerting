package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

type FileMetrics struct {
	Metrics map[string]string `json:"metrics"`
}

func (fm *FileMetrics) SetFileMetrics(ctx context.Context, mName string, mValue string) error {
	fm.Metrics[mName] = mValue
	return nil
}

func NewFileMetrics(ctx context.Context) *FileMetrics {
	return &FileMetrics{Metrics: make(map[string]string)}
}

func SaveToFile(ctx context.Context, path string, ms interfaces.MetricStorage) error {
	// сериализуем структуру в JSON формат
	metrics, status := ms.GetAll(ctx)
	if status != http.StatusOK {
		return fmt.Errorf("SaveToFile: metrics get error %v", status)
	}

	storage := NewFileMetrics(ctx)
	for m, v := range metrics {
		storage.SetFileMetrics(ctx, m, v)
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
		// Сейчас все пусть будут gauge, чтобы ошибок с конвертацией не было, он не записывает тип в storage
		// Впоследствии сделаю, чтобы в storage хранились отдельно gauge и counter, не все string
		if status := ms.Put(ctx, "gauge", m, v); status != http.StatusOK {
			return fmt.Errorf("LoadFromFile: put metric status %v", status)
		}
	}

	return nil
}
