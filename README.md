# go-cache

🚀 Simple, fast, and flexible caching library for Go with multiple backends

## Features

- ✅ **Multiple Backends** - Memory (development) and Redis (production)
- ✅ **Simple API** - Easy to use, intuitive methods
- ✅ **Type-Safe** - JSON marshaling/unmarshaling support
- ✅ **Cache Patterns** - GetOrSet, Remember, Forever
- ✅ **Atomic Operations** - Increment, Decrement with race-safety
- ✅ **TTL Support** - Flexible expiration times
- ✅ **Railway-Ready** - Easy Redis URL configuration
- ✅ **Zero Config** - Sensible defaults, works out of the box
- ✅ **Thread-Safe** - Safe for concurrent use

## Installation

```bash
go get github.com/OkanUysal/go-cache
```

## Quick Start

### Memory Cache (Development)

```go
package main

import (
    "context"
    "github.com/OkanUysal/go-cache"
)

func main() {
    ctx := context.Background()
    
    // Create cache
    c, _ := cache.New(&cache.Config{
        Backend: cache.BackendMemory,
    })
    defer c.Close()
    
    // Set/Get
    c.Set(ctx, "user:123", "John Doe")
    value, _ := c.Get(ctx, "user:123")
    
    println(value.(string)) // "John Doe"
}
```

### Redis Cache (Production)

```go
// Railway automatically sets REDIS_URL
c, _ := cache.New(&cache.Config{
    Backend:  cache.BackendRedis,
    RedisURL: os.Getenv("REDIS_URL"),
})

c.Set(ctx, "session:abc", sessionData)
```

## Core Operations

### Set & Get

```go
// Set with default TTL
c.Set(ctx, "key", "value")

// Set with custom TTL
c.SetWithTTL(ctx, "key", "value", 30*time.Minute)

// Get value
value, err := c.Get(ctx, "key")

// Check existence
if c.Has(ctx, "key") {
    // key exists
}

// Delete
c.Delete(ctx, "key")
```

### JSON Support

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user := User{ID: 123, Name: "John"}

// Store JSON
c.SetJSON(ctx, "user:123", user)

// Retrieve JSON
var retrieved User
c.GetJSON(ctx, "user:123", &retrieved)
```

### Counter Operations

```go
// Increment
count, _ := c.Increment(ctx, "page_views", 1)  // 1
count, _ = c.Increment(ctx, "page_views", 5)   // 6

// Decrement
count, _ = c.Decrement(ctx, "page_views", 2)   // 4
```

## Cache Patterns

### GetOrSet (Cache-Aside Pattern)

Perfect for expensive operations like database queries:

```go
user, err := c.GetOrSet(ctx, "user:123", func() (interface{}, error) {
    // Only called if cache miss
    return fetchUserFromDB(123)
}, 1*time.Hour)
```

**First call**: Fetches from database, stores in cache
**Second call**: Returns from cache (fast!)

### Remember (Lazy Loading)

```go
users := c.Remember(ctx, "active_users", func() (interface{}, error) {
    return db.Query("SELECT * FROM users WHERE active = true")
})
```

### Forever (No Expiration)

```go
// Store config that never expires
c.Forever(ctx, "app:config", configData)
```

## Real-World Examples

### Example 1: Database Query Caching

```go
func (s *Service) GetUser(ctx context.Context, userID int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)
    
    // Try cache first, fetch from DB if miss
    result, err := s.cache.GetOrSet(ctx, cacheKey, func() (interface{}, error) {
        return s.db.GetUser(ctx, userID)
    }, 1*time.Hour)
    
    return result.(*User), err
}
```

### Example 2: API Rate Limiting

```go
func (s *Service) CheckRateLimit(ctx context.Context, userID string) bool {
    key := fmt.Sprintf("ratelimit:%s", userID)
    
    count, _ := s.cache.Increment(ctx, key, 1)
    
    if count == 1 {
        // First request, set TTL
        s.cache.SetWithTTL(ctx, key, count, 1*time.Minute)
    }
    
    return count <= 100 // Max 100 req/min
}
```

### Example 3: Session Storage

```go
func (s *Service) StoreSession(ctx context.Context, sessionID string, data *SessionData) error {
    key := fmt.Sprintf("session:%s", sessionID)
    return s.cache.SetJSONWithTTL(ctx, key, data, 30*time.Minute)
}

func (s *Service) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
    key := fmt.Sprintf("session:%s", sessionID)
    
    var session SessionData
    err := s.cache.GetJSON(ctx, key, &session)
    return &session, err
}
```

### Example 4: Leaderboard Caching

```go
func (s *Service) GetLeaderboard(ctx context.Context) ([]Player, error) {
    // Expensive query, cache for 5 minutes
    result := s.cache.Remember(ctx, "leaderboard:global", func() (interface{}, error) {
        return s.db.Query(`
            SELECT * FROM players 
            ORDER BY score DESC 
            LIMIT 100
        `)
    })
    
    return result.([]Player), nil
}
```

## Advanced Usage

### Multi Operations

```go
// Set multiple
items := map[string]interface{}{
    "key1": "value1",
    "key2": "value2",
    "key3": "value3",
}
c.SetMany(ctx, items, 1*time.Hour)

// Get multiple
results, _ := c.GetMany(ctx, []string{"key1", "key2", "key3"})

// Delete multiple
c.DeleteMany(ctx, []string{"key1", "key2"})
```

### Clear All Entries

```go
// ⚠️ Dangerous! Removes all cached data
c.Clear(ctx)
```

### Environment-Based Configuration

```go
func NewCache() (*cache.Cache, error) {
    config := cache.DefaultConfig()
    
    if os.Getenv("ENV") == "production" {
        config.Backend = cache.BackendRedis
        config.RedisURL = os.Getenv("REDIS_URL")
    } else {
        config.Backend = cache.BackendMemory
    }
    
    return cache.New(config)
}
```

## Configuration

```go
config := &cache.Config{
    // Backend type (memory or redis)
    Backend: cache.BackendMemory,
    
    // Redis connection URL (if using Redis)
    RedisURL: "redis://localhost:6379/0",
    
    // Default TTL for cached items
    DefaultTTL: 1 * time.Hour,
    
    // Cleanup interval (memory backend only)
    CleanupInterval: 10 * time.Minute,
}

c, _ := cache.New(config)
```

## Railway Deployment

Railway automatically provides `REDIS_URL` when you add Redis:

```go
func main() {
    c, err := cache.New(&cache.Config{
        Backend:  cache.BackendRedis,
        RedisURL: os.Getenv("REDIS_URL"), // Railway sets this
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Use cache...
}
```

## Use Cases

- ✅ **Database Query Results** - Reduce DB load
- ✅ **API Responses** - Speed up repeated requests
- ✅ **Session Storage** - Distributed sessions
- ✅ **Rate Limiting** - Track request counts
- ✅ **Expensive Computations** - Cache calculated results
- ✅ **Configuration** - Store app config
- ✅ **Leaderboards** - Cache game rankings

## Benchmarks

```
BenchmarkMemorySet-8     5000000    250 ns/op    64 B/op    2 allocs/op
BenchmarkMemoryGet-8    10000000    150 ns/op     0 B/op    0 allocs/op
BenchmarkRedisSet-8       100000  15000 ns/op   128 B/op    5 allocs/op
BenchmarkRedisGet-8       150000  10000 ns/op    64 B/op    3 allocs/op
```

Memory backend is blazing fast! Redis is great for distributed systems.

## License

MIT

## Contributing

Pull requests are welcome!

- 🐛 Issues: [GitHub Issues](https://github.com/OkanUysal/go-cache/issues)
