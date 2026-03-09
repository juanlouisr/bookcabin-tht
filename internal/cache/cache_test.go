package cache

import (
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	// Set a value
	cache.Set("key1", "value1", 5*time.Minute)

	// Get the value
	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestMemoryCache_Get_NotFound(t *testing.T) {
	cache := NewMemoryCache()

	// Try to get non-existent key
	val, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}

func TestMemoryCache_Get_Expired(t *testing.T) {
	cache := NewMemoryCache()

	// Set a value with very short TTL
	cache.Set("key1", "value1", 1*time.Millisecond)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Try to get expired key
	val, found := cache.Get("key1")
	if found {
		t.Error("Expected not to find expired key")
	}
	if val != nil {
		t.Errorf("Expected nil for expired key, got %v", val)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()

	// Set and then delete
	cache.Set("key1", "value1", 5*time.Minute)
	cache.Delete("key1")

	// Try to get deleted key
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected not to find deleted key")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()

	// Set multiple values
	cache.Set("key1", "value1", 5*time.Minute)
	cache.Set("key2", "value2", 5*time.Minute)
	cache.Set("key3", "value3", 5*time.Minute)

	// Clear all
	cache.Clear()

	// Verify all keys are gone
	for _, key := range []string{"key1", "key2", "key3"} {
		if _, found := cache.Get(key); found {
			t.Errorf("Expected not to find %s after clear", key)
		}
	}
}

func TestMemoryCache_Overwrite(t *testing.T) {
	cache := NewMemoryCache()

	// Set initial value
	cache.Set("key1", "value1", 5*time.Minute)

	// Overwrite with new value
	cache.Set("key1", "value2", 5*time.Minute)

	// Get the value
	val, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if val != "value2" {
		t.Errorf("Expected value2, got %v", val)
	}
}
