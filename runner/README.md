# Runner Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Thread-safe lifecycle management for long-running services and periodic tasks with context cancellation, error tracking, and uptime monitoring.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Subpackages](#subpackages)
  - [startStop - Service Lifecycle Management](#startstop-subpackage)
  - [ticker - Periodic Task Execution](#ticker-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The runner package provides production-ready lifecycle management for Go services and periodic tasks. It handles complex scenarios like graceful shutdown, error collection, concurrent operations, and automatic cleanup with minimal boilerplate code.

### Design Philosophy

1. **Thread-Safe**: All operations use atomic primitives and mutexes for concurrent safety
2. **Context-Aware**: Full support for context cancellation and timeout propagation
3. **Error Tracking**: Automatic error collection with retrieval interfaces
4. **Clean Shutdown**: Exponential backoff polling for graceful cleanup
5. **Panic Recovery**: Built-in panic recovery to prevent process crashes

---

## Key Features

- **Service Lifecycle Management**: Start, stop, and restart long-running services with proper cleanup
- **Periodic Execution**: Execute functions at regular intervals with automatic lifecycle management
- **Thread-Safe Operations**: Atomic operations (`atomic.Value`), mutex protection (`sync.Mutex`)
- **Context Cancellation**: Proper propagation and handling of context cancellation
- **Error Collection**: Track all errors from operations with `ErrorsLast()` and `ErrorsList()`
- **Uptime Tracking**: Monitor how long a service has been running
- **Panic Recovery**: Automatic recovery with stack traces to prevent crashes
- **Idempotent Operations**: Safe to call Start/Stop multiple times
- **Clean State Transitions**: Automatic cleanup of previous instances on restart

---

## Installation

```bash
go get github.com/nabbar/golib/runner
```

---

## Architecture

### Package Structure

The package is organized into specialized subpackages for different execution patterns:

```
runner/
├── interface.go         # Core Runner interface and function types
├── tools.go            # Utility functions (RecoveryCaller, RunNbr, RunTick)
├── startStop/          # Service lifecycle management (Start/Stop pattern)
│   ├── interface.go    # StartStop interface and constructor
│   └── model.go        # Implementation with state management
└── ticker/             # Periodic execution (ticker pattern)
    ├── interface.go    # Ticker interface and constructor
    └── model.go        # Implementation with time.Ticker
```

### Component Overview

```
┌──────────────────────────────────────────────────────┐
│                  Runner Interface                     │
│  Start() Stop() Restart() IsRunning() Uptime()      │
└───────────┬────────────────────────────┬─────────────┘
            │                            │
   ┌────────▼──────────┐      ┌─────────▼──────────┐
   │    startStop      │      │      ticker        │
   │                   │      │                    │
   │ Service lifecycle │      │ Periodic execution │
   │ Start/Stop funcs  │      │ time.Ticker based  │
   │ Blocking pattern  │      │ Regular intervals  │
   └───────────────────┘      └────────────────────┘
```

| Component | Purpose | Pattern | Thread-Safe |
|-----------|---------|---------|-------------|
| **Runner Interface** | Common lifecycle operations | Interface | N/A |
| **startStop** | Long-running services (HTTP server, listeners) | Start blocks, Stop cleans up | ✅ |
| **ticker** | Periodic tasks (cron-like, health checks) | Executes function every N duration | ✅ |
| **Utilities** | Helper functions (recovery, polling) | Standalone functions | ✅ |

### Execution Patterns

**startStop Pattern** (Blocking Service)
- Start function blocks until service terminates
- Stop function triggers graceful shutdown
- Use case: HTTP servers, database connections, message consumers

**ticker Pattern** (Periodic Execution)
- Function executes at regular intervals
- Continues until stopped or context cancelled
- Use case: Health checks, metrics collection, data synchronization

---

## Performance

### Memory Efficiency

The runner package maintains minimal memory overhead:

- **Atomic Operations**: Lock-free reads for state checks (`IsRunning()`, `Uptime()`)
- **Shared Context**: Single context per runner instance
- **Error Pooling**: Efficient error collection using `github.com/nabbar/golib/errors/pool`
- **Zero Allocations**: State checks use atomic operations without heap allocations

### Thread Safety

All operations are thread-safe through:

- **Atomic Values**: `libatm.Value[T]` for lock-free reads (start time, cancel function)
- **Mutex Protection**: `sync.Mutex` for Start/Stop/Restart serialization
- **Exponential Backoff**: Efficient polling for cleanup completion (1ms → 10ms)
- **Goroutine Safety**: Multiple goroutines can safely call operations concurrently

### Cleanup Guarantees

```
Stop Operation Flow:
├─ Cancel context (immediate)
├─ Poll for cleanup (exponential backoff)
│  └─ 1ms → 2ms → 4ms → 8ms → 10ms (max)
└─ Return after max 2 seconds
```

**Stop Guarantees**:
- Context cancellation: Immediate
- Cleanup detection: Up to 2 seconds with exponential backoff
- Idempotent: Safe to call multiple times
- No goroutine leaks: Verified with race detector

---

## Use Cases

This package is designed for scenarios requiring reliable lifecycle management:

**HTTP Servers**
- Start server with graceful shutdown support
- Track uptime and running state
- Collect startup/shutdown errors
- Restart on configuration changes

**Background Workers**
- Process message queues with lifecycle control
- Graceful shutdown on termination signals
- Error tracking for debugging
- Uptime monitoring for health checks

**Periodic Tasks**
- Execute health checks every N seconds
- Scheduled data synchronization
- Metrics collection at regular intervals
- Cache cleanup and maintenance jobs

**Database Connections**
- Maintain connection pools with lifecycle management
- Automatic reconnection with restart
- Monitor connection uptime
- Clean shutdown of connections

**Scheduled Jobs**
- Cron-like periodic execution
- Log rotation and archival
- Report generation
- Data backup operations

---

## Quick Start

### HTTP Server with startStop

Manage an HTTP server lifecycle with graceful shutdown:

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/nabbar/golib/runner/startStop"
)

func main() {
    srv := &http.Server{
        Addr:    ":8080",
        Handler: http.DefaultServeMux,
    }
    
    // Create runner with start and stop functions
    runner := startStop.New(
        func(ctx context.Context) error {
            // Start function blocks until server stops
            fmt.Println("Starting HTTP server on :8080")
            return srv.ListenAndServe()
        },
        func(ctx context.Context) error {
            // Stop function performs graceful shutdown
            fmt.Println("Shutting down HTTP server")
            return srv.Shutdown(ctx)
        },
    )
    
    // Start the server
    if err := runner.Start(context.Background()); err != nil {
        panic(err)
    }
    
    fmt.Printf("Server running (uptime: %v)\n", runner.Uptime())
    
    // Let it run for a while
    time.Sleep(5 * time.Second)
    
    // Stop the server
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := runner.Stop(ctx); err != nil {
        fmt.Printf("Stop error: %v\n", err)
    }
    
    // Check for errors during lifecycle
    if err := runner.ErrorsLast(); err != nil {
        fmt.Printf("Server errors: %v\n", err)
    }
}
```

### Periodic Health Check with ticker

Execute a function at regular intervals:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/nabbar/golib/runner/ticker"
)

func main() {
    checkCount := 0
    
    // Create ticker that runs every 2 seconds
    tick := ticker.New(2*time.Second, func(ctx context.Context, t *time.Ticker) error {
        checkCount++
        fmt.Printf("Health check #%d at %v\n", checkCount, time.Now())
        
        // Simulate occasional errors
        if checkCount%5 == 0 {
            return fmt.Errorf("health check warning at count %d", checkCount)
        }
        return nil
    })
    
    // Start the ticker
    if err := tick.Start(context.Background()); err != nil {
        panic(err)
    }
    
    fmt.Printf("Ticker started (running: %v)\n", tick.IsRunning())
    
    // Let it run for 10 seconds
    time.Sleep(10 * time.Second)
    
    // Stop the ticker
    if err := tick.Stop(context.Background()); err != nil {
        fmt.Printf("Stop error: %v\n", err)
    }
    
    // Check collected errors
    errors := tick.ErrorsList()
    fmt.Printf("Total errors: %d\n", len(errors))
    for i, err := range errors {
        fmt.Printf("  Error %d: %v\n", i+1, err)
    }
}
```

### Context Cancellation

Automatic shutdown when context is cancelled:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/nabbar/golib/runner/ticker"
)

