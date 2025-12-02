# BufferReadCloser Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Lightweight I/O wrappers that add `io.Closer` support to standard Go buffer types with automatic resource cleanup and custom close callbacks.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Wrapper Behavior](#wrapper-behavior)
  - [Data Flow](#data-flow)
  - [Important Considerations](#important-considerations)
- [Performance](#performance)
  - [Benchmark Results](#benchmark-results)
  - [Memory Usage](#memory-usage)
  - [Throughput](#throughput)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [File Reading with Automatic Cleanup](#file-reading-with-automatic-cleanup)
  - [Writer with Auto-Flush](#writer-with-auto-flush)
  - [Network Connection Management](#network-connection-management)
  - [Buffer Pool Integration](#buffer-pool-integration)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Interfaces](#interfaces)
  - [Constructors](#constructors)
  - [Configuration](#configuration)
  - [Error Handling](#error-handling)
  - [Monitoring](#monitoring)
- [Contributing](#contributing)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

### Design Philosophy

This package extends Go's standard buffered I/O types (`bytes.Buffer`, `bufio.Reader`, `bufio.Writer`, `bufio.ReadWriter`) by adding `io.Closer` support, enabling automatic resource cleanup and custom close callbacks.

**Core Principles:**

1. **Minimal Overhead**: Thin wrappers with zero-copy passthrough to underlying buffers
2. **Lifecycle Management**: Automatic reset and cleanup on close
3. **Flexibility**: Optional custom close functions for additional cleanup logic
4. **Standard Compatibility**: Implements all relevant `io.*` interfaces
5. **Defensive Programming**: Handles nil parameters gracefully with sensible defaults

### Key Features

- **Buffer Wrapper**: `bytes.Buffer` + `io.Closer` with automatic reset
- **Reader Wrapper**: `bufio.Reader` + `io.Closer` with resource release
- **Writer Wrapper**: `bufio.Writer` + `io.Closer` with auto-flush and error propagation
- **ReadWriter Wrapper**: `bufio.ReadWriter` + `io.Closer` for bidirectional I/O
- **Custom Close Callbacks**: Optional `FuncClose` for chaining cleanup operations
- **Full Interface Support**: All standard `io.*` interfaces preserved
- **100% Test Coverage**: 69 specs + 23 benchmarks + 13 examples
- **Zero Dependencies**: Only standard library
- **Thread-Safe Patterns**: Documented concurrent usage with external synchronization

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│          bufferReadCloser Package               │
└─────────────────┬───────────────────────────────┘
                  │
     ┌────────────┼────────────┬─────────────┐
     │            │            │             │
┌────▼─────┐ ┌───▼────┐ ┌────▼─────┐ ┌─────▼────────┐
│  Buffer  │ │ Reader │ │  Writer  │ │ ReadWriter   │
├──────────┤ ├────────┤ ├──────────┤ ├──────────────┤
│bytes.    │ │bufio.  │ │bufio.    │ │bufio.        │
│Buffer    │ │Reader  │ │Writer    │ │ReadWriter    │
│    +     │ │   +    │ │    +     │ │      +       │
│io.Closer │ │io.     │ │io.Closer │ │io.Closer     │
│          │ │Closer  │ │          │ │              │
└──────────┘ └────────┘ └──────────┘ └──────────────┘
```

### Wrapper Behavior

| Wrapper | Underlying Type | On Close | Nil Handling |
|---------|----------------|----------|--------------|
| **Buffer** | `bytes.Buffer` | Reset + custom close | Creates empty buffer |
| **Reader** | `bufio.Reader` | Reset + custom close | Returns EOF immediately |
| **Writer** | `bufio.Writer` | Flush + Reset + custom close | Writes to `io.Discard` |
| **ReadWriter** | `bufio.ReadWriter` | Flush + custom close (no reset*) | Empty source + `io.Discard` |

\* *ReadWriter cannot call Reset() due to ambiguous methods in `bufio.ReadWriter`*

### Data Flow

```
User Code
    ↓
Wrapper (Buffer/Reader/Writer/ReadWriter)
    ↓
Underlying stdlib type (bytes.Buffer/bufio.*)
    ↓
Actual I/O destination/source

On Close():
    1. Flush buffered data (Writer/ReadWriter)
    2. Reset buffer (Buffer/Reader/Writer only)
    3. Execute custom FuncClose if provided
    4. Return any error
```

### Important Considerations

**Writer Flush Error Handling**: Since the correction of the flush error handling bug, `Writer.Close()` and `ReadWriter.Close()` now properly return flush errors. Always check the error:

```go
writer := bufferReadCloser.NewWriter(bw, nil)
if err := writer.Close(); err != nil {
    // Handle flush error - data may not have been written
    log.Printf("flush failed: %v", err)
}
```

**Thread Safety**: Like stdlib buffers, these wrappers are NOT thread-safe. Use external synchronization (mutex, channels) for concurrent access.

**ReadWriter Limitation**: Cannot reset on close due to ambiguous `Reset()` methods in `bufio.ReadWriter`. Only flush is performed.

---

## Performance

### Benchmark Results

Measured on AMD Ryzen 9 7900X3D with Go 1.25:

| Operation | Wrapper | stdlib | Overhead | Allocations |
|-----------|---------|--------|----------|-------------|
| Buffer.Read | 31.53 ns/op | 27.74 ns/op | +14% | 0 B/op |
| Buffer.Write | 29.69 ns/op | 29.47 ns/op | +1% | 0 B/op |
| Reader.Read | 1013 ns/op | 1026 ns/op | -1% | 4144 B/op |
| Writer.Write | 1204 ns/op | 1263 ns/op | -5% | 5168 B/op |
| Close (no func) | 2.47 ns/op | N/A | N/A | 0 B/op |
| Close (with func) | 6.90 ns/op | N/A | N/A | 0 B/op |

### Memory Usage

- **Wrapper Overhead**: 24 bytes per instance (2 pointers)
- **Zero Additional Buffering**: Uses existing stdlib buffers
- **Allocation Pattern**: Single allocation for wrapper struct
- **GC Pressure**: Minimal - only wrapper allocation

### Throughput

Large data transfers (1MB):
- **Buffer**: 2,598 MB/s
- **Reader/Writer**: ~2,500 MB/s
- **Identical to stdlib** for practical purposes

### Scalability

- **O(1) memory overhead** regardless of data size
- **Linear performance** with data size
- **No contention** (not thread-safe by design)
- **Suitable for high-throughput** applications

---

## Use Cases

### 1. File Processing with Automatic Cleanup

**Scenario**: Reading configuration files where you want both the buffer and file handle closed together.

**Advantages**:
- Single `defer` statement handles both buffer and file cleanup
- No risk of forgetting to close the file
- Automatic buffer reset prevents memory leaks

**How it's suited**: The custom close function chains file closure with buffer cleanup, ensuring proper resource management even on error paths.

### 2. Network Protocol Implementation

**Scenario**: Implementing request/response protocols over TCP connections with buffered I/O.

**Advantages**:
- Automatic flush on close ensures all data is sent
- Connection tracking via custom close callbacks
- Simplified error handling with single close point

**How it's suited**: ReadWriter wrapper provides bidirectional buffering with guaranteed flush, perfect for protocols requiring both reading and writing.

### 3. Buffer Pool Management

**Scenario**: High-performance applications using `sync.Pool` for buffer reuse.

**Advantages**:
- Automatic reset before returning to pool
- Prevents buffer state leakage between uses
- Simplified pool integration

**How it's suited**: Custom close function returns buffer to pool after reset, eliminating manual cleanup code.

### 4. Testing and Mocking

**Scenario**: Unit tests requiring lifecycle tracking of I/O operations.

**Advantages**:
- Easy verification of close calls
- Trackable resource cleanup
- Simplified test setup

**How it's suited**: Custom close callbacks enable precise tracking of when and how resources are released.

### 5. Middleware and Logging

**Scenario**: Adding cross-cutting concerns (logging, metrics, tracing) to I/O operations.

**Advantages**:
- Composable wrappers
- Non-invasive instrumentation
- Centralized cleanup logic

**How it's suited**: Custom close functions provide hook points for logging and metrics without modifying core logic.

### 6. Temporary Data Processing

**Scenario**: Processing data streams where intermediate buffers should be automatically cleaned up.

**Advantages**:
- Guaranteed cleanup even on panic (with defer)
- Memory efficiency through automatic reset
- Simplified error paths

**How it's suited**: Automatic reset on close prevents memory accumulation in long-running processes.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/bufferReadCloser
```

### Basic Usage

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

func main() {
    // Create a closeable buffer
    buf := bytes.NewBufferString("Hello, World!")
    wrapped := bufferReadCloser.NewBuffer(buf, nil)
    defer wrapped.Close() // Automatic cleanup
    
    // Use like normal buffer
    data := make([]byte, 5)
    n, _ := wrapped.Read(data)
    fmt.Printf("Read %d bytes: %s\n", n, string(data))
}
```

### File Reading with Automatic Cleanup

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

func readFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    
    // Wrap with custom close that closes the file
    reader := bufferReadCloser.NewReader(
        bufio.NewReader(file),
        file.Close, // Chains file close
    )
    defer reader.Close() // Closes both reader and file
    
    // Read data
    data := make([]byte, 1024)
    n, _ := reader.Read(data)
    fmt.Printf("Read %d bytes\n", n)
    
    return nil
}
```

### Writer with Auto-Flush

```go
package main

import (
    "bufio"
    "bytes"
    "fmt"
    "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

func main() {
    dest := &bytes.Buffer{}
    bw := bufio.NewWriter(dest)
    
    writer := bufferReadCloser.NewWriter(bw, nil)
    defer func() {
        if err := writer.Close(); err != nil {
            fmt.Printf("Flush error: %v\n", err)
        }
    }()
    
    // Write data (buffered)
    writer.WriteString("buffered data")
    // Data not yet in dest
    
    // Close flushes automatically
    writer.Close()
    fmt.Println(dest.String()) // "buffered data"
}
```

### Network Connection Management

```go
package main

import (
    "bufio"
    "net"
    "sync/atomic"
    "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

var activeConnections int64

func handleConnection(conn net.Conn) {
    rw := bufio.NewReadWriter(
        bufio.NewReader(conn),
        bufio.NewWriter(conn),
    )
    
    atomic.AddInt64(&activeConnections, 1)
    
    wrapper := bufferReadCloser.NewReadWriter(rw, func() error {
        atomic.AddInt64(&activeConnections, -1)
        return conn.Close()
    })
    defer wrapper.Close() // Auto-flush, metrics, close
    
    // Handle protocol
    // ...
}
```

### Buffer Pool Integration

```go
package main

import (
    "bytes"
    "sync"
    "github.com/nabbar/golib/ioutils/bufferReadCloser"
)

var bufferPool = sync.Pool{
    New: func() interface{} {
        return bytes.NewBuffer(make([]byte, 0, 4096))
    },
}

func processData(data []byte) {
    // Get buffer from pool
    buf := bufferPool.Get().(*bytes.Buffer)
    
    // Wrap with custom close that returns to pool
    wrapped := bufferReadCloser.NewBuffer(buf, func() error {
        bufferPool.Put(buf)
        return nil
    })
    defer wrapped.Close() // Resets and returns to pool
    
    // Use buffer
    wrapped.Write(data)
    // Process...
}
```

---

## Best Practices

### 1. Always Check Close Errors

Since flush errors are now properly returned, always check the error from `Close()`:

```go
writer := bufferReadCloser.NewWriter(bw, nil)
if err := writer.Close(); err != nil {
    return fmt.Errorf("failed to flush: %w", err)
}
```

### 2. Use defer for Guaranteed Cleanup

```go
reader := bufferReadCloser.NewReader(br, file.Close)
defer reader.Close() // Always executes, even on panic
```

### 3. Chain Resource Cleanup

```go
file, _ := os.Open("data.txt")
reader := bufferReadCloser.NewReader(
    bufio.NewReader(file),
    file.Close, // Chains cleanup
)
defer reader.Close() // Single defer handles both
```

### 4. Handle Nil Parameters Intentionally

The package handles nil gracefully, but be aware of the behavior:
- `NewBuffer(nil, nil)` creates an empty buffer
- `NewReader(nil, nil)` returns EOF immediately
- `NewWriter(nil, nil)` writes to `io.Discard`

### 5. Use External Synchronization for Concurrent Access

```go
var mu sync.Mutex
buf := bufferReadCloser.NewBuffer(bytes.NewBuffer(nil), nil)

// Concurrent access
go func() {
    mu.Lock()
    defer mu.Unlock()
    buf.WriteString("data")
}()
```

### Testing

The package includes comprehensive tests with 100% code coverage:

- **69 Ginkgo specs**: Unit tests covering all functionality
- **23 benchmarks**: Performance validation against stdlib
- **13 examples**: Executable documentation
- **Concurrency tests**: Thread-safety patterns
- **Race detection**: All tests pass with `-race`

For detailed testing information, see [TESTING.md](TESTING.md).

Run tests:
```bash
go test -v -cover ./...
go test -race ./...
go test -bench=. -benchmem
```

---

## API Reference

### Interfaces

#### Buffer
```go
type Buffer interface {
    io.Reader
    io.ReaderFrom
    io.ByteReader
    io.RuneReader
    io.Writer
    io.WriterTo
    io.ByteWriter
    io.StringWriter
    io.Closer
}
```

Wraps `bytes.Buffer` with automatic reset on close.

#### Reader
```go
type Reader interface {
    io.Reader
    io.WriterTo
    io.Closer
}
```

Wraps `bufio.Reader` with automatic reset on close.

#### Writer
```go
type Writer interface {
    io.Writer
    io.StringWriter
    io.ReaderFrom
    io.Closer
}
```

Wraps `bufio.Writer` with automatic flush and reset on close.

#### ReadWriter
```go
type ReadWriter interface {
    Reader
    Writer
}
```

Wraps `bufio.ReadWriter` with automatic flush on close (no reset due to API limitation).

### Constructors

#### NewBuffer
```go
func NewBuffer(b *bytes.Buffer, fct FuncClose) Buffer
```

Creates a Buffer wrapper. If `b` is nil, creates an empty buffer.

#### NewReader
```go
func NewReader(b *bufio.Reader, fct FuncClose) Reader
```

Creates a Reader wrapper. If `b` is nil, creates a reader from empty source.

#### NewWriter
```go
func NewWriter(b *bufio.Writer, fct FuncClose) Writer
```

Creates a Writer wrapper. If `b` is nil, creates a writer to `io.Discard`.

#### NewReadWriter
```go
func NewReadWriter(b *bufio.ReadWriter, fct FuncClose) ReadWriter
```

Creates a ReadWriter wrapper. If `b` is nil, creates a readwriter with empty source and `io.Discard` destination.

### Configuration

#### FuncClose
```go
type FuncClose func() error
```

Optional custom close function called after wrapper's internal cleanup. Use for:
- Closing underlying resources (files, connections)
- Returning buffers to pools
- Updating metrics or logging
- Releasing external resources

### Error Handling

**Close Errors**: `Close()` returns errors from:
1. Flush operations (Writer, ReadWriter)
2. Custom close functions (FuncClose)

**Best Practice**: Always check and handle close errors:
```go
if err := wrapper.Close(); err != nil {
    // Handle error - data may not have been written/flushed
}
```

**Nil Parameters**: Handled gracefully with sensible defaults (no errors).

### Monitoring

Track wrapper usage with custom close functions:

```go
var (
    activeWrappers int64
    totalClosed    int64
)

wrapper := bufferReadCloser.NewBuffer(buf, func() error {
    atomic.AddInt64(&totalClosed, 1)
    atomic.AddInt64(&activeWrappers, -1)
    return nil
})
atomic.AddInt64(&activeWrappers, 1)
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Code Contributions

**AI Usage Policy**: AI assistance should be **limited to**:
- ✅ Testing (writing tests, test cases, test data)
- ✅ Debugging (identifying bugs, suggesting fixes)
- ✅ Documentation (comments, README, examples)

**AI must NEVER be used to**:
- ❌ Generate package implementation code
- ❌ Design core functionality
- ❌ Make architectural decisions

All core package code must be human-designed and validated.

### Development Guidelines

- Maintain 100% test coverage
- All tests must pass with `-race`
- Follow existing code style
- Add GoDoc comments for all public elements
- Update documentation for new features
- Include examples for new functionality

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test -v -race -cover ./...`
5. Update documentation
6. Submit pull request with clear description

---

## Resources

### Documentation

- **[GoDoc Reference](https://pkg.go.dev/github.com/nabbar/golib/ioutils/bufferReadCloser)**: Complete API documentation with examples
- **[TESTING.md](TESTING.md)**: Detailed testing guide, coverage analysis, and test execution instructions
- **[Go bufio Package](https://pkg.go.dev/bufio)**: Standard library buffered I/O documentation
- **[Go bytes Package](https://pkg.go.dev/bytes)**: Standard library bytes buffer documentation
- **[Go io Package](https://pkg.go.dev/io)**: Standard library I/O interfaces

### Related Packages

- **[github.com/nabbar/golib/ioutils](../)**: Parent package with additional I/O utilities

### Community

- **[GitHub Repository](https://github.com/nabbar/golib)**: Source code, issues, and discussions
- **[GitHub Issues](https://github.com/nabbar/golib/issues)**: Bug reports and feature requests

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2020 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/ioutils/bufferReadCloser`  
**Minimum Go Version**: 1.18+
