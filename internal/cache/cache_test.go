package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	c := New(5 * time.Minute)
	defer c.Stop()

	data := []byte("png-image-data")
	c.Set("key1", data)

	got, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to be found in cache")
	}
	if string(got) != string(data) {
		t.Fatalf("expected %q, got %q", data, got)
	}
}

func TestGetMiss(t *testing.T) {
	c := New(5 * time.Minute)
	defer c.Stop()

	got, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected cache miss for nonexistent key")
	}
	if got != nil {
		t.Fatalf("expected nil data on miss, got %v", got)
	}
}

func TestTTLExpiration(t *testing.T) {
	c := New(100 * time.Millisecond)
	defer c.Stop()

	c.Set("expire-me", []byte("data"))

	// Verify entry exists.
	if _, ok := c.Get("expire-me"); !ok {
		t.Fatal("expected entry to exist immediately after set")
	}

	// Wait for TTL to pass, then manually trigger eviction.
	time.Sleep(200 * time.Millisecond)
	c.evict()

	if _, ok := c.Get("expire-me"); ok {
		t.Fatal("expected entry to be evicted after TTL")
	}
}

func TestGetResetsTTL(t *testing.T) {
	c := New(150 * time.Millisecond)
	defer c.Stop()

	c.Set("keep-alive", []byte("data"))

	// Access the key before TTL expires to reset it.
	time.Sleep(100 * time.Millisecond)
	if _, ok := c.Get("keep-alive"); !ok {
		t.Fatal("expected entry to still exist before TTL")
	}

	// Wait another 100ms (total 200ms from Set, but only 100ms from last Get).
	time.Sleep(100 * time.Millisecond)
	c.evict()

	// Entry should still be alive because Get reset the TTL.
	if _, ok := c.Get("keep-alive"); !ok {
		t.Fatal("expected entry to survive because Get reset the TTL")
	}

	// Now let it actually expire.
	time.Sleep(200 * time.Millisecond)
	c.evict()

	if _, ok := c.Get("keep-alive"); ok {
		t.Fatal("expected entry to be evicted after idle TTL")
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New(5 * time.Minute)
	defer c.Stop()

	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", id)
			for j := 0; j < iterations; j++ {
				c.Set(key, []byte(fmt.Sprintf("data-%d-%d", id, j)))
				c.Get(key)
				c.Len()
			}
		}(i)
	}

	wg.Wait()

	// Verify no data corruption: each goroutine's key should exist.
	for i := 0; i < goroutines; i++ {
		key := fmt.Sprintf("key-%d", i)
		if _, ok := c.Get(key); !ok {
			t.Errorf("expected key %s to exist after concurrent writes", key)
		}
	}
}

func TestStopShutsDown(t *testing.T) {
	c := New(1 * time.Second)

	c.Set("key", []byte("data"))
	c.Stop()

	// After Stop, the cache should still be readable but the eviction
	// goroutine should have exited. We verify Stop does not panic on
	// subsequent operations.
	if _, ok := c.Get("key"); !ok {
		t.Fatal("expected data to still be accessible after Stop")
	}
}

func TestStopIdempotentPanicsOnDouble(t *testing.T) {
	// Closing an already-closed channel panics. This test documents that
	// Stop should only be called once.
	c := New(1 * time.Second)
	c.Stop()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on double Stop")
		}
	}()
	c.Stop()
}

func TestHitsAndMisses(t *testing.T) {
	c := New(5 * time.Minute)
	defer c.Stop()

	// Initial counters should be zero.
	if c.Hits() != 0 {
		t.Fatalf("expected 0 hits, got %d", c.Hits())
	}
	if c.Misses() != 0 {
		t.Fatalf("expected 0 misses, got %d", c.Misses())
	}

	// Miss.
	c.Get("absent")
	if c.Misses() != 1 {
		t.Fatalf("expected 1 miss, got %d", c.Misses())
	}
	if c.Hits() != 0 {
		t.Fatalf("expected 0 hits after miss, got %d", c.Hits())
	}

	// Set then hit.
	c.Set("present", []byte("data"))
	c.Get("present")
	if c.Hits() != 1 {
		t.Fatalf("expected 1 hit, got %d", c.Hits())
	}

	// Another hit.
	c.Get("present")
	if c.Hits() != 2 {
		t.Fatalf("expected 2 hits, got %d", c.Hits())
	}

	// Another miss.
	c.Get("missing")
	if c.Misses() != 2 {
		t.Fatalf("expected 2 misses, got %d", c.Misses())
	}
}

func TestLen(t *testing.T) {
	c := New(5 * time.Minute)
	defer c.Stop()

	if c.Len() != 0 {
		t.Fatalf("expected empty cache, got len %d", c.Len())
	}

	c.Set("a", []byte("1"))
	c.Set("b", []byte("2"))
	if c.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", c.Len())
	}

	// Overwrite should not increase count.
	c.Set("a", []byte("updated"))
	if c.Len() != 2 {
		t.Fatalf("expected 2 entries after overwrite, got %d", c.Len())
	}
}

func TestEvictionSelectivity(t *testing.T) {
	c := New(100 * time.Millisecond)
	defer c.Stop()

	c.Set("old", []byte("old-data"))
	time.Sleep(150 * time.Millisecond)

	// Add a fresh entry after the old one has expired.
	c.Set("new", []byte("new-data"))
	c.evict()

	if _, ok := c.Get("old"); ok {
		t.Fatal("expected old entry to be evicted")
	}
	if _, ok := c.Get("new"); !ok {
		t.Fatal("expected new entry to survive eviction")
	}
}
