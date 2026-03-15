package checker

import (
	"sync"
	"time"
)

type cacheEntry struct {
	result Result
	expiry time.Time
}

type resultCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func newCache(ttl time.Duration) *resultCache {
	return &resultCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *resultCache) get(domain string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[domain]
	if !ok || time.Now().After(e.expiry) {
		return Result{}, false
	}
	return e.result, true
}

func (c *resultCache) set(domain string, result Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[domain] = cacheEntry{
		result: result,
		expiry: time.Now().Add(c.ttl),
	}
}
