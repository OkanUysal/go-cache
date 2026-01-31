package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore implements a Redis-backed cache
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Redis-backed cache
func NewRedisStore(redisURL string) (*RedisStore, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisStore{
		client: client,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisStore) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return val, nil
}

// Set stores a value in Redis
func (r *RedisStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Serialize value to JSON if it's not a string
	var data interface{}
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = v
	default:
		jsonData, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data = jsonData
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a value from Redis
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Has checks if a key exists in Redis
func (r *RedisStore) Has(ctx context.Context, key string) bool {
	count, err := r.client.Exists(ctx, key).Result()
	return err == nil && count > 0
}

// Increment increments a numeric value in Redis
func (r *RedisStore) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.IncrBy(ctx, key, delta).Result()
}

// Decrement decrements a numeric value in Redis
func (r *RedisStore) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return r.client.DecrBy(ctx, key, delta).Result()
}

// Clear removes all entries from Redis (dangerous!)
func (r *RedisStore) Clear(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// GetJSON retrieves and unmarshals JSON data
func (r *RedisStore) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// SetJSON marshals and stores JSON data
func (r *RedisStore) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

// Expire sets a new TTL for a key
func (r *RedisStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// TTL returns the remaining time to live for a key
func (r *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Ping checks if Redis is available
func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// IncrementWithExpiry increments and sets expiry
func (r *RedisStore) IncrementWithExpiry(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	pipe := r.client.Pipeline()
	incrCmd := pipe.IncrBy(ctx, key, delta)
	pipe.Expire(ctx, key, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return incrCmd.Val(), nil
}

// GetClient returns the underlying Redis client for advanced operations
func (r *RedisStore) GetClient() *redis.Client {
	return r.client
}

var (
	ErrRedisUnavailable = errors.New("redis unavailable")
)
