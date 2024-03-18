package metrics

import (
	"context"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	metricsservice "github.com/Muscaw/GitFortress/internal/domain/metrics/service"
)

var service *metricsService

type metricsService struct {
	handlers []metricsservice.MetricsPort
}

func (m *metricsService) Push(metric entity.Metric) {
	for _, handler := range m.handlers {
		handler.Handle(metric)
	}
}

func (m *metricsService) TrackCounter(name string) entity.Counter {
	c := entity.NewCounter(name, m)
	return c
}

func (m *metricsService) RegisterHandler(handler metricsservice.MetricsPort) {
	m.handlers = append(m.handlers, handler)
}

func (m *metricsService) Start(ctx context.Context) {
	for _, handler := range m.handlers {
		go handler.Start(ctx)
	}
}

func GetMetricsService() metricsservice.MetricsService {
	if service == nil {
		service = &metricsService{handlers: []metricsservice.MetricsPort{}}
	}
	return service
}
