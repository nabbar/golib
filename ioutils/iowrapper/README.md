# IOWrapper Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)]()

Flexible I/O wrapper for intercepting, transforming, and monitoring I/O operations with customizable read, write, seek, and close behavior.

---

## Overview

This package provides a flexible wrapper that allows you to intercept and customize I/O operations (read, write, seek, close) on any underlying object without modifying its implementation. It implements all standard Go I/O interfaces and uses atomic operations for thread-safe function updates.

### Design Philosophy

1. **Transparency**: Wrap any I/O object without changing its interface
2. **Flexibility**: Customize operations at runtime with custom functions
3. **Thread Safety**: All operations and updates are thread-safe via atomics
4. **Zero Overhead**: Minimal performance cost when no customization is used
5. **Composability**: Wrappers can be chained for complex transformations

### Key Features

- **Universal Wrapping**: Wrap any object (io.Reader, io.Writer, io.Seeker, io.Closer, or combinations)
- **Runtime Customization**: Change behavior dynamically with custom functions
- **Thread-Safe**: All operations safe for concurrent use via atomic operations
- **Full Interface Support**: Implements io.Reader, io.Writer, io.Seeker, io.Closer
- **Zero Dependencies**: Only standard library + internal atomics
- **100% Test Coverage**: 114 specs covering all scenarios
- **Production Ready**: Thoroughly tested for concurrency and edge cases

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

### 1. Logging and Monitoring

**Problem**: Track I/O operations without modifying existing code.

```go
wrapper := iowrapper.New(file)

var bytesRead, bytesWritten atomic.Int64

wrapper.SetRead(func(p []byte) []byte {
    n, err := file.Read(p)
    if err != nil || n == 0 {
        return nil
    }
    data := p[:n]
    bytesRead.Add(int64(len(data)))
    log.Printf("Read %d bytes (total: %d)", len(data), bytesRead.Load())
    return data
})

wrapper.SetWrite(func(p []byte) []byte {
    n, err := file.Write(p)
    if err != nil {
        return nil
    }
    bytesWritten.Add(int64(n))
    metrics.RecordWrite(n)
    return p[:n]
})
```

**Advantages**: Observability without changes to underlying I/O logic, real-time monitoring of throughput and operation counts.

### 2. Data Transformation

**Problem**: Apply transformations transparently during I/O.

```go
// ROT13 cipher on read
wrapper := iowrapper.New(reader)
wrapper.SetRead(func(p []byte) []byte {
    n, err := reader.Read(p)
    if err != nil || n == 0 {
        return nil
    }
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
```

**Advantages**: Process data transparently without exposing transformation logic to callers, chainable transformations for complex pipelines.

### 3. Data Validation

**Problem**: Validate data before writing.

```go
wrapper := iowrapper.New(writer)

wrapper.SetWrite(func(p []byte) []byte {
    if !utf8.Valid(p) {
        log.Printf("Invalid UTF-8 data rejected")
        return nil // Returns io.ErrUnexpectedEOF
    }
    if len(p) > maxSize {
        return p[:maxSize] // Truncate
    }
    n, err := writer.Write(p)
    if err != nil {
        return nil
    }
    return p[:n]
})
```

**Advantages**: Enforce invariants at the I/O layer without scattering validation logic, centralized validation reduces bugs.

### 4. Checksumming and Integrity

**Problem**: Calculate checksums while reading data.

```go
wrapper := iowrapper.New(reader)
hasher := sha256.New()

wrapper.SetRead(func(p []byte) []byte {
    n, err := reader.Read(p)
    if err != nil || n == 0 {
        return nil
    }
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

**Advantages**: Compute checksums without explicit hash tracking in business logic, supports any hash.Hash implementation.

### 5. Rate Limiting and Throttling

**Problem**: Control I/O throughput.

```go
wrapper := iowrapper.New(reader)
limiter := rate.NewLimiter(rate.Limit(1024*1024), 4096) // 1 MB/s

