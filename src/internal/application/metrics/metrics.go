package metrics

import "github.com/Muscaw/GitFortress/internal/domain/metrics/entity"

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

func newCounter(name string) entity.Counter {
	return &counter{
		name:  name,
		value: 0,
	}
}
