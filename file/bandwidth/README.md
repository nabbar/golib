# Bandwidth Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-84.4%25-brightgreen)]()

Lightweight, thread-safe bandwidth throttling and rate limiting for file I/O operations with seamless progress tracking integration.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The `bandwidth` package provides bandwidth throttling and rate limiting for file I/O operations through seamless integration with the `github.com/nabbar/golib/file/progress` package. It enforces bytes-per-second transfer limits using time-based throttling with atomic operations for thread-safe concurrent usage.

### Design Philosophy

1. **Zero-Cost Unlimited**: Setting limit to 0 disables throttling with no overhead
2. **Atomic Operations**: Thread-safe concurrent access without mutexes
3. **Callback Integration**: Seamless integration with progress tracking callbacks
4. **Time-Based Limiting**: Enforces rate limits by introducing sleep delays when needed
5. **Simple API**: Minimal learning curve with straightforward registration pattern

### Why Use This Package?

- **Network Bandwidth Control**: Prevent overwhelming network connections during uploads/downloads
- **Disk I/O Rate Limiting**: Avoid disk saturation during large file operations
- **Shared Bandwidth Management**: Control aggregate bandwidth across multiple concurrent transfers
- **Progress Monitoring**: Combine bandwidth limiting with real-time progress tracking
- **Production-Ready**: Thread-safe, tested, and battle-hardened implementation

### Key Features

- **Configurable Limits**: Any bytes-per-second rate from 1 byte/s to unlimited
- **Thread-Safe**: Safe for concurrent use across multiple goroutines
- **Zero Overhead**: No performance penalty when unlimited (limit = 0)
- **Atomic Operations**: Lock-free timestamp storage for minimal contention
- **84.4% Test Coverage**: Comprehensive test suite with race detection
- **Integration Ready**: Works seamlessly with progress package
- **Callback Support**: Optional user callbacks for increment and reset events
- **No External Dependencies**: Only standard library + golib packages

---

## Architecture

### Package Structure

```
file/bandwidth/
├── interface.go         # BandWidth interface and New() constructor
├── model.go            # Internal bw struct implementation
├── doc.go              # Package documentation
└── *_test.go           # Test files
```

### Component Overview

