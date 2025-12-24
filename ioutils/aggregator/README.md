# IOUtils Aggregator

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-84.8%25-brightgreen)](TESTING.md)

Thread-safe write aggregator that serializes concurrent write operations to a single output function with optional periodic callbacks and real-time monitoring.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Buffer Sizing](#buffer-sizing)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Example](#basic-example)
  - [File Writing](#file-writing)
  - [Socket to File](#socket-to-file)
  - [With Callbacks](#with-callbacks)
  - [Real-time Monitoring](#real-time-monitoring)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Aggregator Interface](#aggregator-interface)
  - [Configuration](#configuration)
  - [Metrics](#metrics)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **aggregator** package provides a high-performance, thread-safe solution for aggregating multiple concurrent write operations into a single sequential output stream. It's designed to handle scenarios where multiple goroutines need to write to a resource that doesn't support concurrent access (e.g., files, network sockets, databases).

### Design Philosophy

1. **Thread Safety First**: All operations are safe for concurrent use without external synchronization
2. **Performance Oriented**: Minimal overhead with atomic operations and lock-free metrics
3. **Predictable Backpressure**: Explicit buffer sizing allows controlled memory usage and flow control
4. **Observable**: Real-time metrics for monitoring buffer state and memory consumption
5. **Context-Aware**: Full integration with Go's context for cancellation and deadline propagation

### Key Features

- ✅ **Concurrent Write Aggregation**: Multiple goroutines can write safely
- ✅ **Buffered Channel**: Configurable buffer size for performance tuning
- ✅ **Periodic Callbacks**: Optional async/sync functions triggered by timers
- ✅ **Real-time Metrics**: Four monitoring metrics (count and size based)
- ✅ **Context Integration**: Implements `context.Context` for lifecycle management
- ✅ **Lifecycle Management**: Implements `librun.StartStop` for controlled start/stop
- ✅ **Error Recovery**: Automatic panic recovery with detailed logging
- ✅ **Zero Dependencies**: Only standard library and internal golib packages

---

## Architecture

### Component Diagram

```
┌───────────────────────────────────────────────────────────────┐
│                       Aggregator                              │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────┐    ┌──────────────┐    ┌─────────────┐       │
│  │  Goroutine  │───▶│              │───▶│             │       │
│  │      1      │    │              │    │             │       │
│  └─────────────┘    │              │    │             │       │
│                     │   Buffered   │    │  FctWriter  │───▶ Output
│  ┌─────────────┐    │   Channel    │    │  (serial)   │       │
│  │  Goroutine  │───▶│  (BufWriter) │───▶│             │       │
│  │      2      │    │              │    │             │       │
│  └─────────────┘    │              │    │             │       │
│                     │              │    │             │       │
│  ┌─────────────┐    │              │    │             │       │
│  │  Goroutine  │───▶│              │───▶│             │       │
│  │      N      │    │              │    │             │       │
│  └─────────────┘    └──────────────┘    └─────────────┘       │
│                                                               │
│       ▲                    │                    │             │
│       │                    │                    │             │
│       │                    ▼                    ▼             │
│  ┌─────────┐    ┌─────────────────┐  ┌─────────────────┐      │
│  │ Metrics │    │   AsyncFct      │  │    SyncFct      │      │
│  │ - Count │    │ (concurrent)    │  │  (blocking)     │      │
│  │ - Size  │    │ Ticker-based    │  │  Ticker-based   │      │
│  └─────────┘    └─────────────────┘  └─────────────────┘      │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Write(data) → Buffer Check → Channel Send → Processing Loop
                    │              │              │
                    │              │              ├─▶ FctWriter(data)
                    ▼              │              │
            Metrics Update         │              ├─▶ AsyncFct (timer)
            - NbWaiting++          │              │
            - SizeWaiting+len      │              └─▶ SyncFct (timer)
                    │              │
                    │              ▼
                    │      Metrics Update
                    │      - NbProcessing++
                    │      - SizeProcessing+len
                    │              │
                    ▼              ▼
            Return Success    Process & Update
                              - NbProcessing--
                              - SizeProcessing-len
```

### Buffer Sizing

The `BufWriter` parameter is critical for performance and behavior. It defines the capacity of the internal buffered channel.

**Sizing Formula:**

```
BufWriter = (WriteRate × MaxProcessingTime) × 1.5
```

Where:
- **WriteRate**: Expected writes/second under typical load
- **MaxProcessingTime**: max(SyncTimer, FctWriter execution time)
- **1.5**: Safety margin (20-50%) for burst handling

**Memory Estimation:**

```
MaxMemory = BufWriter × AverageMessageSize
```

**Trade-offs:**

| Buffer Size | Pros | Cons |
|-------------|------|------|
| **Too Small** | Low memory usage | Write() blocks, backpressure cascades |
| **Optimal** | No blocking under normal load | Balanced memory usage |
| **Too Large** | Absorbs large bursts | Excessive memory, hides performance issues |

See [doc.go](doc.go) for detailed buffer sizing guidelines and example calculations.

---

## Performance

### Benchmarks

Based on actual test results from the comprehensive test suite:

| Operation | Median | Mean | Max |
|-----------|--------|------|-----|
| **Start Time** | 10.7ms | 11.0ms | 15.2ms |
| **Stop Time** | 12.1ms | 12.4ms | 16.9ms |
| **Restart Time** | 33.8ms | 34.2ms | 42.1ms |
| **Write Latency** | <1ms | <1ms | <5ms |
| **Complete Cycle** | 34.2ms | 34.6ms | 40.4ms |
| **Metrics Read** | <1µs | <5µs | <10µs |

**Throughput:**
- Single writer: **~1000 writes/second**
- Concurrent (10 writers): **~5000-10000 writes/second** (depends on FctWriter speed)
- Network I/O scenarios: **limited by FctWriter**, not aggregator overhead

### Memory Usage

```
Base overhead:        ~2KB (struct + atomics)
Per buffered item:    len([]byte) + ~48 bytes (slice header + channel overhead)
Total at capacity:    BufWriter × (AvgMessageSize + 48 bytes)
```

**Example:**
- BufWriter = 1000
- Average message = 512 bytes
- Peak memory ≈ 1000 × 560 = 560KB

### Scalability

- **Concurrent Writers**: Tested with up to 100 concurrent goroutines
- **Buffer Sizes**: Validated from 1 to 10,000 items
- **Message Sizes**: Tested from 1 byte to 1MB
- **Zero Race Conditions**: All tests pass with `-race` detector

---

## Use Cases

### 1. Socket Server to File Logger

**Problem**: Multiple concurrent socket connections writing to a single log file (non-concurrent filesystem).

```go
// Socket handler writes are serialized to file
agg, _ := aggregator.New(ctx, aggregator.Config{
    BufWriter: 1000,  // Handle bursts
    FctWriter: func(p []byte) (int, error) {
        return logFile.Write(p)
    },
})
```

**Real-world**: Used with `github.com/nabbar/golib/socket/server` for high-traffic socket applications.

### 2. Database Connection Pool Writer

**Problem**: Serialize writes to a single DB connection from multiple producers.

```go
agg, _ := aggregator.New(ctx, aggregator.Config{
    BufWriter: 500,
    FctWriter: func(p []byte) (int, error) {
        _, err := db.Exec("INSERT INTO logs VALUES (?)", string(p))
        return len(p), err
    },
    SyncTimer: 5 * time.Second,
    SyncFct: func(ctx context.Context) {
        db.Exec("COMMIT")  // Periodic commit
    },
})
```

### 3. Network Stream Multiplexer

**Problem**: Multiple data sources writing to a single network connection.

```go
agg, _ := aggregator.New(ctx, aggregator.Config{
    BufWriter: 100,
    FctWriter: func(p []byte) (int, error) {
        return networkConn.Write(p)
    },
    AsyncTimer: 30 * time.Second,
    AsyncFct: func(ctx context.Context) {
        // Send keepalive
        networkConn.Write([]byte("PING\n"))
    },
})
```

### 4. Metrics Collection Pipeline

**Problem**: Collect metrics from many sources and write to time-series database.

```go
agg, _ := aggregator.New(ctx, aggregator.Config{
    BufWriter: 10000,  // High-frequency metrics
    FctWriter: func(p []byte) (int, error) {
        return metricsDB.Write(p)
    },
    SyncTimer: 10 * time.Second,
    SyncFct: func(ctx context.Context) {
        metricsDB.Flush()  // Batch flush
    },
})
```

### 5. Temporary File Accumulator

**Problem**: Accumulate data from concurrent sources into a temp file, then process atomically.

```go
tmpFile, _ := os.CreateTemp("", "accumulated-*.dat")

agg, _ := aggregator.New(ctx, aggregator.Config{
    BufWriter: 200,
    FctWriter: func(p []byte) (int, error) {
        return tmpFile.Write(p)
    },
    SyncTimer: 1 * time.Minute,
    SyncFct: func(ctx context.Context) {
        tmpFile.Sync()  // Ensure data is flushed
    },
})
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/aggregator
```

### Basic Example

```go
package main

import (
    "context"
    "fmt"
	"time"
    "github.com/nabbar/golib/ioutils/aggregator"
)

func main() {
    ctx := context.Background()

    // Create aggregator
    cfg := aggregator.Config{
        BufWriter: 100,
        FctWriter: func(p []byte) (int, error) {
            fmt.Printf("Writing: %s\n", string(p))
            return len(p), nil
        },
    }

    agg, err := aggregator.New(ctx, cfg)
    if err != nil {
        panic(err)
    }

    // Start processing
    agg.Start(ctx)
    defer agg.Close()

    // Write from multiple goroutines
    for i := 0; i < 10; i++ {
        go func(id int) {
            data := fmt.Sprintf("Message from goroutine %d", id)
            agg.Write([]byte(data))
        }(i)
    }

    // Wait for processing
    time.Sleep(1 * time.Second)
}
```

### File Writing

```go
file, _ := os.Create("output.log")
defer file.Close()

cfg := aggregator.Config{
    BufWriter: 500,
    FctWriter: func(p []byte) (int, error) {
        return file.Write(p)
    },
    SyncTimer: 5 * time.Second,
    SyncFct: func(ctx context.Context) {
        file.Sync()  // Periodic fsync
    },
}

agg, _ := aggregator.New(ctx, cfg)
agg.Start(ctx)
defer agg.Close()

// Write safely from multiple goroutines
agg.Write([]byte("Log entry 1\n"))
agg.Write([]byte("Log entry 2\n"))
```

### Socket to File

```go
// Temporary file for socket data
tmpFile, _ := os.CreateTemp("", "socket-data-*.tmp")
defer os.Remove(tmpFile.Name())

cfg := aggregator.Config{
    BufWriter: 1000,
    FctWriter: func(p []byte) (int, error) {
        return tmpFile.Write(p)
    },
}

agg, _ := aggregator.New(ctx, cfg)
agg.Start(ctx)

// In socket server handler (github.com/nabbar/golib/socket/server)
func handleConnection(conn net.Conn) {
    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        agg.Write(scanner.Bytes())
    }
}
```

### With Callbacks

```go
cfg := aggregator.Config{
    BufWriter: 100,
    FctWriter: func(p []byte) (int, error) {
        return database.Insert(p)
    },
    AsyncTimer: 1 * time.Minute,
    AsyncMax:   3,  // Max 3 concurrent async calls
    AsyncFct: func(ctx context.Context) {
        database.Cleanup()  // Background cleanup
    },
    SyncTimer: 10 * time.Second,
    SyncFct: func(ctx context.Context) {
        database.Commit()  // Periodic commit
    },
}

agg, _ := aggregator.New(ctx, cfg)
agg.Start(ctx)
defer agg.Close()
```

### Real-time Monitoring

```go
agg, _ := aggregator.New(ctx, cfg)
agg.Start(ctx)

// Monitor loop
go func() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            waiting := agg.NbWaiting()
            processing := agg.NbProcessing()
            sizeWaiting := agg.SizeWaiting()
            sizeProcessing := agg.SizeProcessing()

            totalMemory := sizeWaiting + sizeProcessing
            bufferUsage := float64(processing) / float64(bufWriter) * 100

            log.Printf("Buffer: %.1f%% | Memory: %d bytes | Waiting: %d",
                bufferUsage, totalMemory, waiting)

            // Alert if backpressure
            if waiting > 0 {
                log.Warn("Backpressure detected!")
            }
        case <-ctx.Done():
            return
        }
    }
}()
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **84.8% code coverage** and **124 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and lifecycle operations
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks (throughput, latency, memory)
- ✅ Error handling and edge cases
- ✅ Context integration and cancellation

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Buffer Sizing:**
```go
// Calculate based on actual metrics
writeRate := 100  // writes/second
maxTime := 5      // seconds (SyncTimer or FctWriter max)
bufWriter := int(float64(writeRate * maxTime) * 1.5)  // 750
```

**Context Management:**
```go
// Use context for lifecycle
ctx, cancel := context.WithCancel(parent)
defer cancel()

agg, _ := aggregator.New(ctx, cfg)
agg.Start(ctx)
defer agg.Close()  // Always close
```

**Error Monitoring:**
```go
// Periodic error checking
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        if err := agg.ErrorsLast(); err != nil {
            log.Error("Aggregator error:", err)
        }
    }
}()
```

**Graceful Shutdown:**
```go
// Wait for completion
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt)

<-sigChan
log.Info("Shutting down...")

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := agg.Stop(); err != nil {
    log.Error("Stop error:", err)
}
```

**Metric-Based Alerting:**
```go
// Alert on sustained backpressure
const threshold = 10  // seconds
backpressureStart := time.Time{}

for range ticker.C {
    if agg.NbWaiting() > 0 {
        if backpressureStart.IsZero() {
            backpressureStart = time.Now()
        } else if time.Since(backpressureStart) > threshold*time.Second {
            alert.Send("Backpressure sustained for %v", time.Since(backpressureStart))
        }
    } else {
        backpressureStart = time.Time{}
    }
}
```

### ❌ DON'T

**Don't ignore buffer sizing:**
```go
// ❌ BAD: Default buffer (1) causes blocking
cfg := aggregator.Config{
    FctWriter: slowWriter,  // Takes 100ms
}

// ✅ GOOD: Sized for throughput
cfg := aggregator.Config{
    FctWriter: slowWriter,
    BufWriter: 1000,  // Can handle 10 writes/sec for 100s
}
```

**Don't write after Close:**
```go
// ❌ BAD: Use after close
agg.Close()
agg.Write(data)  // Returns ErrClosedResources

// ✅ GOOD: Check running state
if agg.IsRunning() {
    agg.Write(data)
}
```

**Don't Start() multiple times:**
```go
// ❌ BAD: Concurrent starts
go agg.Start(ctx)
go agg.Start(ctx)  // Returns ErrStillRunning

// ✅ GOOD: Single start with check
if !agg.IsRunning() {
    agg.Start(ctx)
}
```

**Don't block in FctWriter without buffer:**
```go
// ❌ BAD: Slow writer with small buffer
cfg := aggregator.Config{
    BufWriter: 1,
    FctWriter: func(p []byte) (int, error) {
        time.Sleep(1 * time.Second)  // Blocks everything
        return len(p), nil
    },
}

// ✅ GOOD: Buffer sized for latency
cfg := aggregator.Config{
    BufWriter: 100,  // Absorbs 100 writes during 1s processing
    FctWriter: slowWriter,
}
```

**Don't ignore metrics:**
```go
// ❌ BAD: No monitoring
agg.Write(data)

// ✅ GOOD: Monitor and adapt
if agg.NbWaiting() > 0 {
    log.Warn("Backpressure, consider increasing BufWriter")
}
if float64(agg.SizeProcessing()) > memoryBudget {
    log.Error("Memory limit exceeded")
}
```

---

## API Reference

### Aggregator Interface

```go
type Aggregator interface {
    context.Context
    librun.StartStop
    io.Closer
    io.Writer

    // Monitoring metrics
    NbWaiting() int64
    NbProcessing() int64
    SizeWaiting() int64
    SizeProcessing() int64
}
```

**Methods:**

- **`Write(p []byte) (int, error)`**: Write data to aggregator (thread-safe)
- **`Start(ctx context.Context) error`**: Start processing loop
- **`Stop() error`**: Stop processing and wait for completion
- **`Restart(ctx context.Context) error`**: Stop and restart
- **`Close() error`**: Stop and release all resources
- **`IsRunning() bool`**: Check if aggregator is running
- **`Uptime() time.Duration`**: Get running duration
- **`ErrorsLast() error`**: Get most recent error
- **`ErrorsList() []error`**: Get all errors

### Configuration

```go
type Config struct {
    // Core
    FctWriter  func(p []byte) (n int, err error)  // Required: write function
    BufWriter  int                                 // Buffer size (default: 1)

    // Async callback
    AsyncTimer time.Duration                       // Async callback interval
    AsyncMax   int                                 // Max concurrent async calls
    AsyncFct   func(ctx context.Context)           // Async callback function

    // Sync callback
    SyncTimer  time.Duration                       // Sync callback interval
    SyncFct    func(ctx context.Context)           // Sync callback function
}
```

**Validation:**
- `FctWriter` is required (returns `ErrInvalidWriter` if nil)
- Default `BufWriter` is 1 if not specified
- Timers of 0 disable callbacks
- `AsyncMax` of -1 means unlimited concurrency

### Metrics

#### Count-Based Metrics

**`NbWaiting() int64`**
- Number of `Write()` calls currently blocked waiting for buffer space
- **Healthy**: Always 0
- **Warning**: > 0 indicates backpressure
- **Critical**: Growing value indicates buffer too small

**`NbProcessing() int64`**
- Number of items buffered in channel awaiting processing
- **Healthy**: Varies with load but < BufWriter
- **Warning**: Consistently near BufWriter
- **Critical**: Always at BufWriter (buffer saturated)

#### Size-Based Metrics

**`SizeWaiting() int64`**
- Total bytes in blocked `Write()` calls
- **Healthy**: 0
- **Warning**: > 0 indicates memory pressure from blocking
- **Use**: Detect memory buildup before it becomes critical

**`SizeProcessing() int64`**
- Total bytes in buffer awaiting processing
- **Healthy**: Varies with load
- **Use**: Actual memory consumption of buffer
- **Formula**: `AvgMsgSize = SizeProcessing / NbProcessing`

#### Derived Metrics

```go
// Buffer utilization percentage
bufferUsage := float64(agg.NbProcessing()) / float64(bufWriter) * 100

// Total memory in flight
totalMemory := agg.SizeWaiting() + agg.SizeProcessing()

// Average message size
avgSize := agg.SizeProcessing() / max(agg.NbProcessing(), 1)

// Estimated max memory
maxMemory := bufWriter * avgSize
```

### Error Codes

```go
var (
    ErrInvalidWriter   = errors.New("invalid writer")      // FctWriter is nil
    ErrInvalidInstance = errors.New("invalid instance")    // Internal corruption
    ErrStillRunning    = errors.New("still running")       // Start() while running
    ErrClosedResources = errors.New("closed resources")    // Write() after Close()
)
```

**Error Handling:**

- Errors from `FctWriter` are logged internally but don't stop processing
- Use `ErrorsLast()` and `ErrorsList()` to retrieve logged errors
- Context errors propagate through `Err()` method
- Panics in callbacks are recovered automatically

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >85%)
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
   - Use `gmeasure` (not `measure`) for benchmarks
   - Ensure zero race conditions

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

- ✅ **84.8% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Panic recovery** in all critical paths
- ✅ **Memory-safe** with proper resource cleanup

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Configurable Panic Handling**: Allow users to provide custom panic handlers instead of automatic recovery
2. **Metrics Export**: Optional integration with Prometheus or other metrics systems
3. **Dynamic Buffer Resizing**: Automatic buffer size adjustment based on runtime metrics
4. **Write Batching**: Optional batching of multiple small writes into larger chunks for efficiency

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/aggregator)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, buffer sizing formulas, and performance considerations. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/runner/startStop](https://pkg.go.dev/github.com/nabbar/golib/runner/startStop)** - Lifecycle management interface implemented by the aggregator. Provides standardized Start/Stop/Restart operations with state tracking and error handling. Used for controlled service lifecycle management.

- **[github.com/nabbar/golib/atomic](https://pkg.go.dev/github.com/nabbar/golib/atomic)** - Thread-safe atomic value storage used internally for context and logger management. Provides lock-free atomic operations for better performance in concurrent scenarios.

- **[github.com/nabbar/golib/semaphore](https://pkg.go.dev/github.com/nabbar/golib/semaphore)** - Concurrency control mechanism used for limiting parallel async function executions. Prevents resource exhaustion when AsyncMax is configured.

- **[github.com/nabbar/golib/socket/server](https://pkg.go.dev/github.com/nabbar/golib/socket/server)** - Socket server implementation that commonly uses aggregator for thread-safe logging and data collection from multiple client connections. Real-world use case example.

### External References

- **[Go Concurrency Patterns: Pipelines](https://go.dev/blog/pipelines)** - Official Go blog article explaining pipeline patterns and fan-in/fan-out techniques. Relevant for understanding how the aggregator implements the fan-in pattern to merge multiple write streams.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for concurrency, error handling, and interface design. The aggregator follows these conventions for idiomatic Go code.

- **[Context Package](https://pkg.go.dev/context)** - Standard library documentation for context.Context. The aggregator fully implements this interface for cancellation propagation and deadline management in concurrent operations.

- **[Go Memory Model](https://go.dev/ref/mem)** - Official specification of Go's memory consistency guarantees. Essential for understanding the thread-safety guarantees provided by atomic operations and channels used in the aggregator.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/ioutils/aggregator`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
