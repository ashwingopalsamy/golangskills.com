package versionedcache

import "sync"

type Event struct {
	Key     string
	Value   string
	Version uint64
	Deleted bool
}

type Cache struct {
	mu     sync.RWMutex
	values map[string]string
}

func New() *Cache {
	return &Cache{values: make(map[string]string)}
}

func (c *Cache) Apply(event Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if event.Deleted {
		delete(c.values, event.Key)
		return
	}
	c.values[event.Key] = event.Value
}

func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.values[key]
	return value, ok
}
