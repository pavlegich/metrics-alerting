package config

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

// ServerConfig содержит значения флагов и переменных окружения сервера.
type ServerConfig struct {
	Address       string `env:"ADDRESS" json:"address"`
	StoragePath   string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Database      string `env:"DATABASE_DSN" json:"database_dsn"`
	Key           string `env:"KEY" json:"key"`
	CryptoKey     string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config        string `env:"CONFIG"`
	Restore       bool   `env:"RESTORE" json:"restore"`
	StoreInterval int    `env:"STORE_INTERVAL" json:"store_interval"`
}

// ServerParseFlags обрабатывает введённые значения флагов и переменных окружения
// при запуск сервера.
func ServerParseFlags(ctx context.Context) (*ServerConfig, error) {
	cfg := &ServerConfig{}

	flag.StringVar(&cfg.Config, "config", "", "Path to config")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "alias for -config")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.StoreInterval, "i", 3, "Frequency of storing on disk")
	flag.StringVar(&cfg.StoragePath, "f", "/tmp/metrics-db.json", "Full path of values storage")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore values from the disk")
	flag.StringVar(&cfg.Database, "d", "postgresql://localhost:5432/metrics", "URI (DSN) to database")
	flag.StringVar(&cfg.Key, "k", "", "Key for sign")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to private key")

	flag.Parse()

	// Проверка наличия пути к файлу конфигурации для флагов
	if cfg.Config != "" {
		cfg.parseConfig(ctx)
	}

	flag.Parse()

	// Проверяем переменные окружения
	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: wrong environment values %w", err)
	}

	return cfg, nil
}

// parseConfig обрабатывает файл конфигурации для сервера.
func (cfg *ServerConfig) parseConfig(ctx context.Context) error {
	f, err := os.Open(cfg.Config)
	if err != nil {
		return fmt.Errorf("parseConfig: open file failed %w", err)
	}
	defer f.Close()

	data, err := os.ReadFile(cfg.Config)
	if err != nil {
		return fmt.Errorf("parseConfig: read file failed %w", err)
	}

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return fmt.Errorf("parseConfig: unmarshal flags failed %w", err)
	}

	return nil
}
