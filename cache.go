package gocache

import (
	"gocache/twoqueues"
	"sync"
)

type cache struct {
	mu          sync.Mutex
	twoq        *twoqueues.Cache
	lEntries    int
	fEntries    int
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.twoq == nil{
		c.twoq = twoqueues.New(c.lEntries, c.fEntries)
	}
	c.twoq.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.twoq == nil {
		return
	}

	if v, ok := c.twoq.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
