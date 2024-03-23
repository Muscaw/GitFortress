package service

import (
	"context"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
)

type MetricsPort interface {
	Start(ctx context.Context)
	Handle(metric entity.Metric, valueNames []string)
}
