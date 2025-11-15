# MapCloser Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-80.2%25-green)]()

Thread-safe, context-aware manager for multiple io.Closer instances with automatic cleanup and error aggregation.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides a robust solution for managing multiple resources that implement `io.Closer`. It automatically closes all registered closers when a context is cancelled or when manually triggered, making resource cleanup safe and predictable in concurrent applications.

### Design Philosophy

1. **Automatic Cleanup**: Context-driven resource lifecycle management
2. **Thread Safety**: All operations safe for concurrent use via atomics
3. **Error Aggregation**: Collect all errors instead of failing fast
4. **Simplicity**: Single responsibility - manage closer lifecycle
5. **Fail-Safe**: Handles nil closers and double-close gracefully

---

## Key Features

- **Thread-Safe**: All operations use atomic primitives for safe concurrent access
- **Context-Aware**: Automatic cleanup when context is cancelled, timed out, or done
- **Error Aggregation**: Collects errors from all closers, returns combined error
- **Cloneable**: Create independent copies for hierarchical resource management
- **Nil-Safe**: Gracefully handles nil closers without panics
- **100ms Polling**: Background goroutine monitors context every 100ms
- **Production Ready**: 80.2% test coverage, 29 specs

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils/mapCloser
```

---

## Architecture

### Package Structure

```
mapCloser/
├── interface.go    # Public API and Closer interface
└── model.go        # Internal implementation with atomics
```

### Component Diagram

```
┌──────────────────────────────────────────┐
│         mapCloser.Closer                 │
│    (Thread-Safe Resource Manager)       │
└──────────────┬───────────────────────────┘
               │
    ┌──────────▼──────────┐
    │  Atomic State       │
    ├─────────────────────┤
    │ • atomic.Bool       │ ← Closed flag
    │ • atomic.Uint64     │ ← Counter
    │ • context.Context   │ ← Monitored context
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  Closer Storage     │
    │  (indexed map)      │
    ├─────────────────────┤
    │ 1 → io.Closer       │
    │ 2 → io.Closer       │
    │ 3 → io.Closer       │
    │ ...                 │
    └─────────────────────┘
```

### Operation Flow

```
User calls New(ctx)
       ↓
1. Create cancellable context
       ↓
2. Initialize atomic state
   ├─ Bool (closed=false)
   ├─ Uint64 (counter=0)
   └─ Config storage (empty)
       ↓
3. Start monitoring goroutine
   └─ Poll context every 100ms
       ↓
4. Return Closer instance
       ↓
User calls Add(closer1, closer2, ...)
       ↓
5. Check if closed/context done
   ├─ Yes → no-op
   └─ No  → Store each closer
       ↓
6. Increment counter (atomic)
   Store closer at index
       ↓
Context cancelled OR Close() called
       ↓
7. Set closed flag (atomic)
       ↓
8. Iterate all stored closers
   ├─ Call Close() on each
   ├─ Collect errors
   └─ Continue even if some fail
       ↓
9. Cancel context
   Return aggregated error
```

### Thread Safety Model

```
Closer Instance
├─ atomic.Bool (closed) ← Read/Write atomic
├─ atomic.Uint64 (counter) ← Increment atomic
├─ libctx.Config (storage) ← Thread-safe map
└─ func() (cancel) ← Called once

Concurrent Operations:
✓ Multiple Add() calls
✓ Add() + Get() simultaneously
✓ Add() + Len() simultaneously
✓ Close() from context goroutine + manual Close()
✓ Clone() + operations on original
```

---

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "os"
    "github.com/nabbar/golib/ioutils/mapCloser"
)

func main() {
    // Create closer manager
    ctx := context.Background()
    closer := mapCloser.New(ctx)
    defer closer.Close()
    
    // Register resources
    file1, _ := os.Open("file1.txt")
    file2, _ := os.Open("file2.txt")
    
    closer.Add(file1, file2)
    
    // Use resources...
    // All automatically closed on defer
}
```

### Context-Based Cleanup

```go
// Automatic cleanup on timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

closer := mapCloser.New(ctx)
closer.Add(conn1, conn2, conn3)

// When timeout occurs, all resources auto-close
// No manual Close() needed
```

### Error Handling

```go
closer := mapCloser.New(context.Background())
closer.Add(resource1, resource2, resource3)

if err := closer.Close(); err != nil {
    // err contains aggregated errors: "error1, error2"
    log.Printf("Some resources failed to close: %v", err)
}
```

