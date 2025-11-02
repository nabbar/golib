# Multi Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-112%20Passed-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-81.7%25-brightgreen)]()

Thread-safe I/O multiplexer for broadcasting writes to multiple destinations with atomic operations and zero-allocation streaming.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The `multi` package provides a production-ready I/O multiplexer that enables broadcasting write operations to multiple destinations while managing a single input source. It emphasizes thread-safe concurrent access, atomic operations, and seamless integration with Go's standard `io` interfaces.

### Design Philosophy

1. **Thread-Safe**: All operations use atomic primitives and synchronization for safe concurrent access
2. **Zero-Allocation**: Steady-state operations avoid heap allocations for optimal performance
3. **Streaming-First**: Built on `io.Reader`/`io.Writer` for continuous data flow
4. **Type-Safe Atomics**: Consistent type wrappers prevent `atomic.Value` panics
5. **Standard Interfaces**: Implements `io.ReadWriteCloser` and `io.StringWriter`

---

## Key Features

- **Broadcast Writes**: Automatically duplicate writes to all registered destinations via `io.MultiWriter`
- **Thread-Safe Operations**: Atomic operations (`atomic.Value`, `atomic.Int64`) and `sync.Map` for concurrent access
- **Dynamic Writers**: Add and remove write destinations on-the-fly without disrupting operations
- **Input Management**: Single input source with thread-safe replacement
- **Zero Data Races**: Verified with `go test -race` across all concurrent scenarios
- **Memory Efficient**: Constant memory usage with no allocations in write path
- **Standard Interfaces**: Drop-in replacement for `io.ReadWriteCloser`

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils/multi
```

---

## Architecture

### Package Structure

```
ioutils/multi/
├── doc.go           # Package documentation
├── interface.go     # Multi interface and New() constructor
├── model.go         # Implementation with atomic operations
├── error.go         # ErrInstance error definition
└── discard.go       # DiscardCloser no-op implementation
```

### Component Diagram

```
┌──────────────────────────────────────────┐
│           Multi Interface                │
│  (io.ReadWriteCloser + io.StringWriter)  │
└────────────┬─────────────────────────────┘
             │
             ▼
┌──────────────────────────────────────────┐
│          mlt (implementation)            │
│                                          │
│  ┌─────────────────────────────────────┐ │
│  │ i: *atomic.Value (readerWrapper)    │ │ ◄─── Input Source
│  │    └─ Wrapped io.ReadCloser         │ │
│  └─────────────────────────────────────┘ │
│                                          │
│  ┌─────────────────────────────────────┐ │
│  │ d: *atomic.Value (io.Writer)        │ │ ◄─── Output Destinations
│  │    └─ io.MultiWriter([...writers])  │ │
│  └─────────────────────────────────────┘ │
│                                          │
│  ┌─────────────────────────────────────┐ │
│  │ w: sync.Map (writer storage)        │ │ ◄─── Writer Registry
│  │ c: *atomic.Int64 (counter)          │ │
│  └─────────────────────────────────────┘ │
└──────────────────────────────────────────┘
```

### Type Safety Architecture

The package uses wrapper types to maintain consistent concrete types in `atomic.Value`:

```
Input Path:
io.ReadCloser → readerWrapper → atomic.Value.Store()
                └─ Same concrete type for all stores

Output Path:
[]io.Writer → io.MultiWriter → atomic.Value.Store()
              └─ Always io.Writer from MultiWriter (even single/discard)
```

| Component | Mechanism | Purpose |
|-----------|-----------|---------|
| **`readerWrapper`** | Wraps `io.ReadCloser` | Type consistency for input in `atomic.Value` |
| **`io.MultiWriter`** | Always used for output | Type consistency for destination in `atomic.Value` |
| **`sync.Map`** | Writer registry | Thread-safe writer storage with unique keys |
| **`atomic.Int64`** | Key generator | Lock-free counter for writer identifiers |

---

## Quick Start

### Basic Broadcasting

Broadcast writes to multiple destinations:

```go
package main

