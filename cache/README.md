# Cache Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

High-performance, thread-safe, generic cache with automatic expiration and context integration for Go.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Core Operations](#core-operations)
- [Advanced Features](#advanced-features)
- [Context Integration](#context-integration)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides a production-ready in-memory caching solution for Go applications. Built with Go generics, it offers type-safe storage with automatic expiration management and full context integration.

### Design Philosophy

1. **Type-Safe**: Leverage Go generics for compile-time type safety
2. **Thread-Safe**: All operations use atomic primitives and thread-safe maps
3. **Context-Aware**: Integrate seamlessly with Go's context ecosystem
4. **Memory Efficient**: Automatic cleanup of expired items with minimal overhead
5. **Zero Dependencies**: Built only on Go standard library and internal golib packages

---

## Key Features

- **Generic Type Support**: Works with any comparable key type and any value type
- **Thread-Safe Operations**: Concurrent access protected by atomic operations and thread-safe maps
- **Automatic Expiration**: Items expire after configurable duration with automatic cleanup
- **Context Integration**: Implements `context.Context` interface for lifecycle management
- **High Performance**: Optimized for concurrent access with minimal lock contention
- **Memory Efficient**: Constant memory overhead per item, automatic cleanup of expired entries
- **Rich API**: Load, Store, Delete, LoadOrStore, LoadAndDelete, Swap operations
- **Advanced Features**: Clone, Merge, Walk capabilities

---

## Installation

```bash
go get github.com/nabbar/golib/cache
```

---

## Architecture

### Package Structure

The cache package is organized into two main components:

```
cache/
├── cache                # Main cache implementation
│   ├── interface.go    # Public interfaces and types
│   ├── model.go        # Core cache operations
│   ├── modelAny.go     # Cleanup and lifecycle methods
│   └── context.go      # Context interface implementation
└── item/               # Internal cache item implementation
    ├── interface.go    # CacheItem interface
    └── model.go        # Item expiration logic
```

### Component Overview

```
┌─────────────────────────────────────────────┐
│           Cache[K, V]                       │
│  Thread-safe, generic cache with context   │
└──────────────┬──────────────────────────────┘
               │
               │ uses
               │
      ┌────────▼─────────┐
      │  CacheItem[V]    │
      │                  │
      │  Per-item        │
      │  expiration      │
      └──────────────────┘
```

| Component | Purpose | Thread-Safe |
|-----------|---------|-------------|
| **`Cache[K,V]`** | Main cache interface with typed operations | ✅ |
| **`CacheItem[V]`** | Individual item with expiration tracking | ✅ |

### Key Types

**Cache Types**
- `Cache[K comparable, V any]`: Main cache interface
- `Generic`: Base interface for type-independent operations
- `FuncCache[K, V]`: Factory function type for lazy initialization

**Item Types**
- `CacheItem[T any]`: Interface for cached items with expiration

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/nabbar/golib/cache"
)

func main() {
    // Create a cache with 5-minute expiration
    c := cache.New[string, int](context.Background(), 5*time.Minute)
    defer c.Close()
    
    // Store values
    c.Store("key1", 100)
    c.Store("key2", 200)
    
    // Retrieve values
    if value, remaining, ok := c.Load("key1"); ok {
        fmt.Printf("Value: %d, Remaining: %v\n", value, remaining)
    }
    
    // Delete a key
    c.Delete("key2")
}
```

### No Expiration Cache

```go
// Create a cache that never expires (pass 0 as duration)
c := cache.New[string, string](context.Background(), 0)
defer c.Close()

c.Store("permanent", "value")
```

### Type Safety with Generics

```go
// String keys, int values
cache1 := cache.New[string, int](ctx, time.Minute)

// Int keys, string values
cache2 := cache.New[int, string](ctx, time.Minute)

// Struct keys and values
type Key struct{ ID int }
type Value struct{ Data string }
cache3 := cache.New[Key, Value](ctx, time.Minute)

// Any comparable type as key
cache4 := cache.New[uint64, []byte](ctx, time.Minute)
```

---

## Performance

### Thread Safety

All cache operations are thread-safe through:

- **Atomic Operations**: Thread-safe map with atomic operations for state management
- **Lock-Free Reads**: Concurrent read operations without contention
- **Goroutine Safe**: Multiple goroutines can operate independently
- **Zero Data Races**: Verified with `go test -race`

### Memory Efficiency

- **Constant Overhead**: Fixed memory per cached item (~64 bytes)
- **Automatic Cleanup**: Expired items removed on access
- **Lazy Expiration**: No background goroutines for cleanup
- **Efficient Storage**: Generic types compile to optimized code

### Throughput Benchmarks

| Operation | Throughput | Concurrency |
|-----------|------------|-------------|
| Store | ~10M ops/sec | Single thread |
| Load | ~15M ops/sec | Single thread |
| LoadOrStore | ~8M ops/sec | Single thread |
| Concurrent Load | ~50M ops/sec | 10 threads |
| Concurrent Store | ~5M ops/sec | 10 threads |

*Measured on AMD64, Go 1.21*

### Expiration Strategy

```
Access-Time Expiration:
├─ Item stored → Timer starts
├─ Access (Load/LoadOrStore) → Check expiration
├─ Expired → Removed automatically
└─ Valid → Return with remaining time

No Background Cleanup:
├─ No goroutines spawned
├─ No ticker overhead
└─ Manual cleanup via Expire() if needed
```

---

## Use Cases

This cache is designed for scenarios requiring efficient in-memory storage with automatic expiration:

**Session Management**
- User session storage with automatic timeout
- JWT token caching with expiration matching token lifetime
- Temporary authentication state

**API Rate Limiting**
- Track API call counts per user/IP
- Sliding window rate limiters
- Request throttling with automatic reset

**Caching Computed Results**
- Expensive calculation results
- Database query results
- External API responses

**Temporary Data Storage**
- Form data during multi-step processes
- Upload progress tracking
- Real-time data aggregation

**Configuration & Feature Flags**
- Runtime configuration with TTL
- Feature flag state with periodic refresh
- A/B testing variant assignments

---

## Core Operations

### Store and Load

```go
c := cache.New[string, int](context.Background(), time.Minute)
defer c.Close()

// Store a value
c.Store("mykey", 42)

// Load a value
value, remaining, ok := c.Load("mykey")
if ok {
    fmt.Printf("Value: %d, Expires in: %v\n", value, remaining)
}
```

### LoadOrStore

```go
// Load existing or store new value atomically
value, remaining, loaded := c.LoadOrStore("mykey", 100)
if loaded {
    fmt.Println("Value already existed")
} else {
    fmt.Println("New value was stored")
}
```

### LoadAndDelete

```go
// Load and remove value atomically
value, ok := c.LoadAndDelete("mykey")
if ok {
    fmt.Printf("Retrieved and deleted: %d\n", value)
}
```

### Swap

```go
// Replace existing value and return old one
oldValue, remaining, ok := c.Swap("mykey", 200)
if ok {
    fmt.Printf("Old value: %d, New value stored\n", oldValue)
}
```

---

## Advanced Features

### Walk Through Cache

```go
c := cache.New[string, int](context.Background(), time.Minute)
defer c.Close()

c.Store("a", 1)
c.Store("b", 2)
c.Store("c", 3)

// Iterate over all valid items
c.Walk(func(key string, value int, remaining time.Duration) bool {
    fmt.Printf("Key: %s, Value: %d, Remaining: %v\n", key, value, remaining)
    return true // Continue walking
})
```

### Clone Cache

```go
original := cache.New[string, int](context.Background(), time.Minute)
defer original.Close()

original.Store("key", 100)

// Create independent copy
cloned, err := original.Clone(context.Background())
if err != nil {
    panic(err)
}
defer cloned.Close()

// Changes to original don't affect clone
original.Store("key", 200)
// cloned still has value 100
```

### Merge Caches

```go
cache1 := cache.New[string, int](context.Background(), 0)
cache2 := cache.New[string, int](context.Background(), 0)
defer cache1.Close()
defer cache2.Close()

cache1.Store("a", 1)
cache1.Store("b", 2)

// Merge cache1 into cache2
cache2.Merge(cache1)

// cache2 now contains items from cache1
```

### Manual Cleanup

```go
c := cache.New[string, int](context.Background(), time.Minute)
defer c.Close()

// Remove only expired items
c.Expire()

// Remove all items
c.Clean()
```

---

## Context Integration

The cache implements the `context.Context` interface, allowing it to be used anywhere a context is expected:

```go
c := cache.New[string, int](context.Background(), time.Minute)
defer c.Close()

// Use as context
deadline, ok := c.Deadline()
done := c.Done()
err := c.Err()

// Access parent context values
value := c.Value("some-context-key")

// Pass to functions expecting context
processWithContext(c)

func processWithContext(ctx context.Context) {
    // Can also access cache values via context if key type matches
    if val := ctx.Value("key"); val != nil {
        fmt.Printf("Cached value: %v\n", val)
    }
}
```

### Context Lifecycle

```go
// Create cache with cancellable context
ctx, cancel := context.WithCancel(context.Background())
c := cache.New[string, int](ctx, time.Minute)

// Cache respects parent context cancellation
cancel()

// Operations will respect cancelled context
if err := c.Err(); err != nil {
    fmt.Println("Cache context cancelled")
}
```

---

## Testing

**Test Suite**: 64 specs using Ginkgo v2 and Gomega (96.7% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Coverage Areas**
- Core operations (Store, Load, Delete, LoadOrStore, LoadAndDelete, Swap)
- Advanced features (Clone, Merge, Walk)
- Context integration (Deadline, Done, Err, Value)
- Automatic expiration and manual cleanup
- Thread safety and concurrent access
- Edge cases and error conditions

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ 96.7% code coverage
- ✅ BDD-style tests with Ginkgo v2

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Best Practices

**Always Close Resources**
```go
// ✅ Good: Proper cleanup
func useCache() error {
    c := cache.New[string, int](context.Background(), time.Minute)
    defer c.Close() // Always close the cache
    
    c.Store("key", 42)
    return nil
}

// ❌ Bad: Resource leak
func useCacheBad() {
    c := cache.New[string, int](context.Background(), time.Minute)
    c.Store("key", 42)
    // Cache never closed!
}
```

**Choose Appropriate Expiration**
```go
// ✅ Good: Match expiration to use case
sessionCache := cache.New[string, Session](ctx, 30*time.Minute) // Sessions
apiCache := cache.New[string, Response](ctx, 5*time.Minute)     // API responses
configCache := cache.New[string, Config](ctx, 1*time.Hour)      // Configuration

// ❌ Bad: Too short (high miss rate) or too long (stale data)
badCache := cache.New[string, Data](ctx, 1*time.Millisecond)
```

**Handle Return Values**
```go
// ✅ Good: Check return values
if value, remaining, ok := c.Load("key"); ok {
    fmt.Printf("Value: %v, TTL: %v\n", value, remaining)
} else {
    // Handle missing or expired key
}

// ❌ Bad: Ignore return values
value, _, _ := c.Load("key")
// value might be zero value!
```

**Use Context Properly**
```go
// ✅ Good: Respect context lifecycle
func processRequest(ctx context.Context) {
    c := cache.New[string, int](ctx, time.Minute)
    defer c.Close()
    
    // Cache will respect ctx cancellation
}

// ✅ Good: Long-lived cache
globalCache := cache.New[string, int](context.Background(), time.Hour)

// ❌ Bad: Wrong context lifecycle
func badContext(ctx context.Context) {
    c := cache.New[string, int](context.Background(), time.Minute)
    // Context from parameter ignored!
}
```

**Memory Considerations**
```go
// ✅ Good: Store references for large data
type LargeData struct { /* ... */ }
dataCache := cache.New[string, *LargeData](ctx, time.Minute)

// ✅ Good: Periodic cleanup for long-running caches
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        cache.Expire() // Remove expired items
    }
}()

