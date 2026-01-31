package cache

import (
	"context"
	"time"
)

// Store is the interface that all storage backends must implement
type Store interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Set stores a value in the cache with the given TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Has checks if a key exists in the cache
	Has(ctx context.Context, key string) bool

	// Increment increments a numeric value
	Increment(ctx context.Context, key string, delta int64) (int64, error)

	// Decrement decrements a numeric value
	Decrement(ctx context.Context, key string, delta int64) (int64, error)

	// Clear removes all entries from the cache
	Clear(ctx context.Context) error

	// Close closes the connection
	Close() error
}
