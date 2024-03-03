package metrics

import (
	"github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	metricsservice "github.com/Muscaw/GitFortress/internal/domain/metrics/service"
)

var service *metricsService

type metricsService struct {
	handlers []entity.MetricHandler
	metrics  []entity.Metric
}

func (m *metricsService) TrackCounter(name string) entity.Counter {
	c := newCounter(name)
	m.metrics = append(m.metrics, c)
	return c
}

func (m *metricsService) RegisterHandler(handler entity.MetricHandler) {
	m.handlers = append(m.handlers, handler)
}

func (m *metricsService) Start() {
	for _, handler := range m.handlers {
		go handler.Start()
	}
}

func GetMetricsService() metricsservice.MetricsService {
	if service == nil {
		service = &metricsService{handlers: []entity.MetricHandler{}}
	}
	return service
}