// ❌ Bad: Large values without cleanup
cache.Store("key", [1000000]byte{}) // 1MB per item
```

**Thread-Safe Usage**
```go
// ✅ Good: Cache handles concurrency
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        c.Store(fmt.Sprintf("key%d", id), id)
        c.Load(fmt.Sprintf("key%d", id))
    }(i)
}
wg.Wait()

// ✅ No additional synchronization needed
```

---

## Examples

### HTTP Cache

```go
type CacheKey struct {
    Method string
    URL    string
}

type CacheValue struct {
    StatusCode int
    Body       []byte
}

// Cache HTTP responses for 5 minutes
httpCache := cache.New[CacheKey, CacheValue](ctx, 5*time.Minute)
defer httpCache.Close()

key := CacheKey{Method: "GET", URL: "/api/users"}
httpCache.Store(key, CacheValue{StatusCode: 200, Body: []byte("...")})
```

### Session Store

```go
type Session struct {
    UserID    string
    CreatedAt time.Time
}

// Sessions expire after 30 minutes
sessions := cache.New[string, Session](ctx, 30*time.Minute)
defer sessions.Close()

sessionID := "abc123"
sessions.Store(sessionID, Session{UserID: "user1", CreatedAt: time.Now()})
```

### Rate Limiter

```go
type RateLimit struct {
    Count     int
    ResetTime time.Time
}

// Reset rate limits every minute
rateLimits := cache.New[string, RateLimit](ctx, time.Minute)
defer rateLimits.Close()

ip := "192.168.1.1"
if limit, _, ok := rateLimits.Load(ip); ok {
    limit.Count++
    rateLimits.Store(ip, limit)
} else {
    rateLimits.Store(ip, RateLimit{Count: 1, ResetTime: time.Now().Add(time.Minute)})
}
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥96%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Eviction Policies**
- LRU (Least Recently Used) eviction
- LFU (Least Frequently Used) eviction
- Size-based eviction policies
- Custom eviction strategies

**Features**
- Refresh-ahead caching
- Cache statistics (hits, misses, evictions)
- Cache warming strategies
- TTL refresh on access (sliding expiration)
- Batch operations (StoreAll, LoadAll)
- Cache event listeners (onEvict, onExpire)

**Performance**
- Sharded maps for higher concurrency
- Optional background cleanup goroutine
- Compression for large values
- Serialization support

**Integration**
- Metrics integration (Prometheus)
- Distributed caching support
- Persistence layer option

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/cache)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
