package cache

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned when a key is not found
	ErrNotFound = errors.New("key not found")
)

// item represents a cached item
type item struct {
	value      interface{}
	expiration int64
}

// isExpired checks if the item has expired
func (i *item) isExpired() bool {
	if i.expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > i.expiration
}

// MemoryStore implements an in-memory cache
type MemoryStore struct {
	items   map[string]*item
	mu      sync.RWMutex
	cleanup time.Duration
	stop    chan bool
}

// NewMemoryStore creates a new in-memory cache
func NewMemoryStore(cleanupInterval time.Duration) *MemoryStore {
	store := &MemoryStore{
		items:   make(map[string]*item),
		cleanup: cleanupInterval,
		stop:    make(chan bool),
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// Get retrieves a value from the cache
func (m *MemoryStore) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, found := m.items[key]
	if !found {
		return nil, ErrNotFound
	}

	if item.isExpired() {
		return nil, ErrNotFound
	}

	return item.value, nil
}

// Set stores a value in the cache
func (m *MemoryStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	m.items[key] = &item{
		value:      value,
		expiration: expiration,
	}

	return nil
}

// Delete removes a value from the cache
func (m *MemoryStore) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.items, key)
	return nil
}

// Has checks if a key exists
func (m *MemoryStore) Has(ctx context.Context, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, found := m.items[key]
	if !found {
		return false
	}

	return !item.isExpired()
}

// Increment increments a numeric value
func (m *MemoryStore) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, found := m.items[key]
	var current int64

	if found && !item.isExpired() {
		if val, ok := item.value.(int64); ok {
			current = val
		}
	}

	newValue := current + delta
	m.items[key] = &item{
		value:      newValue,
		expiration: 0,
	}

	return newValue, nil
}

// Decrement decrements a numeric value
func (m *MemoryStore) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	return m.Increment(ctx, key, -delta)
}

// Clear removes all entries
func (m *MemoryStore) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make(map[string]*item)
	return nil
}

// Close stops the cleanup goroutine
func (m *MemoryStore) Close() error {
	m.stop <- true
	return nil
}

// cleanupExpired removes expired entries periodically
func (m *MemoryStore) cleanupExpired() {
	ticker := time.NewTicker(m.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			for key, item := range m.items {
				if item.isExpired() {
					delete(m.items, key)
				}
			}
			m.mu.Unlock()

		case <-m.stop:
			return
		}
	}
}

// MarshalJSON for JSON encoding support
func (m *MemoryStore) marshalValue(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

// UnmarshalJSON for JSON decoding support
func (m *MemoryStore) unmarshalValue(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
