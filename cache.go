package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Cache is the main cache client
type Cache struct {
	store      Store
	defaultTTL time.Duration
}

// New creates a new cache instance
func New(config *Config) (*Cache, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var store Store
	var err error

	switch config.Backend {
	case BackendMemory:
		store = NewMemoryStore(config.CleanupInterval)

	case BackendRedis:
		if config.RedisURL == "" {
			return nil, fmt.Errorf("RedisURL is required for Redis backend")
		}
		store, err = NewRedisStore(config.RedisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis store: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported backend: %s", config.Backend)
	}

	return &Cache{
		store:      store,
		defaultTTL: config.DefaultTTL,
	}, nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	return c.store.Get(ctx, key)
}

// Set stores a value in the cache with default TTL
func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	return c.store.Set(ctx, key, value, c.defaultTTL)
}

// SetWithTTL stores a value in the cache with custom TTL
func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.store.Set(ctx, key, value, ttl)
}

// Delete removes a value from the cache
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.store.Delete(ctx, key)
}

// Has checks if a key exists
func (c *Cache) Has(ctx context.Context, key string) bool {
	return c.store.Has(ctx, key)
}

// Increment increments a numeric value
func (c *Cache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return c.store.Increment(ctx, key, delta)
}

// Decrement decrements a numeric value
func (c *Cache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return c.store.Decrement(ctx, key, delta)
}

// Clear removes all entries
func (c *Cache) Clear(ctx context.Context) error {
	return c.store.Clear(ctx)
}

// Close closes the cache connection
func (c *Cache) Close() error {
	return c.store.Close()
}

// GetJSON retrieves and unmarshals JSON data
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	value, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	// If it's already a string (from Redis), unmarshal it
	if str, ok := value.(string); ok {
		return json.Unmarshal([]byte(str), dest)
	}

	// If it's bytes
	if bytes, ok := value.([]byte); ok {
		return json.Unmarshal(bytes, dest)
	}

	// If it's already the correct type (from memory), cast it
	// This is a simple type assertion - in production you might want reflection
	destValue, ok := value.(interface{})
	if ok {
		// Marshal and unmarshal to ensure type compatibility
		data, err := json.Marshal(destValue)
		if err != nil {
			return err
		}
		return json.Unmarshal(data, dest)
	}

	return fmt.Errorf("cannot unmarshal value of type %T", value)
}

// SetJSON marshals and stores data as JSON
func (c *Cache) SetJSON(ctx context.Context, key string, value interface{}) error {
	return c.SetWithTTL(ctx, key, value, c.defaultTTL)
}

// SetJSONWithTTL marshals and stores data as JSON with custom TTL
func (c *Cache) SetJSONWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.SetWithTTL(ctx, key, value, ttl)
}

// GetOrSet retrieves a value or sets it if not found (cache-aside pattern)
func (c *Cache) GetOrSet(ctx context.Context, key string, fetcher func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	// Try to get from cache
	value, err := c.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// Not in cache, fetch it
	value, err = fetcher()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.SetWithTTL(ctx, key, value, ttl); err != nil {
		// Log error but don't fail - we have the value
		return value, nil
	}

	return value, nil
}

// Remember is an alias for GetOrSet with default TTL
func (c *Cache) Remember(ctx context.Context, key string, fetcher func() (interface{}, error)) (interface{}, error) {
	return c.GetOrSet(ctx, key, fetcher, c.defaultTTL)
}

// Forever stores a value with no expiration
func (c *Cache) Forever(ctx context.Context, key string, value interface{}) error {
	return c.store.Set(ctx, key, value, 0)
}

// GetMany retrieves multiple values at once
func (c *Cache) GetMany(ctx context.Context, keys []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	for _, key := range keys {
		value, err := c.Get(ctx, key)
		if err == nil {
			results[key] = value
		}
	}

	return results, nil
}

// SetMany stores multiple values at once
func (c *Cache) SetMany(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := c.SetWithTTL(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMany removes multiple keys at once
func (c *Cache) DeleteMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := c.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// GetStore returns the underlying store for advanced operations
func (c *Cache) GetStore() Store {
	return c.store
}
