package service

import (
	"context"
	"sync"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
)

type MetricsService interface {
	RegisterHandler(handler MetricsPort)
	Start(wg *sync.WaitGroup, ctx context.Context)
	TrackCounter(name string) entity.Counter
	TrackGauge(name string) entity.Gauge
}

type MetricsPort interface {
	Start(ctx context.Context, doneFunc DoneFunc)
	Handle(metric entity.MetricInformation, valueNames []string)
}

type DoneFunc func()
