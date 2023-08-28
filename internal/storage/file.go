package storage

import (
	"encoding/json"
	"os"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

func SaveStorage(path string, m *interfaces.MetricStorage) error {
	// сериализуем структуру в JSON формат
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// сохраняем данные в файл
	return os.WriteFile(path, data, 0666)
}

func LoadStorage(path string) (*MemStorage, error) {
	storage := &MemStorage{}

	data, err := os.ReadFile(path)
	if err != nil {
		return storage, err
	}

	if err := json.Unmarshal(data, storage); err != nil {
		return storage, err
	}

	return storage, nil
}
