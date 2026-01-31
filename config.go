package cache

import (
	"time"
)

// Backend represents the storage backend type
type Backend string

const (
	// BackendMemory uses in-memory storage (single instance)
	BackendMemory Backend = "memory"
	
	// BackendRedis uses Redis storage (distributed)
	BackendRedis Backend = "redis"
)

// Config holds the cache configuration
type Config struct {
	// Backend specifies the storage backend
	// Default: BackendMemory
	Backend Backend

	// RedisURL is the Redis connection URL
	// Format: redis://[:password@]host[:port][/db]
	// Required if Backend is BackendRedis
	RedisURL string

	// DefaultTTL is the default expiration time for cache entries
	// Default: 1 hour
	DefaultTTL time.Duration

	// CleanupInterval is how often to clean expired entries (memory backend only)
	// Default: 10 minutes
	CleanupInterval time.Duration
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Backend:         BackendMemory,
		DefaultTTL:      1 * time.Hour,
		CleanupInterval: 10 * time.Minute,
	}
}

// WithBackend sets the storage backend
func (c *Config) WithBackend(backend Backend) *Config {
	c.Backend = backend
	return c
}

// WithRedisURL sets the Redis connection URL
func (c *Config) WithRedisURL(url string) *Config {
	c.RedisURL = url
	return c
}

// WithDefaultTTL sets the default TTL
func (c *Config) WithDefaultTTL(ttl time.Duration) *Config {
	c.DefaultTTL = ttl
	return c
}
