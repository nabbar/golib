# Context Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/context)

**Thread-safe context management with generic key-value storage, extending Go's standard context.Context.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Operations](#operations)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Advanced Usage](#advanced-usage)
- [Gin Integration](#gin-integration)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **context** package provides a thread-safe, generic context management system that extends Go's standard `context.Context` with advanced key-value storage capabilities. It enables you to store, retrieve, and manage configuration data within contexts while maintaining full compatibility with the standard library.

### Design Philosophy

1. **Generic & Type-Safe**: Use any comparable type as keys (string, int, custom types)
2. **Thread-Safe**: All operations are concurrent-safe using sync.Map and sync.RWMutex
3. **Context Compatible**: Implements context.Context interface fully
4. **Zero Dependencies**: Core package has no external dependencies
5. **Extensible**: Framework integration packages (e.g., Gin) built on top

### Why Use This Package?

- ✅ **Type-safe key-value storage** within contexts
- ✅ **Thread-safe operations** for concurrent access
- ✅ **Context cloning** with independent storage
- ✅ **Configuration merging** from multiple sources
- ✅ **Atomic operations** (LoadOrStore, LoadAndDelete)
- ✅ **Flexible iteration** with Walk and WalkLimit
- ✅ **Full context.Context compatibility**
- ✅ **Framework integrations** (Gin, and more)

---

## Key Features

| Feature | Description | Benefit |
|---------|-------------|---------|
| **Generic Keys** | Use any comparable type as keys | Type safety and flexibility |
| **Thread-Safe** | Concurrent-safe storage and retrieval | Safe for goroutines |
| **Context Cloning** | Create independent copies | Isolated configuration spaces |
| **Configuration Merging** | Combine multiple configs | Flexible configuration composition |
| **Walk Operations** | Iterate over stored values | Inspection and transformation |
| **Atomic Operations** | LoadOrStore, LoadAndDelete | Race-free updates |
| **Cancellation** | Proper context cancellation handling | Resource cleanup |
| **Gin Integration** | Seamless Gin framework support | Web application context |

---

## Architecture

### Package Structure

```
context/
├── interface.go      # Core interfaces and factory functions
├── model.go          # Config implementation
├── map.go            # MapManage implementation helpers
├── helper.go         # Utility functions
├── context.go        # Context management
└── gin/              # Gin framework integration
    ├── interface.go  # Gin-specific context wrapper
    └── model.go      # Implementation
```

### Component Architecture

```
┌─────────────────────────────────────────────────────────┐
│              Config[T comparable]                        │
│                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   context.   │  │  MapManage   │  │   Context    │ │
│  │   Context    │  │  [T]         │  │              │ │
│  └──────────────┘  └──────────────┘  └──────────────┘ │
│         │                  │                  │          │
│         ▼                  ▼                  ▼          │
│  ┌──────────────────────────────────────────────────┐  │
│  │         Thread-Safe Storage                       │  │
│  │  (sync.Map + sync.RWMutex)                       │  │
│  └──────────────────────────────────────────────────┘  │
│                          │                               │
│                          ▼                               │
│  ┌──────────────────────────────────────────────────┐  │
│  │      Underlying context.Context                   │  │
│  │  (Cancellation, Deadlines, Values)               │  │
│  └──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Type System

```
Config[T comparable]
    │
    ├─ MapManage[T] interface
    │   ├─ Clean()
    │   ├─ Load(key T)
    │   ├─ Store(key T, val interface{})
    │   ├─ Delete(key T)
    │   ├─ LoadOrStore(key T, val interface{})
    │   └─ LoadAndDelete(key T)
    │
    ├─ Context interface
    │   └─ GetContext() context.Context
    │
    ├─ context.Context (embedded)
    │   ├─ Deadline()
    │   ├─ Done()
    │   ├─ Err()
    │   └─ Value(key interface{})
    │
    └─ Additional methods
        ├─ SetContext(ctx FuncContext)
        ├─ Clone(ctx context.Context)
        ├─ Merge(cfg Config[T])
        ├─ Walk(fct FuncWalk[T])
        └─ WalkLimit(fct FuncWalk[T], keys...)
```

---

## Installation

```bash
go get github.com/nabbar/golib/context
```

For Gin integration:

```bash
go get github.com/nabbar/golib/context/gin
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    libctx "github.com/nabbar/golib/context"
)