```
┌─────────────────────────────────────────────┐
│           BandWidth Interface               │
│  ┌───────────────────────────────────────┐  │
│  │  RegisterIncrement(fpg, callback)     │  │
│  │  RegisterReset(fpg, callback)         │  │
│  └───────────────────────────────────────┘  │
└──────────────────┬──────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│           bw Implementation                 │
│  ┌───────────────────────────────────────┐  │
│  │  t: atomic.Value (timestamp)          │  │
│  │  l: Size (bytes per second limit)     │  │
│  └───────────────────────────────────────┘  │
│  ┌───────────────────────────────────────┐  │
│  │  Increment(size) - enforce limit      │  │
│  │  Reset(size, current) - clear state   │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────┐
│      Progress Package Integration           │
│  ┌───────────────────────────────────────┐  │
│  │  FctIncrement callbacks               │  │
│  │  FctReset callbacks                   │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

| Component | Memory | Complexity | Thread-Safe |
|-----------|--------|------------|-------------|
| **BandWidth** | O(1) | Simple | ✅ always |
| **bw** | O(1) | Internal | ✅ atomic ops |
| **Timestamp** | 8 bytes | Minimal | ✅ atomic.Value |

### Rate Limiting Algorithm

```
1. Store timestamp when bytes are transferred
2. On next transfer, calculate elapsed time since last timestamp
3. Calculate current rate: rate = bytes / elapsed_seconds
4. If rate > limit, calculate required sleep: sleep = (rate / limit) * 1s
5. Sleep to enforce limit (capped at 1 second maximum)
6. Store new timestamp
```

This approach provides smooth rate limiting without strict per-operation delays, allowing burst transfers when the average rate is below the limit.

---

## Performance

### Memory Efficiency

**Constant Memory Usage** - The package maintains O(1) memory regardless of transfer size:

```
Base overhead:        ~100 bytes (struct)
Timestamp storage:    8 bytes (atomic.Value)
Total:                ~108 bytes per instance
Memory Growth:        ZERO (no additional allocation per operation)
```

### Throughput Impact

Performance impact depends on the configured limit:

| Limit | Overhead | Impact |
|-------|----------|--------|
| **0 (unlimited)** | ~0µs | Zero overhead |
| **1 MB/s** | <1ms | Minimal for normal files |
| **100 KB/s** | <10ms | Noticeable for small transfers |
| **1 KB/s** | Variable | Significant throttling |

*Measured with default buffer sizes, actual performance varies with file size and transfer patterns*

### Concurrency Performance

The package scales well with concurrent instances:

| Goroutines | Throughput | Latency | Memory |
|------------|-----------|---------|--------|
| 1 | Native speed | <1ms | ~100B |
| 10 | Native speed | <1ms | ~1KB |
| 100 | Native speed | <1ms | ~10KB |

**Thread Safety:**
- ✅ Lock-free atomic operations
- ✅ Zero contention on timestamp storage
- ✅ Safe for concurrent RegisterIncrement/RegisterReset calls

---

## Use Cases

### 1. Network Upload Rate Limiting

**Problem**: Control upload speed to avoid overwhelming network connections.

```go
bw := bandwidth.New(size.SizeMiB) // 1 MB/s limit
fpg, _ := progress.Open("upload.dat")
bw.RegisterIncrement(fpg, nil)
io.Copy(networkConn, fpg) // Throttled to 1 MB/s
```

**Real-world**: Used for cloud backup uploads, file synchronization services.

### 2. Disk I/O Throttling

**Problem**: Prevent disk saturation during large file operations.

```go
bw := bandwidth.New(10 * size.SizeMiB) // 10 MB/s
fpg, _ := progress.Open("large_backup.tar")
bw.RegisterIncrement(fpg, func(sz int64) {
    fmt.Printf("Progress: %d bytes\n", sz)
})
io.Copy(destination, fpg)
```

**Real-world**: Database backups, log archiving, bulk data processing.

### 3. Multi-File Shared Bandwidth

**Problem**: Control aggregate bandwidth across multiple concurrent transfers.

```go
sharedBW := bandwidth.New(5 * size.SizeMiB) // Shared 5 MB/s
for _, file := range files {
    go func(f string) {
        fpg, _ := progress.Open(f)
        sharedBW.RegisterIncrement(fpg, nil)
        io.Copy(destination, fpg)
    }(file)
}
```

**Real-world**: Distributed file systems, CDN uploads, multi-stream downloads.

### 4. Progress Monitoring with Rate Limiting

**Problem**: Combine bandwidth limiting with user-visible progress tracking.

```go
bw := bandwidth.New(size.SizeMiB)
fpg, _ := progress.Open("data.bin")
bw.RegisterIncrement(fpg, func(sz int64) {
    pct := float64(sz) / float64(fileSize) * 100
    fmt.Printf("Progress: %.1f%%\n", pct)
})
io.Copy(writer, fpg) // 1 MB/s with progress updates
```

**Real-world**: File managers, backup software, download clients.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/file/bandwidth
```

### Basic Usage (Unlimited)

```go
package main

import (
    "io"
    "os"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
)

func main() {
    // Create bandwidth limiter (unlimited)
    bw := bandwidth.New(0)
    
    // Open file with progress tracking
    fpg, _ := progress.Open("file.dat")
    defer fpg.Close()
    
    // Register bandwidth limiting
    bw.RegisterIncrement(fpg, nil)
    
    // Transfer with no throttling
    io.Copy(destination, fpg)
}
```

### With Bandwidth Limit

```go
package main

import (
    "io"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

func main() {
    // Create bandwidth limiter: 1 MB/s
    bw := bandwidth.New(size.SizeMiB)
    
    // Open file with progress tracking
    fpg, _ := progress.Open("large-file.dat")
    defer fpg.Close()
    
    // Register bandwidth limiting
    bw.RegisterIncrement(fpg, nil)
    
    // Transfer throttled to 1 MB/s
    io.Copy(destination, fpg)
}
```

### With Progress Callback

