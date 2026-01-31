package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/OkanUysal/go-cache"
)

func TestMemoryCache(t *testing.T) {
	ctx := context.Background()

	c, err := cache.New(&cache.Config{
		Backend: cache.BackendMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	// Test Set/Get
	err = c.Set(ctx, "key1", "value1")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	value, err := c.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test Has
	if !c.Has(ctx, "key1") {
		t.Error("Key should exist")
	}

	// Test Delete
	err = c.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	if c.Has(ctx, "key1") {
		t.Error("Key should not exist after delete")
	}
}

func TestExpiration(t *testing.T) {
	ctx := context.Background()

	c, err := cache.New(&cache.Config{
		Backend: cache.BackendMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	// Set with short TTL
	err = c.SetWithTTL(ctx, "expire_key", "expire_value", 100*time.Millisecond)
	if err != nil {
		t.Errorf("SetWithTTL failed: %v", err)
	}

	// Should exist immediately
	if !c.Has(ctx, "expire_key") {
		t.Error("Key should exist")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after expiration
	if c.Has(ctx, "expire_key") {
		t.Error("Key should have expired")
	}
}

func TestIncrement(t *testing.T) {
	ctx := context.Background()

	c, err := cache.New(&cache.Config{
		Backend: cache.BackendMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	// Increment non-existent key
	val, err := c.Increment(ctx, "counter", 1)
	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}
	if val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	// Increment again
	val, err = c.Increment(ctx, "counter", 5)
	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}
	if val != 6 {
		t.Errorf("Expected 6, got %d", val)
	}

	// Decrement
	val, err = c.Decrement(ctx, "counter", 2)
	if err != nil {
		t.Errorf("Decrement failed: %v", err)
	}
	if val != 4 {
		t.Errorf("Expected 4, got %d", val)
	}
}

func TestJSON(t *testing.T) {
	ctx := context.Background()

	c, err := cache.New(&cache.Config{
		Backend: cache.BackendMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	type User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := User{ID: 123, Name: "John"}

	// Set JSON
	err = c.SetJSON(ctx, "user:123", user)
	if err != nil {
		t.Errorf("SetJSON failed: %v", err)
	}

	// Get JSON
	var retrieved User
	err = c.GetJSON(ctx, "user:123", &retrieved)
	if err != nil {
		t.Errorf("GetJSON failed: %v", err)
	}

	if retrieved.ID != user.ID || retrieved.Name != user.Name {
		t.Errorf("Expected %+v, got %+v", user, retrieved)
	}
}

func TestGetOrSet(t *testing.T) {
	ctx := context.Background()

	c, err := cache.New(&cache.Config{
		Backend: cache.BackendMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	fetchCount := 0
	fetcher := func() (interface{}, error) {
		fetchCount++
		return "fetched_value", nil
	}

	// First call should fetch
	value, err := c.GetOrSet(ctx, "lazy_key", fetcher, 1*time.Hour)
	if err != nil {
		t.Errorf("GetOrSet failed: %v", err)
	}
	if value != "fetched_value" {
		t.Errorf("Expected fetched_value, got %v", value)
	}
	if fetchCount != 1 {
		t.Errorf("Expected fetch count 1, got %d", fetchCount)
	}

	// Second call should use cache
	value, err = c.GetOrSet(ctx, "lazy_key", fetcher, 1*time.Hour)
	if err != nil {
		t.Errorf("GetOrSet failed: %v", err)
	}
	if value != "fetched_value" {
		t.Errorf("Expected fetched_value, got %v", value)
	}
	if fetchCount != 1 {
		t.Errorf("Expected fetch count still 1, got %d", fetchCount)
	}
}

func TestMultiOperations(t *testing.T) {
	ctx := context.Background()

	c, err := cache.New(&cache.Config{
		Backend: cache.BackendMemory,
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	// SetMany
	items := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	err = c.SetMany(ctx, items, 1*time.Hour)
	if err != nil {
		t.Errorf("SetMany failed: %v", err)
	}

	// GetMany
	results, err := c.GetMany(ctx, []string{"key1", "key2", "key3"})
	if err != nil {
		t.Errorf("GetMany failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	// DeleteMany
	err = c.DeleteMany(ctx, []string{"key1", "key2"})
	if err != nil {
		t.Errorf("DeleteMany failed: %v", err)
	}

	if c.Has(ctx, "key1") || c.Has(ctx, "key2") {
		t.Error("Keys should be deleted")
	}

	if !c.Has(ctx, "key3") {
		t.Error("key3 should still exist")
	}
}
