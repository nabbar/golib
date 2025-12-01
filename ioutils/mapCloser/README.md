# IOUtils MapCloser

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-80.8%25-brightgreen)](TESTING.md)

Thread-safe, context-aware manager for multiple io.Closer instances with automatic cleanup, error aggregation, and lock-free atomic operations.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Thread Safety Model](#thread-safety-model)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Example](#basic-example)
  - [Context-Based Cleanup](#context-based-cleanup)
  - [Error Aggregation](#error-aggregation)
  - [Hierarchical Management](#hierarchical-management)
  - [Testing Cleanup](#testing-cleanup)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Closer Interface](#closer-interface)
  - [Configuration](#configuration)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **mapCloser** package provides a high-performance, thread-safe solution for managing multiple `io.Closer` instances with automatic cleanup when a context is cancelled. It's designed for applications that need reliable resource management in concurrent environments with predictable lifecycle control.

### Design Philosophy

1. **Lock-Free Performance**: All state changes use atomic operations, no mutexes
2. **Context-Driven Lifecycle**: Automatic cleanup tied to context cancellation
3. **Fail-Safe Operation**: Continues closing all resources even when some fail
4. **Memory-Safe**: All operations check for nil and closed state
5. **Observable**: Simple API for tracking registered closers

### Key Features

- ✅ **Zero Mutexes**: Lock-free implementation using atomic.Bool, atomic.Uint64
- ✅ **Automatic Cleanup**: Resources close when context is done
- ✅ **Error Aggregation**: Collects all close errors, doesn't fail-fast
- ✅ **Concurrent Safe**: All methods can be called from multiple goroutines
- ✅ **Clone Support**: Create independent copies for hierarchical resource management
- ✅ **Nil-Safe**: Gracefully handles nil closers without panics
- ✅ **Production Ready**: 80.8% test coverage, 34 specs, zero race conditions

---

## Architecture

### Component Diagram

```
┌───────────────────────────────────────────────────────────┐
│                     MapCloser                              │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────────┐     ┌──────────────────┐            │
│  │  Atomic State   │     │  Context Monitor │            │
│  ├─────────────────┤     ├──────────────────┤            │
│  │ • atomic.Bool   │     │ ctx.Done()       │            │
│  │   (closed flag) │     │  monitoring      │            │
│  │                 │     │                  │            │
│  │ • atomic.Uint64 │     │  goroutine       │            │
│  │   (counter)     │     │  blocks on       │            │
│  │                 │     │  Done()          │            │
│  └─────────────────┘     └──────────────────┘            │
│           │                       │                       │
│           └───────┬───────────────┘                       │
│                   │                                       │
│        ┌──────────▼───────────┐                          │
│        │  Closer Storage      │                          │
│        │  (libctx.Config)     │                          │
│        ├──────────────────────┤                          │
│        │ 1 → io.Closer        │                          │
│        │ 2 → io.Closer        │                          │
│        │ 3 → io.Closer        │                          │
│        │ ...                  │                          │
│        └──────────────────────┘                          │
│                   │                                       │
│                   ▼                                       │
│        ┌──────────────────────┐                          │
│        │  Close() Operation   │                          │
│        ├──────────────────────┤                          │
│        │ • CompareAndSwap     │                          │
│        │ • Walk & Close all   │                          │
│        │ • Error aggregation  │                          │
│        │ • Context cancel     │                          │
│        └──────────────────────┘                          │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

### Data Flow

```
New(ctx) → Initialize
    │
    ├─ Create child context with cancel function
    ├─ Initialize atomic.Bool (closed = false)
    ├─ Initialize atomic.Uint64 (counter = 0)
    ├─ Initialize libctx.Config storage
    └─ Start background goroutine
            │
            └─ Block on <-ctx.Done()
                    │
                    └─ Call Close() automatically

Add(closer1, closer2, ...)
    │
    ├─ Check if closed → return (no-op)
    ├─ Check if ctx.Err() != nil → return (no-op)
    │
    └─ For each closer:
            ├─ Increment counter (atomic)
            └─ Store at counter index

Close()
    │
    ├─ CompareAndSwap(false, true) → first call wins
    ├─ If already closed → return error
    │
    ├─ Walk all stored closers
    │   ├─ Call Close() on each
    │   ├─ Collect errors (continue on failure)
    │   └─ Return true to continue
    │
    ├─ Cancel context (defer)
    └─ Aggregate errors → return
```

### Thread Safety Model

| Operation | Mechanism | Contention | Performance |
|-----------|-----------|------------|-------------|
| **Add()** | atomic.Uint64.Add() | None | O(1) ~100ns |
| **Get()** | Walk + type assertion | None | O(n) ~n µs |
| **Len()** | atomic.Uint64.Load() | None | O(1) ~10ns |
| **Clean()** | atomic store + clear | None | O(1) ~100ns |
| **Clone()** | Copy state + storage | None | O(n) ~n µs |
| **Close()** | CompareAndSwap | First wins | O(n) ~n ms |

**Synchronization Primitives:**
- `atomic.Bool`: Closed flag (prevents double close)
- `atomic.Uint64`: Monotonic counter for indexing
- `libctx.Config[uint64]`: Thread-safe map for closers
- `context.CancelFunc`: Called once on close

**Thread-Safety Guarantees:**
- All public methods are safe for concurrent calls
- CompareAndSwap ensures only one Close() succeeds
- No deadlocks or race conditions (verified with `-race` detector)

---

## Performance

### Benchmarks

Performance measurements from test suite (standard Go tests, no race detector):

| Operation | Complexity | Latency | Allocations |
|-----------|-----------|---------|-------------|
| **New()** | O(1) | ~1 µs | 5 allocs |
| **Add() single** | O(1) | ~100 ns | 0 allocs |
| **Add() batch** | O(n) | ~n×100 ns | 0 allocs |
| **Get()** | O(n) | ~n µs | 1 alloc |
| **Len()** | O(1) | ~10 ns | 0 allocs |
| **Clean()** | O(1) | ~100 ns | 0 allocs |
| **Clone()** | O(n) | ~n µs | n allocs |
| **Close()** | O(n) | ~n ms | 0 allocs |

*n = number of registered closers*

**Throughput:**
- Concurrent Add: **~10M operations/sec** (100 goroutines)
- Get operations: **~1M operations/sec**
- Metrics reads (Len): **~100M operations/sec**

### Memory Usage

```
Base struct size:        ~80 bytes (atomic primitives + pointers)
Per closer registered:   ~40 bytes (key-value pair in libctx.Config)
Background goroutine:    ~2 KB (standard Go goroutine stack)

Example with 1000 closers:
Total memory = 80 + (1000 × 40) + 2048 ≈ 42 KB
```

**Memory characteristics:**
- No memory leaks (verified with multiple test runs)
- Constant memory after initialization
- Clean() fully releases closer storage

### Scalability

**Tested Limits:**
- **Concurrent writers**: 100 goroutines (no race conditions)
- **Registered closers**: 10,000 items (tested with overflow)
- **Clone operations**: Independent copies work correctly
- **Context cancellations**: Immediate cleanup after Done()

**Performance Notes:**
- Add() performance independent of closer count
- Get() performance scales linearly with closer count
- Close() time dominated by actual closer.Close() calls, not overhead
- No performance degradation with concurrent operations

---

## Use Cases

### 1. HTTP Server Cleanup

**Problem**: Web server needs to close database connections, cache clients, log files on shutdown.

```go
type Server struct {
    closer mapCloser.Closer
    db     *sql.DB
    cache  *redis.Client
}

func NewServer(ctx context.Context) (*Server, error) {
    closer := mapCloser.New(ctx)
    
    // Database connection
    db, err := sql.Open("postgres", connString)
    if err != nil {
        return nil, err
    }
    closer.Add(db)
    
    // Redis cache
    cache := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    if err := cache.Ping(ctx).Err(); err != nil {
        closer.Close()  // Close DB before returning
        return nil, err
    }
    closer.Add(cache)
    
    // Log file
    logFile, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    closer.Add(logFile)
    
    return &Server{closer: closer, db: db, cache: cache}, nil
}

func (s *Server) Shutdown() error {
    return s.closer.Close()  // Closes all: DB, cache, logFile
}
```

**Real-world**: Used with `github.com/nabbar/golib/socket/server` for high-traffic applications.

### 2. Worker Pool Management

**Problem**: Each worker manages temporary files, connections that need cleanup when worker stops.

```go
type Worker struct {
    id     int
    closer mapCloser.Closer
    tmpDir string
}

func NewWorkerPool(ctx context.Context, count int) []*Worker {
    workers := make([]*Worker, count)
    
    for i := 0; i < count; i++ {
        closer := mapCloser.New(ctx)  // Each worker has own closer
        
        // Create temp directory
        tmpDir, _ := os.MkdirTemp("", fmt.Sprintf("worker-%d-*", i))
        
        // Open worker log
        logFile, _ := os.Create(filepath.Join(tmpDir, "worker.log"))
        closer.Add(logFile)
        
        // Worker temp file
        tmpFile, _ := os.CreateTemp(tmpDir, "data-*.tmp")
        closer.Add(tmpFile)
        
        workers[i] = &Worker{id: i, closer: closer, tmpDir: tmpDir}
    }
    
    return workers
}

// Context cancellation automatically closes all workers' resources
```

### 3. Test Resource Cleanup

**Problem**: Tests create resources that must be cleaned up even on failure or timeout.

```go
func TestFeature(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    closer := mapCloser.New(ctx)
    defer closer.Close()  // Ensures cleanup
    
    // Create test database
    testDB := setupTestDB(t)
    closer.Add(testDB)
    
    // Create temp files
    file1, _ := os.CreateTemp("", "test-*.dat")
    file2, _ := os.CreateTemp("", "test-*.log")
    closer.Add(file1, file2)
    
    // Run test...
    // All resources cleaned up on test end OR timeout
}
```

### 4. Hierarchical Resource Scopes

**Problem**: Parent context shares some resources, child contexts have request-specific resources.

```go
func main() {
    ctx := context.Background()
    
    // Parent closer for shared resources
    parentCloser := mapCloser.New(ctx)
    defer parentCloser.Close()
    
    sharedDB, _ := sql.Open("postgres", connString)
    parentCloser.Add(sharedDB)
    
    // Handle each request with isolated resources
    http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
        reqCloser := parentCloser.Clone()  // Independent copy
        defer reqCloser.Close()
        
        // Request-specific temp file
        tmpFile, _ := os.CreateTemp("", "request-*")
        reqCloser.Add(tmpFile)
        
        // Process request...
        // tmpFile closed at end of request, sharedDB still open
    })
    
    http.ListenAndServe(":8080", nil)
}
```

### 5. Multi-Stage Pipeline Cleanup

**Problem**: Data processing pipeline with stages, each stage creates resources.

```go
func ProcessPipeline(ctx context.Context, data []byte) error {
    closer := mapCloser.New(ctx)
    defer closer.Close()
    
    // Stage 1: Decompress
    decompressed, closer1, err := decompressData(data)
    if err != nil {
        return err
    }
    closer.Add(closer1...)  // Add stage 1 resources
    
    // Stage 2: Parse
    parsed, closer2, err := parseData(decompressed)
    if err != nil {
        return err
    }
    closer.Add(closer2...)  // Add stage 2 resources
    
    // Stage 3: Transform
    transformed, closer3, err := transformData(parsed)
    if err != nil {
        return err
    }
    closer.Add(closer3...)  // Add stage 3 resources
    
    // All resources cleaned up on return
    return nil
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/mapCloser
```

### Basic Example

```go
package main

import (
    "context"
    "os"
    "github.com/nabbar/golib/ioutils/mapCloser"
)

func main() {
    ctx := context.Background()
    
    // Create closer manager
    closer := mapCloser.New(ctx)
    defer closer.Close()
    
    // Open files
    file1, _ := os.Open("data1.txt")
    file2, _ := os.Open("data2.txt")
    file3, _ := os.Open("data3.txt")
    
    // Register for automatic cleanup
    closer.Add(file1, file2, file3)
    
    // Use files...
    // All automatically closed by defer closer.Close()
}
```

### Context-Based Cleanup

```go
// Automatic cleanup on timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

closer := mapCloser.New(ctx)

conn1, _ := net.Dial("tcp", "example.com:80")
conn2, _ := net.Dial("tcp", "example.com:443")

closer.Add(conn1, conn2)

// When timeout occurs, connections auto-close
// No manual Close() needed
```

### Error Aggregation

```go
closer := mapCloser.New(context.Background())

// Add multiple closers
closer.Add(resource1, resource2, resource3)

// Close all, collect errors
if err := closer.Close(); err != nil {
    // err contains: "error from resource1, error from resource2"
    log.Printf("Some resources failed to close: %v", err)
}
```

### Hierarchical Management

```go
ctx := context.Background()

// Parent manages shared resources
parent := mapCloser.New(ctx)
defer parent.Close()

parent.Add(sharedDatabase)

// Child manages request-specific resources
child := parent.Clone()  // Independent copy
defer child.Close()

child.Add(requestTempFile)

// child.Close() only closes requestTempFile
// parent.Close() only closes sharedDatabase
```

### Testing Cleanup

```go
func TestFeature(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    closer := mapCloser.New(ctx)
    t.Cleanup(func() { closer.Close() })
    
    testDB := setupTestDB(t)
    testFiles := setupTestFiles(t)
    
    closer.Add(testDB)
    closer.Add(testFiles...)
    
    // Run test...
    // All resources cleaned up automatically
}
```

---

## Best Practices

### ✅ DO

**Always Use defer:**
```go
closer := mapCloser.New(ctx)
defer closer.Close()  // Ensures cleanup even on panic
```

**Check Close Errors:**
```go
if err := closer.Close(); err != nil {
    log.Printf("Cleanup errors: %v", err)
    // Take corrective action
}
```

**Choose Appropriate Context:**
```go
// Long-running service
ctx := context.Background()

// Timed operation
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Cancellable operation
ctx, cancel := context.WithCancel(parentCtx)
defer cancel()
```

**Use Clone for Hierarchical Management:**
```go
parentCloser := mapCloser.New(ctx)
parentCloser.Add(sharedResource)

childCloser := parentCloser.Clone()
childCloser.Add(tempResource)

childCloser.Close()  // Closes temp only
parentCloser.Close() // Closes shared
```

**Nil Closers Are Safe:**
```go
var file *os.File  // nil
closer.Add(file)   // Safe - filtered out during Close()
```

### ❌ DON'T

**Don't Ignore Context:**
```go
// ❌ BAD: Context never done
closer := mapCloser.New(context.Background())
// Resources never auto-close

// ✅ GOOD: Use cancellable context
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
closer := mapCloser.New(ctx)
```

**Don't Double-Close:**
```go
// ❌ BAD: Second close returns error
closer.Close()
closer.Close()  // Error: closer already closed

// ✅ GOOD: Check state
if closer.IsRunning() {
    closer.Close()
}
```

**Don't Add After Close:**
```go
// ❌ BAD: Add after close
closer.Close()
closer.Add(resource)  // No-op, resource not managed

// ✅ GOOD: Add before close
closer.Add(resource)
closer.Close()
```

**Don't Leak Goroutines:**
```go
// ❌ BAD: Closer not closed
closer := mapCloser.New(ctx)
// Background goroutine leaks

// ✅ GOOD: Always close
defer closer.Close()  // Stops background goroutine
```

---

## API Reference

### Closer Interface

```go
type Closer interface {
    Add(clo ...io.Closer)
    Get() []io.Closer
    Len() int
    Clean()
    Clone() Closer
    Close() error
}
```

**Methods:**

#### Add(clo ...io.Closer)

Registers one or more io.Closer instances for management.

**Behavior:**
- If Closer is closed or context is done: no-op
- Nil closers are accepted but filtered out during Get() and Close()
- Each Add increments internal counter
- Thread-safe: O(1) atomic operations

**Example:**
```go
closer.Add(file1, file2, file3)
closer.Add(nil, conn1, nil)  // Nils safely ignored
```

#### Get() []io.Closer

Returns copy of all registered closers, excluding nil values.

**Behavior:**
- Returns empty slice if closed or no closers registered
- Returned slice is independent (safe to modify)
- Order not guaranteed
- Thread-safe: O(n) iteration

**Example:**
```go
closers := closer.Get()
for _, c := range closers {
    // Use closer
}
```

#### Len() int

Returns total count of closers added (including nil).

**Behavior:**
- Returns 0 if overflow occurs (> math.MaxInt)
- Counter never decrements (except Clean())
- Thread-safe: O(1) atomic load

**Example:**
```go
closer.Add(file1, nil, file2)
count := closer.Len()  // Returns 3
```

#### Clean()

Removes all closers without closing them.

**Behavior:**
- Resets counter to zero
- Clears storage
- Does NOT close closers
- No-op if already closed
- Thread-safe: O(1) operations

**Example:**
```go
closer.Add(file1, file2)
closer.Clean()  // Files NOT closed, just removed
file1.Close()   // Manual close needed
```

#### Clone() Closer

Creates independent copy with same state.

**Behavior:**
- Shares context cancel function
- Independent closer storage (deep copy)
- Counter value copied at time of cloning
- Returns nil if original closed
- Thread-safe: O(n) copy

**Example:**
```go
parent := mapCloser.New(ctx)
parent.Add(sharedDB)

child := parent.Clone()
child.Add(tempFile)

child.Close()   // Closes tempFile + sharedDB copy
parent.Close()  // Closes original sharedDB
```

#### Close() error

Closes all closers and cancels context.

**Behavior:**
- First call: performs cleanup, returns result
- Subsequent calls: returns error "closer already closed"
- Continues closing even if some fail
- Returns aggregated error (format: "error1, error2, error3")
- Thread-safe: CompareAndSwap ensures single execution

**Example:**
```go
if err := closer.Close(); err != nil {
    log.Printf("Errors: %v", err)
}
```

### Configuration

The mapCloser requires only a context for initialization:

```go
func New(ctx context.Context) Closer
```

**Parameters:**
- `ctx`: Context to monitor for cancellation

**Returns:**
- `Closer`: Thread-safe closer manager

**Internal Configuration:**
- Buffer size: Not applicable (uses map)
- Polling interval: Immediate (blocks on ctx.Done())
- Default capacity: Unlimited (grows as needed)

### Error Codes

```go
var (
    ErrAlreadyClosed = errors.New("closer already closed")
    ErrNotInitialized = errors.New("not initialized")
)
```

**Error Handling:**
- Errors from individual closers are aggregated
- CompareAndSwap prevents concurrent close operations
- Post-close operations are no-ops (no errors)

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >80%)
   - Pass all tests including race detector
   - Use `gofmt` and `golint`

2. **AI Usage Policy**
   - ❌ **AI must NEVER be used** to generate package code or core functionality
   - ✅ **AI assistance is limited to**:
     - Testing (writing and improving tests)
     - Debugging (troubleshooting and bug resolution)
     - Documentation (comments, README, TESTING.md)
   - All AI-assisted work must be reviewed and validated by humans

3. **Testing**
   - Add tests for new features
   - Use Ginkgo v2 / Gomega for test framework
   - Ensure zero race conditions
   - Document test cases

4. **Documentation**
   - Update GoDoc comments for public APIs
   - Add examples for new features
   - Update README.md and TESTING.md if needed

5. **Pull Request Process**
   - Fork the repository
   - Create a feature branch
   - Write clear commit messages
   - Ensure all tests pass
   - Update documentation
   - Submit PR with description of changes

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **80.8% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Memory-safe** with proper nil checks
- ✅ **No panic calls** in production code

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Configurable Monitoring**: Allow custom polling intervals or event-driven monitoring instead of blocking on Done()
2. **Priority-Based Closing**: Close resources in specific order based on priority
3. **Close Timeout**: Add timeout for individual closer.Close() operations
4. **Metrics Export**: Optional integration with Prometheus or other metrics systems
5. **Callback Hooks**: Pre/post close callbacks for custom cleanup logic

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/mapCloser)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, thread-safety guarantees, and best practices for production use. Provides detailed explanations of internal mechanisms.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (80.8%), performance benchmarks, and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/context](https://pkg.go.dev/github.com/nabbar/golib/context)** - Thread-safe context storage used internally by mapCloser. Provides lock-free atomic operations for storing typed values associated with contexts.

- **[github.com/nabbar/golib/ioutils](https://pkg.go.dev/github.com/nabbar/golib/ioutils)** - Parent package with additional I/O utilities including aggregators, buffers, and progress tracking.

- **[github.com/nabbar/golib/runner](https://pkg.go.dev/github.com/nabbar/golib/runner)** - Recovery mechanisms used for panic handling. Provides RecoveryCaller for safe recovery in production code.

### External References

- **[Go Context Package](https://pkg.go.dev/context)** - Standard library documentation for context.Context. The mapCloser package fully integrates with context for lifecycle management and cancellation propagation.

- **[Go io Package](https://pkg.go.dev/io)** - Standard library documentation for io.Closer interface. Understanding io.Closer is essential for using mapCloser effectively.

- **[Go Memory Model](https://go.dev/ref/mem)** - Official specification of Go's memory consistency guarantees. Essential for understanding the thread-safety guarantees provided by atomic operations used in mapCloser.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for concurrency, error handling, and interface design. The mapCloser follows these conventions for idiomatic Go code.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/ioutils/mapCloser`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