import (
    "bytes"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    m := multi.New()

    // Add multiple write destinations
    var buf1, buf2, buf3 bytes.Buffer
    m.AddWriter(&buf1, &buf2, &buf3)

    // Write once, broadcast to all
    m.Write([]byte("broadcast data"))
    m.WriteString(" more data")

    // All buffers now contain: "broadcast data more data"
}
```

### Input and Copy Operations

Set an input source and copy to all outputs:

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

    // Configure output destinations
    var out1, out2 bytes.Buffer
    m.AddWriter(&out1, &out2)

    // Set input source
    input := io.NopCloser(strings.NewReader("source data"))
    m.SetInput(input)

    // Copy from input to all outputs
    n, err := m.Copy()
    // out1 and out2 both contain "source data"
}
```

### Dynamic Writer Management

Add and remove writers on-the-fly:

```go
package main

import (
    "bytes"
    "github.com/nabbar/golib/ioutils/multi"
)

func main() {
    m := multi.New()

    // Add initial writers
    var buf1, buf2 bytes.Buffer
    m.AddWriter(&buf1, &buf2)

    m.Write([]byte("message 1"))
    // buf1 and buf2 have "message 1"

    // Add more writers dynamically
    var buf3 bytes.Buffer
    m.AddWriter(&buf3)

    m.Write([]byte("message 2"))
    // All three buffers have "message 2"

    // Reset all writers
    m.Clean()

    m.Write([]byte("message 3"))
    // Discarded (no writers)

    // Add new writer
    var buf4 bytes.Buffer
    m.AddWriter(&buf4)
    m.Write([]byte("message 4"))
    // Only buf4 has "message 4"
}
```

### Thread-Safe Concurrent Operations

Safe concurrent writes from multiple goroutines:

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

    // Concurrent writes (thread-safe)
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            m.Write([]byte(fmt.Sprintf("msg%d ", id)))
        }(i)
    }

    // Concurrent writer additions (thread-safe)
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            var b bytes.Buffer
            m.AddWriter(&b)
        }()
    }

    wg.Wait()
    // All operations completed safely
}
```

---

## Performance

### Memory Efficiency

The package achieves **zero heap allocations** in the steady state:

- **Atomic Operations**: Direct load/store without locks on read/write path
- **Pre-allocated Structures**: Writer registry uses `sync.Map` (lock-free reads)
- **No Intermediate Buffers**: Direct passthrough to `io.MultiWriter`
- **Constant Memory**: Memory usage independent of data volume

**Example**: Stream 10GB through multiplexer using only ~100 bytes for internal state.

### Thread Safety Mechanisms

All operations are thread-safe through:

- **Atomic Values**: `atomic.Value` for reader/writer with consistent types
- **Lock-Free Reads**: `sync.Map` allows concurrent read access
- **Atomic Counter**: `atomic.Int64` for generating unique writer keys
- **Type Wrappers**: `readerWrapper` ensures `atomic.Value` type consistency

### Benchmark Results

Performance measurements from test suite:

| Operation | Mean Duration | Memory | Notes |
|-----------|--------------|--------|-------|
| **Constructor** | <100µs | O(1) | Pre-allocated structures |
| **Write (single)** | <10µs | 0 allocs | Direct passthrough |
| **Write (multiple)** | <50µs | 0 allocs | io.MultiWriter |
| **WriteString** | <10µs | 0 allocs | Optimized for StringWriter |
| **Read** | <10µs | 0 allocs | Direct delegation |
| **AddWriter** | <50µs | O(1) | sync.Map store + rebuild |
| **Clean** | <50µs | O(n) | Iterate and delete |
| **SetInput** | <10µs | O(1) | Atomic store |
| **Copy (1MB)** | ~400µs | O(1) | io.Copy throughput |

*Measured on AMD64, Go 1.21, with race detector enabled*

### Race Detection Results

```bash
# With race detector
CGO_ENABLED=1 go test -race ./...
# Result: 112 passed, 0 failed, 0 data races (1.18s)

