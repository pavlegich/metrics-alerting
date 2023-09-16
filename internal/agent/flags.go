package agent

import (
	"context"
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
}

func ParseFlags(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Frequency of metrics polling from the runtime package")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Frequency of sending metrics to HTTP-server")
	flag.StringVar(&cfg.Key, "k", "", "Key for sign")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return cfg, fmt.Errorf("ParseFlags: environment values not parsed %w", err)
	}

	return cfg, nil
}
