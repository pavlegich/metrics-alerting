package storage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

type FileMetrics struct {
	Metrics map[string]string `json:"metrics"`
}

func (fm *FileMetrics) Set(mName string, mValue string) error {
	fm.Metrics[mName] = mValue
	return nil
}

func NewFileMetrics() *FileMetrics {
	return &FileMetrics{Metrics: make(map[string]string)}
}

func Save(path string, ms interfaces.MetricStorage) error {
	// сериализуем структуру в JSON формат
	metrics, status := ms.GetAll()
	if status != http.StatusOK {
		return fmt.Errorf("Save: metrics get error %v", status)
	}

	storage := NewFileMetrics()
	for m, v := range metrics {
		storage.Set(m, v)
	}

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("Save: data marshal %w", err)
	}
	// сохраняем данные в файл
	if err := os.WriteFile(path, data, 0666); err != nil {
		return fmt.Errorf("Save: write file error %w", err)
	}
	return nil
}

func Load(path string, ms interfaces.MetricStorage) error {
	if _, err := os.Stat(path); err != nil {
		if _, err := os.Create(path); err != nil {
			return fmt.Errorf("Load: file create %w", err)
		}
		if err := Save(path, ms); err != nil {
			return fmt.Errorf("Load: data save %w", err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Load: read file error %w", err)
	}

	storage := NewFileMetrics()

	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("Load: data unmarshal %w", err)
	}

	for m, v := range storage.Metrics {
		// Сейчас все пусть будут gauge, чтобы ошибок с конвертацией не было, он не записывает тип в storage
		// Впоследствии сделаю, чтобы в storage хранились отдельно gauge и counter, не все string
		if status := ms.Put("gauge", m, v); status != http.StatusOK {
			return fmt.Errorf("Load: get all metrics status %v", status)
		}
	}

	return nil
}
