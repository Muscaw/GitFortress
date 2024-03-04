package service

import (
	"context"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
)

type MetricsHandler interface {
	Start(ctx context.Context)
	AddMetric(metric entity.Metric)
}