---

## Performance

### Operation Metrics

The package adds **minimal overhead** to resource management:

| Operation | Complexity | Time | Notes |
|-----------|-----------|------|-------|
| New() | O(1) | ~1 µs | Creates context + goroutine |
| Add() | O(1) | ~100 ns | Atomic increment + store |
| Get() | O(n) | ~n µs | Iterates all closers |
| Len() | O(1) | ~10 ns | Atomic load |
| Clean() | O(n) | ~n µs | Clears storage |
| Clone() | O(n) | ~n µs | Copies all closers |
| Close() | O(n) | ~n ms | Closes each closer sequentially |

*n = number of registered closers*

### Memory Efficiency

- **Instance Size**: ~80 bytes (atomic values + pointers)
- **Per Closer**: ~16 bytes (key-value pair in storage)
- **Allocations**: Minimal (atomic operations, no per-call allocation)
- **Goroutine**: 1 background goroutine per instance

### Context Monitoring

- **Polling Interval**: 100ms
- **Latency**: Up to 100ms delay for automatic cleanup
- **CPU Usage**: Negligible (~0.001% per instance)

---

## Use Cases

### HTTP Server with Multiple Connections

Manage database connections, cache connections, and other resources:

```go
type Server struct {
    closer mapCloser.Closer
}

func NewServer(ctx context.Context) (*Server, error) {
    closer := mapCloser.New(ctx)
    
    // Connect to database
    db, err := sql.Open("postgres", connString)
    if err != nil {
        return nil, err
    }
    closer.Add(db)
    
    // Connect to Redis
    redis, err := redisClient.Connect()
    if err != nil {
        closer.Close() // Close DB
        return nil, err
    }
    closer.Add(redis)
    
    // Open log file
    logFile, _ := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE, 0644)
    closer.Add(logFile)
    
    return &Server{closer: closer}, nil
}

func (s *Server) Shutdown() error {
    return s.closer.Close() // Closes all: DB, Redis, logFile
}
```

**Why**: Single point of cleanup for all server resources.

### Worker Pool Management

Each worker manages its own resources with independent cleanup:

```go
type Worker struct {
    id     int
    closer mapCloser.Closer
}

func NewWorker(ctx context.Context, id int) *Worker {
    closer := mapCloser.New(ctx)
    
    // Open worker-specific resources
    logFile, _ := os.Create(fmt.Sprintf("worker-%d.log", id))
    tempFile, _ := os.CreateTemp("", fmt.Sprintf("worker-%d-*", id))
    
    closer.Add(logFile, tempFile)
    
    return &Worker{id: id, closer: closer}
}

func (w *Worker) Shutdown() {
    w.closer.Close()
}

// Start multiple workers
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // Cancels context, triggers all workers to close

for i := 0; i < 10; i++ {
    worker := NewWorker(ctx, i)
    go worker.Run()
}
```

**Why**: Each worker's resources auto-clean when context is cancelled.

### Hierarchical Resource Management with Clone

Parent-child resource relationships:

```go
func main() {
    ctx := context.Background()
    
    // Parent closer for shared resources
    parentCloser := mapCloser.New(ctx)
    sharedDB, _ := sql.Open("postgres", connString)
    parentCloser.Add(sharedDB)
    
    // Child closers for per-request resources
    handleRequest := func(req *Request) {
        reqCloser := parentCloser.Clone() // Independent copy
        
        tempFile, _ := os.CreateTemp("", "request-*")
        reqCloser.Add(tempFile)
        
        // Process request...
        
        reqCloser.Close() // Closes tempFile only, not sharedDB
    }
    
    // Process multiple requests
    for _, req := range requests {
        go handleRequest(req)
    }
    
    parentCloser.Close() // Finally close sharedDB
}
```

**Why**: Clone enables independent resource scopes while sharing some closers.

### Testing Cleanup

Automatic cleanup of test resources:

```go
func TestFeature(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    closer := mapCloser.New(ctx)
    t.Cleanup(func() { closer.Close() })
    
    // Create test resources
    testDB := setupTestDB(t)
    testServer := setupTestServer(t)
    testFiles := setupTestFiles(t)
    
    closer.Add(testDB, testServer, testFiles...)
    
    // Run test...
    // All resources auto-clean on test end or timeout
}
```

**Why**: Ensures test resources are always cleaned up, even on timeout or panic.

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

### New

```go
func New(ctx context.Context) Closer
```