```go
package main

import (
    "fmt"
    "io"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

func main() {
    // Create bandwidth limiter: 2 MB/s
    bw := bandwidth.New(2 * size.SizeMiB)
    
    // Open file with progress tracking
    fpg, _ := progress.Open("file.dat")
    defer fpg.Close()
    
    // Register with progress callback
    var totalBytes int64
    bw.RegisterIncrement(fpg, func(size int64) {
        totalBytes += size
        fmt.Printf("Transferred: %d bytes\n", totalBytes)
    })
    
    // Transfer with progress updates
    io.Copy(destination, fpg)
}
```

### With Reset Callback

```go
package main

import (
    "fmt"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

func main() {
    bw := bandwidth.New(size.SizeMiB)
    fpg, _ := progress.Open("file.dat")
    defer fpg.Close()
    
    // Register reset callback
    bw.RegisterReset(fpg, func(size, current int64) {
        fmt.Printf("Reset: max=%d current=%d\n", size, current)
    })
    
    // Operations that may trigger reset
    buffer := make([]byte, 512)
    fpg.Read(buffer)
    fpg.Reset(1024)
}
```

### Network Transfer Example

```go
package main

import (
    "net"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

func main() {
    // Connect to server
    conn, _ := net.Dial("tcp", "example.com:8080")
    defer conn.Close()
    
    // Bandwidth limit: 500 KB/s
    bw := bandwidth.New(500 * size.SizeKilo)
    
    // Open local file
    fpg, _ := progress.Open("upload.dat")
    defer fpg.Close()
    
    // Register bandwidth limiting with progress
    var uploaded int64
    bw.RegisterIncrement(fpg, func(size int64) {
        uploaded += size
        fmt.Printf("Uploaded: %d bytes\n", uploaded)
    })
    
    // Upload with rate limiting
    io.Copy(conn, fpg)
}
```

---

## API Reference

### Types

#### BandWidth Interface

```go
type BandWidth interface {
    RegisterIncrement(fpg Progress, fi FctIncrement)
    RegisterReset(fpg Progress, fr FctReset)
}
```

Primary interface for bandwidth control and rate limiting.

**Methods**:
- `RegisterIncrement(fpg, callback)` - Register bandwidth-limited increment callback
- `RegisterReset(fpg, callback)` - Register reset callback that clears tracking state

### Functions

#### New

```go
func New(bytesBySecond Size) BandWidth
```

Creates a new BandWidth instance with the specified rate limit.

**Parameters**:
- `bytesBySecond` - Maximum transfer rate in bytes per second
  - Use `0` for unlimited bandwidth (no throttling)
  - Common values: `size.SizeKilo` (1KB/s), `size.SizeMega` (1MB/s)

**Returns**: BandWidth instance

**Example**:
```go
bw := bandwidth.New(0)                    // Unlimited
bw := bandwidth.New(size.SizeMega)        // 1 MB/s
bw := bandwidth.New(512 * size.SizeKilo)  // 512 KB/s
```

### Behavior

| Configuration | Behavior |
|---------------|----------|
| `bytesBySecond = 0` | No throttling, zero overhead |
| `bytesBySecond > 0` | Enforces rate by sleep delays |
| Rate calculation | `bytes / elapsed_seconds` |
| Sleep duration | Capped at 1 second maximum |

### Thread Safety

All methods are safe for concurrent use:
- ✅ Safe for concurrent RegisterIncrement/RegisterReset calls
- ✅ Internal state protected by atomic operations
- ✅ No mutexes required for concurrent access

---

## Best Practices

### Resource Management

**Always close resources**:
```go
// ✅ Good
func processFile(path string) error {
    fpg, err := progress.Open(path)
    if err != nil {
        return err
    }
    defer fpg.Close()  // Ensure file is closed
    
    bw := bandwidth.New(size.SizeMiB)
    bw.RegisterIncrement(fpg, nil)
    
    return processData(fpg)
}

// ❌ Bad
func processBad(path string) {
    fpg, _ := progress.Open(path)  // Never closed!
    bw := bandwidth.New(size.SizeMiB)
    bw.RegisterIncrement(fpg, nil)
    processData(fpg)
}
```

### Bandwidth Limit Selection

