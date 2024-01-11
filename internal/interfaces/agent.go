package interfaces

import (
	"context"

	"github.com/pavlegich/metrics-alerting/internal/infra/config"
)

// Agent содержит методы для работы агента.
type Agent interface {
	SendStats(ctx context.Context, statsStorage StatsStorage, cfg *config.AgentConfig)
}
