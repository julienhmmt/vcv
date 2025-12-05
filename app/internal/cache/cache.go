package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with TTL
type CacheEntry struct {
	Data      any
	ExpiresAt time.Time
}

// Cache provides thread-safe in-memory caching with TTL
type Cache struct {
	mu   sync.RWMutex
	data map[string]*CacheEntry
	ttl  time.Duration
}

// New creates a new cache with specified TTL
func New(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]*CacheEntry),
		ttl:  ttl,
	}
}

// Get retrieves cached data if not expired
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[key]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Data, true
}

// Set stores data with TTL
func (c *Cache) Set(key string, data any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a specific key from cache
func (c *Cache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear removes all cached entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*CacheEntry)
}

// Cleanup removes expired entries (should be called periodically)
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.ExpiresAt) {
			delete(c.data, key)
		}
	}
}
