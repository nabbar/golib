# File Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Advanced file management toolkit for Go with bandwidth throttling, permission handling, and progress tracking capabilities.

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
  - [bandwidth - Rate Limiting](#bandwidth-subpackage)
  - [perm - Permission Management](#perm-subpackage)
  - [progress - Progress Tracking](#progress-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The file package provides production-ready file management utilities for Go applications requiring advanced I/O control. It emphasizes real-time monitoring, bandwidth control, and cross-platform permission management through three specialized subpackages that integrate seamlessly while remaining independently usable.

### Design Philosophy

1. **Composable**: Independent subpackages that work together naturally
2. **Observable**: Real-time progress tracking with callback support
3. **Controlled**: Bandwidth throttling for network and disk I/O
4. **Portable**: Cross-platform permission management with multiple encoding formats
5. **Type-Safe**: Strong typing with comprehensive error handling

---

## Key Features

- **Bandwidth Throttling**: Rate-limit file I/O operations (bytes-per-second control)
- **Progress Tracking**: Real-time monitoring with increment, reset, and EOF callbacks
- **Permission Management**: Type-safe permissions with JSON/YAML/TOML/CBOR support
- **Viper Integration**: Seamless configuration file integration
- **Temporary Files**: Managed temporary and unique file creation
- **Thread-Safe**: Concurrent-safe operations across all subpackages
- **Zero Overhead**: Minimal performance impact when features not used
- **Standard Interfaces**: Implements all `io` standard interfaces

---

## Installation

```bash
go get github.com/nabbar/golib/file
```

Or install individual subpackages:

```bash
go get github.com/nabbar/golib/file/bandwidth
go get github.com/nabbar/golib/file/perm
go get github.com/nabbar/golib/file/progress
```

---

## Architecture

### Package Structure

The package is organized into three specialized subpackages with distinct responsibilities:

```
file/
├── bandwidth/           # Rate limiting and throttling
│   ├── interface.go     # BandWidth interface
│   └── model.go         # Implementation
├── perm/                # Permission management
│   ├── interface.go     # Perm type and parsing
│   ├── format.go        # Conversion methods
│   ├── encode.go        # Marshal/unmarshal
│   └── parse.go         # String parsing
└── progress/            # Progress tracking
    ├── interface.go     # Progress interface
    ├── model.go         # Core implementation
    ├── progress.go      # Callback management
    ├── ioreader.go      # Read operations
    ├── iowriter.go      # Write operations
    └── ioseeker.go      # Seek operations
```

### Component Overview

```
┌──────────────────────────────────────────────────┐
│              File Package Root                    │
│    Integrated file management toolkit            │
└──────┬────────────┬────────────┬─────────────────┘
       │            │            │
   ┌───▼────┐  ┌───▼────┐  ┌───▼────────┐
   │bandwidth│  │  perm  │  │  progress  │
   │         │  │        │  │            │
   │Rate     │  │Type-   │  │Callback-   │
   │limiting │  │safe    │  │based       │
   │         │  │perms   │  │monitoring  │
   └─────────┘  └────────┘  └────────────┘
```

| Component | Purpose | Coverage | Thread-Safe |
|-----------|---------|----------|-------------|
| **`bandwidth`** | I/O rate limiting | 77.8% | ✅ |
| **`perm`** | Permission handling | 88.9% | ✅ |
| **`progress`** | I/O monitoring | 71.1% | ✅ |
| **Overall** | File management | ~79% | ✅ |

### Integration Model

The subpackages integrate naturally:

```
┌────────────────────────────────────────┐
│         Application Layer              │
└──────────────┬─────────────────────────┘
               │
        ┌──────▼────────┐
        │   progress    │  ← File operations
        └──┬─────────┬──┘
           │         │
      ┌────▼───┐  ┌─▼────┐
      │bandwidth│  │ perm │
      │limiting │  │ mode │
      └─────────┘  └──────┘
```

---

## Performance

### Throughput Benchmarks

| Operation | Overhead | Notes |
|-----------|----------|-------|
| Progress tracking (no callbacks) | ~0% | Zero cost when unused |
| Progress tracking (with callbacks) | <2% | Negligible overhead |
| Bandwidth limiting | Variable | Depends on limit |
| Permission parsing | ~100ns | O(n) string length |
| Permission formatting | ~50ns | O(1) operation |

### Memory Efficiency

- **bandwidth**: Minimal (atomic operations only)
- **perm**: Zero heap allocations for conversions
- **progress**: Single file descriptor + atomics

### Thread Safety

All subpackages use lock-free atomic operations where possible:

- **Atomic Operations**: `atomic.Value`, `atomic.Int32`, `atomic.Bool`
- **Synchronization**: No mutexes in hot paths
- **Concurrent Safe**: Independent instances safe across goroutines

---

## Use Cases

This package is designed for scenarios requiring advanced file management:

**Backup Systems**
- Rate-limited backup operations to prevent disk saturation
- Real-time progress reporting for backup UI
- Cross-platform permission preservation

**File Transfer Applications**
- Bandwidth-controlled uploads/downloads
- Progress bars with percentage calculation
- Resume capability with position tracking

**Log Management**
- Managed log rotation with permission control
- Temporary file handling for log processing
- Progress tracking for log compression

**Configuration Management**
- Type-safe permission parsing from config files
- Viper integration for seamless configuration
- Validation of user-provided permissions

**Media Processing**
- Progress tracking for large media file operations
- Bandwidth limiting to prevent network congestion
- Temporary file management for transcoding

---

## Quick Start

### Basic File Operations with Progress

```go
package main

import (
    "fmt"
    "io"
    
    "github.com/nabbar/golib/file/progress"
)

func main() {
    // Open file with progress tracking
    p, err := progress.Open("largefile.dat")
    if err != nil {
        panic(err)
    }
    defer p.Close()
    
    // Track progress
    var totalBytes int64
    p.RegisterFctIncrement(func(bytes int64) {
        totalBytes = bytes
        fmt.Printf("\rRead: %d bytes", totalBytes)
    })
    
    p.RegisterFctEOF(func() {
        fmt.Println("\nReading complete!")
    })
    
    // Read file
    io.Copy(io.Discard, p)
}
```

### Bandwidth-Limited File Copy

```go
package main

import (
    "io"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

func main() {
    // Create bandwidth limiter (1 MB/s)
    bw := bandwidth.New(size.SizeMiB)
    
    // Open source with progress
    src, _ := progress.Open("source.dat")
    defer src.Close()
    
    // Create destination
    dst, _ := progress.Create("dest.dat")
    defer dst.Close()
    
    // Apply bandwidth limiting
    bw.RegisterIncrement(src, func(bytes int64) {
        fmt.Printf("\rCopied: %d bytes", bytes)
    })
    
    // Copy with rate limiting
    io.Copy(dst, src)
}
```

### Permission Management

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/nabbar/golib/file/perm"
)

func main() {
    // Parse permission from string
    p, err := perm.Parse("0644")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Octal:", p.String())        // Output: 0644
    fmt.Println("Decimal:", p.Uint64())      // Output: 420
    fmt.Println("FileMode:", p.FileMode())   // Output: -rw-r--r--
    
    // Use with file operations
    file, _ := os.OpenFile("data.txt", os.O_CREATE|os.O_WRONLY, p.FileMode())
    defer file.Close()
}
```

### Complete Integration Example

Combining all three subpackages:

```go
package main

import (
    "fmt"
    "io"
    "os"
    
    "github.com/nabbar/golib/file/bandwidth"
    "github.com/nabbar/golib/file/perm"
    "github.com/nabbar/golib/file/progress"
    "github.com/nabbar/golib/size"
)

func main() {
    // Parse file permissions
    perms, _ := perm.Parse("0644")
    
    // Open source file
    src, _ := progress.Open("video.mp4")
    defer src.Close()
    
    // Create destination with permissions
    dst, _ := progress.New("output.mp4", os.O_CREATE|os.O_WRONLY, perms.FileMode())
    defer dst.Close()
    
    // Set up bandwidth limiting (10 MB/s)
    bw := bandwidth.New(size.Size(10 * 1024 * 1024))
    bw.RegisterIncrement(src, nil)
    
    // Track progress
    fileInfo, _ := src.Stat()
    fileSize := fileInfo.Size()
    var copied int64
    
    src.RegisterFctIncrement(func(bytes int64) {
        copied = bytes
        percent := float64(copied) / float64(fileSize) * 100
        fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", 
            percent, copied, fileSize)
    })
    
    src.RegisterFctEOF(func() {
        fmt.Println("\nTransfer complete!")
    })
    
    // Copy with all features active
    io.Copy(dst, src)
}
```

---

## Subpackages

### `bandwidth` Subpackage

Rate limiting and throttling for file I/O operations.

**Purpose**: Control file transfer rates to prevent resource saturation.

**Key Features**
- Configurable bytes-per-second limits
- Seamless progress integration
- Zero-cost when set to unlimited
- Time-based throttling with sleep intervals
- Thread-safe atomic operations

**API Overview**

```go
// Create limiter
bw := bandwidth.New(size.SizeMiB) // 1 MB/s

// Register with progress-enabled file
bw.RegisterIncrement(fpg, func(bytes int64) {
    // Optional progress callback
})

bw.RegisterReset(fpg, func(max, current int64) {
    // Optional reset callback
})
```

**Use Cases**
- Network transfer rate limiting
- Disk I/O throttling
- Multi-tenant resource sharing
- QoS enforcement

**Performance**: <1μs overhead per operation

**Thread Safety**: Safe for concurrent use

---

### `perm` Subpackage

Type-safe, portable file permission handling.

**Purpose**: Cross-platform permission management with validation and encoding.

**Key Features**
- Octal string parsing (e.g., "0644", "0755")
- Multiple format encoding (JSON, YAML, TOML, CBOR, Text)
- Type conversions (int, uint, FileMode)
- Viper decoder hook for configuration files
- Special permissions support (setuid, setgid, sticky bit)
- Quote handling and validation

**API Overview**

```go
// Parsing
p, err := perm.Parse("0644")        // From string
p, err := perm.ParseInt(420)        // From decimal
p, err := perm.ParseByte([]byte("0755")) // From bytes

// Conversion
mode := p.FileMode()                // os.FileMode
str := p.String()                   // "0644"
num := p.Uint64()                   // 420

// Encoding
data, _ := json.Marshal(p)          // JSON
data, _ := yaml.Marshal(p)          // YAML
```

**Common Permissions**

| Permission | Octal | Usage |
|------------|-------|-------|
| Private file | `0600` | User read/write only |
| Standard file | `0644` | User write, all read |
| Executable | `0755` | Standard executable |
| Group writable | `0664` | Group write access |
| Directory | `0755` | Standard directory |

**Performance**
- Parsing: ~100ns (O(n) where n = string length)
- Formatting: ~50ns (O(1))
- Zero heap allocations for most operations

**Thread Safety**: Immutable type, safe for concurrent reads

---

### `progress` Subpackage

File I/O with real-time progress tracking and callbacks.

**Purpose**: Monitor file operations with event-driven architecture.

**Key Features**
- Increment callbacks (bytes read/written)
- Reset callbacks (operation restart)
- EOF callbacks (completion detection)
- Configurable buffer sizes
- Temporary file management
- Full `io` interface implementation
- Position tracking (BOF/EOF)

**API Overview**

```go
// File creation
p, _ := progress.New(name, flags, perm)  // Custom flags/perms
p, _ := progress.Open(name)              // Read existing
p, _ := progress.Create(name)            // Create new
p, _ := progress.Temp(pattern)           // Temporary file
p, _ := progress.Unique(dir, pattern)    // Unique file

// Callback registration
p.RegisterFctIncrement(func(bytes int64) {
    // Called on each I/O operation
})

p.RegisterFctReset(func(max, current int64) {
    // Called on progress reset
})

p.RegisterFctEOF(func() {
    // Called at end of file
})

// Buffer management
p.SetBufferSize(64 * 1024)               // 64KB buffer

// File operations
info, _ := p.Stat()                      // File info
pos, _ := p.SizeBOF()                    // Current position
remaining, _ := p.SizeEOF()              // Bytes to EOF
p.Truncate(size)                         // Truncate
p.Sync()                                 // Sync to disk
p.Close()                                // Close file
p.CloseDelete()                          // Close and delete
```

**Callback Signatures**

```go
type FctIncrement func(size int64)           // Progress update
type FctReset func(size, current int64)      // Reset event
type FctEOF func()                           // End of file
```

**Use Cases**
- Progress bars and UI updates
- Transfer rate calculation
- Resume capability
- Logging and monitoring
- Resource cleanup

**Performance**: Near-native I/O with <2% overhead when callbacks used

**Thread Safety**: Each instance independent, safe across goroutines

---

## Best Practices

### Always Close Resources

```go
// ✅ Good: Proper cleanup
p, err := progress.Open("file.txt")
if err != nil {
    return err
}
defer p.Close()
```

### Handle Errors Explicitly

```go
// ✅ Good: Check all errors
p, err := perm.Parse(userInput)
if err != nil {
    return fmt.Errorf("invalid permission: %w", err)
}

file, err := progress.Create("data.txt")
if err != nil {
    return fmt.Errorf("create file: %w", err)
}
defer file.Close()
```

### Use Appropriate Buffer Sizes

```go
// ✅ Good: Tune for workload
p, _ := progress.Open("huge-file.dat")
p.SetBufferSize(256 * 1024)  // 256KB for large files

p2, _ := progress.Open("small.txt")
p2.SetBufferSize(4 * 1024)   // 4KB for small files
```

### Safe Bandwidth Limiting

```go
// ✅ Good: Reasonable limits
bw := bandwidth.New(size.SizeMiB * 10)  // 10 MB/s

// ✅ Good: Unlimited when needed
bw := bandwidth.New(0)  // No limit

// ⚠️  Caution: Very low limits may impact performance
bw := bandwidth.New(1024)  // 1 KB/s - very slow
```

### Validate Permissions

```go
// ✅ Good: Validate user input
func setFilePerms(path, permStr string) error {
    p, err := perm.Parse(permStr)
    if err != nil {
        return fmt.Errorf("invalid permission: %w", err)
    }
    
    return os.Chmod(path, p.FileMode())
}

// ❌ Bad: No validation
func setFilePermsBad(path, permStr string) error {
    mode := os.FileMode(0777)  // Hardcoded, insecure
    return os.Chmod(path, mode)
}
```

### Use Callbacks Wisely

```go
// ✅ Good: Lightweight callbacks
p.RegisterFctIncrement(func(bytes int64) {
    atomic.StoreInt64(&totalBytes, bytes)
})

// ❌ Bad: Heavy operations in callbacks
p.RegisterFctIncrement(func(bytes int64) {
    // Don't do slow operations here
    database.UpdateProgress(bytes)  // Blocking!
    sendNotification(bytes)         // Network call!
})
```

### Temporary File Cleanup

```go
// ✅ Good: Auto-delete temp files
temp, _ := progress.Temp("process-*.tmp")
defer temp.CloseDelete()  // Deletes on close

// ✅ Good: Manual cleanup if needed
temp, _ := progress.Temp("backup-*.tmp")
defer func() {
    if shouldKeep {
        temp.Close()
    } else {
        temp.CloseDelete()
    }
}()
```

---

## Testing

The file package uses **Ginkgo v2** and **Gomega** for comprehensive testing.

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection
CGO_ENABLED=1 go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Test Summary**

| Subpackage | Specs | Coverage | Status |
|------------|-------|----------|--------|
| bandwidth | 25 | 77.8% | ✅ Pass |
| perm | 141 | 88.9% | ✅ Pass |
| progress | 89 | 71.1% | ✅ Pass |
| **Total** | **255** | **~79%** | ✅ Pass |

**Coverage Areas**
- File creation and operations
- Progress tracking with callbacks
- Bandwidth limiting and throttling
- Permission parsing and encoding
- Error handling and edge cases
- Temporary file management
- Concurrent operations

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI** to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥75%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Use English for all documentation and comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Bandwidth Features**
- Adaptive throttling based on system load
- Burst mode support
- Multiple bandwidth profiles
- Bandwidth pooling across multiple files

**Permission Features**
- Symbolic notation support (e.g., "rwxr-xr-x")
- Permission comparison and validation
- ACL (Access Control List) support
- Windows security descriptor integration

**Progress Features**
- Async callbacks with goroutines
- Rate calculation (bytes/second)
- ETA (Estimated Time of Arrival) calculation
- Progress persistence for resume capability
- Multiple progress listeners

**Integration**
- Cloud storage progress tracking (S3, GCS, Azure)
- Network stream progress monitoring
- Compression progress tracking
- Encryption progress tracking

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/file)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Related Packages**:
  - [`github.com/nabbar/golib/size`](../size) - Size constants and utilities
  - [`github.com/nabbar/golib/archive`](../archive) - Archive and compression
