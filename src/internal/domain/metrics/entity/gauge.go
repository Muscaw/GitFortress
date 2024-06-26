package entity

type Gauge interface {
	Metric

	Values() map[string]any
	SetFloat(valueName string, value float64)
	SetInt(valueName string, value int)
	SetInts(values map[string]int)
}

type gauge struct {
	name     string
	values   map[string]any
	registry MetricsRegistry
}

func (g *gauge) Values() map[string]any {
	return g.values
}

func (g *gauge) Name() string {
	return g.name
}

func (g *gauge) SetFloat(valueName string, value float64) {
	g.values[valueName] = value
	g.pushToRegistry([]string{valueName})
}

func (g *gauge) SetInt(valueName string, value int) {
	g.values[valueName] = value
	g.pushToRegistry([]string{valueName})
}

func (g *gauge) SetInts(values map[string]int) {
	var keys []string
	for k, v := range values {
		g.values[k] = v
		keys = append(keys, k)
	}
	g.pushToRegistry(keys)
}

func (g *gauge) pushToRegistry(keys []string) {

	g.registry.Push(MetricInformation{metricType: GAUGE_METRIC_TYPE, metricName: g.name, values: g.values}, keys)
}

func NewGauge(name string, registry MetricsRegistry) Gauge {
	return &gauge{
		name:     name,
		values:   map[string]any{},
		registry: registry,
	}
}