**Choose appropriate limits**:
```go
// Fast local disk
bw := bandwidth.New(100 * size.SizeMega)  // 100 MB/s

// Network connection (1 Mbps)
bw := bandwidth.New(125 * size.SizeKilo)  // 125 KB/s ≈ 1 Mbps

// Slow network / cloud backup
bw := bandwidth.New(500 * size.SizeKilo)  // 500 KB/s

// Unlimited (no throttling)
bw := bandwidth.New(0)
```

### Error Handling

**Check all errors**:
```go
// ✅ Good
fpg, err := progress.Open(path)
if err != nil {
    return fmt.Errorf("open failed: %w", err)
}

n, err := io.Copy(dest, fpg)
if err != nil {
    return fmt.Errorf("copy failed: %w", err)
}

// ❌ Bad
fpg, _ := progress.Open(path)  // Ignoring errors!
io.Copy(dest, fpg)
```

### Concurrency

**One instance per goroutine or shared**:
```go
// ✅ Good: Shared instance for aggregate limiting
sharedBW := bandwidth.New(5 * size.SizeMega)

for _, file := range files {
    go func(f string) {
        fpg, _ := progress.Open(f)
        defer fpg.Close()
        sharedBW.RegisterIncrement(fpg, nil)
        io.Copy(dest, fpg)
    }(file)
}

// ✅ Good: Separate instances for independent limiting
for _, file := range files {
    go func(f string) {
        bw := bandwidth.New(size.SizeMega)  // Per-file limit
        fpg, _ := progress.Open(f)
        defer fpg.Close()
        bw.RegisterIncrement(fpg, nil)
        io.Copy(dest, fpg)
    }(file)
}
```

### Testing

The package includes a comprehensive test suite with **84.4% code coverage** and race detection. All tests pass with `-race` flag enabled.

**Quick test commands:**
```bash
go test ./...                          # Run all tests
go test -cover ./...                   # With coverage
CGO_ENABLED=1 go test -race ./...      # With race detection
```

See **[TESTING.md](TESTING.md)** for comprehensive testing documentation.

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

- ✅ **84.4% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** with atomic operations
- ✅ **Memory-safe** with proper resource management
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**Algorithm Improvements:**
1. Token bucket algorithm for more precise rate limiting
2. Configurable burst allowance for transient spikes
3. Moving average calculation for smoother limiting
4. Adaptive rate adjustment based on system load

**Feature Additions:**
1. Multiple rate limits (e.g., per-second and per-minute)
2. Dynamic limit adjustment during runtime
3. Rate limiting statistics and reporting
4. Integration with system network QoS

**API Extensions:**
1. Rate limit getter method for monitoring
2. Pause/resume functionality
3. Bandwidth usage statistics
4. Event hooks for limit exceeded

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/file/bandwidth)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, rate limiting algorithm, buffer sizing, performance considerations, and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 84.4% coverage analysis, and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/file/progress](https://pkg.go.dev/github.com/nabbar/golib/file/progress)** - Progress tracking for file I/O operations. The bandwidth package integrates seamlessly with progress for rate-limited file transfers with real-time monitoring.

- **[github.com/nabbar/golib/size](https://pkg.go.dev/github.com/nabbar/golib/size)** - Size constants and utilities (KiB, MiB, GiB, etc.) used for configuring bandwidth limits. Provides type-safe size constants to avoid magic numbers.

### Standard Library References

- **[io](https://pkg.go.dev/io)** - Standard I/O interfaces. The bandwidth package works with `io.Reader`, `io.Writer`, and `io.Copy` for seamless integration with Go's I/O ecosystem.

- **[sync/atomic](https://pkg.go.dev/sync/atomic)** - Atomic operations used for lock-free timestamp storage. Understanding atomic operations helps in appreciating the thread-safety guarantees.

- **[time](https://pkg.go.dev/time)** - Time operations for rate calculation and sleep delays. The package uses `time.Since()` and `time.Sleep()` for rate limiting implementation.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and concurrency patterns. The bandwidth package follows these conventions for idiomatic Go code.

- **[Rate Limiting](https://en.wikipedia.org/wiki/Rate_limiting)** - Wikipedia article explaining rate limiting concepts, algorithms, and use cases. Provides background on the general approach to rate limiting.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the bandwidth package. Check existing issues before creating new ones.

- **[Contributing Guide](../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements, testing procedures, and pull request process.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/file/bandwidth`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