# Without race detector  
go test ./...
# Result: 112 passed, 0 failed (0.13s)
```

**Status**: ✅ Zero data races detected across all concurrent scenarios

---

## Use Cases

This package is designed for scenarios requiring efficient I/O broadcasting:

**Log Aggregation**
- Write logs simultaneously to file, stdout, and remote syslog
- Add/remove log destinations without restarting
- Thread-safe logging from multiple goroutines

**Monitoring and Metrics**
- Broadcast metrics to multiple monitoring backends
- Stream data to both storage and real-time analytics
- Concurrent metric collection

**Data Pipelines**
- Fan-out data processing to multiple consumers
- Duplicate streams for backup and processing
- Multi-destination ETL operations

**Debugging and Development**
- Tee network traffic to capture files
- Debug HTTP requests by duplicating to logger
- Record test data while processing

**Streaming Services**
- Multicast streams to multiple clients
- Duplicate media streams for redundancy
- Real-time data distribution

---

## Best Practices

**Always Close Resources**
```go
// ✅ Good
func process(input io.ReadCloser) error {
    m := multi.New()
    defer m.Close() // Close input

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

**Handle Errors Properly**
```go
// ✅ Good
func broadcast(data []byte, m multi.Multi) error {
    n, err := m.Write(data)
    if err != nil {
        return fmt.Errorf("write failed: %w", err)
    }
    if n != len(data) {
        return fmt.Errorf("short write: %d/%d", n, len(data))
    }
    return nil
}

// ❌ Bad: Silent failures
func broadcastBad(data []byte, m multi.Multi) {
    m.Write(data) // Ignore errors
}
```

**Manage Writer Lifecycle**
```go
// ✅ Good: Clean before reuse
func switchWriters(m multi.Multi, newWriters ...io.Writer) {
    m.Clean() // Remove old writers
    m.AddWriter(newWriters...)
}

// ❌ Bad: Writers accumulate
func switchWritersBad(m multi.Multi, newWriters ...io.Writer) {
    m.AddWriter(newWriters...) // Old writers still active
}
```

**Thread-Safe Concurrent Access**
```go
// ✅ Good: Independent operations
func concurrentOps(m multi.Multi) {
    var wg sync.WaitGroup

    // Writes are thread-safe
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            m.Write([]byte(fmt.Sprintf("msg%d", id)))
        }(i)
    }

    wg.Wait()
}

// ⚠️  Note: Underlying io.ReadCloser may not be thread-safe
func concurrentReads(m multi.Multi) {
    // If input is strings.Reader, concurrent reads cause data race
    // The Multi wrapper is thread-safe, but wrapped objects may not be
}
```

**Efficient Streaming**
```go
// ✅ Good: Streaming
func stream(src io.Reader, m multi.Multi) error {
    buf := make([]byte, 32*1024) // Reasonable buffer
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

// ❌ Bad: Load entire file
func streamBad(path string, m multi.Multi) error {
    data, _ := os.ReadFile(path) // Full file in memory
    m.Write(data)
    return nil
}
```

---

## Testing

**Test Suite**: 112 specs using Ginkgo v2 and Gomega

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Results**
- Total Specs: 112 passed, 1 skipped, 0 failed
- Coverage: 81.7% of statements
- Race Detection: ✅ Zero data races
- Execution Time: ~0.13s (without race), ~1.18s (with race)

**Coverage Areas**
- Constructor and interface compliance
- Write operations (single, multiple, large data)
- Read operations and error propagation
- Copy operations and integration
- Concurrent operations (writes, AddWriter, Clean, SetInput)
- Edge cases (nil values, zero-length, state transitions)
- Error handling (ErrInstance, writer errors)
- Performance benchmarks with gmeasure

**Quality Assurance**
- ✅ Thread-safe concurrent operations
- ✅ Zero data races verified with `-race`
- ✅ Atomic operation correctness
- ✅ Type-safe atomic.Value usage

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
- Document thread-safety guarantees

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Use Ginkgo v2 BDD style for consistency

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Features**
- Writer filtering by predicate
- Writer priority/ordering control
- Conditional broadcast (write to subset)
- Writer groups/tags for selective operations
- Statistics tracking (bytes written, error counts)

**Performance**
- Writer pooling for temporary destinations
- Batch operations (write multiple at once)
- Async writes with buffering
- Writer delegation patterns

**Observability**
- Writer health checking
- Operation metrics (throughput, latency)
- Error callback hooks
- Debug/trace modes

**Integration**
- Helper functions for common patterns
- Integration with logging frameworks
- Network multicast support
- Cloud storage backends (S3, GCS, Azure)

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Package Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/multi)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Parent Package**: [ioutils](../README.md)
