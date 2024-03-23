package entity

type Metric interface {
	Name() string
}

type MetricsRegistry interface {
	Push(metric Metric, valueNames []string)
}
