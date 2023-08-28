package server

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	StoragePath   string `env:"FILE_STORAGE_PATH"`
	Restore       bool   `env:"RESTORE"`
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "Frequency of storing on disk")
	flag.StringVar(&cfg.StoragePath, "f", "/tmp/metrics-db.json", "Full path of values storage")
	flag.BoolVar(&cfg.Restore, "r", true, "Restore values from the disk")

	flag.Parse()

	// Проверяем переменные окружения
	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