wrapper.SetRead(func(p []byte) []byte {
    // Wait for rate limiter
    if err := limiter.WaitN(context.Background(), len(p)); err != nil {
        return nil
    }
    
    n, err := reader.Read(p)
    if err != nil || n == 0 {
        return nil
    }
    return p[:n]
})
```

**Advantages**: Throttle I/O operations to respect bandwidth constraints, transparent to application code.

### 6. Wrapper Chaining (Advanced)

**Problem**: Combine multiple transformations.

```go
// Chain: File → Logging → Compression → Encryption → Output
file, _ := os.Open("data.txt")

// Layer 1: Logging
logged := iowrapper.New(file)
logged.SetRead(makeLoggingRead(file))

// Layer 2: Compression
compressed := iowrapper.New(logged)
compressed.SetRead(makeCompressRead(logged))

// Layer 3: Encryption
encrypted := iowrapper.New(compressed)
encrypted.SetRead(makeEncryptRead(compressed))

// Use encrypted wrapper
io.Copy(destination, encrypted)
```

**Advantages**: Composable transformations with clear separation of concerns, each layer has single responsibility, testable in isolation.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/iowrapper
```

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

## Best Practices

### Testing

The package includes a comprehensive test suite with **100% code coverage** and **114 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and I/O operations
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks (throughput, latency, memory)
- ✅ Error handling and edge cases
- ✅ Custom function behavior and atomic updates

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Handle EOF Correctly:**
```go
wrapper.SetRead(func(p []byte) []byte {
    n, err := reader.Read(p)
    if err == io.EOF || n == 0 {
        return nil // Signal EOF properly
    }
    return p[:n]
})
```

**Minimize Allocations:**
```go
// ✅ GOOD: Reuses provided buffer
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    return p[:n] // No allocation
})
```

**Use Thread-Safe Operations:**
```go
// ✅ GOOD: Thread-safe counter using atomics
var counter atomic.Int64
wrapper.SetRead(func(p []byte) []byte {
    counter.Add(1) // Atomic, thread-safe
    n, _ := reader.Read(p)
    return p[:n]
})
```

**Reset to Default When Done:**
```go
// ✅ GOOD: Reset custom function
wrapper.SetRead(customFunc)
// ... use custom behavior ...
wrapper.SetRead(nil) // Reset to default delegation
```

**Compose Wrappers for Complexity:**
```go
// ✅ GOOD: Chain wrappers for separation of concerns
logged := iowrapper.New(file)
logged.SetRead(logFunc)

compressed := iowrapper.New(logged)
compressed.SetRead(compressFunc)

encrypted := iowrapper.New(compressed)
encrypted.SetRead(encryptFunc)
```

### ❌ DON'T

**Don't Allocate on Every Call:**
```go
// ❌ BAD: Allocates on every call
wrapper.SetRead(func(p []byte) []byte {
    data := make([]byte, len(p)) // Allocation!
    copy(data, p)
    return data
})
```

**Don't Use Mutexes (Use Atomics):**
```go
// ❌ BAD: Mutex overhead
var mu sync.Mutex
var counter int
wrapper.SetRead(func(p []byte) []byte {
    mu.Lock()
    counter++
    mu.Unlock()
    return /* ... */
})

// ✅ GOOD: Use atomic instead
var counter atomic.Int64
wrapper.SetRead(func(p []byte) []byte {
    counter.Add(1)
    return /* ... */
})
```

**Don't Mix Multiple Concerns:**
```go
// ❌ BAD: One function doing everything
wrapper.SetRead(func(p []byte) []byte {
    // logging + compression + encryption all mixed
    log.Println("reading")
    compressed := compress(data)
    encrypted := encrypt(compressed)
    return encrypted
})

// ✅ GOOD: Use wrapper chaining instead
```