Creates a new Closer that monitors the context for cancellation.

**Parameters:**
- `ctx context.Context`: Context to monitor

**Returns:**
- `Closer`: Thread-safe closer manager

**Example:**
```go
ctx, cancel := context.WithCancel(context.Background())
closer := mapCloser.New(ctx)
defer closer.Close()
```

### Methods

**Add(clo ...io.Closer)**
- Registers one or more io.Closer for management
- No-op if closer is already closed or context is done
- Nil closers are accepted and filtered out
- Thread-safe

**Get() []io.Closer**
- Returns copy of all registered closers (excluding nil)
- Safe to modify returned slice
- Returns empty slice if closed
- Thread-safe

**Len() int**
- Returns count of added closers (including nil)
- Returns 0 on overflow (>math.MaxInt)
- Thread-safe

**Clean()**
- Removes all closers without closing them
- Resets counter to zero
- No-op if already closed
- Thread-safe

**Clone() Closer**
- Creates independent copy with same state
- Shares context, independent storage
- Returns nil if already closed
- Thread-safe

**Close() error**
- Closes all closers and cancels context
- Returns aggregated error if any fail
- Subsequent calls return error
- Thread-safe

---

## Best Practices

### 1. Always Use defer for Cleanup

```go
closer := mapCloser.New(ctx)
defer closer.Close() // Ensures cleanup even on panic
```

### 2. Check Errors from Close()

```go
if err := closer.Close(); err != nil {
    log.Printf("Some resources failed to close: %v", err)
    // Take corrective action
}
```

### 3. Choose Appropriate Context

```go
// For long-running servers
ctx := context.Background()

// For operations with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// For cancellable operations
ctx, cancel := context.WithCancel(parent Context)
defer cancel()
```

### 4. Nil Closers Are OK

```go
var file *os.File // nil
closer.Add(file) // Safe - filtered out during Get() and Close()
```

### 5. Don't Double-Close

```go
closer.Close()
closer.Close() // Returns error: already closed
```

### 6. Use Clone for Hierarchical Management

```go
parentCloser := mapCloser.New(ctx)
parentCloser.Add(sharedResource)

childCloser := parentCloser.Clone()
childCloser.Add(tempResource)

childCloser.Close() // Closes temp only
parentCloser.Close() // Closes shared
```

---

## Testing

The package has **80.2% test coverage** with 29 comprehensive specs using Ginkgo v2 and Gomega.

### Run Tests

```bash
# Standard go test
go test -v -cover .

# With Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo -v -cover

# With race detector
go test -race .
```

### Test Statistics

| Metric | Value |
|--------|-------|
| Total Specs | 29 |
| Coverage | 80.2% |
| Execution Time | ~5ms |
| Success Rate | 100% |

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI** to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass tests: `go test ./...`
- Maintain test coverage (≥80%)
- Follow existing code style and patterns
- Add GoDoc comments for all public elements

**Documentation**
- Update README.md for new features
- Add practical examples for common use cases
- Keep TESTING.md synchronized with test changes
- Use clear, concise English

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Use Ginkgo/Gomega BDD style
- Test thread safety for concurrent operations

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results and coverage
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for project-wide guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Features**
- Callback hooks: Pre/post close callbacks
- Prioritized closing: Close in specific order
- Graceful shutdown: Timeout-based forced close
- Metrics: Track close success/failure rates

**Performance**
- Configurable polling interval (currently 100ms)
- Batch close operations
- Async close support

**Developer Experience**
- Helper for HTTP server shutdown
- Integration with signal handling
- Structured error reporting

Suggestions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License © Nicolas JUHEL

All source files in this package are licensed under the MIT License. See individual files for the full license header.

---

## Resources

**Documentation**
- [GoDoc Reference](https://pkg.go.dev/github.com/nabbar/golib/ioutils/mapCloser)
- [Testing Guide](TESTING.md)
- [Go Context Package](https://pkg.go.dev/context)
- [Go io Package](https://pkg.go.dev/io)

**Related Packages**
- [bufferReadCloser](../bufferReadCloser) - I/O wrappers with close support
- [iowrapper](../iowrapper) - Flexible I/O wrapper with custom functions
- [ioutils](../) - Parent package with additional I/O utilities

**Community**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [Contributing Guide](../../CONTRIBUTING.md)

---

**Version**: Go 1.19+ on Linux, macOS, Windows  
**Maintained By**: mapCloser Package Contributors
