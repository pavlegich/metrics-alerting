package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

// FileMetrics содержит метрики для хранения в файле.
type FileMetrics struct {
	Metrics map[string]string `json:"metrics"`
}

// NewFileMetrics создаёт новое хранилище метрик для файла.
func NewFileMetrics(ctx context.Context) *FileMetrics {
	return &FileMetrics{
		Metrics: make(map[string]string),
	}
}

// File содержит информацию о пути к файлу.
type File struct {
	path string
	mu   *sync.Mutex
}

// NewFile создаёт новый объект File для хранения метрик сервера.
func NewFile(path string) *File {
	return &File{
		path: path,
		mu:   &sync.Mutex{},
	}
}

// Save получает все текущие метрики из хранилища сервера,
// преобразует их в JSON формат и сохраняет в файл.
func (f *File) Save(ctx context.Context, ms interfaces.MetricStorage) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// сериализуем структуру в JSON формат
	storage := ms.GetAll(ctx)

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("SaveToFile: data marshal %w", err)
	}
	// сохраняем данные в файл
	if err := os.WriteFile(f.path, data, 0666); err != nil {
		return fmt.Errorf("SaveToFile: write file error %w", err)
	}
	return nil
}

// Load получает и конвертирует метрики из JSON формата,
// сохраняет в хранилище сервера.
func (f *File) Load(ctx context.Context, ms interfaces.MetricStorage) error {
	if _, err := os.Stat(f.path); err != nil {
		if _, err := os.Create(f.path); err != nil {
			return fmt.Errorf("LoadFromFile: file create %w", err)
		}
		if err := f.Save(ctx, ms); err != nil {
			return fmt.Errorf("LoadFromFile: data save %w", err)
		}
	}
	data, err := os.ReadFile(f.path)
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

// Ping проверяет наличие файла.
func (f *File) Ping(ctx context.Context) error {
	_, err := os.Stat(f.path)
	return err
}
