package metrics

import (
	"context"
	"sync"

	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	metricsservice "github.com/Muscaw/GitFortress/internal/domain/metrics/service"
)

var service *metricsService

type metricsService struct {
	handlers []metricsservice.MetricsPort
}

func (m *metricsService) Push(metric entity.MetricInformation, valueNames []string) {
	for _, handler := range m.handlers {
		handler.Handle(metric, valueNames)
	}
}

func (m *metricsService) TrackCounter(name string) entity.Counter {
	c := entity.NewCounter(name, m)
	return c
}

func (m *metricsService) TrackGauge(name string) entity.Gauge {
	g := entity.NewGauge(name, m)
	return g
}

func (m *metricsService) RegisterHandler(handler metricsservice.MetricsPort) {
	m.handlers = append(m.handlers, handler)
}

func (m *metricsService) Start(wg *sync.WaitGroup, ctx context.Context) {
	for _, handler := range m.handlers {
		wg.Add(1)
		go handler.Start(ctx, func() {
			wg.Done()
		})
	}
}

func newMetricsService() *metricsService {
	return &metricsService{handlers: make([]metricsservice.MetricsPort, 0)}
}

func GetMetricsService() metricsservice.MetricsService {
	if service == nil {
		service = newMetricsService()
	}
	return service
}
