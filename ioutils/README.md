# IOUtils Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Tests](https://img.shields.io/badge/Tests-657%20Specs-green)]()
[![Coverage](https://img.shields.io/badge/Coverage-90.8%25-brightgreen)]()

Production-ready I/O utilities and abstractions for Go applications with streaming-first design, thread-safe operations, and comprehensive subpackages for file management, stream processing, progress tracking, and resource management.

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
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The `ioutils` package provides a comprehensive suite of I/O utilities for Go applications, organized into 8 specialized subpackages plus root-level path management utilities. Each subpackage focuses on a specific aspect of I/O operations, from delimiter-based stream processing to file descriptor limit management, all designed for production use with extensive testing and thread safety guarantees.

### Design Philosophy

1. **Streaming-First**: All operations use `io.Reader`/`io.Writer` for continuous data flow with constant memory usage
2. **Thread-Safe**: Atomic operations and proper synchronization primitives for safe concurrent access
3. **Standard Interfaces**: Full compatibility with Go's `io` package interfaces
4. **Resource Management**: Proper cleanup with context-aware closers and automatic resource lifecycle
5. **Zero-Copy Operations**: Direct passthrough where possible to minimize allocations
6. **Production-Ready**: 657 comprehensive test specs with 90.8% average coverage, zero race conditions

---

## Key Features

- **Path Management**: File and directory creation with permission validation and automatic parent creation
- **Delimiter Processing**: High-performance buffered reading for any delimiter character with 100% test coverage ([delim](#delim))
- **I/O Multiplexing**: Thread-safe broadcasting of writes to multiple destinations with atomic operations ([multi](#multi))
- **Progress Tracking**: Real-time monitoring of read/write operations with customizable callbacks ([ioprogress](#ioprogress))
- **Stream Wrappers**: Flexible I/O transformation and interception with runtime customization ([iowrapper](#iowrapper))
- **Buffer Management**: Buffered I/O with proper resource lifecycle and automatic cleanup ([bufferReadCloser](#bufferreadcloser))
- **Resource Pooling**: Context-aware management of multiple closers with automatic cleanup ([mapCloser](#mapcloser))
- **File Descriptor Limits**: Cross-platform FD limit management with safe increase operations ([fileDescriptor](#filedescriptor))
- **Testing Tools**: No-op writers for test stubbing and mock implementations ([nopwritecloser](#nopwritecloser))

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils
```

**Requirements**:
- Go 1.18 or higher
- Compatible with Linux, macOS, Windows

**Subpackages** are imported individually:
```go
import (
    "github.com/nabbar/golib/ioutils"
    "github.com/nabbar/golib/ioutils/ioprogress"
    "github.com/nabbar/golib/ioutils/iowrapper"
    // ... other subpackages as needed
)
```

---

## Architecture

### Package Structure

The `ioutils` package is organized into focused subpackages, each solving specific I/O challenges:

```
ioutils/
├── tools.go                   # Path management utilities (91.7% coverage)
├── bufferReadCloser/          # Buffered I/O with close support (100% coverage, 57 specs)
├── delim/                     # Delimiter-based buffered reading (100% coverage, 198 specs)
├── fileDescriptor/            # File descriptor limits management (85.7% coverage, 20 specs)
├── ioprogress/                # I/O progress tracking (84.7% coverage, 42 specs)
├── iowrapper/                 # I/O operation wrappers (100% coverage, 114 specs)
├── mapCloser/                 # Multiple resource management (80.2% coverage, 29 specs)
├── maxstdio/                  # Windows stdio limits (cgo, Windows-only)
├── multi/                     # I/O multiplexing to multiple writers (81.7% coverage, 112 specs)
└── nopwritecloser/            # No-op write closer for testing (100% coverage, 54 specs)
```

### Component Overview

```
┌────────────────────────────────────────────────────────────────┐
│                    Application Layer                           │
│         (Your code using standard io interfaces)               │
└───────────┬────────────────────────────────────────────────────┘
            │
            ▼
┌────────────────────────────────────────────────────────────────┐
│                     IOUtils Package                            │
│  ┌──────────────┬──────────────┬─────────────┬──────────────┐  │
│  │   delim      │     multi    │ ioprogress  │  iowrapper   │  │
│  │  Delimiter   │  Broadcast   │  Progress   │  Transform   │  │
│  │  Buffering   │   Writes     │  Tracking   │   I/O Ops    │  │
│  └──────────────┴──────────────┴─────────────┴──────────────┘  │
│  ┌──────────────┬──────────────┬─────────────┬──────────────┐  │
│  │ mapCloser    │ bufferRdCls  │fileDescrptr │nopwritecloser│  │
│  │  Resource    │  Buffered    │  FD Limits  │   Testing    │  │
│  │  Lifecycle   │   I/O        │  Management │   Stub       │  │
│  └──────────────┴──────────────┴─────────────┴──────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │       PathCheckCreate (Root) + maxstdio (Windows)        │  │
│  │      File & Directory Management + Windows FD Limits     │  │
│  └──────────────────────────────────────────────────────────┘  │
└────────────┬───────────────────────────────────────────────────┘
             │
             ▼
┌────────────────────────────────────────────────────────────────┐
│                Standard Go I/O Interfaces                      │
│        io.Reader, io.Writer, io.Closer, io.Seeker              │
└────────────────────────────────────────────────────────────────┘
```

| Subpackage | Purpose | Specs | Coverage | Thread-Safe |
|------------|---------|-------|----------|-------------|
| **Root** | Path management | 31 | 91.7% | ✅ |
| **bufferReadCloser** | Buffered I/O with lifecycle | 57 | 100% | ✅ |
| **delim** | Delimiter-based buffering | 198 | 100% | ✅ (per instance) |
| **fileDescriptor** | FD limit management | 20 | 85.7% | ✅ |
| **ioprogress** | Progress callbacks | 42 | 84.7% | ✅ |
| **iowrapper** | I/O transformation | 114 | 100% | ✅ |
| **mapCloser** | Multi-resource cleanup | 29 | 80.2% | ✅ |
| **maxstdio** | Windows FD limits | N/A | Windows-only | ✅ |
| **multi** | I/O multiplexing | 112 | 81.7% | ✅ |
| **nopwritecloser** | Testing stub | 54 | 100% | ✅ |

---

## Quick Start

### Path Management

```go
package main

import (
    "os"
    "github.com/nabbar/golib/ioutils"
)

func main() {
    // Ensure directory exists with permissions
    err := ioutils.PathCheckCreate(false, "/var/app/data", 0644, 0755)
    if err != nil {
        panic(err)
    }
    
    // Ensure file exists with permissions
    err = ioutils.PathCheckCreate(true, "/var/app/config.json", 0644, 0755)
    if err != nil {
        panic(err)
    }
}
```

### Progress Tracking

```go
package main

import (
    "fmt"
    "io"
    "os"
    "sync/atomic"
    
    "github.com/nabbar/golib/ioutils/ioprogress"
)

func main() {
    file, _ := os.Open("largefile.dat")
    defer file.Close()
    
    // Wrap with progress tracking
    reader := ioprogress.NewReadCloser(file)
    defer reader.Close()
    
    var totalBytes int64
    reader.RegisterFctIncrement(func(size int64) {
        atomic.AddInt64(&totalBytes, size)
        fmt.Printf("\rRead: %d bytes", atomic.LoadInt64(&totalBytes))
    })
    
    reader.RegisterFctEOF(func() {
        fmt.Println("\nDone!")
    })
    
    io.Copy(io.Discard, reader)
}
```

### I/O Transformation

```go
package main

import (
    "bytes"
    "strings"
    
    "github.com/nabbar/golib/ioutils/iowrapper"
)

func main() {
    data := "hello world"
    reader := strings.NewReader(data)
    
    // Wrap reader with transformation
    wrapped := iowrapper.New(reader)
    
    // Transform data to uppercase on read
    wrapped.SetRead(func(p []byte) []byte {
        return bytes.ToUpper(p)
    })
    
    buf := make([]byte, 128)
    n, _ := wrapped.Read(buf)
    
    // Output: HELLO WORLD
    println(string(buf[:n]))
}
```

### Resource Management

```go
package main

import (
    "context"
    "os"
    
    "github.com/nabbar/golib/ioutils/mapCloser"
)

func main() {
    ctx := context.Background()
    closer := mapCloser.New(ctx)
    
    // Add multiple resources
    file1, _ := os.Open("file1.txt")
    file2, _ := os.Open("file2.txt")
    file3, _ := os.Open("file3.txt")
    
    closer.Add(file1, file2, file3)
    
    // Close all at once
    defer closer.Close()
    
    // Use files...
}
```

---

## Performance

### Overhead Analysis

The package maintains minimal overhead across all subpackages:

| Operation | Overhead | Impact |
|-----------|----------|--------|
| **Path check/create** | ~100μs | File system bound |
| **Progress tracking** | <100ns/op | <0.1% for typical I/O |
| **I/O wrapper** | ~50ns/op | Negligible |
| **Buffer management** | 0 allocs | Zero GC pressure |
| **Resource closing** | O(n) | Linear with resource count |

### Test Performance

```
Package              Tests  Duration  Coverage
ioutils              31     0.03s     91.7%
bufferReadCloser     57     0.02s     100.0%
fileDescriptor       20     0.01s     90.0%
ioprogress           42     0.01s     84.7%
iowrapper            114    0.07s     100.0%
mapCloser            29     0.01s     80.2%
nopwritecloser       54     0.22s     100.0%
────────────────────────────────────────────
Total                347    0.37s     ~93%
```

**Key Findings**:
- Zero race conditions across all subpackages
- Fast test execution (<0.5s total)
- Minimal memory allocations
- Production-ready performance

---

## Use Cases

The `ioutils` package is designed for diverse I/O scenarios:

**Application Configuration**
- Ensure config files and directories exist
- Set appropriate file permissions
- Validate directory structure on startup

**File Transfer Applications**
- Track upload/download progress in real-time
- Monitor bandwidth utilization
- Provide user feedback with progress bars

**Stream Processing**
- Transform data on-the-fly during I/O
- Intercept and log I/O operations
- Implement custom compression/encryption

**Resource Management**
- Manage multiple file handles in batch operations
- Context-aware cleanup on cancellation
- Prevent resource leaks in error scenarios

**High-Concurrency Servers**
- Adjust file descriptor limits dynamically
- Monitor resource utilization
- Handle thousands of concurrent connections

**Testing & Mocking**
- Stub I/O operations for unit tests
- Simulate write operations without side effects
- Test error handling paths

**Log Management**
- Route log streams to multiple destinations
- Buffer log output efficiently with proper resource cleanup
- Manage log file rotation and archiving

**Cross-Platform Development**
- Handle platform-specific file descriptor limits
- Ensure consistent behavior across OS
- Manage Windows-specific I/O constraints

---

## Subpackages

Detailed documentation for each subpackage:

### bufferReadCloser

**Purpose**: Buffered I/O with proper resource lifecycle management

**Key Features**:
- Wraps `bytes.Buffer`, `bufio.Reader`, `bufio.Writer`, `bufio.ReadWriter`
- Implements `io.Closer` for all buffer types
- Optional custom close function
- Automatic flush on close for writers

**Use Cases**:
- In-memory buffering with explicit cleanup
- Complex I/O pipelines requiring resource management
- Testing scenarios needing controlled buffer lifecycle

**Documentation**: [bufferReadCloser/README.md](bufferReadCloser/README.md)

---

### delim

**Purpose**: High-performance buffered reading for delimiter-separated data streams

**Key Features**:
- Custom delimiter support (any rune character)
- Constant memory usage regardless of data size  
- Zero-copy operations
- 100% test coverage with 198 specs

**Use Cases**:
- CSV/TSV processing with custom separators
- Log file parsing with custom markers
- Null-terminated string handling
- Stream processing with delimiter-separated events

**Documentation**: [delim/README.md](delim/README.md)  
**Testing**: [delim/TESTING.md](delim/TESTING.md)

---

### fileDescriptor

**Purpose**: Query and manage file descriptor limits

**Key Features**:
- Get current and maximum FD limits
- Increase FD limits (up to system maximum)
- Cross-platform support (Linux/Unix, Windows)

**Use Cases**:
- High-concurrency applications
- Dynamic resource limit adjustment
- Pre-flight checks for server applications

**Documentation**: [fileDescriptor/README.md](fileDescriptor/README.md)

---

### ioprogress

**Purpose**: Real-time I/O progress tracking

**Key Features**:
- Transparent wrappers for `io.ReadCloser` and `io.WriteCloser`
- Three callback types (Increment, Reset, EOF)
- Thread-safe atomic counters
- Zero performance overhead

**Use Cases**:
- File transfer progress indicators
- Bandwidth monitoring
- ETA calculation
- Multi-stage processing feedback

**Documentation**: [ioprogress/README.md](ioprogress/README.md)  
**Testing**: [ioprogress/TESTING.md](ioprogress/TESTING.md)

---

### iowrapper

**Purpose**: Flexible I/O operation interception and transformation

**Key Features**:
- Wrap any I/O-compatible object
- Custom read/write/seek/close functions
- Runtime behavior modification
- Standard interface compliance

**Use Cases**:
- Data transformation during I/O
- Operation logging and instrumentation
- Testing and mocking
- Legacy code adaptation

**Documentation**: [iowrapper/README.md](iowrapper/README.md)  
**Testing**: [iowrapper/TESTING.md](iowrapper/TESTING.md)

---

### mapCloser

**Purpose**: Manage multiple `io.Closer` instances as a group

**Key Features**:
- Batch addition and closing
- Context-aware automatic cleanup
- Thread-safe operations
- Error aggregation

**Use Cases**:
- Multi-file operations
- Resource pooling
- Guaranteed cleanup on context cancellation
- Simplified error handling

**Documentation**: [mapCloser/README.md](mapCloser/README.md)

---

### maxstdio

**Purpose**: Windows stdio file descriptor management (cgo required)

**Key Features**:
- Get/set max stdio FD limit on Windows
- Direct C runtime integration
- Platform-specific optimization

**Use Cases**:
- Windows server applications
- High-concurrency Windows services
- Resource limit configuration

**Documentation**: [maxstdio/README.md](maxstdio/README.md)

**Note**: Windows-only, requires CGO_ENABLED=1

---

### multi

**Purpose**: Thread-safe I/O multiplexer for broadcasting writes to multiple destinations

**Key Features**:
- Broadcast writes to multiple `io.Writer` destinations
- Thread-safe with atomic operations and `sync.Map`
- Dynamic writer management (add/remove on-the-fly)
- Zero allocations in write path

**Use Cases**:
- Log aggregation to multiple outputs
- Data fan-out to multiple consumers
- Debugging and development (tee operations)
- Streaming to multiple clients

**Documentation**: [multi/README.md](multi/README.md)  
**Testing**: [multi/TESTING.md](multi/TESTING.md)

---

### nopwritecloser

**Purpose**: No-op `io.WriteCloser` implementation for testing

**Key Features**:
- Wraps any `io.Writer`
- Close() always returns nil
- Zero overhead
- Testing-friendly

**Use Cases**:
- Unit test stubbing
- API adaptation (requires closer but none needed)
- Mock implementations

**Documentation**: [nopwritecloser/README.md](nopwritecloser/README.md)  
**Testing**: [nopwritecloser/TESTING.md](nopwritecloser/TESTING.md)

---

## API Reference

### Root Package

#### PathCheckCreate

```go
func PathCheckCreate(isFile bool, path string, permFile os.FileMode, permDir os.FileMode) error
```

Ensures a file or directory exists at the given path with correct permissions.

**Parameters**:
- `isFile`: `true` for file, `false` for directory
- `path`: Target filesystem path
- `permFile`: Permissions for files (e.g., `0644`)
- `permDir`: Permissions for directories (e.g., `0755`)

**Returns**:
- `error`: Error if operation fails

**Behavior**:
- Creates missing directories in path
- Creates file if it doesn't exist
- Updates permissions if incorrect
- Returns error if file/dir type mismatch

**Example**:
```go
// Ensure log directory exists
err := ioutils.PathCheckCreate(false, "/var/log/app", 0644, 0755)

// Ensure config file exists
err := ioutils.PathCheckCreate(true, "/etc/app/config.json", 0600, 0755)
```

---

## Best Practices

### 1. Always Check Errors

```go
// ✅ Good: Check all errors
err := ioutils.PathCheckCreate(true, path, 0644, 0755)
if err != nil {
    return fmt.Errorf("path setup: %w", err)
}

// ❌ Bad: Ignore errors
_ = ioutils.PathCheckCreate(true, path, 0644, 0755)
```

### 2. Use Appropriate Permissions

```go
// ✅ Good: Restrictive permissions for sensitive data
ioutils.PathCheckCreate(true, "/etc/app/secrets.key", 0600, 0700)

// ✅ Good: Standard permissions for general files
ioutils.PathCheckCreate(true, "/var/log/app.log", 0644, 0755)

// ❌ Bad: Overly permissive
ioutils.PathCheckCreate(true, "/etc/app/secrets.key", 0777, 0777)
```

### 3. Close Resources Properly

```go
// ✅ Good: Use defer for cleanup
ctx := context.Background()
closer := mapCloser.New(ctx)
defer closer.Close()

file, _ := os.Open("data.txt")
closer.Add(file)

// ❌ Bad: Manual cleanup (easy to forget)
file1, _ := os.Open("file1.txt")
file2, _ := os.Open("file2.txt")
file1.Close()  // May be skipped on error
file2.Close()
```

### 4. Use Context for Cancellation

```go
// ✅ Good: Context-aware resource management
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

closer := mapCloser.New(ctx)
// Resources auto-close on context cancellation

// ❌ Bad: No cancellation support
closer := mapCloser.New(context.Background())
// Resources never auto-close
```

### 5. Choose Appropriate Subpackage

```go
// ✅ Good: Use specific tool for the job
// Progress tracking
reader := ioprogress.NewReadCloser(file)

// I/O transformation
wrapper := iowrapper.New(reader)

// Resource management
closer.Add(reader, wrapper)

// ❌ Bad: Reimplementing existing functionality
// Don't write custom progress tracking if ioprogress exists
```

---

## Testing

The package includes comprehensive test suites using **Ginkgo v2** and **Gomega**.

### Test Execution

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (requires CGO_ENABLED=1)
CGO_ENABLED=1 go test -race ./...

# Using Ginkgo CLI
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo -r -cover
```

### Test Statistics

| Package | Specs | Coverage | Race Detection | Duration |
|---------|-------|----------|----------------|----------|
| `ioutils` | 31 | 91.7% | ✅ Clean | ~30ms |
| `bufferReadCloser` | 57 | 100% | ✅ Clean | ~10ms |
| `delim` | 198 | 100% | ✅ Clean | ~170ms (~2.1s with -race) |
| `fileDescriptor` | 20 | 85.7% | ✅ Clean | ~10ms |
| `ioprogress` | 42 | 84.7% | ✅ Clean | ~10ms |
| `iowrapper` | 114 | 100% | ✅ Clean | ~50ms |
| `mapCloser` | 29 | 80.2% | ✅ Clean | ~10ms |
| `multi` | 112 | 81.7% | ✅ Clean | ~180ms (~1.3s with -race) |
| `nopwritecloser` | 54 | 100% | ✅ Clean | ~250ms (~1.2s with -race) |
| **Total** | **657** | **90.8%** | **✅ Clean** | **~720ms (~8s with -race)** |

### Quality Assurance

- ✅ All tests pass with race detector
- ✅ Thread-safe concurrent operations
- ✅ Comprehensive edge case coverage
- ✅ Integration scenario testing
- ✅ Error handling validation
- ✅ Platform-specific tests (where applicable)

For detailed testing documentation, see [TESTING.md](TESTING.md).

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI** to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must maintain thread safety
- Pass all tests including race detection: `CGO_ENABLED=1 go test -race ./...`
- Maintain or improve test coverage (currently 90.8%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features or API changes
- Add practical code examples for common use cases
- Keep TESTING.md synchronized with test suite changes
- Ensure all public APIs have comprehensive GoDoc comments

**Testing Requirements**
- Write tests for all new features using Ginkgo v2 and Gomega
- Test edge cases, error conditions, and concurrent scenarios
- Verify thread safety with `-race` flag
- Add benchmarks for performance-critical changes
- Ensure zero race conditions detected

**Pull Request Process**
1. Provide clear description of changes and motivation
2. Reference related issues or feature requests
3. Include test results (unit tests, race detection, coverage report)
4. Update documentation (README.md, TESTING.md, GoDoc comments)
5. Ensure CI passes (all tests, race detection, linting)

---

## Future Enhancements

Potential improvements for future versions:

**Core Utilities**
- Recursive directory walking with callbacks
- Atomic file operations (write-then-rename pattern)
- File locking utilities (cross-platform)
- Temporary file/directory management with auto-cleanup

**Performance**
- Memory-mapped file support
- Zero-copy I/O operations
- Batch file operations optimization
- Buffer pooling for high-throughput scenarios

**Stream Processing**
- Streaming compression/decompression wrappers
- Encryption/decryption I/O wrappers
- Checksum calculation during I/O
- Rate limiting wrappers

**Resource Management**
- Connection pooling utilities
- Weighted closer (priority-based cleanup)
- Resource usage monitoring
- Automatic resource leak detection

**Platform Support**
- Enhanced Windows support (non-cgo alternatives)
- macOS-specific optimizations
- Linux io_uring integration
- BSD platform support

**Integration**
- Cloud storage abstractions (S3, GCS, Azure)
- Network stream utilities
- Protocol-specific wrappers (HTTP, gRPC)
- Database connection management

Suggestions and feature requests are welcome via [GitHub Issues](https://github.com/nabbar/golib/issues).

---

## License

**MIT License** © Nicolas JUHEL

All source files in this package are licensed under the MIT License. See the LICENSE file in the repository root and individual source files for the complete license text.

### AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, this package's development utilized AI assistance for testing, documentation, and bug fixing under human supervision. AI was **not** used for core package implementation.

---

## Resources

**Documentation**
- [Go Package Documentation (GoDoc)](https://pkg.go.dev/github.com/nabbar/golib/ioutils)
- [Testing Guide](TESTING.md)
- [Contributing Guidelines](../CONTRIBUTING.md)

**Related Go Documentation**
- [io Package](https://pkg.go.dev/io) - Standard I/O interfaces
- [os Package](https://pkg.go.dev/os) - File and directory operations
- [bufio Package](https://pkg.go.dev/bufio) - Buffered I/O
- [context Package](https://pkg.go.dev/context) - Cancellation and deadlines

**Related golib Packages**
- [github.com/nabbar/golib/file](https://pkg.go.dev/github.com/nabbar/golib/file) - File utilities
- [github.com/nabbar/golib/atomic](https://pkg.go.dev/github.com/nabbar/golib/atomic) - Atomic operations

**Testing Frameworks**
- [Ginkgo v2](https://onsi.github.io/ginkgo/) - BDD testing framework
- [Gomega](https://onsi.github.io/gomega/) - Matcher/assertion library

**Community & Support**
- [GitHub Repository](https://github.com/nabbar/golib)
- [Issue Tracker](https://github.com/nabbar/golib/issues)
- [Project Documentation](https://github.com/nabbar/golib/blob/main/README.md)

---

## Summary

The `ioutils` package provides a production-ready suite of I/O utilities:

- **9 Specialized Subpackages**: Each solving specific I/O challenges with streaming-first design
- **657 Test Specs**: Comprehensive test coverage across all subpackages
- **90.8% Coverage**: Production-ready quality with extensive edge case testing
- **Zero Race Conditions**: All packages thread-safe with atomic operations
- **Fast Execution**: ~720ms test time (~8s with race detector)
- **Cross-Platform**: Linux, macOS, Windows support

**Quick Links**:
- [delim](delim/README.md) - Delimiter-based buffering (100% coverage, 198 specs)
- [multi](multi/README.md) - I/O multiplexing (81.7% coverage, 112 specs)
- [ioprogress](ioprogress/README.md) - Progress tracking
- [iowrapper](iowrapper/README.md) - I/O transformation (100% coverage, 114 specs)
- [mapCloser](mapCloser/README.md) - Resource management
- [bufferReadCloser](bufferReadCloser/README.md) - Buffered I/O (100% coverage, 57 specs)
- [All Subpackages](#subpackages)

For questions, issues, or contributions, visit the [GitHub repository](https://github.com/nabbar/golib).
