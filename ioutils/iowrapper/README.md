# IOWrapper Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Flexible I/O wrapper for intercepting, transforming, and monitoring I/O operations with customizable read, write, seek, and close behavior.

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

This package provides a flexible wrapper that allows you to intercept and customize I/O operations (read, write, seek, close) on any underlying object without modifying its implementation. It implements all standard Go I/O interfaces and uses atomic operations for thread-safe function updates.

### Design Philosophy

1. **Transparency**: Wrap any I/O object without changing its interface
2. **Flexibility**: Customize operations at runtime with custom functions
3. **Thread Safety**: All operations and updates are thread-safe via atomics
4. **Zero Overhead**: Minimal performance cost when no customization is used
5. **Composability**: Wrappers can be chained for complex transformations

---

## Key Features

- **Universal Wrapping**: Wrap any object (io.Reader, io.Writer, io.Seeker, io.Closer, or combinations)
- **Runtime Customization**: Change behavior dynamically with custom functions
- **Thread-Safe**: All operations safe for concurrent use via atomic operations
- **Full Interface Support**: Implements io.Reader, io.Writer, io.Seeker, io.Closer
- **Zero Dependencies**: Only standard library + internal atomics
- **100% Test Coverage**: 114 specs covering all scenarios
- **Production Ready**: Thoroughly tested for concurrency and edge cases

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils/iowrapper
```

---

## Architecture

### Package Structure

```
iowrapper/
├── interface.go    # Public API and type definitions
└── model.go        # Internal implementation with atomic operations
```

### Component Diagram

```
┌─────────────────────────────────────────────┐
│          iowrapper.IOWrapper                │
│     (io.Reader + io.Writer + ...)          │
└──────────────┬──────────────────────────────┘
               │
    ┌──────────▼──────────┐
    │   Atomic Functions  │
    │  (thread-safe)      │
    ├─────────────────────┤
    │ • FuncRead          │
    │ • FuncWrite         │
    │ • FuncSeek          │
    │ • FuncClose         │
    └──────────┬──────────┘
               │
    ┌──────────▼──────────┐
    │  Underlying Object  │
    │  (any type)         │
    ├─────────────────────┤
    │ io.Reader    (opt)  │
    │ io.Writer    (opt)  │
    │ io.Seeker    (opt)  │
    │ io.Closer    (opt)  │
    └─────────────────────┘
```

### Operation Flow

```
User calls Read(p []byte)
       ↓
1. Load custom function (atomic)
       ↓
2. Custom function set?
   ├─ Yes → Call custom FuncRead(p)
   └─ No  → Check underlying object
       ↓
3. Underlying implements io.Reader?
   ├─ Yes → Call underlying.Read(p)
   └─ No  → Return io.ErrUnexpectedEOF
       ↓
4. Process result
   ├─ nil → Return io.ErrUnexpectedEOF
   ├─ []byte{} → Return 0, nil
   └─ data → Copy to p, return len, nil
```

### Thread Safety Model

```
Wrapper Instance
├─ Underlying object (immutable after creation)
└─ Atomic function storage
   ├─ atomic.Value[FuncRead]  ← SetRead (atomic store)
   ├─ atomic.Value[FuncWrite] ← SetWrite (atomic store)
   ├─ atomic.Value[FuncSeek]  ← SetSeek (atomic store)
   └─ atomic.Value[FuncClose] ← SetClose (atomic store)

Concurrent Operations:
✓ Multiple Read() calls
✓ Multiple Write() calls  
✓ Read() + SetRead() simultaneously
✓ Write() + SetWrite() simultaneously
```

---

## Quick Start

### Basic Wrapping

```go
package main

import (
    "bytes"
    "fmt"
    "github.com/nabbar/golib/ioutils/iowrapper"
)

func main() {
    // Wrap any I/O object
    buf := bytes.NewBuffer([]byte("hello world"))
    wrapper := iowrapper.New(buf)

    // Use as normal io.Reader
    data := make([]byte, 5)
    n, _ := wrapper.Read(data)
    fmt.Println(string(data[:n])) // Output: "hello"
}
```

### Custom Read Function

```go
wrapper := iowrapper.New(reader)

// Transform data to uppercase on read
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    data := p[:n]
    for i := range data {
        if data[i] >= 'a' && data[i] <= 'z' {
            data[i] -= 32 // Convert to uppercase
        }
    }
    return data
})

