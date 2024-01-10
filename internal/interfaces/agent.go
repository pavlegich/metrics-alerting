package interfaces

import (
	"context"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
)

type Agent interface {
	SendStats(ctx context.Context, statsStorage StatsStorage, cfg *config.AgentConfig)
}
