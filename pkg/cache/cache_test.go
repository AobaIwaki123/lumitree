package cache

import (
	"testing"
	"time"
)

func TestMemoryCache(t *testing.T) {
	c := NewMemoryCache(50 * time.Millisecond)

	c.Set("key1", "val1")

	val, ok := c.Get("key1")
	if !ok || val != "val1" {
		t.Errorf("expected val1, got %v (ok=%v)", val, ok)
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}

	// Test custom TTL
	c.SetWithTTL("key2", "val2", 200*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	val, ok = c.Get("key2")
	if !ok || val != "val2" {
		t.Errorf("expected key2 to still exist, got %v (ok=%v)", val, ok)
	}

	// Test Delete
	c.Delete("key2")
	_, ok = c.Get("key2")
	if ok {
		t.Error("expected key2 to be deleted")
	}
}