func main() {
    // Create context with 5 second timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    tick := ticker.New(1*time.Second, func(ctx context.Context, t *time.Ticker) error {
        fmt.Printf("Tick at %v\n", time.Now())
        return nil
    })
    
    if err := tick.Start(ctx); err != nil {
        panic(err)
    }
    
    // Wait for context to expire
    <-ctx.Done()
    
    // Ticker automatically stops when context is cancelled
    time.Sleep(100 * time.Millisecond)
    fmt.Printf("Ticker running: %v (stopped automatically)\n", tick.IsRunning())
}
```

### Utility Functions

The package provides utility functions for common patterns:

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/nabbar/golib/runner"
)

func main() {
    // RunNbr: Retry up to N times with custom check and action
    success := runner.RunNbr(5,
        func() bool {
            // Check if condition is met
            return serverIsReady()
        },
        func() {
            // Action to perform between checks
            time.Sleep(100 * time.Millisecond)
        },
    )
    fmt.Printf("Server ready: %v\n", success)
    
    // RunTick: Poll with timeout and ticker interval
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    ready := runner.RunTick(ctx, 100*time.Millisecond, 5*time.Second,
        func() bool {
            return databaseIsConnected()
        },
        func() {
            fmt.Println("Waiting for database...")
        },
    )
    fmt.Printf("Database ready: %v\n", ready)
}

func serverIsReady() bool {
    // Check server readiness
    return true
}

func databaseIsConnected() bool {
    // Check database connection
    return true
}
```