**Don't Ignore Errors:**
```go
// ❌ BAD: Ignoring errors
wrapper.SetRead(func(p []byte) []byte {
    reader.Read(p) // Ignoring error
    return p
})

// ✅ GOOD: Check errors
wrapper.SetRead(func(p []byte) []byte {
    n, err := reader.Read(p)
    if err != nil || n == 0 {
        return nil // Signal error/EOF
    }
    return p[:n]
})
```

**Don't Modify Input Buffer Carelessly:**
```go
// ⚠️ CAUTION: Modifies caller's buffer
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    modifyInPlace(p[:n]) // Mutates caller's buffer!
    return p[:n]
})

// ✅ BETTER: Return new slice if transformation needed
wrapper.SetRead(func(p []byte) []byte {
    n, _ := reader.Read(p)
    return transform(p[:n]) // New slice
})
```

---

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

## Improvements

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **100% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Zero mutexes** for maximum performance
- ✅ **Memory-safe** with proper nil checks and bounds validation

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Context Support**: Optional context.Context parameter for I/O operations to enable cancellation and deadline propagation in custom functions
2. **Metrics Integration**: Built-in metrics collection (bytes read/written, operation counts) with optional export to Prometheus or other systems
3. **Middleware Chain**: Simplified wrapper composition API with pre-built transformation utilities
4. **Error Wrapping**: Enhanced error context with stack traces and operation metadata for better debugging
5. **Zero-allocation Paths**: Optimize hot paths further to eliminate remaining allocations in edge cases
6. **Batch Operations**: Support for vectorized I/O operations to improve throughput
7. **Helper Functions**: Common transformation utilities (base64, compression, encryption) as ready-to-use functions
8. **Debug Mode**: Built-in debug logging with verbosity levels for troubleshooting I/O behavior

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/ioutils/iowrapper)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns. Includes type definitions for FuncRead, FuncWrite, FuncSeek, and FuncClose with detailed behavior documentation.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, operation flow charts, thread-safety model, and detailed error handling patterns. Provides comprehensive explanations of internal mechanisms, atomic operations, and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 100% coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting, CI integration examples, and helper function documentation.

### Related golib Packages

- **[github.com/nabbar/golib/atomic](https://pkg.go.dev/github.com/nabbar/golib/atomic)** - Thread-safe atomic value storage used internally for storing custom functions (FuncRead, FuncWrite, etc.). Provides lock-free atomic operations for better performance in concurrent scenarios. This package is critical for the wrapper's zero-mutex, thread-safe architecture.

- **[github.com/nabbar/golib/ioutils/bufferReadCloser](../bufferReadCloser)** - I/O wrappers with close support for buffered operations. Can be combined with iowrapper for advanced I/O patterns requiring buffer management and proper cleanup.

- **[github.com/nabbar/golib/ioutils/fileDescriptor](../fileDescriptor)** - File descriptor limit management for applications handling many concurrent files. Useful when wrapping file operations to track and limit resource usage.

- **[github.com/nabbar/golib/ioutils](../)** - Parent package containing additional I/O utilities including aggregator (concurrent write serialization), bufferReadCloser, and fileDescriptor. Comprehensive toolkit for advanced I/O patterns.

### External References

- **[Go io Package](https://pkg.go.dev/io)** - Standard library I/O interfaces (io.Reader, io.Writer, io.Seeker, io.Closer) that the wrapper implements. Essential reference for understanding the contracts and error handling patterns used by the wrapper.

- **[Effective Go - Interfaces](https://go.dev/doc/effective_go#interfaces)** - Official Go programming guide covering interface design and composition patterns. The wrapper follows these conventions for idiomatic Go interface implementation.

- **[Go Concurrency Patterns](https://go.dev/blog/pipelines)** - Official Go blog article explaining concurrency patterns. Relevant for understanding how custom functions can be used in pipeline architectures with concurrent I/O.

- **[Go Memory Model](https://go.dev/ref/mem)** - Official specification of Go's memory consistency guarantees. Essential for understanding the thread-safety guarantees provided by atomic operations used in the wrapper.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License © Nicolas JUHEL

All source files in this package are licensed under the MIT License. See individual files for the full license header.
