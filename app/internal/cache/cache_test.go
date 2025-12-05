package cache_test

import (
	"testing"
	"time"

	"vcv/internal/cache"
)

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New(1 * time.Minute)

	c.Set("key1", "value1")

	got, found := c.Get("key1")
	if !found {
		t.Error("expected key1 to be found")
	}
	if got != "value1" {
		t.Errorf("expected value1, got %v", got)
	}
}

func TestCache_GetMissing(t *testing.T) {
	c := cache.New(1 * time.Minute)

	_, found := c.Get("nonexistent")
	if found {
		t.Error("expected nonexistent key to not be found")
	}
}

func TestCache_Expiration(t *testing.T) {
	c := cache.New(10 * time.Millisecond)

	c.Set("key1", "value1")

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be expired")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := cache.New(1 * time.Minute)

	c.Set("key1", "value1")
	c.Invalidate("key1")

	_, found := c.Get("key1")
	if found {
		t.Error("expected key1 to be invalidated")
	}
}

func TestCache_Clear(t *testing.T) {
	c := cache.New(1 * time.Minute)

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Clear()

	_, found1 := c.Get("key1")
	_, found2 := c.Get("key2")
	if found1 || found2 {
		t.Error("expected all keys to be cleared")
	}
}

func TestCache_Cleanup(t *testing.T) {
	c := cache.New(10 * time.Millisecond)

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Add a fresh key
	c.Set("key3", "value3")

	// Cleanup should remove expired entries
	c.Cleanup()

	_, found1 := c.Get("key1")
	_, found2 := c.Get("key2")
	_, found3 := c.Get("key3")

	if found1 || found2 {
		t.Error("expected expired keys to be cleaned up")
	}
	if !found3 {
		t.Error("expected fresh key to remain after cleanup")
	}
}
