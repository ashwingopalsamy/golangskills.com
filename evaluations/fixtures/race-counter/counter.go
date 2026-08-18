package racecounter

type Counter struct {
	accepted int
	posted   int
	version  int
}

func (c *Counter) Add() {
	c.accepted++
	c.posted++
	c.version++
}

func (c *Counter) Snapshot() (accepted, posted, version int) {
	return c.accepted, c.posted, c.version
}
