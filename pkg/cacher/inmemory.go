package cacher

import (
	"fmt"
	"sync"

	"pulley.com/shakesearch/pkg/searcher"
)

type inMemoryCacher struct {
	mu    sync.RWMutex
	store map[string]searcher.Response
}

func NewInMemoryCacher() *inMemoryCacher {
	return &inMemoryCacher{
		store: make(map[string]searcher.Response),
	}
}

func (c *inMemoryCacher) Get(key string) (searcher.Response, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.store == nil {
		return searcher.Response{}, fmt.Errorf("InMemoryCacher: no store")
	}

	content, ok := c.store[key]
	if !ok {
		return searcher.Response{}, fmt.Errorf("InMemoryCacher: no cache for %s", key)
	}

	return content, nil
}

func (c *inMemoryCacher) Set(key string, content searcher.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = content
}