### Panic Recovery

Built-in panic recovery prevents crashes:

```go
package main

import (
    "context"
    
    "github.com/nabbar/golib/runner/startStop"
)

func main() {
    runner := startStop.New(
        func(ctx context.Context) error {
            // This panic will be recovered automatically
            panic("something went wrong!")
        },
        func(ctx context.Context) error {
            return nil
        },
    )
    
    // Start will not crash the process
    _ = runner.Start(context.Background())
    
    // Recovery message is printed to stderr with stack trace
    // Process continues running
}
```

---

## Subpackages

### `startStop` Subpackage

Lifecycle management for long-running services with blocking start and graceful stop operations.

**Features**
- Start function executes asynchronously (runs in goroutine)
- Stop function triggers graceful shutdown
- Automatic cleanup detection with exponential backoff
- Context cancellation support
- Error tracking for both start and stop operations
- Uptime monitoring

**Interface**

```go
type StartStop interface {
    // Runner interface methods
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    IsRunning() bool
    Uptime() time.Duration
    
    // Error tracking methods
    ErrorsLast() error
    ErrorsList() []error
}
```

**State Management**

```
State Transitions:
┌─────────┐  Start()  ┌─────────┐  Stop()   ┌─────────┐
│ Stopped │ ───────> │ Running │ ───────> │ Stopped │
└─────────┘          └─────────┘          └─────────┘
     ▲                                          │
     └──────────── Restart() ───────────────────┘
```

**Example: Message Queue Worker**

```go
import (
    "context"
    "github.com/nabbar/golib/runner/startStop"
)

func NewWorker(queue MessageQueue) startStop.StartStop {
    return startStop.New(
        func(ctx context.Context) error {
            // Start consuming messages (blocks)
            for {
                select {
                case <-ctx.Done():
                    return nil
                case msg := <-queue.Messages():
                    if err := processMessage(msg); err != nil {
                        return err
                    }
                }
            }
        },
        func(ctx context.Context) error {
            // Stop consuming and cleanup
            return queue.Close()
        },
    )
}
```

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/runner/startStop) for complete API.

---

### `ticker` Subpackage

Execute functions at regular intervals with automatic lifecycle management.

**Features**
- Executes function every N duration using `time.Ticker`
- Automatic goroutine lifecycle management
- Context cancellation support
- Error collection from all executions
- Configurable tick interval (minimum 1ms, default 30s)
- Graceful shutdown with cleanup detection

**Interface**

```go
type Ticker interface {
    // Runner interface methods
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    IsRunning() bool
    Uptime() time.Duration
    
    // Error tracking methods
    ErrorsLast() error
    ErrorsList() []error
}
```

**Execution Flow**

```
Ticker Lifecycle:
Start() ──> goroutine created
             ↓
         time.Ticker created
             ↓
         ┌───────────────┐
         │  Tick Loop    │ ──> Execute function
         │  <-ticker.C   │ ──> Collect errors
         │  <-ctx.Done() │ ──> Check cancellation
         └───────────────┘
             ↓
Stop() ──> Context cancelled
             ↓
         Cleanup: ticker.Stop(), clear uptime
```

**Example: Metrics Collection**

```go
import (
    "context"
    "time"
    "github.com/nabbar/golib/runner/ticker"
)

func StartMetricsCollector() ticker.Ticker {
    return ticker.New(30*time.Second, func(ctx context.Context, t *time.Ticker) error {
        // Collect metrics
        metrics := collectSystemMetrics()
        
        // Send to monitoring service
        if err := sendMetrics(metrics); err != nil {
            return err // Error is collected automatically
        }
        
        return nil
    })
}

// Usage
collector := StartMetricsCollector()
collector.Start(context.Background())

// Runs every 30 seconds until stopped
time.Sleep(5 * time.Minute)

collector.Stop(context.Background())

// Check for errors during collection
if err := collector.ErrorsLast(); err != nil {
    log.Printf("Metrics collection errors: %v", err)
}
```

**Example: Cache Cleanup**