// All reads now return uppercase
buf := make([]byte, 100)
wrapper.Read(buf) // Returns uppercase data
```

### Custom Write with Logging

```go
buf := &bytes.Buffer{}
wrapper := iowrapper.New(buf)

// Log all writes
var bytesWritten int
wrapper.SetWrite(func(p []byte) []byte {
    bytesWritten += len(p)
    log.Printf("Writing %d bytes (total: %d)", len(p), bytesWritten)
    buf.Write(p)
    return p
})

wrapper.Write([]byte("data")) // Logs the write
```

### Reset to Default Behavior

```go
wrapper := iowrapper.New(reader)
wrapper.SetRead(customFunc)

// Later, reset to default (delegate to underlying reader)
wrapper.SetRead(nil)
```

---

## Performance

### Operation Metrics

The wrapper adds **minimal overhead** to I/O operations:

| Operation | Overhead | Notes |
|-----------|----------|-------|
| Read (default) | ~0-100 ns | Atomic load + delegation |
| Write (default) | ~0-100 ns | Atomic load + delegation |
| Read (custom) | ~0-100 ns + custom | Custom function cost |
| Write (custom) | ~0-100 ns + custom | Custom function cost |
| SetRead/SetWrite | ~100-200 ns | Atomic store |
| Creation | ~5-7 ms / 10k ops | One-time cost |

*Measured on AMD64, Go 1.21*

### Memory Efficiency

- **Wrapper Size**: ~64 bytes (1 pointer + 4 atomic values)
- **Allocations**: 0 per I/O operation
- **Custom Function**: Stored once, no per-call allocation
- **Data Copying**: Only when custom function returns different slice

### Benchmark Results

From actual test runs:

```
Wrapper creation:     ~5.7 ms / 10,000 operations
Default read:         ~0 ns/op  (indistinguishable from baseline)
Default write:        ~0 ns/op  (indistinguishable from baseline)
Custom read:          ~0-100 ns/op (function call overhead)
Custom write:         ~0 ns/op
Function update:      ~100 ns/op (atomic store)
Seek:                 ~0 ns/op
Mixed operations:     ~100 ns/op
```

---

## Use Cases

### Logging and Monitoring

Track I/O operations without modifying the underlying implementation:

```go
wrapper := iowrapper.New(file)

var bytesRead, bytesWritten atomic.Int64

wrapper.SetRead(func(p []byte) []byte {
    n, _ := file.Read(p)
    data := p[:n]
    bytesRead.Add(int64(len(data)))
    log.Printf("Read %d bytes (total: %d)", len(data), bytesRead.Load())
    return data
})

wrapper.SetWrite(func(p []byte) []byte {
    file.Write(p)
    bytesWritten.Add(int64(len(p)))
    metrics.RecordWrite(len(p))
    return p
})
```

**Why**: Observability without code changes to underlying I/O logic.

### Data Transformation

Transform data on-the-fly during read/write operations:

```go
// ROT13 cipher on read
wrapper := iowrapper.New(reader)
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    data := p[:n]
    for i, b := range data {
        if b >= 'a' && b <= 'z' {
            data[i] = 'a' + (b-'a'+13)%26
        } else if b >= 'A' && b <= 'Z' {
            data[i] = 'A' + (b-'A'+13)%26
        }
    }
    return data
})

// Compression on write
wrapper.SetWrite(func(p []byte) []byte {
    compressed := compress(p) // Your compression logic
    writer.Write(compressed)
    return compressed
})
```

**Why**: Process data transparently without exposing transformation logic to callers.

### Data Validation

Validate data before it's processed:

```go
wrapper := iowrapper.New(writer)

wrapper.SetWrite(func(p []byte) []byte {
    if !utf8.Valid(p) {
        log.Printf("Invalid UTF-8 data rejected")
        return nil // Causes Write() to return io.ErrUnexpectedEOF
    }
    if len(p) > maxSize {
        return p[:maxSize] // Truncate
    }
    writer.Write(p)
    return p
})
```

**Why**: Enforce invariants at the I/O layer without scattering validation logic.

### Checksumming

Calculate checksums while reading or writing:

```go
wrapper := iowrapper.New(reader)
hasher := sha256.New()

wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    data := p[:n]
    hasher.Write(data) // Update checksum continuously
    return data
})

// Read all data
io.Copy(io.Discard, wrapper)

// Get final checksum
checksum := hex.EncodeToString(hasher.Sum(nil))
fmt.Printf("SHA256: %s\n", checksum)
```

**Why**: Compute checksums without explicit hash tracking in business logic.

### Rate Limiting

Control I/O throughput:

```go
wrapper := iowrapper.New(reader)

limiter := rate.NewLimiter(rate.Limit(1024*1024), 4096) // 1 MB/s

wrapper.SetRead(func(p []byte) []byte {
    // Wait for rate limiter
    if err := limiter.WaitN(context.Background(), len(p)); err != nil {
        return nil
    }
    
    n, _ := reader.Read(p)
    return p[:n]
})
```

**Why**: Throttle I/O operations to respect bandwidth constraints.

### Wrapper Chaining

Combine multiple transformations:

```go
// Chain: File → Logging → Compression → Encryption → Output
file, _ := os.Open("data.txt")

// Layer 1: Logging
logged := iowrapper.New(file)
logged.SetRead(func(p []byte) []byte {
    data := /* read from file */
    log.Printf("Read %d bytes", len(data))
    return data
})

// Layer 2: Compression
compressed := iowrapper.New(logged)
compressed.SetRead(func(p []byte) []byte {
    data := /* read from logged */
    return compress(data)
})

// Layer 3: Encryption
encrypted := iowrapper.New(compressed)
encrypted.SetRead(func(p []byte) []byte {
    data := /* read from compressed */
    return encrypt(data)
})

// Use encrypted wrapper
io.Copy(destination, encrypted)
```

**Why**: Composable transformations with clear separation of concerns.

## API Reference

### IOWrapper Interface

```go
type IOWrapper interface {
    io.Reader
    io.Writer
    io.Seeker
    io.Closer

    SetRead(read FuncRead)
    SetWrite(write FuncWrite)
    SetSeek(seek FuncSeek)
    SetClose(close FuncClose)
}
```

Implements all standard Go I/O interfaces with customizable behavior via Set* methods.

### Function Types

```go
// FuncRead: Custom read function
// Return nil for EOF/error, empty slice for 0 bytes, or data slice
type FuncRead func(p []byte) []byte

// FuncWrite: Custom write function
// Return nil for error, or slice of bytes written
type FuncWrite func(p []byte) []byte

// FuncSeek: Custom seek function
// Return new position and any error
type FuncSeek func(offset int64, whence int) (int64, error)

// FuncClose: Custom close function  
// Return error if close fails
type FuncClose func() error
```

### New

```go
func New(in any) IOWrapper
```

Creates a wrapper for any object. Delegates to underlying interfaces when available.

**Parameters:**
- `in any`: Object to wrap (can be nil, any I/O interface, or combination)

**Returns:**
- `IOWrapper`: Thread-safe wrapper with default delegation behavior

**Example:**
```go
wrapper := iowrapper.New(bytes.NewBuffer([]byte("data")))
```

### Default Behavior

| Operation | Underlying Implements | Behavior |
|-----------|----------------------|----------|
| Read | io.Reader | Delegates to underlying.Read() |
| Read | Not io.Reader | Returns io.ErrUnexpectedEOF |
| Write | io.Writer | Delegates to underlying.Write() |
| Write | Not io.Writer | Returns io.ErrUnexpectedEOF |
| Seek | io.Seeker | Delegates to underlying.Seek() |
| Seek | Not io.Seeker | Returns io.ErrUnexpectedEOF |
| Close | io.Closer | Delegates to underlying.Close() |
| Close | Not io.Closer | Returns nil (no error) |

### Custom Function Behavior

**SetRead(func(p []byte) []byte)**
- `nil` return → `Read()` returns `0, io.ErrUnexpectedEOF`
- `[]byte{}` return → `Read()` returns `0, nil`
- `data` return → `Read()` copies to p, returns `len(data), nil`
- Pass `nil` to reset to default

**SetWrite(func(p []byte) []byte)**
- `nil` return → `Write()` returns `0, io.ErrUnexpectedEOF`
- `data` return → `Write()` returns `len(data), nil`
- Pass `nil` to reset to default

**SetSeek(func(offset int64, whence int) (int64, error))**
- Direct control over position and errors
- Pass `nil` to reset to default

**SetClose(func() error)**
- Direct control over cleanup
- Pass `nil` to reset to default

---

## Best Practices

### 1. Reset to Default When Done

```go
wrapper.SetRead(customFunc)
// ... use custom behavior ...
wrapper.SetRead(nil) // Reset to default
```

### 2. Handle EOF Correctly

```go
wrapper.SetRead(func(p []byte) []byte {
    n, err := reader.Read(p)
    if err == io.EOF || n == 0 {
        return nil // Signal EOF
    }
    return p[:n]
})
```

### 3. Minimize Allocations

```go
// ❌ Bad - allocates on every call
wrapper.SetRead(func(p []byte) []byte {
    data := make([]byte, len(p)) // Allocation!
    copy(data, p)
    return data
})

