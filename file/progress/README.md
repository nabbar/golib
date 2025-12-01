# File Progress

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-76.1%25-brightgreen)](TESTING.md)

Thread-safe file I/O wrapper with progress tracking callbacks, supporting standard `io` interfaces for seamless integration with existing Go code.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Buffer Configuration](#buffer-configuration)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Example](#basic-example)
  - [With Progress Callbacks](#with-progress-callbacks)
  - [File Upload Simulation](#file-upload-simulation)
  - [Temporary Files](#temporary-files)
  - [File Copying](#file-copying)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Progress Interface](#progress-interface)
  - [Configuration](#configuration)
  - [Callbacks](#callbacks)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **progress** package provides a production-ready file I/O wrapper that tracks read/write progress through callback functions. It implements all standard Go `io` interfaces (`Reader`, `Writer`, `Seeker`, `Closer`, etc.) while adding transparent progress monitoring capabilities.

### Design Philosophy

1. **Standard Library Compatibility**: Fully implements Go's standard `io` interfaces
2. **Zero Overhead When Unused**: Progress tracking adds minimal overhead when no callbacks are registered
3. **Thread-Safe Callbacks**: Atomic operations ensure safe concurrent callback invocation
4. **Transparent Integration**: Drop-in replacement for `*os.File` in existing code
5. **Flexible File Creation**: Multiple constructors for different use cases (open, create, temp)

### Key Features

- ✅ **Progress Tracking**: Real-time callbacks for read/write operations, EOF, and position resets
- ✅ **Standard io Interfaces**: Implements `Reader`, `Writer`, `Seeker`, `Closer`, `ReaderFrom`, `WriterTo`, and more
- ✅ **Temporary File Support**: Auto-deletion of temporary files with `IsTemp()` indicator
- ✅ **Atomic Callbacks**: Thread-safe callback storage and invocation using atomic operations
- ✅ **Buffer Configuration**: Configurable buffer sizes for optimal I/O performance
- ✅ **Position Tracking**: `SizeBOF()` and `SizeEOF()` methods for current position and remaining bytes
- ✅ **Error Propagation**: Comprehensive error codes for debugging and error handling
- ✅ **Zero Dependencies**: Only standard library packages

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                      Progress Wrapper                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐         ┌──────────────┐                  │
│  │   os.File    │◀────────│   progress   │                  │
│  │  (underlying)│         │   (wrapper)  │                  │
│  └──────────────┘         └──────┬───────┘                  │
│                                  │                          │
│                    ┌─────────────┼─────────────┐            │
│                    │             │             │            │
│            ┌───────▼──────┐ ┌────▼─────┐ ┌────▼─────┐       │
│            │ FctIncrement │ │ FctReset │ │  FctEOF  │       │
│            │  (atomic)    │ │ (atomic) │ │ (atomic) │       │
│            └──────────────┘ └──────────┘ └──────────┘       │
│                    │             │             │            │
│                    └─────────────┼─────────────┘            │
│                                  │                          │
│                          User Callbacks                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Read(p []byte) → os.File.Read() → analyze() → callbacks
     │                │                │
     │                │                ├─▶ FctIncrement(bytes_read)
     │                │                │
     │                │                └─▶ FctEOF() [if io.EOF]
     │                │
     │                └─▶ return (n, err)
     │
     └─▶ return (n, err)

Write(p []byte) → os.File.Write() → analyze() → FctIncrement(bytes_written)
                        │
                        └─▶ return (n, err)

Seek(offset, whence) → os.File.Seek() → FctReset(max_size, current_pos)
                            │
                            └─▶ return (pos, err)

Truncate(size) → os.File.Truncate() → FctReset(max_size, current_pos)
                      │
                      └─▶ return err
```

### Buffer Configuration

The `SetBufferSize()` method allows optimizing I/O performance for specific use cases:

**Default Buffer Size**: `32 KB` (DefaultBuffSize)

**Sizing Guidelines:**

```
Small files (< 1 MB):    16 KB - 64 KB
Medium files (1-100 MB): 64 KB - 256 KB  
Large files (> 100 MB):  256 KB - 1 MB
Network I/O:             8 KB - 32 KB
SSD/NVMe:                64 KB - 512 KB
HDD:                     256 KB - 1 MB
```

**Trade-offs:**
- **Larger buffers**: Fewer I/O operations, higher memory usage
- **Smaller buffers**: More frequent callbacks, lower memory footprint

**Memory Estimation:**

```go
maxMemory := bufferSize + overhead  // ~200 bytes overhead
```

---

## Performance

### Benchmarks

Based on test suite measurements using `gmeasure`:

| Operation | Throughput | Latency (p50) | Latency (p99) |
|-----------|------------|---------------|---------------|
| Read (32KB buffer) | ~2.5 GB/s | 12 µs | 45 µs |
| Write (32KB buffer) | ~2.2 GB/s | 14 µs | 52 µs |
| Seek | N/A | 1 µs | 3 µs |
| Callback Invocation | N/A | 50 ns | 200 ns |

**Note**: Benchmarks are hardware-dependent and measured on modern SSD hardware.

### Memory Usage

**Per-File Instance:**
- Base overhead: ~200 bytes
- Callback storage: 24 bytes per callback (atomic.Value)
- Buffer (when set): Configurable (default 32 KB)

**Total Memory:**
```
Memory = BaseOverhead + (NumCallbacks × 24) + BufferSize
```

**Example:**
```go
// Typical usage: ~32.5 KB per file instance
memory := 200 + (3 * 24) + 32768  // 32,840 bytes
```

### Scalability

- **Concurrent Files**: Scales linearly up to OS file descriptor limit
- **Callback Overhead**: < 1% when using atomic operations
- **Thread Safety**: Safe for concurrent callback registration from multiple goroutines
- **Memory Footprint**: O(1) per file, independent of file size

**Limits:**
- OS file descriptor limit (typically 1024-65536)
- Available memory for buffers
- Disk I/O bandwidth

---

## Use Cases

### 1. File Download with Progress Bar

Monitor download progress in real-time:

```go
func downloadWithProgress(url, dest string) error {
    resp, _ := http.Get(url)
    defer resp.Body.Close()
    
    p, _ := progress.Create(dest)
    defer p.Close()
    
    total := resp.ContentLength
    var downloaded int64
    
    p.RegisterFctIncrement(func(n int64) {
        downloaded += n
        fmt.Printf("\rDownloading: %d%%", (downloaded*100)/total)
    })
    
    io.Copy(p, resp.Body)
    return nil
}
```

### 2. Large File Processing with Status Updates

Track processing progress for long-running operations:

```go
func processLargeFile(path string) error {
    p, _ := progress.Open(path)
    defer p.Close()
    
    size, _ := p.SizeEOF()
    
    p.RegisterFctIncrement(func(n int64) {
        current, _ := p.SizeBOF()
        log.Printf("Processed: %.2f%%", float64(current*100)/float64(size))
    })
    
    scanner := bufio.NewScanner(p)
    for scanner.Scan() {
        // Process line
    }
    return nil
}
```

### 3. Temporary File Management

Automatic cleanup of temporary files:

```go
func processTempData(data []byte) error {
    p, _ := progress.Temp("process-*.tmp")
    defer p.Close()  // Auto-deleted if IsTemp()
    
    p.Write(data)
    // Process temp file
    return nil
}
```

### 4. File Upload with Bandwidth Monitoring

Track upload speed and estimate completion time:

```go
func uploadWithMetrics(path, url string) error {
    p, _ := progress.Open(path)
    defer p.Close()
    
    var (
        start = time.Now()
        bytes int64
    )
    
    p.RegisterFctIncrement(func(n int64) {
        bytes += n
        elapsed := time.Since(start).Seconds()
        speed := float64(bytes) / elapsed / 1024 / 1024
        fmt.Printf("Upload speed: %.2f MB/s\n", speed)
    })
    
    http.Post(url, "application/octet-stream", p)
    return nil
}
```

### 5. Batch File Operations

Monitor progress across multiple files:

```go
func processBatch(files []string) error {
    for i, file := range files {
        p, _ := progress.Open(file)
        
        p.RegisterFctEOF(func() {
            fmt.Printf("Completed %d/%d: %s\n", i+1, len(files), file)
        })
        
        // Process file
        p.Close()
    }
    return nil
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/file/progress
```

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/file/progress"
)

func main() {
    // Open existing file
    p, err := progress.Open("data.txt")
    if err != nil {
        panic(err)
    }
    defer p.Close()
    
    // Read file
    buf := make([]byte, 1024)
    n, err := p.Read(buf)
    fmt.Printf("Read %d bytes\n", n)
}
```

### With Progress Callbacks

```go
p, _ := progress.Open("largefile.dat")
defer p.Close()

var totalBytes int64

// Track each read operation
p.RegisterFctIncrement(func(n int64) {
    totalBytes += n
    fmt.Printf("Read: %d bytes total\n", totalBytes)
})

// Detect when EOF is reached
p.RegisterFctEOF(func() {
    fmt.Println("File reading completed!")
})

// Detect position resets (e.g., after Seek)
p.RegisterFctReset(func(max, current int64) {
    fmt.Printf("Position reset: %d/%d\n", current, max)
})

io.Copy(io.Discard, p)
```

### File Upload Simulation

```go
p, _ := progress.Create("upload.dat")
defer p.Close()

data := make([]byte, 10*1024*1024) // 10 MB

var uploaded int64
p.RegisterFctIncrement(func(n int64) {
    uploaded += n
    percentage := (uploaded * 100) / int64(len(data))
    if percentage%10 == 0 {
        fmt.Printf("Upload: %d%%\n", percentage)
    }
})

p.Write(data)
```

### Temporary Files

```go
// Create unique temporary file
p, _ := progress.Temp("myapp-*.tmp")
defer p.Close()  // Auto-deleted on close

fmt.Printf("Temp file: %s\n", p.Path())
fmt.Printf("Is temporary: %v\n", p.IsTemp())

p.Write([]byte("temporary data"))
```

### File Copying

```go
src, _ := progress.Open("source.bin")
defer src.Close()

dst, _ := progress.Create("dest.bin")
defer dst.Close()

var copied int64
src.RegisterFctIncrement(func(n int64) {
    copied += n
})

io.Copy(dst, src)
fmt.Printf("Copied: %d bytes\n", copied)
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **76.1% code coverage** and **140 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and standard interfaces
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks (throughput, latency, memory)
- ✅ Error handling and edge cases
- ✅ Progress callback mechanisms

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Progress Tracking:**
```go
// Register callbacks before I/O
p.RegisterFctIncrement(func(n int64) {
    // Update progress bar
})

// Use SizeEOF for percentage calculation
total, _ := p.SizeEOF()
current, _ := p.SizeBOF()
percentage := float64(current) * 100 / float64(total)
```

**Resource Management:**
```go
// Always close files
p, _ := progress.Open("file.txt")
defer p.Close()

// Check IsTemp before manual deletion
if !p.IsTemp() {
    os.Remove(p.Path())
}
```

**Error Handling:**
```go
// Handle all errors
if n, err := p.Read(buf); err != nil {
    if errors.Is(err, io.EOF) {
        // Normal end of file
    } else {
        return fmt.Errorf("read error: %w", err)
    }
}
```

**Buffer Sizing:**
```go
// Set appropriate buffer for workload
p.SetBufferSize(256 * 1024)  // 256 KB for large files

// Smaller for network I/O
p.SetBufferSize(8 * 1024)    // 8 KB for network
```

**Data Persistence:**
```go
// Sync after critical writes
p.Write(criticalData)
if err := p.Sync(); err != nil {
    return fmt.Errorf("sync failed: %w", err)
}
```

### ❌ DON'T

**Don't ignore errors:**
```go
// ❌ BAD: Ignoring errors
p.Read(buf)
p.Write(data)

// ✅ GOOD: Proper error handling
if _, err := p.Read(buf); err != nil {
    return err
}
```

**Don't perform heavy work in callbacks:**
```go
// ❌ BAD: Blocking callback
p.RegisterFctIncrement(func(n int64) {
    time.Sleep(100 * time.Millisecond)  // BLOCKS I/O!
    database.UpdateProgress(n)
})

// ✅ GOOD: Async processing
updates := make(chan int64, 100)
p.RegisterFctIncrement(func(n int64) {
    select {
    case updates <- n:
    default:
    }
})
go func() {
    for n := range updates {
        database.UpdateProgress(n)
    }
}()
```

**Don't use after Close:**
```go
// ❌ BAD: Use after close
p.Close()
p.Read(buf)  // Returns ErrorNilPointer

// ✅ GOOD: Check before use
if p != nil {
    p.Read(buf)
}
```

**Don't share across goroutines without sync:**
```go
// ❌ BAD: Concurrent access
for i := 0; i < 10; i++ {
    go func() {
        p.Write(data)  // RACE!
    }()
}

// ✅ GOOD: Use separate files or synchronize
var mu sync.Mutex
for i := 0; i < 10; i++ {
    go func() {
        mu.Lock()
        defer mu.Unlock()
        p.Write(data)
    }()
}
```

**Don't panic in callbacks:**
```go
// ❌ BAD: Panic in callback
p.RegisterFctIncrement(func(n int64) {
    if n == 0 {
        panic("zero bytes!")  // Crashes program
    }
})

// ✅ GOOD: Error logging
p.RegisterFctIncrement(func(n int64) {
    if n == 0 {
        log.Error("Warning: zero bytes processed")
    }
})
```

**Don't set extreme buffer sizes:**
```go
// ❌ BAD: Excessive buffer
p.SetBufferSize(100 * 1024 * 1024)  // 100 MB!

// ✅ GOOD: Reasonable buffer
p.SetBufferSize(256 * 1024)  // 256 KB
```

**Don't forget callback propagation:**
```go
// ❌ BAD: Lose callbacks when copying
src, _ := progress.Open("src.txt")
src.RegisterFctIncrement(callback)
dst, _ := progress.Create("dst.txt")
io.Copy(dst, src)  // src callbacks not on dst!

// ✅ GOOD: Propagate callbacks
src.SetRegisterProgress(dst)
io.Copy(dst, src)
```

---

## API Reference

### Progress Interface

```go
type Progress interface {
    io.Reader
    io.Writer
    io.Seeker
    io.Closer
    io.ReaderAt
    io.WriterAt
    io.ReaderFrom
    io.WriterTo
    io.ByteReader
    io.ByteWriter
    io.StringWriter
    
    // Progress-specific methods
    RegisterFctIncrement(fct FctIncrement)
    RegisterFctReset(fct FctReset)
    RegisterFctEOF(fct FctEOF)
    SetRegisterProgress(f Progress)
    
    // File operations
    Path() string
    Stat() (os.FileInfo, error)
    SizeBOF() (int64, error)
    SizeEOF() (int64, error)
    Truncate(size int64) error
    Sync() error
    IsTemp() bool
    SetBufferSize(size int32)
    CloseDelete() error
}
```

**Methods:**

- **`Read(p []byte) (int, error)`**: Read bytes into buffer
- **`Write(p []byte) (int, error)`**: Write bytes from buffer
- **`Seek(offset int64, whence int) (int64, error)`**: Change file position
- **`Close() error`**: Close file and release resources
- **`Path() string`**: Get cleaned file path
- **`Stat() (os.FileInfo, error)`**: Get file metadata
- **`SizeBOF() (int64, error)`**: Bytes from start to current position
- **`SizeEOF() (int64, error)`**: Bytes from current position to end
- **`Truncate(size int64) error`**: Resize file
- **`Sync() error`**: Flush to disk
- **`IsTemp() bool`**: Check if temporary file
- **`SetBufferSize(size int32)`**: Configure I/O buffer size
- **`CloseDelete() error`**: Close and delete file

### Configuration

**Constructors:**

```go
// Open existing file (read-only by default)
func Open(path string) (Progress, error)

// Create new file (write-only, truncate if exists)
func Create(path string) (Progress, error)

// Open/create with custom flags
func New(path string, flag int, perm os.FileMode) (Progress, error)

// Create temporary file (auto-deleted on close)
func Temp(pattern string) (Progress, error)

// Create unique file with auto-generated name
func Unique(path, pattern string) (Progress, error)
```

**Examples:**

```go
// Read-only
p, _ := progress.Open("readonly.txt")

// Write-only (create or truncate)
p, _ := progress.Create("output.txt")

// Read-write, append mode
p, _ := progress.New("file.txt", os.O_RDWR|os.O_APPEND, 0644)

// Temporary file
p, _ := progress.Temp("temp-*.dat")

// Unique file in directory
p, _ := progress.Unique("/tmp", "unique-*.log")
```

**Buffer Configuration:**

```go
p.SetBufferSize(256 * 1024)  // 256 KB buffer
```

### Callbacks

```go
// Called after each read/write with bytes processed
type FctIncrement func(size int64)

// Called when file position is reset (Seek, Truncate)
type FctReset func(maxSize, currentPos int64)

// Called when EOF is reached during read
type FctEOF func()
```

**Registration:**

```go
p.RegisterFctIncrement(func(n int64) {
    log.Printf("Processed %d bytes", n)
})

p.RegisterFctReset(func(max, cur int64) {
    log.Printf("Reset to %d/%d bytes", cur, max)
})

p.RegisterFctEOF(func() {
    log.Println("End of file reached")
})
```

**Callback Behavior:**

- **FctIncrement**: Called after **each** `Read`/`Write` operation with bytes for that operation
- **FctReset**: Called after `Seek` or `Truncate` with file size and new position
- **FctEOF**: Called once when `io.EOF` is detected during read
- All callbacks are optional (nil-safe)
- Callbacks are invoked serially (no concurrent calls per file)
- Callbacks should be fast (avoid blocking operations)

### Error Codes

```go
const (
    ErrorParamEmpty       // Empty parameter provided
    ErrorNilPointer       // Nil pointer or closed file
    ErrorIOFileOpen       // File open failed
    ErrorIOFileCreate     // File creation failed
    ErrorIOFileStat       // File stat failed
    ErrorIOFileSeek       // Seek operation failed
    ErrorIOFileTruncate   // Truncate operation failed
    ErrorIOFileSync       // Sync operation failed
    ErrorIOTempFile       // Temporary file creation failed
    ErrorIOTempClose      // Temporary file close failed
    ErrorIOTempRemove     // Temporary file removal failed
)
```

**Usage:**

```go
p, err := progress.Open("file.txt")
if err != nil {
    if errors.Is(err, progress.ErrorIOFileOpen.Error(nil)) {
        log.Fatal("Cannot open file")
    }
}
```

**Error Handling:**

- Errors from `os.File` operations are wrapped with package-specific codes
- All methods that can fail return `error` as last return value
- Use `errors.Is()` for error type checking
- Use `errors.As()` for extracting wrapped errors

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >75%)
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

- ✅ **76.1% test coverage** (target: >75%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** callback operations using atomic operations
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard compliant** implements all relevant `io` interfaces

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Context Integration**: Add `context.Context` support for cancellable I/O operations
2. **Rate Limiting**: Implement bandwidth control through callback rate limiting
3. **Metrics Export**: Optional integration with Prometheus or OpenTelemetry
4. **Custom Error Handlers**: Allow users to provide custom error handlers in callbacks
5. **Convenience Methods**: Add `ReadAll()`, `WriteAll()`, and similar helpers

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/file/progress)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns. Automatically generated from source code comments with live example code execution.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, callback mechanisms, buffer sizing guidelines, and performance considerations. Provides detailed explanations of internal mechanisms, thread-safety guarantees, and best practices for production use. Essential reading for understanding implementation details.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (76.1%), performance benchmarks, and guidelines for writing new tests. Includes troubleshooting, concurrency testing strategies, and CI integration examples. Critical resource for contributors and quality assurance.

### Related golib Packages

- **[github.com/nabbar/golib/ioutils/delim](../../ioutils/delim)** - Buffered reader for delimiter-separated data streams with custom delimiter support and constant memory usage. Useful for processing CSV, log files, or any delimited data. Complements progress tracking for structured data processing.

- **[github.com/nabbar/golib/ioutils/aggregator](../../ioutils/aggregator)** - Thread-safe write aggregator that serializes concurrent write operations. Useful for collecting output from multiple goroutines into a single file with progress tracking. Can be combined with progress package for concurrent data collection scenarios.

- **[github.com/nabbar/golib/file/bandwidth](../bandwidth)** - Bandwidth limiting for file I/O operations. Controls read/write speeds to prevent network or disk saturation. Can be used alongside progress package for controlled, monitored file transfers.

- **[github.com/nabbar/golib/errors](../../errors)** - Enhanced error handling with error codes and structured error information. Used internally by progress package for comprehensive error reporting. Provides error chaining and classification capabilities.

### External References

- **[Go io Package](https://pkg.go.dev/io)** - Standard library documentation for `io` interfaces. The progress package fully implements these interfaces for seamless integration with Go's I/O ecosystem. Essential reading for understanding Go's I/O model.

- **[Go os Package](https://pkg.go.dev/os)** - Standard library documentation for file operations. The progress package wraps `os.File` while maintaining full compatibility. Important for understanding underlying file operations and permissions.

- **[Effective Go - Files](https://go.dev/doc/effective_go#files)** - Official Go programming guide covering file handling best practices. Demonstrates idiomatic patterns that the progress package follows. Recommended reading for proper file resource management.

- **[Go Memory Model](https://go.dev/ref/mem)** - Official specification of Go's memory consistency guarantees. Essential for understanding the thread-safety guarantees provided by atomic operations used in callback storage. Relevant for concurrent usage scenarios.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/file/progress`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
