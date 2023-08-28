package storage

import (
	"encoding/json"
	"os"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

func Save(path string, m *interfaces.MetricStorage) error {
	// сериализуем структуру в JSON формат
	data, err := json.Marshal(&m)
	if err != nil {
		return err
	}
	// сохраняем данные в файл
	return os.WriteFile(path, data, 0666)
}

func Load(path string, m *interfaces.MetricStorage) error {
	if _, err := os.Stat(path); err != nil {
		if _, err := os.Create(path); err != nil {
			return err
		}
		if err := Save(path, m); err != nil {
			return err
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	return nil
}
