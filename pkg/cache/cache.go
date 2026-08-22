// Package cache provides in-memory TTL caching mechanisms.
package cache

import (
	"sync"
	"time"
)

type item struct {
	value      any
	expiration int64 // Unix nano
}

// MemoryCache is a thread-safe in-memory key-value cache with TTL expiration.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]item
	ttl   time.Duration
}

// NewMemoryCache creates a new in-memory cache with the default TTL.
func NewMemoryCache(defaultTTL time.Duration) *MemoryCache {
	return &MemoryCache{
		items: make(map[string]item),
		ttl:   defaultTTL,
	}
}

// Get retrieves an item by key if it exists and has not expired.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}

	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		// Clean up expired item
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}

	return it.value, true
}

// Set stores an item with the default TTL.
func (c *MemoryCache) Set(key string, val any) {
	c.SetWithTTL(key, val, c.ttl)
}

// SetWithTTL stores an item with a custom TTL.
func (c *MemoryCache) SetWithTTL(key string, val any, ttl time.Duration) {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	c.items[key] = item{
		value:      val,
		expiration: exp,
	}
	c.mu.Unlock()
}

// Delete removes an item by key.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Clear removes all items from the cache.
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	c.items = make(map[string]item)
	c.mu.Unlock()
}
