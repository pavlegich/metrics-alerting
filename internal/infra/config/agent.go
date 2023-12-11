package config

import (
	"context"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

// AgentConfig содержит значения флагов и переменных окружения агента.
type AgentConfig struct {
	Address        string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
}

// AgentParseFlags обрабатывает введённые значения флагов и переменных окружения
// при запуске агента.
func AgentParseFlags(ctx context.Context) (*AgentConfig, error) {
	cfg := &AgentConfig{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Frequency of metrics polling from the runtime package")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Frequency of sending metrics to HTTP-server")
	flag.StringVar(&cfg.Key, "k", "", "Key for sign")
	flag.IntVar(&cfg.RateLimit, "l", 1, "Number of simultaneous requests to the server")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "Path to public key")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: environment values not parsed %w", err)
	}

	return cfg, nil
}
