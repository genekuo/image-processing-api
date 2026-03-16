// Package cache provides a thread-safe in-memory cache with TTL-based eviction.
// Cached entries store PNG-encoded image bytes and are evicted when they have
// not been accessed within the configured TTL duration.
package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

// entry represents a single cached item with its data and last access timestamp.
type entry struct {
	data       []byte
	lastAccess time.Time
}

// Cache is a thread-safe in-memory cache that evicts entries after a
// configurable TTL of idle time. It tracks cache hits and misses via
// atomic counters.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]*entry
	ttl     time.Duration
	done    chan struct{}
	hits    atomic.Int64
	misses  atomic.Int64
}

// New creates a new Cache with the given TTL and starts a background
// eviction goroutine that runs every 30 seconds. Call Stop to cleanly
// shut down the eviction goroutine when the cache is no longer needed.
func New(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]*entry),
		ttl:     ttl,
		done:    make(chan struct{}),
	}
	go c.evictionLoop()
	return c
}

// Get retrieves the cached data for the given key. If the key exists, the
// entry's last access time is updated (resetting the TTL) and the hit counter
// is incremented. If the key does not exist, the miss counter is incremented.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		c.misses.Add(1)
		return nil, false
	}

	// Update last access time under a write lock.
	c.mu.Lock()
	e.lastAccess = time.Now()
	c.mu.Unlock()

	c.hits.Add(1)
	return e.data, true
}

// Set stores data in the cache under the given key with the current timestamp.
func (c *Cache) Set(key string, data []byte) {
	c.mu.Lock()
	c.entries[key] = &entry{
		data:       data,
		lastAccess: time.Now(),
	}
	c.mu.Unlock()
}

// Len returns the current number of entries in the cache.
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Hits returns the total number of cache hits since creation.
func (c *Cache) Hits() int64 {
	return c.hits.Load()
}

// Misses returns the total number of cache misses since creation.
func (c *Cache) Misses() int64 {
	return c.misses.Load()
}

// Stop cleanly shuts down the background eviction goroutine.
func (c *Cache) Stop() {
	close(c.done)
}

// evictionLoop runs in a background goroutine and removes entries that have
// not been accessed within the configured TTL. It checks every 30 seconds.
func (c *Cache) evictionLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.evict()
		}
	}
}

// evict removes all entries whose last access time exceeds the TTL.
func (c *Cache) evict() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, e := range c.entries {
		if now.Sub(e.lastAccess) > c.ttl {
			delete(c.entries, key)
		}
	}
}
