package service

import (
	"context"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
)

type MetricsService interface {
	RegisterHandler(handler MetricsPort)
	Start(ctx context.Context)
	TrackCounter(name string) entity.Counter
	TrackGauge(name string) entity.Gauge
}

type MetricsPort interface {
	Start(ctx context.Context)
	Handle(metric entity.Metric, valueNames []string)
}
