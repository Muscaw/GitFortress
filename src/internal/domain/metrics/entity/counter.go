package entity

type Metric interface {
	Name() string
}

type Counter interface {
	Metric

	Increment()
}

type counter struct {
	name  string
	value int
}

func (c *counter) Name() string {
	return c.name
}

func (c *counter) Increment() {
	c.value += 1
}

func NewCounter(name string) Counter {
	return &counter{name: name, value: 0}
}