func main() {
    // Create a new config with string keys
    cfg := libctx.New[string](nil)
    
    // Store values
    cfg.Store("user", "john_doe")
    cfg.Store("role", "admin")
    cfg.Store("count", 42)
    
    // Load values
    user, ok := cfg.Load("user")
    if ok {
        fmt.Printf("User: %s\n", user.(string))
    }
    
    // Iterate over all values
    cfg.Walk(func(key string, val interface{}) bool {
        fmt.Printf("%s: %v\n", key, val)
        return true
    })
}
```

### With Custom Key Types

```go
type ConfigKey string

const (
    KeyDatabase ConfigKey = "database"
    KeyCache    ConfigKey = "cache"
    KeyAPI      ConfigKey = "api"
)

cfg := libctx.New[ConfigKey](nil)
cfg.Store(KeyDatabase, dbConnection)
cfg.Store(KeyCache, cacheClient)

// Type-safe access
if db, ok := cfg.Load(KeyDatabase); ok {
    // Use database connection
}
```

### With Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

cfg := libctx.New[string](func() context.Context {
    return ctx
})

cfg.Store("key", "value")

// When context is cancelled, cleanup occurs
select {
case <-cfg.Done():
    fmt.Println("Context cancelled")
}
```

---

## Core Concepts

### Config Interface

The `Config[T]` interface extends `context.Context` with key-value storage:

```go
type Config[T comparable] interface {
    context.Context
    MapManage[T]
    Context
    
    SetContext(ctx FuncContext)
    Clone(ctx context.Context) Config[T]
    Merge(cfg Config[T]) bool
    Walk(fct FuncWalk[T]) bool
    WalkLimit(fct FuncWalk[T], validKeys ...T) bool
}
```

### MapManage Interface

Thread-safe map operations:

```go
type MapManage[T comparable] interface {
    Clean()
    Load(key T) (val interface{}, ok bool)
    Store(key T, cfg interface{})
    Delete(key T)
    LoadOrStore(key T, cfg interface{}) (val interface{}, loaded bool)
    LoadAndDelete(key T) (val interface{}, loaded bool)
}
```

### Function Types

```go
// FuncContext provides the underlying context
type FuncContext func() context.Context

// FuncWalk is called for each key-value pair during iteration
type FuncWalk[T comparable] func(key T, val interface{}) bool
```

---

## Operations

### Store and Load

```go
cfg := libctx.New[string](nil)

// Store a value
cfg.Store("key", "value")

// Load a value
if val, ok := cfg.Load("key"); ok {
    fmt.Println(val.(string))
}

// Store nil removes the key
cfg.Store("key", nil)
```

### Atomic Operations

```go
// LoadOrStore: load existing or store new
val, loaded := cfg.LoadOrStore("counter", 0)
if loaded {
    fmt.Println("Found existing:", val)
} else {
    fmt.Println("Stored new:", val)
}

// LoadAndDelete: retrieve and remove in one operation
if val, ok := cfg.LoadAndDelete("temporary"); ok {
    fmt.Println("Removed:", val)
}
```

### Clone

```go
// Create independent copy
original := libctx.New[string](nil)
original.Store("key1", "value1")
original.Store("key2", "value2")

clone := original.Clone(nil)
clone.Store("key3", "value3")

// original doesn't have key3
// clone has key1, key2, and key3
```

### Merge

```go
cfg1 := libctx.New[string](nil)
cfg1.Store("a", 1)

cfg2 := libctx.New[string](nil)
cfg2.Store("b", 2)

// Merge cfg2 into cfg1
if cfg1.Merge(cfg2) {
    // cfg1 now has both "a" and "b"
}
```

### Walk

