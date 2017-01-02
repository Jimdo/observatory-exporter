package main

import "sync"

type Cache struct {
	data map[string]Metrics
	mu   sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		data: map[string]Metrics{},
	}
}

func (c *Cache) ReadAll() map[string]Metrics {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.data
}

func (c *Cache) Write(key string, value Metrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}
