# Multi I/O MultiWriter

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-80.8%25-brightgreen)](TESTING.md)

Thread-safe I/O multi-writer that broadcasts writes to multiple destinations and manages a single input source, featuring adaptive sequential/parallel write strategies based on latency monitoring.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Adaptive Strategy](#adaptive-strategy)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [Broadcasting Writes](#broadcasting-writes)
  - [Stream Copying](#stream-copying)
  - [Adaptive Configuration](#adaptive-configuration)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Multi Interface](#multi-interface)
  - [Configuration](#configuration)
  - [Statistics](#statistics)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **multi** package extends Go's standard `io.MultiWriter` with adaptive sequential/parallel execution, thread-safe dynamic writer management, and comprehensive concurrency support. While `io.MultiWriter` provides basic write fan-out, **multi** adds intelligent performance optimization, input source management, and real-time monitoring capabilities.

### Why Not Just Use io.MultiWriter?

The standard library's `io.MultiWriter` has several limitations that **multi** addresses:

**Limitations of io.MultiWriter:**
- ❌ **No thread safety**: Cannot safely add/remove writers during operation
- ❌ **No input management**: Only handles writes, no reader support
- ❌ **No adaptive strategy**: Always sequential execution, even with slow writers
- ❌ **No observability**: No metrics, statistics, or performance monitoring
- ❌ **Blocking behavior**: One slow writer blocks all writes
- ❌ **Static configuration**: Writers must be known at creation time

**How multi Extends io.MultiWriter:**
- ✅ **Thread-safe operations**: Add/remove writers atomically during execution
- ✅ **Complete I/O interface**: Manages both input (Reader) and outputs (Writers)
- ✅ **Adaptive execution**: Automatically switches between sequential and parallel modes
- ✅ **Real-time metrics**: Latency monitoring, writer counts, mode statistics
- ✅ **Non-blocking option**: Parallel mode prevents slow writers from blocking
- ✅ **Dynamic management**: Hot-swap writers without service interruption

**Internally**, multi uses `io.MultiWriter` for sequential write operations, but adds a parallel execution mode and adaptive switching based on observed latency. This gives you the efficiency of `io.MultiWriter` when appropriate, with automatic fallback to parallel execution when writers exhibit high latency.

### Design Philosophy

1.  **Extend Standard Library**: Build on `io.MultiWriter` while addressing its limitations with adaptive strategies and thread safety.
2.  **Adaptive Performance**: Dynamically optimize write strategies (Sequential via io.MultiWriter vs Parallel via goroutines) based on observed latency.
3.  **Thread Safety First**: All operations safe for concurrent use via atomic operations and concurrent maps.
4.  **Interface Compliance**: Fully implements `io.ReadWriteCloser`, `io.StringWriter`, extending beyond io.MultiWriter's write-only nature.
5.  **Zero Panic**: Uses defensive programming and safe defaults (e.g., `DiscardCloser`) to prevent nil pointer exceptions.
6.  **Observability**: Provides real-time statistics on latency, writer counts, and operational modes for production monitoring.

### Key Features

-   ✅ **Write Broadcasting**: Writes are sent to all registered writers efficiently.
-   ✅ **Input Management**: Manages a single input source with thread-safe replacement.
-   ✅ **Adaptive Execution**: Automatically switches to parallel writes if latency exceeds thresholds.
-   ✅ **Thread-Safe**: Safe for concurrent AddWriter, SetInput, Write, and Read operations.
-   ✅ **Atomic State**: Uses `atomic.Value` and typed atomic maps for lock-free state management.
-   ✅ **Safe Defaults**: Initializes with discarders to ensure immediate usability.
-   ✅ **Extensive Testing**: 80.8% coverage with race detection and extensive benchmarks.

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────────┐
│                           Multi                            │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────┐           ┌─────────────────────┐        │
│  │ Input Source │           │ Output Destinations │        │
│  │ (io.Reader)  │           │ (io.Writer Map)     │        │
│  └──────┬───────┘           └──────────┬──────────┘        │
│         │                              │                   │
│         ▼                              ▼                   │
│  ┌──────────────┐           ┌─────────────────────┐        │
│  │ ReaderWrap   │           │    WriteWrapper     │        │
│  └──────┬───────┘           └──────────┬──────────┘        │
│         │                              │                   │
│         │                   ┌──────────┴──────────┐        │
│         │                   │                     │        │
│         │           ┌───────▼──────┐      ┌───────▼──────┐ │
│         │           │ Sequential   │      │   Parallel   │ │
│         │           │ (io.Multi)   │      │ (Goroutines) │ │
│         │           └──────────────┘      └──────────────┘ │
│         │                   │                     │        │
│         ▼                   │                     │        │
│   Client Read() <─────────────────────────────────┘        │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### Data Flow

1.  **Write Operation**:
    *   Data arrives at `Write()`.
    *   Latency start time is recorded.
    *   Operation is delegated to current strategy (Sequential or Parallel).
    *   **Sequential**: Iterates writers using `io.MultiWriter`.
    *   **Parallel**: Spawns goroutines for each writer if size/count thresholds met.
    *   Latency is measured and added to atomic stats.
    *   Sampler checks if mode switch is needed (Adaptive Mode).

2.  **Read Operation**:
    *   Read is delegated to the current atomic `readerWrapper`.
    *   If no reader is set, defaults to `DiscardCloser` (returns EOF/0).

### Adaptive Strategy

The multi-writer monitors write latency to decide between sequential and parallel execution:

*   **Sampling**: Every `SampleWrite` operations (default 100).
*   **Switch to Parallel**: If `AverageLatency > ThresholdLatency` AND `WriterCount >= MinimalWriter`.
*   **Switch to Sequential**: If `AverageLatency < ThresholdLatency`.
*   **Parallel Execution**: Only triggers for writes larger than `MinimalSize` (default 512 bytes).

---

## Performance

### Benchmarks

Based on benchmark results (AMD64, Go 1.25):

| Operation | Median | Mean | Max |
|-----------|--------|------|-----|
| **Multi Creation** | 4.6µs | 5.3µs | 8.8µs |
| **SetInput** | <1µs | <1µs | 100µs |
| **Sequential Write** | 400µs | 400µs | 500µs |
| **Parallel Write** | 200µs | 233µs | 300µs |
| **Adaptive Write** | 200µs | 266µs | 500µs |
| **Copy Operation** | 200µs | 266µs | 500µs |
| **Read Operations** | 300µs | 366µs | 800µs |

*Note: Parallel writes show significant improvement (~50% latency reduction) under load.*

### Memory Usage

-   **Base Overhead**: Minimal (structs + atomic pointers).
-   **Sequential**: Zero allocation per write (uses standard `io.MultiWriter`).
-   **Parallel**: Allocates goroutines and error channels per write operation.
-   **Optimization**: Parallel mode only activates when beneficial (latency/size thresholds), minimizing overhead for small writes.

### Scalability

-   **Writers**: Tested with dynamic addition/removal of writers.
-   **Concurrency**: Thread-safe implementation allows concurrent readers/writers without locks (using `sync.Map` and `atomic.Value`).
-   **Throughput**: scales linearly with available CPU for parallel mode.

---

## Use Cases

### 1. Broadcasting Logs

Send application logs to stdout, a file, and a network socket simultaneously.

```go
m := multi.New(false, false, multi.DefaultConfig())
m.AddWriter(os.Stdout, logFile, netConn)
log.SetOutput(m)
```

### 2. Stream Replication with Monitoring

Copy an incoming stream to multiple storage backends while monitoring throughput.

```go
m := multi.New(true, false, multi.DefaultConfig()) // Adaptive mode
m.SetInput(sourceStream)
m.AddWriter(s3Uploader, localDisk, backupServer)
m.Copy() // Efficiently copies to all destinations
```

### 3. Adaptive High-Throughput Writing

For systems with variable writer latency (e.g., slow network writers mixed with fast files), adaptive mode prevents the slowest writer from blocking the main thread entirely by switching to parallel execution.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/multi
```

### Basic Usage

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    // Create new multi instance (adaptive=false, parallel=false)
    m := multi.New(false, false, multi.DefaultConfig())
    defer m.Close()

    var buf1, buf2 bytes.Buffer
    
    // Add writers
    m.AddWriter(&buf1, &buf2)
    
    // Write data
    m.Write([]byte("Hello World"))
    
    fmt.Println(buf1.String()) // "Hello World"
    fmt.Println(buf2.String()) // "Hello World"
}
```

### Broadcasting Writes

```go
m := multi.New(true, false, multi.DefaultConfig()) // Enable adaptive mode
m.AddWriter(w1, w2, w3)
m.WriteString("Broadcast message")
```

### Stream Copying

```go
m.SetInput(inputStream)
m.AddWriter(output1, output2)
n, err := m.Copy()
```

### Adaptive Configuration

```go
cfg := multi.Config{
    SampleWrite:      100,  // Check every 100 writes
    ThresholdLatency: 5000, // 5µs threshold
    MinimalWriter:    2,    // Need at least 2 writers for parallel
    MinimalSize:      1024, // Min 1KB for parallel
}
m := multi.New(true, false, cfg)
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **80.8% code coverage** and **120 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and operations (AddWriter, SetInput, Write, Read, Copy)
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks (sequential vs parallel throughput)
- ✅ Error handling and edge cases
- ✅ Adaptive mode switching logic

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Use Defaults:**
```go
// ✅ GOOD: Start with sensible defaults
cfg := multi.DefaultConfig()
m := multi.New(true, false, cfg) // Adaptive mode
```

**Close Resources:**
```go
// ✅ GOOD: Always close when done
m := multi.New(false, false, multi.DefaultConfig())
defer m.Close() // Closes input reader

// Note: Writers are NOT closed by Close()
// You must manage writer lifecycle separately
```

**Monitor Stats:**
```go
// ✅ GOOD: Monitor in production
stats := m.Stats()
log.Printf("Mode: %v, Latency: %dns, Writers: %d",
    stats.WriterMode, stats.AverageLatency, stats.WriterCounter)
```

**Thread Safety:**
```go
// ✅ GOOD: Safe for concurrent use
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        m.Write([]byte("data"))
    }()
}
wg.Wait()
```

### ❌ DON'T

**Don't use blocking writers without consideration:**
```go
// ❌ BAD: Blocking writer without buffering
slowWriter := &SlowWriter{delay: 10 * time.Second}
m.AddWriter(slowWriter) // Blocks all writes

// ✅ GOOD: Use parallel mode for slow writers
m := multi.New(false, true, cfg) // Force parallel
m.AddWriter(slowWriter)
```

**Don't force parallel for tiny writes:**
```go
// ❌ BAD: Goroutine overhead > benefit
cfg := multi.Config{MinimalSize: 10} // Too small
m := multi.New(false, true, cfg)

// ✅ GOOD: Use default thresholds
cfg := multi.DefaultConfig() // MinimalSize: 512
m := multi.New(false, true, cfg)
```

**Don't check for nil:**
```go
// ❌ BAD: Unnecessary nil check
m := multi.New(false, false, multi.DefaultConfig())
if m != nil { // Always true
    m.Write(data)
}

// ✅ GOOD: New() always returns valid instance
m := multi.New(false, false, multi.DefaultConfig())
m.Write(data)
```

---

## API Reference

### Multi Interface

```go
type Multi interface {
    io.ReadWriteCloser
    io.StringWriter

    AddWriter(w ...io.Writer)
    Clean()
    SetInput(i io.Reader)
    Stats() Stats
    IsParallel() bool
    IsSequential() bool
    IsAdaptive() bool
    Reader() io.ReadCloser
    Writer() io.Writer
    Copy() (n int64, err error)
}
```

### Configuration

```go
type Config struct {
    SampleWrite      int   // Writes between checks
    ThresholdLatency int64 // Nanoseconds
    MinimalWriter    int   // Min writers for parallel
    MinimalSize      int   // Min bytes for parallel
}
```

### Statistics

```go
type Stats struct {
    AdaptiveMode    bool
    WriterMode      bool  // true = parallel
    AverageLatency  int64 // ns
    SampleCollected int64
    WriterCounter   int
}
```

### Error Codes

```go
var (
    ErrInstance = fmt.Errorf("invalid instance")
)
```

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
   - Use `gmeasure` for performance benchmarks
   - Ensure zero race conditions with `go test -race`

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
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Adaptive strategy** for optimal performance

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Weighted Writes**: Prioritize certain writers based on importance
2. **Async Buffer**: Add buffered channels for non-blocking writes (fire-and-forget mode)
3. **Error Policy**: Configurable error handling (ignore errors vs fail fast vs retry)
4. **Writer Health Checks**: Monitor writer health and auto-disable failing writers
5. **Metrics Export**: Optional integration with Prometheus or other metrics systems

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

-   **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/multi)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, thread-safety guarantees, adaptive strategy explanation, and implementation details. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 80.8% coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/atomic](https://pkg.go.dev/github.com/nabbar/golib/atomic)** - Atomic primitives used for thread-safe state management without locks. Provides lock-free atomic operations for better performance in concurrent scenarios.

- **[github.com/nabbar/golib/ioutils/aggregator](https://pkg.go.dev/github.com/nabbar/golib/ioutils/aggregator)** - Write aggregator that can be used as a destination for the multi-writer. Serializes concurrent write operations to a single output function.

### External References

- **[io.MultiWriter](https://pkg.go.dev/io#MultiWriter)** - Go standard library's write fan-out function. The multi package uses io.MultiWriter internally for sequential write operations and extends it with adaptive strategies, thread safety, and input management.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and concurrency patterns. The multi package follows these conventions for idiomatic Go code.

- **[Go Concurrency Patterns](https://go.dev/blog/pipelines)** - Official Go blog article explaining pipeline patterns and fan-in/fan-out techniques. Relevant for understanding how the multi-writer implements concurrent I/O broadcasting and extends io.MultiWriter with parallel execution.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)
**Package**: `github.com/nabbar/golib/ioutils/multi`
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
