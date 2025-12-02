# IOUtils Multi

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-81.7%25-brightgreen)](TESTING.md)

Thread-safe I/O multiplexer for broadcasting writes to multiple destinations with single input source management, atomic operations, and zero-allocation streaming.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Type Safety](#type-safety)
  - [Concurrency Model](#concurrency-model)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Broadcasting](#basic-broadcasting)
  - [Input and Copy](#input-and-copy)
  - [Dynamic Writer Management](#dynamic-writer-management)
  - [Thread-Safe Concurrent Operations](#thread-safe-concurrent-operations)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Multi Interface](#multi-interface)
  - [DiscardCloser](#discardcloser)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **multi** package provides a production-ready, thread-safe I/O multiplexer that enables broadcasting write operations to multiple destinations while managing a single input source. It's designed for scenarios where data needs to be duplicated to multiple outputs (logging, monitoring, data pipelines) with concurrent access from multiple goroutines.

### Design Philosophy

1. **Thread Safety First**: All operations use atomic primitives (`atomic.Value`, `atomic.Int64`) and `sync.Map` for safe concurrent access without external synchronization
2. **Type Safety**: Wrapper types (`readerWrapper`) ensure `atomic.Value` never stores inconsistent types, preventing panics
3. **Zero Allocation**: Steady-state read/write operations avoid heap allocations for optimal performance
4. **Standard Interfaces**: Implements `io.ReadWriteCloser` and `io.StringWriter` for seamless integration with Go's I/O ecosystem
5. **Predictable Behavior**: Explicit control over writer management and input source lifecycle

### Key Features

- ✅ **Broadcast Writes**: Automatically duplicate writes to all registered destinations via `io.MultiWriter`
- ✅ **Dynamic Writers**: Add and remove write destinations on-the-fly with `AddWriter()` and `Clean()`
- ✅ **Single Input Source**: Manage one input reader with thread-safe replacement via `SetInput()`
- ✅ **Thread-Safe Operations**: Zero data races verified with `go test -race` across all concurrent scenarios
- ✅ **Atomic Operations**: Lock-free reads of reader/writer with `atomic.Value`
- ✅ **Memory Efficient**: Constant memory usage, no allocations in write path
- ✅ **Standard Compliance**: Implements `io.ReadWriteCloser`, `io.StringWriter`
- ✅ **Default Safety**: Initialized with `DiscardCloser` to prevent nil panics

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Multi Interface                         │
│        (io.ReadWriteCloser + io.StringWriter)                │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                  mlt (implementation)                        │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ i: *atomic.Value                                     │   │
│  │    └─ readerWrapper{io.ReadCloser}                   │   │
│  │       └─ Type consistency for atomic stores          │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ d: *atomic.Value                                     │   │
│  │    └─ io.Writer (from io.MultiWriter)                │   │
│  │       └─ Always MultiWriter (even single/discard)    │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │ w: sync.Map                                          │   │
│  │    └─ map[int64]io.Writer (writer registry)          │   │
│  │                                                       │   │
│  │ c: *atomic.Int64                                     │   │
│  │    └─ Writer key generator (lock-free counter)       │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘

Write Flow:
  Goroutine 1 ─┐
  Goroutine 2 ─┼─> Write() ──> atomic.Value.Load() ──> io.MultiWriter
  Goroutine N ─┘                                         │
                                                          ├─> Writer 1
                                                          ├─> Writer 2
                                                          └─> Writer N

Read Flow:
  Read() ──> atomic.Value.Load() ──> readerWrapper ──> io.ReadCloser
```

### Type Safety

The package uses wrapper types to maintain type consistency in `atomic.Value`:

**Input Path** (prevents `atomic.Value` panic):
```
io.ReadCloser → readerWrapper → atomic.Value.Store()
                └─ Same concrete type for all stores
```

**Output Path** (prevents `atomic.Value` panic):
```
[]io.Writer → io.MultiWriter(writers...) → atomic.Value.Store()
              └─ Always returns io.Writer (even for single writer or io.Discard)
```

| Component | Type | Purpose | Thread-Safe |
|-----------|------|---------|-------------|
| **`readerWrapper`** | Wrapper struct | Wraps `io.ReadCloser` to ensure consistent type in `atomic.Value` | ✅ Atomic |
| **`io.MultiWriter`** | `io.Writer` | Always used for output, even with single writer | ✅ Always |
| **`sync.Map`** | Key-value store | Thread-safe writer registry with unique keys | ✅ Built-in |
| **`atomic.Int64`** | Counter | Lock-free key generation for writers | ✅ Atomic |

### Concurrency Model

**Thread-Safe Operations:**
- ✅ `Write()` - Concurrent writes are safe (io.MultiWriter is thread-safe)
- ✅ `WriteString()` - Same as Write()
- ✅ `Read()` - Safe accessor (underlying reader may have its own requirements)
- ✅ `AddWriter()` - Concurrent additions are safe
- ✅ `Clean()` - Concurrent cleaning is safe
- ✅ `SetInput()` - Concurrent replacement is safe
- ✅ `Reader()`, `Writer()`, `Close()` - All safe for concurrent access

**Important Notes:**
- ⚠️ The `Multi` wrapper is thread-safe, but underlying `io.ReadCloser` may not support concurrent reads
- ⚠️ Use one `Multi` instance per goroutine for reading, or synchronize externally
- ✅ Multiple goroutines can safely write to the same `Multi` instance

---

## Performance

### Benchmarks

Performance measurements from test suite (AMD64, Go 1.21+, with race detector):

| Operation | Median | Mean | Max | Notes |
|-----------|--------|------|-----|-------|
| **Constructor** | N/A | N/A | <100µs | Initializes structures |
| **Write (single)** | N/A | N/A | <10µs | Zero allocations |
| **Write (3 writers)** | N/A | N/A | <50µs | io.MultiWriter overhead |
| **WriteString** | N/A | N/A | <10µs | Optimized path |
| **Read** | N/A | N/A | <10µs | Delegated to wrapper |
| **AddWriter** | N/A | N/A | <50µs | Rebuilds MultiWriter |
| **Clean** | N/A | N/A | <50µs | Map iteration + cleanup |
| **SetInput** | N/A | N/A | <10µs | Atomic store |
| **Copy (1KB)** | N/A | N/A | <100µs | io.Copy overhead |
| **Copy (1MB)** | N/A | N/A | ~400µs | ~2.5 GB/s throughput |

### Memory Usage

**Base Overhead:**
```
Multi instance:       ~100 bytes (struct + atomics)
Reader wrapper:       ~24 bytes (wrapper overhead)
Writer registry:      sync.Map overhead (~48 bytes + entries)
Total (empty):        ~200 bytes
```

**Per-Writer Overhead:**
```
Single writer:        +0 bytes (uses io.MultiWriter but no extra data)
Multiple writers:     +24 bytes per writer (map entry)
```

**Scaling Example:**
```
10 writers:           ~200 + (10 × 24) = ~440 bytes
100 writers:          ~200 + (100 × 24) = ~2.6 KB
1000 writers:         ~200 + (1000 × 24) = ~24 KB
```

**Memory Characteristics:**
- ✅ O(1) memory per write operation (zero allocations)
- ✅ O(n) memory for n registered writers (constant per writer)
- ✅ No buffering (data flows directly through)
- ✅ No intermediate copies

### Scalability

**Concurrent Writers Tested:**
- ✅ 10 concurrent goroutines: Zero races
- ✅ 100 concurrent goroutines: Zero races
- ✅ 1000 concurrent goroutines: Zero races (stress test)

**Writer Count Tested:**
- ✅ 1 writer: Optimal performance
- ✅ 10 writers: <10% overhead
- ✅ 100 writers: Linear scaling
- ✅ 1000 writers: Linear scaling (map iteration)

**Data Volume Tested:**
- ✅ Small writes (1-100 bytes): Optimal
- ✅ Medium writes (1-100 KB): Good performance
- ✅ Large writes (1-10 MB): Linear with data size
- ✅ Streaming (GB+): Constant memory

---

## Use Cases

### 1. Multi-Destination Logging

**Problem**: Write logs to file, stdout, and syslog simultaneously.

```go
logFile, _ := os.Create("app.log")
defer logFile.Close()

m := multi.New()
m.AddWriter(os.Stdout, logFile, syslogWriter)

m.WriteString("[INFO] Application started\n")
// Written to all three destinations
```

**Real-world**: Used in production servers for simultaneous console and file logging.

### 2. Data Backup and Processing

**Problem**: Process data while simultaneously backing it up.

```go
m := multi.New()

processingPipe := startProcessingPipeline()
backupFile, _ := os.Create("backup.dat")

m.AddWriter(processingPipe, backupFile)

// Data goes to both processing and backup
io.Copy(m, dataSource)
```

### 3. Monitoring and Metrics

**Problem**: Send metrics to multiple monitoring backends (Prometheus, Datadog, local file).

```go
m := multi.New()
m.AddWriter(prometheusExporter, datadogExporter, metricsFile)

// Broadcast metrics to all systems
m.Write(encodeMetric("requests.count", 42))
```

### 4. Debugging Network Traffic

**Problem**: Capture network traffic while also processing it.

```go
captureFile, _ := os.Create("traffic.pcap")
defer captureFile.Close()

m := multi.New()
m.AddWriter(networkProcessor, captureFile)

// Traffic flows to processor and capture file
io.Copy(m, networkConnection)
```

### 5. Stream Replication

**Problem**: Replicate streaming data to multiple consumers.

```go
m := multi.New()

var consumer1, consumer2, consumer3 bytes.Buffer
m.AddWriter(&consumer1, &consumer2, &consumer3)

// Stream copied to all consumers
m.SetInput(videoStream)
m.Copy()
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/multi
```

### Basic Broadcasting

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    m := multi.New()

    // Add write destinations
    var buf1, buf2, buf3 bytes.Buffer
    m.AddWriter(&buf1, &buf2, &buf3)

    // Write once, broadcast to all
    m.Write([]byte("Hello, "))
    m.WriteString("World!")

    fmt.Println("Buffer 1:", buf1.String())
    fmt.Println("Buffer 2:", buf2.String())
    fmt.Println("Buffer 3:", buf3.String())
}
```

**Output:**
```
Buffer 1: Hello, World!
Buffer 2: Hello, World!
Buffer 3: Hello, World!
```

### Input and Copy

```go
package main

import (
    "bytes"
    "io"
    "strings"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    m := multi.New()

    // Setup outputs
    var out1, out2 bytes.Buffer
    m.AddWriter(&out1, &out2)

    // Setup input
    input := io.NopCloser(strings.NewReader("source data"))
    m.SetInput(input)

    // Copy from input to all outputs
    n, err := m.Copy()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Copied %d bytes to %d outputs\n", n, 2)
}
```

### Dynamic Writer Management

```go
package main

import (
    "bytes"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    m := multi.New()

    // Phase 1: Initial writers
    var buf1, buf2 bytes.Buffer
    m.AddWriter(&buf1, &buf2)
    m.Write([]byte("phase 1\n"))

    // Phase 2: Add more writers
    var buf3 bytes.Buffer
    m.AddWriter(&buf3)
    m.Write([]byte("phase 2\n"))

    // Phase 3: Reset and start fresh
    m.Clean()
    var buf4 bytes.Buffer
    m.AddWriter(&buf4)
    m.Write([]byte("phase 3\n"))

    // buf1, buf2: "phase 1\nphase 2\n"
    // buf3: "phase 2\n"
    // buf4: "phase 3\n"
}
```

### Thread-Safe Concurrent Operations

```go
package main

import (
    "bytes"
    "fmt"
    "sync"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    m := multi.New()

    var buf bytes.Buffer
    m.AddWriter(&buf)

    var wg sync.WaitGroup

    // 100 concurrent writes (thread-safe)
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            m.WriteString(fmt.Sprintf("msg%d ", id))
        }(i)
    }

    wg.Wait()
    fmt.Printf("Wrote %d bytes from 100 goroutines\n", buf.Len())
}
```

---

## Best Practices

### Resource Management

**Always close resources:**
```go
// ✅ Good
func process(input io.ReadCloser) error {
    m := multi.New()
    defer m.Close() // Closes input

    var output bytes.Buffer
    m.AddWriter(&output)
    m.SetInput(input)

    _, err := m.Copy()
    return err
}

// ❌ Bad: Resource leak
func processBad(input io.ReadCloser) {
    m := multi.New()
    m.SetInput(input)
    m.Copy() // Input never closed
}
```

### Error Handling

**Check all errors:**
```go
// ✅ Good
n, err := m.Write(data)
if err != nil {
    return fmt.Errorf("write failed: %w", err)
}
if n != len(data) {
    return fmt.Errorf("short write: %d/%d", n, len(data))
}

// ❌ Bad: Silent failures
m.Write(data) // Ignoring errors
```

### Writer Management

**Clean before switching:**
```go
// ✅ Good
m.Clean() // Remove old writers
m.AddWriter(newWriter1, newWriter2)

// ❌ Bad: Writers accumulate
m.AddWriter(newWriter1, newWriter2) // Old writers still active
```

### Thread Safety

**Safe concurrent writes:**
```go
// ✅ Good: Independent goroutines writing
var wg sync.WaitGroup
for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        m.Write([]byte(fmt.Sprintf("msg%d\n", id)))
    }(i)
}
wg.Wait()

// ⚠️  Note: Underlying readers may not be thread-safe
// If input is strings.Reader, concurrent reads = data race
// Solution: One Multi per goroutine for reading
```

### Memory Efficiency

**Stream large data:**
```go
// ✅ Good: Streaming (constant memory)
func stream(src io.Reader, m multi.Multi) error {
    buf := make([]byte, 32*1024)
    for {
        n, err := src.Read(buf)
        if n > 0 {
            if _, wErr := m.Write(buf[:n]); wErr != nil {
                return wErr
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
    }
    return nil
}

// ❌ Bad: Load entire file (high memory)
data, _ := os.ReadFile(path)
m.Write(data)
```

---

## API Reference

### Multi Interface

```go
type Multi interface {
    io.ReadWriteCloser
    io.StringWriter
    
    Clean()
    AddWriter(w ...io.Writer)
    SetInput(i io.ReadCloser)
    Reader() io.ReadCloser
    Writer() io.Writer
    Copy() (n int64, err error)
}
```

**Methods:**

- **`Write(p []byte) (n int, err error)`** - Write data to all registered writers (from `io.Writer`)
- **`WriteString(s string) (n int, err error)`** - Write string to all registered writers (from `io.StringWriter`)
- **`Read(p []byte) (n int, err error)`** - Read from input source (from `io.Reader`)
- **`Close() error`** - Close input source (from `io.Closer`)
- **`Clean()`** - Remove all registered writers and reset to `io.Discard`
- **`AddWriter(w ...io.Writer)`** - Add one or more write destinations (nil writers skipped)
- **`SetInput(i io.ReadCloser)`** - Set or replace input source (nil becomes `DiscardCloser`)
- **`Reader() io.ReadCloser`** - Get current input source
- **`Writer() io.Writer`** - Get current output writer (MultiWriter)
- **`Copy() (n int64, err error)`** - Copy from input to all outputs

**Constructor:**

```go
func New() Multi
```

Creates a new Multi instance initialized with safe defaults.

### DiscardCloser

```go
type DiscardCloser struct{}
```

No-op implementation of `io.ReadWriteCloser` used as default input source.

**Methods:**
- **`Read(p []byte) (n int, err error)`** - Always returns `(0, nil)`
- **`Write(p []byte) (n int, err error)`** - Always returns `(len(p), nil)`
- **`Close() error`** - Always returns `nil`

### Error Codes

```go
var ErrInstance = fmt.Errorf("invalid instance")
```

Returned when operations are attempted on invalid or corrupted internal state.

**When it occurs:**
- Internal `atomic.Value` contains unexpected types
- Internal state corruption (extremely rare)

**Typical causes:**
- Should not occur during normal usage when using `New()`
- May indicate a bug in the package

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

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **81.7% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Type-safe** atomic.Value usage with wrappers

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**Performance Optimizations:**
1. Writer pooling for temporary destinations
2. Batch write operations (write multiple at once)
3. Async buffered writes with configurable buffer sizes
4. Custom io.MultiWriter with performance optimizations

**Feature Additions:**
1. Writer filtering by predicate (conditional broadcast)
2. Writer priority/ordering control
3. Writer groups/tags for selective operations
4. Statistics tracking (bytes written, error counts per writer)
5. Writer health checking and auto-removal of failed writers

**API Extensions:**
1. Context integration for cancellation
2. Error callback hooks for write failures
3. Operation metrics (throughput, latency per writer)
4. Debug/trace modes for troubleshooting

**Quality of Life:**
1. Helper functions for common patterns (tee, duplicate, mirror)
2. Integration with logging frameworks (logrus, zap, zerolog)
3. Network multicast support for UDP/TCP
4. Cloud storage backends (S3, GCS, Azure Blob)

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/multi)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, type safety mechanisms, thread-safe operations, and best practices for production use. Provides detailed explanations of internal mechanisms and atomic operation guarantees.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (81.7%), performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Standard Library References

- **[io Package](https://pkg.go.dev/io)** - Standard I/O interfaces implemented by `multi`. The package fully implements `io.ReadWriteCloser` and `io.StringWriter` for seamless integration with Go's I/O ecosystem.

- **[io.MultiWriter](https://pkg.go.dev/io#MultiWriter)** - Core mechanism used for broadcasting writes. Understanding MultiWriter helps in choosing the right tool for the task and understanding performance characteristics.

- **[sync/atomic](https://pkg.go.dev/sync/atomic)** - Atomic operations used for thread-safe access. The package uses `atomic.Value` and `atomic.Int64` to avoid locks on hot paths.

- **[sync.Map](https://pkg.go.dev/sync#Map)** - Thread-safe map implementation used for writer registry. Provides concurrent-safe operations without explicit locking.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and concurrency. The `multi` package follows these conventions for idiomatic Go code.

- **[Go Concurrency Patterns](https://go.dev/blog/pipelines)** - Official Go blog article explaining pipeline patterns and concurrent data processing. Relevant for understanding how `multi` fits into larger data processing pipelines.

- **[Go Memory Model](https://go.dev/ref/mem)** - Official specification of Go's memory consistency guarantees. Essential for understanding the thread-safety guarantees provided by atomic operations used in the package.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the `multi` package. Check existing issues before creating new ones.

- **[Contributing Guide](../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements, testing procedures, and pull request process.

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
