package metrics

import (
	"context"
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	metricsservice "github.com/Muscaw/GitFortress/internal/domain/metrics/service"
)

var service *metricsService

type metricsService struct {
	handlers []metricsservice.MetricsHandler
	metrics  []entity.Metric
}

func (m *metricsService) RetrieveMetrics() []entity.Metric {
	return m.metrics
}

func (m *metricsService) TrackCounter(name string) entity.Counter {
	c := newCounter(name)
	m.metrics = append(m.metrics, c)
	return c
}

func (m *metricsService) RegisterHandler(handler metricsservice.MetricsHandler) {
	m.handlers = append(m.handlers, handler)
}

func (m *metricsService) Start(ctx context.Context) {
	for _, handler := range m.handlers {
		go handler.Start(ctx)
	}
}

func GetMetricsService() metricsservice.MetricsService {
	if service == nil {
		service = &metricsService{handlers: []metricsservice.MetricsHandler{}}
	}
	return service
}
