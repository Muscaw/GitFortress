package entity

type Metric interface {
	Name() string
}

type Counter interface {
	Metric

	Values() map[string]int
	Increment(valueName string)
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

func (c *counter) Increment(tag string) {
	// No need to check for the key existence. Default value for int is return in case of absence of key
	c.values[tag] += 1
	c.registry.Push(c)
}

func NewCounter(name string, registry MetricsRegistry) Counter {
	return &counter{name: name, values: map[string]int{}, registry: registry}
}

type MetricsRegistry interface {
	Push(metric Metric)
}