```go
// Iterate over all keys
cfg.Walk(func(key string, val interface{}) bool {
    fmt.Printf("%s: %v\n", key, val)
    return true // continue iteration
})

// Iterate over specific keys only
cfg.WalkLimit(func(key string, val interface{}) bool {
    fmt.Printf("%s: %v\n", key, val)
    return true
}, "key1", "key2", "key3")
```

---

## Performance

### Memory Characteristics

- **Config Instance**: ~128 bytes (base)
- **Per Entry**: ~48 bytes overhead (sync.Map entry)
- **RWMutex**: 24 bytes
- **Total per Config**: ~152 bytes + (entries × 48 bytes)

### Thread Safety

All operations are thread-safe through:

- **sync.Map**: Lock-free reads for most operations
- **sync.RWMutex**: Protects context function updates
- **Atomic Guarantees**: LoadOrStore, LoadAndDelete are atomic

### Performance Benchmarks

| Operation | Time | Allocations | Notes |
|-----------|------|-------------|-------|
| New() | ~200ns | 1 | One-time initialization |
| Store() | ~100ns | 0-1 | Amortized, may allocate |
| Load() | ~50ns | 0 | Lock-free read |
| LoadOrStore() | ~150ns | 0-1 | Atomic operation |
| Walk() | ~5µs | 0 | For 100 entries |
| Clone() | ~10µs | 100+ | Depends on entry count |
| Merge() | ~8µs | 0-50 | Depends on source size |

*Benchmarks on AMD64, Go 1.21*

### Concurrency

- **Read Performance**: Excellent (lock-free with sync.Map)
- **Write Performance**: Good (optimized for concurrent writes)
- **Mixed Workload**: Very good (sync.Map optimized for this)
- **Contention**: Low (minimal lock contention)

---

## Use Cases

### Application Configuration

```go
type AppConfig struct {
    Database string
    CacheURL string
    APIKey   string
}

cfg := libctx.New[string](nil)
cfg.Store("app", AppConfig{
    Database: "postgres://...",
    CacheURL: "redis://...",
    APIKey:   "secret",
})
```

### Request Scoped Data (HTTP)

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg := libctx.New[string](r.Context)
        cfg.Store("user_id", getUserID(r))
        cfg.Store("trace_id", generateTraceID())
        
        // Pass to next handler
        next.ServeHTTP(w, r.WithContext(cfg))
    })
}
```

### Service Dependencies

```go
type Services struct {
    DB    *sql.DB
    Cache *redis.Client
    Queue *amqp.Channel
}

cfg := libctx.New[string](nil)
cfg.Store("services", &Services{
    DB:    db,
    Cache: cache,
    Queue: queue,
})
```

### Multi-Tenant Applications

```go
cfg := libctx.New[string](nil)
cfg.Store("tenant_id", "tenant-123")
cfg.Store("tenant_config", tenantConfig)
cfg.Store("tenant_db", tenantDB)
```

### Testing and Mocking

```go
func TestHandler(t *testing.T) {
    cfg := libctx.New[string](nil)
    cfg.Store("db", mockDB)
    cfg.Store("cache", mockCache)
    
    // Test with mocked dependencies
    handler(cfg)
}
```

---

## Advanced Usage

### Context Propagation

```go
func parent() {
    cfg := libctx.New[string](nil)
    cfg.Store("parent_data", "value")
    
    child(cfg)
}

func child(cfg libctx.Config[string]) {
    // Access parent data
    if val, ok := cfg.Load("parent_data"); ok {
        fmt.Println(val)
    }
    
    // Add child-specific data
    cfg.Store("child_data", "child_value")
}
```

### Configuration Layers

```go
// Base configuration
base := libctx.New[string](nil)
base.Store("env", "production")
base.Store("debug", false)

// Environment-specific overrides
envCfg := libctx.New[string](nil)
envCfg.Store("db_host", "prod-db.example.com")

// Merge layers
base.Merge(envCfg)
```

### Cleanup Pattern

```go
cfg := libctx.New[string](nil)