// ✅ Good - reuses provided buffer
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    return p[:n] // No allocation
})
```

### 4. Thread-Safe Custom Functions

```go
// Thread-safe counter using atomics
var counter atomic.Int64
wrapper.SetRead(func(p []byte) []byte {
    counter.Add(1) // Thread-safe
    return /* ... */
})
```

### 5. Don't Modify Input Unless Necessary

```go
// ✅ Good - read-only transformation
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    transformed := transform(p[:n]) // New slice
    return transformed
})

// ⚠️ Caution - modifies input
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    modifyInPlace(p[:n]) // Mutates caller's buffer
    return p[:n]
})
```

### 6. Check for io.ErrUnexpectedEOF

```go
n, err := wrapper.Read(buf)
if err == io.ErrUnexpectedEOF {
    // Custom function returned nil or no underlying reader
}
```

### 7. Compose Wrappers for Complex Logic

```go
// Don't: One function doing everything
wrapper.SetRead(func(p []byte) []byte {
    // logging + compression + encryption all mixed
})

// Do: Chain wrappers
logged := iowrapper.New(file)
logged.SetRead(logFunc)

compressed := iowrapper.New(logged)
compressed.SetRead(compressFunc)

encrypted := iowrapper.New(compressed)
encrypted.SetRead(encryptFunc)
```

---

## Testing

The package has **100% test coverage** with 114 comprehensive specs using Ginkgo v2 and Gomega.

### Run Tests

```bash
# Standard go test
go test -v -cover .

# With Ginkgo CLI (recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo -v -cover

# With race detector
go test -race .
```

### Test Statistics

| Metric | Value |
|--------|-------|
| Total Specs | 114 |
| Coverage | 100.0% |
| Execution Time | ~47ms |
| Success Rate | 100% |

### Test Categories

- **Basic Operations** (basic_test.go): 20 specs
- **Custom Functions** (custom_test.go): 24 specs
- **Edge Cases** (edge_cases_test.go): 18 specs
- **Error Handling** (errors_test.go): 19 specs
- **Concurrency** (concurrency_test.go): 17 specs
- **Integration** (integration_test.go): 8 specs
- **Benchmarks** (benchmark_test.go): 8 specs

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI** to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass tests: `go test ./...`
- Maintain 100% test coverage
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
- Context support: Context-aware I/O operations
- Metrics integration: Built-in metrics collection
- Middleware chain: Simplified wrapper composition
- Error wrapping: Enhanced error context

**Performance**
- Zero-allocation paths: Optimize hot paths further
- Batch operations: Support for vectorized I/O
- Buffer pooling: Reduce allocation pressure

**Developer Experience**
- Helper functions: Common transformation utilities
- Debugging: Built-in debug logging mode
- Examples: More real-world use case examples

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
- [GoDoc Reference](https://pkg.go.dev/github.com/nabbar/golib/ioutils/iowrapper)
- [Testing Guide](TESTING.md)
- [Go io Package](https://pkg.go.dev/io)

**Related Packages**
- [bufferReadCloser](../bufferReadCloser) - I/O wrappers with close support
- [fileDescriptor](../fileDescriptor) - File descriptor limit management
- [ioutils](../) - Parent package with additional I/O utilities

**Community**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [Contributing Guide](../../CONTRIBUTING.md)

---

**Version**: Go 1.19+ on Linux, macOS, Windows  
**Maintained By**: iowrapper Package Contributors
