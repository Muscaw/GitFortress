package entity

type Counter interface {
	Metric

	Values() map[string]int
	Increment(valueName string)
}

func convertMap(originalMap map[string]int) map[string]any {
	convertedMap := make(map[string]any)
	for key, value := range originalMap {
		convertedMap[key] = any(value)
	}
	return convertedMap
}

type counter struct {
	name     string
	values   map[string]int
	registry MetricsRegistry
}

func (c *counter) Values() map[string]int {
	return c.values
}

func (c *counter) Name() string {
	return c.name
}

func (c *counter) Increment(valueName string) {
	// No need to check for the key existence. Default value for int is return in case of absence of key
	c.values[valueName] += 1
	convertedValues := convertMap(c.values)
	c.registry.Push(MetricInformation{metricType: COUNTER_METRIC_TYPE, metricName: c.name, values: convertedValues}, []string{valueName})
}

func NewCounter(name string, registry MetricsRegistry) Counter {
	return &counter{name: name, values: map[string]int{}, registry: registry}
}