```go
func StartCacheCleanup(cache *Cache) ticker.Ticker {
    return ticker.New(10*time.Minute, func(ctx context.Context, t *time.Ticker) error {
        // Remove expired entries
        removed := cache.RemoveExpired()
        log.Printf("Removed %d expired cache entries", removed)
        return nil
    })
}
```

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/runner/ticker) for complete API.

---

## Best Practices

**Always Use Context**
```go
// ✅ Good: Proper context usage
func startService(ctx context.Context) {
    runner := startStop.New(startFunc, stopFunc)
    if err := runner.Start(ctx); err != nil {
        log.Fatal(err)
    }
}

// ❌ Bad: Using background context everywhere
func startServiceBad() {
    runner := startStop.New(startFunc, stopFunc)
    runner.Start(context.Background()) // Can't be cancelled externally
}
```

**Handle Errors**
```go
// ✅ Good: Check and handle errors
func runService(ctx context.Context) error {
    runner := startStop.New(startFunc, stopFunc)
    
    if err := runner.Start(ctx); err != nil {
        return fmt.Errorf("start failed: %w", err)
    }
    
    // ... later ...
    
    if err := runner.Stop(ctx); err != nil {
        return fmt.Errorf("stop failed: %w", err)
    }
    
    // Check for operational errors
    if errs := runner.ErrorsList(); len(errs) > 0 {
        return fmt.Errorf("service errors: %v", errs)
    }
    
    return nil
}

// ❌ Bad: Ignoring errors
func runServiceBad() {
    runner := startStop.New(startFunc, stopFunc)
    runner.Start(context.Background())
    // No error checking!
}
```

**Graceful Shutdown**
```go
// ✅ Good: Proper shutdown with timeout
func main() {
    runner := startStop.New(startFunc, stopFunc)
    runner.Start(context.Background())
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    <-sigChan
    
    // Stop with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := runner.Stop(ctx); err != nil {
        log.Printf("Stop error: %v", err)
    }
}

// ❌ Bad: Abrupt shutdown
func mainBad() {
    runner := startStop.New(startFunc, stopFunc)
    runner.Start(context.Background())
    // Process exits without cleanup
}
```

**Appropriate Ticker Intervals**
```go
// ✅ Good: Reasonable intervals
tick := ticker.New(30*time.Second, healthCheckFunc)  // Every 30 seconds
tick := ticker.New(5*time.Minute, cleanupFunc)       // Every 5 minutes

// ❌ Bad: Too frequent, wastes CPU
tick := ticker.New(10*time.Millisecond, func(ctx context.Context, t *time.Ticker) error {
    // Heavy operation every 10ms = 100 times/second!
    return doHeavyWork()
})
```

**Check Running State**
```go
// ✅ Good: Check before operations
if runner.IsRunning() {
    uptime := runner.Uptime()
    log.Printf("Service uptime: %v", uptime)
} else {
    log.Println("Service is not running")
}

// ✅ Good: Idempotent operations
runner.Stop(ctx) // Safe to call even if not running
```

**Error Collection in Tickers**
```go
// ✅ Good: Errors don't stop the ticker
tick := ticker.New(1*time.Second, func(ctx context.Context, t *time.Ticker) error {
    if err := performCheck(); err != nil {
        // Return error - ticker continues, error is collected
        return fmt.Errorf("check failed: %w", err)
    }
    return nil
})

// Later, review all errors
tick.Stop(context.Background())
for i, err := range tick.ErrorsList() {
    log.Printf("Error %d: %v", i+1, err)
}
```

---

## Testing

**Test Suite**: 100+ specs using Ginkgo v2 and Gomega (≥80% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
ginkgo -race -cover
```

**Coverage Areas**
- Lifecycle operations (Start, Stop, Restart)
- Concurrent operations and thread safety
- Error collection and retrieval
- Context cancellation
- Edge cases (nil contexts, quick exits, panics)
- Uptime tracking

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Proper goroutine cleanup
- ✅ Panic recovery without crashes

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥80%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document all public functions and types

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

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Advanced Scheduling**
- Cron-like scheduling syntax
- Multiple schedule support
- Skip overlapping executions
- Timezone-aware scheduling

**Health Monitoring**
- Built-in health check endpoints
- Automatic restart on failures
- Exponential backoff for retries
- Circuit breaker integration

**Metrics & Observability**
- Execution duration tracking
- Success/failure rate metrics
- Prometheus integration
- OpenTelemetry tracing

**Advanced Features**
- Graceful reload without downtime
- Priority-based execution
- Resource-aware scheduling
- Distributed coordination (etcd/consul)

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/runner)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../CONTRIBUTING.md)
