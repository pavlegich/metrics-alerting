package config

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	conf "github.com/pavlegich/metrics-alerting/internal/utils/config"
)

// AgentConfig содержит значения флагов и переменных окружения агента.
type AgentConfig struct {
	Address        string `env:"ADDRESS" json:"address"`
	Key            string `env:"KEY" json:"key"`
	CryptoKey      string `env:"CRYPTO_KEY" json:"crypto_key"`
	Config         string `env:"CONFIG"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval int    `env:"REPORT_INTERVAL" json:"report_interval"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit"`
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
	flag.StringVar(&cfg.Config, "config", "/Users/Pavel/Desktop/Go.Edu/metrics-alerting/internal/infra/config/agent_config.json", "Path to config")
	flag.StringVar(&cfg.Config, "c", cfg.Config, "alias for -config")

	flag.Parse()

	// Проверка наличия пути к файлу конфигурации для флагов
	if cfg.Config != "" {
		cfg.parseConfig(ctx)
	}

	// Проверяем переменные окружения
	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: environment values not parsed %w", err)
	}

	return cfg, nil
}

func (cfg *AgentConfig) parseConfig(ctx context.Context) error {
	f, err := os.Open(cfg.Config)
	if err != nil {
		return fmt.Errorf("parseConfig: open file failed %w", err)
	}
	defer f.Close()

	data, err := os.ReadFile(cfg.Config)
	if err != nil {
		return fmt.Errorf("parseConfig: read file failed %w", err)
	}

	fc := &AgentConfig{}

	err = json.Unmarshal(data, &fc)
	if err != nil {
		return fmt.Errorf("parseConfig: unmarshal flags failed %w", err)
	}

	if !conf.IsFlagPassed("a") && fc.Address != "" {
		cfg.Address = fc.Address
	}
	if !conf.IsFlagPassed("p") && fc.PollInterval != 0 {
		cfg.PollInterval = fc.PollInterval
	}
	if !conf.IsFlagPassed("r") && fc.ReportInterval != 0 {
		cfg.ReportInterval = fc.ReportInterval
	}
	if !conf.IsFlagPassed("k") && fc.Key != "" {
		cfg.Key = fc.Key
	}
	if !conf.IsFlagPassed("l") && fc.RateLimit != 0 {
		cfg.RateLimit = fc.RateLimit
	}
	if !conf.IsFlagPassed("crypto-key") && fc.CryptoKey != "" {
		cfg.CryptoKey = fc.CryptoKey
	}

	return nil
}
