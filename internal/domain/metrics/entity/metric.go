package entity

const (
	COUNTER_METRIC_TYPE = "counter"
	GAUGE_METRIC_TYPE   = "gauge"
)

type Metric interface {
	Name() string
}

type MetricInformation struct {
	metricType string
	metricName string
	values     map[string]any
}

func (m MetricInformation) MetricType() string {
	return m.metricType
}

func (m MetricInformation) MetricName() string {
	return m.metricName
}

func (m MetricInformation) Values() map[string]any {
	return m.values
}

type MetricsRegistry interface {
	Push(metric MetricInformation, valueNames []string)
}
