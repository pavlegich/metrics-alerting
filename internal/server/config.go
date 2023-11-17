package server

import (
	"context"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// Config содержит значения флагов и переменных окружения сервера.
type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
	Database      string `env:"DATABASE_DSN"`
	Key           string `env:"KEY"`
}

// ParseFlags обрабатывает введённые значения флагов и переменных окружения
// при запуск сервера.
func ParseFlags(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Frequency of storing on disk")
	flag.StringVar(&cfg.StoragePath, "f", "/tmp/metrics-db.json", "Full path of values storage")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore values from the disk")
	flag.StringVar(&cfg.Database, "d", "", "URI (DSN) to database")
	flag.StringVar(&cfg.Key, "k", "", "Key for sign")

	flag.Parse()

	// Проверяем переменные окружения
	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: wrong environment values %w", err)
	}

	return cfg, nil
}
