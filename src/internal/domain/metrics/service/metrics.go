package service

import "github.com/Muscaw/GitFortress/internal/domain/metrics/entity"

type MetricsService interface {
	RegisterHandler(handler MetricsHandler)
	Start()
}

type MetricsRegistry interface {
	RetrieveMetrics() []entity.Metric
}
