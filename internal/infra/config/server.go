// Пакет config содержит объект и методы для инициализации элементов
// кофигурации сервера из указанных при запуске флагов и переменных окружения.
package config

import (
	"context"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// ServerConfig содержит значения флагов и переменных окружения сервера.
type ServerConfig struct {
	Address       string `env:"ADDRESS"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
	Database      string `env:"DATABASE_DSN"`
	Key           string `env:"KEY"`
	CryptoKey     string `env:"CRYPTO_KEY"`
	Restore       bool   `env:"RESTORE"`
	StoreInterval int    `env:"STORE_INTERVAL"`
}

// ServerParseFlags обрабатывает введённые значения флагов и переменных окружения
// при запуск сервера.
func ServerParseFlags(ctx context.Context) (*ServerConfig, error) {
	cfg := &ServerConfig{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.StoreInterval, "i", 10, "Frequency of storing on disk")
	flag.StringVar(&cfg.StoragePath, "f", "/tmp/metrics-db.json", "Full path of values storage")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore values from the disk")
	flag.StringVar(&cfg.Database, "d", "", "URI (DSN) to database")
	flag.StringVar(&cfg.Key, "k", "", "Key for sign")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to private key")

	flag.Parse()

	// Проверяем переменные окружения
	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: wrong environment values %w", err)
	}

	return cfg, nil
}