// Store resources
cfg.Store("db", dbConn)
cfg.Store("file", fileHandle)

// Cleanup
defer func() {
    cfg.Walk(func(key string, val interface{}) bool {
        switch v := val.(type) {
        case io.Closer:
            v.Close()
        }
        return true
    })
    cfg.Clean()
}()
```

---

## Gin Integration

The `context/gin` sub-package provides seamless integration with the Gin web framework, offering a thin wrapper around Gin's context with additional features.

### Installation

```bash
go get github.com/nabbar/golib/context/gin
```

### Features

- ✅ **Full context.Context compatibility** - Implements standard context.Context
- ✅ **Gin context access** - Direct access to underlying gin.Context
- ✅ **Type-safe getters** - Convenient methods for common types
- ✅ **Signal handling** - Cancel context on OS signals (SIGTERM, SIGINT, etc.)
- ✅ **Logger integration** - Built-in logging support
- ✅ **Request-scoped storage** - Store and retrieve request-specific data
- ✅ **Zero additional overhead** - Thin wrapper around Gin context

### GinTonic Interface

```go
type GinTonic interface {
    context.Context
    
    // Gin integration
    GinContext() *gin.Context
    CancelOnSignal(s ...os.Signal)
    
    // Value storage
    Set(key any, value any)
    Get(key any) (value any, exists bool)
    MustGet(key any) any
    
    // Type-safe getters
    GetString(key any) string
    GetBool(key any) bool
    GetInt(key any) int
    GetInt64(key any) int64
    GetFloat64(key any) float64
    GetTime(key any) time.Time
    GetDuration(key any) time.Duration
    GetStringSlice(key any) []string
    GetStringMap(key any) map[string]any
    GetStringMapString(key any) map[string]string
    GetStringMapStringSlice(key any) map[string][]string
    
    // Logger
    SetLogger(log liblog.FuncLog)
}
```

### Basic Usage

```go
import (
    "github.com/gin-gonic/gin"
    ginlib "github.com/nabbar/golib/context/gin"
)

func main() {
    r := gin.Default()
    
    r.GET("/user/:id", func(c *gin.Context) {
        // Create GinTonic context
        gtx := ginlib.New(c, nil)
        
        // Store request-specific data
        gtx.Set("user_id", c.Param("id"))
        gtx.Set("request_time", time.Now())
        
        // Type-safe retrieval
        userID := gtx.GetString("user_id")
        reqTime := gtx.GetTime("request_time")
        
        c.JSON(200, gin.H{
            "user_id": userID,
            "timestamp": reqTime,
        })
    })
    
    r.Run(":8080")
}
```

### Middleware Pattern

```go
func ContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        gtx := ginlib.New(c, nil)
        
        // Store common request data
        gtx.Set("request_id", generateID())
        gtx.Set("user", getCurrentUser(c))
        gtx.Set("start_time", time.Now())
        
        // Continue to next handler
        c.Next()
        
        // Log request duration
        duration := time.Since(gtx.GetTime("start_time"))
        log.Printf("Request took %v", duration)
    }
}

r.Use(ContextMiddleware())
```

### Signal Handling

```go
func gracefulShutdown(c *gin.Context) {
    gtx := ginlib.New(c, nil)
    
    // Cancel context on SIGTERM or SIGINT
    gtx.CancelOnSignal(syscall.SIGTERM, syscall.SIGINT)
    
    select {
    case <-gtx.Done():
        log.Println("Context cancelled by signal")
        return
    case <-processRequest(gtx):
        log.Println("Request completed")
    }
}
```

### Type-Safe Getters

```go
func handler(c *gin.Context) {
    gtx := ginlib.New(c, nil)
    
    // Store various types
    gtx.Set("count", 42)
    gtx.Set("enabled", true)
    gtx.Set("ratio", 3.14)
    gtx.Set("tags", []string{"go", "gin", "web"})
    
    // Type-safe retrieval (returns zero value if not found)
    count := gtx.GetInt("count")           // 42
    enabled := gtx.GetBool("enabled")       // true
    ratio := gtx.GetFloat64("ratio")        // 3.14
    tags := gtx.GetStringSlice("tags")      // ["go", "gin", "web"]
    
    // Use MustGet if you want to panic on missing keys
    value := gtx.MustGet("required_key") // panics if not found
}
```

### Logger Integration

```go
import (
    ginlib "github.com/nabbar/golib/context/gin"
    liblog "github.com/nabbar/golib/logger"
)

