package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/OkanUysal/go-cache"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	ctx := context.Background()

	// Example 1: Memory cache (development)
	fmt.Println("=== Example 1: Memory Cache ===")
	memoryExample(ctx)

	// Example 2: Redis cache (production)
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		fmt.Println("\n=== Example 2: Redis Cache ===")
		redisExample(ctx, redisURL)
	}

	// Example 3: GetOrSet pattern
	fmt.Println("\n=== Example 3: GetOrSet Pattern ===")
	getOrSetExample(ctx)

	// Example 4: JSON caching
	fmt.Println("\n=== Example 4: JSON Caching ===")
	jsonExample(ctx)

	// Example 5: Counter operations
	fmt.Println("\n=== Example 5: Counter Operations ===")
	counterExample(ctx)
}

func memoryExample(ctx context.Context) {
	// Create in-memory cache
	c, err := cache.New(&cache.Config{
		Backend:    cache.BackendMemory,
		DefaultTTL: 5 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Set a value
	c.Set(ctx, "greeting", "Hello, World!")

	// Get the value
	value, err := c.Get(ctx, "greeting")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Cached value: %s\n", value)

	// Check if key exists
	if c.Has(ctx, "greeting") {
		fmt.Println("Key exists!")
	}

	// Delete the key
	c.Delete(ctx, "greeting")
	fmt.Println("Key deleted")
}

func redisExample(ctx context.Context, redisURL string) {
	// Create Redis cache
	c, err := cache.New(&cache.Config{
		Backend:    cache.BackendRedis,
		RedisURL:   redisURL,
		DefaultTTL: 1 * time.Hour,
	})
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
		return
	}
	defer c.Close()

	// Set a value with custom TTL
	c.SetWithTTL(ctx, "session:user123", "session_data", 30*time.Minute)

	// Get the value
	value, err := c.Get(ctx, "session:user123")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Session data: %s\n", value)
}

func getOrSetExample(ctx context.Context) {
	c, err := cache.New(cache.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Simulates expensive database query
	fetchUser := func() (interface{}, error) {
		fmt.Println("  → Fetching from database...")
		time.Sleep(100 * time.Millisecond) // Simulate slow query
		return &User{
			ID:    123,
			Name:  "John Doe",
			Email: "john@example.com",
		}, nil
	}

	// First call - will fetch from "database"
	fmt.Println("First call:")
	start := time.Now()
	user, _ := c.GetOrSet(ctx, "user:123", fetchUser, 1*time.Hour)
	fmt.Printf("  Got user in %v: %+v\n", time.Since(start), user)

	// Second call - will use cache (much faster!)
	fmt.Println("Second call:")
	start = time.Now()
	user, _ = c.GetOrSet(ctx, "user:123", fetchUser, 1*time.Hour)
	fmt.Printf("  Got user in %v: %+v\n", time.Since(start), user)
}

func jsonExample(ctx context.Context) {
	c, err := cache.New(cache.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	user := User{
		ID:    456,
		Name:  "Jane Smith",
		Email: "jane@example.com",
	}

	// Store JSON
	c.SetJSON(ctx, "user:456", user)
	fmt.Printf("Stored user: %+v\n", user)

	// Retrieve JSON
	var retrieved User
	c.GetJSON(ctx, "user:456", &retrieved)
	fmt.Printf("Retrieved user: %+v\n", retrieved)
}

func counterExample(ctx context.Context) {
	c, err := cache.New(cache.DefaultConfig())
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Increment counter
	count, _ := c.Increment(ctx, "page_views", 1)
	fmt.Printf("Page views: %d\n", count)

	count, _ = c.Increment(ctx, "page_views", 5)
	fmt.Printf("Page views: %d\n", count)

	// Decrement counter
	count, _ = c.Decrement(ctx, "page_views", 2)
	fmt.Printf("Page views: %d\n", count)
}
