package agent

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Address, "a", "localhost:8080", "HTTP-server endpoint address host:port")
	flag.IntVar(&cfg.PollInterval, "p", 2, "Frequency of metrics polling from the runtime package")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Frequency of sending metrics to HTTP-server")

	flag.Parse()

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