func handler(c *gin.Context) {
    // Create logger function
    logger := func() liblog.Logger {
        return liblog.New(nil)
    }
    
    // Create GinTonic with logger
    gtx := ginlib.New(c, logger)
    
    // Logger is now available to components
    gtx.Set("user", "alice")
}
```

### Testing

```go
import (
    "testing"
    "net/http/httptest"
    
    "github.com/gin-gonic/gin"
    ginlib "github.com/nabbar/golib/context/gin"
)

func TestHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    gtx := ginlib.New(c, nil)
    gtx.Set("test_key", "test_value")
    
    if val := gtx.GetString("test_key"); val != "test_value" {
        t.Errorf("Expected 'test_value', got '%s'", val)
    }
}
```

---

## Best Practices

### 1. Use Typed Keys

```go
type ContextKey string

const (
    KeyUser    ContextKey = "user"
    KeySession ContextKey = "session"
)

cfg := libctx.New[ContextKey](nil)
cfg.Store(KeyUser, user)
```

### 2. Initialize with Parent Context

```go
cfg := libctx.New[string](func() context.Context {
    return parentCtx
})
```

### 3. Check Load Results

```go
if val, ok := cfg.Load("key"); ok {
    // Use val
} else {
    // Handle missing key
}
```

### 4. Use Type Assertions Safely

```go
if val, ok := cfg.Load("count"); ok {
    if count, ok := val.(int); ok {
        fmt.Println(count)
    }
}
```

### 5. Clean Up Resources

```go
defer cfg.Clean() // Remove all entries
```

### 6. Clone for Independence

```go
// When you need isolated storage
clone := cfg.Clone(nil)
clone.Store("isolated_key", "value")
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd context
go test -v -cover
```

**With Race Detection:**
```bash
CGO_ENABLED=1 go test -race -v
```

**Test Metrics:**
- 60+ test specifications
- >90% code coverage
- Ginkgo v2 + Gomega framework
- Full concurrency testing

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain or improve test coverage
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document all public APIs with GoDoc comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety when applicable
- Add comments explaining complex test scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Performance Optimizations**
- Custom memory pool for frequent allocations
- Batch operations for multiple keys
- Streaming walk operations for large datasets
- Copy-on-write optimization for cloning

**Enhanced Features**
- Typed value getters with automatic type conversion
- JSON/YAML serialization of stored values
- Expiration/TTL support for stored values
- Event hooks for Store/Delete operations
- Structured logging integration

**Additional Integrations**
- Echo framework support
- Chi router integration
- gRPC metadata integration
- Fiber framework support

**Developer Experience**
- Code generation for type-safe wrappers
- Visual debugging tools
- Performance profiling helpers
- Better error messages with context

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Go Standard Library
- **[context](https://pkg.go.dev/context)** - Standard context package
- **[sync.Map](https://pkg.go.dev/sync#Map)** - Concurrent-safe map
- **[sync.RWMutex](https://pkg.go.dev/sync#RWMutex)** - Read-write mutex

### Related Golib Packages
- **[config](../config/README.md)** - Application configuration (uses this package)
- **[atomic](../atomic/README.md)** - Atomic operations

### External Libraries
- **[Gin](https://github.com/gin-gonic/gin)** - Web framework (gin sub-package)
- **[Ginkgo Testing](https://github.com/onsi/ginkgo)** - BDD testing framework
- **[Gomega Matchers](https://github.com/onsi/gomega)** - Matcher library for tests

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2019 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/context)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Gin Integration**: [gin/README.md](gin/README.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Source Code**: [GitHub Repository](https://github.com/nabbar/golib)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
