# File Descriptor Limit Management

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-85.7%25-brightgreen)](TESTING.md)

Cross-platform file descriptor limit management for Go applications with automatic platform detection and safe limit modification.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Platform Behavior](#platform-behavior)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Query](#basic-query)
  - [Increase Limit](#increase-limit)
  - [Server Initialization](#server-initialization)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [SystemFileDescriptor Function](#systemfiledescriptor-function)
  - [Configuration](#configuration)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **fileDescriptor** package provides a unified, cross-platform API for managing file descriptor limits (the maximum number of open files or I/O resources) in Go applications. It abstracts platform differences between Unix/Linux, macOS, and Windows, offering a simple interface for both querying current limits and safely increasing them when needed.

### Design Philosophy

1. **Platform Abstraction**: Single API works across Unix/Linux, macOS, and Windows without conditional compilation in user code
2. **Safety First**: Never decreases limits, respects system constraints, handles permission errors gracefully
3. **Minimal Interface**: One function does everything - query, validate, and modify limits
4. **Zero Dependencies**: Only standard library and platform-specific syscalls
5. **Zero Overhead**: No runtime cost after initial setup, no memory allocations

### Key Features

- ✅ **Cross-Platform**: Unified API for Unix/Linux, macOS, and Windows
- ✅ **Simple Interface**: Single function call for all operations
- ✅ **Safe Operations**: Respects system hard limits, never decreases existing limits
- ✅ **Permission Aware**: Gracefully handles privilege requirements
- ✅ **Thread-Safe**: Kernel-level synchronization, no application-level locks needed
- ✅ **Production Ready**: 85.7% test coverage, 38 comprehensive specs
- ✅ **Zero Allocations**: No heap allocations, ~170ns per call
- ✅ **Well Documented**: Comprehensive GoDoc, examples, and testing guide

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────┐
│        SystemFileDescriptor(newValue)           │
│              Public API                         │
└──────────────────┬──────────────────────────────┘
                   │
         ┌─────────▼─────────┐
         │  Build Tag Check  │
         └─────────┬─────────┘
                   │
       ┌───────────┴───────────┐
       │                       │
┌──────▼──────┐        ┌──────▼──────┐
│Unix/Linux   │        │  Windows    │
│  macOS      │        │             │
├─────────────┤        ├─────────────┤
│syscall.     │        │maxstdio.    │
│ Getrlimit   │        │ GetMaxStdio │
│ Setrlimit   │        │ SetMaxStdio │
│RLIMIT_NOFILE│        │  (max 8192) │
└─────────────┘        └─────────────┘
```

**Build tags** ensure platform-appropriate implementation:
- `fileDescriptor.go`: Public API (all platforms)
- `fileDescriptor_ok.go`: Unix/Linux/macOS (`!windows`)
- `fileDescriptor_ko.go`: Windows (`windows`)

### Data Flow

```
User Application
       ↓
SystemFileDescriptor(newValue)
       ↓
┌──────────────────────┐
│ 1. Query current     │
│    limits (syscall)  │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐     Yes    ┌────────────────┐
│ 2. newValue <= 0 or  ├───────────→│ Return current │
│    <= current?       │            │ (no change)    │
└──────────┬───────────┘            └────────────────┘
           │ No
           ▼
┌──────────────────────┐
│ 3. Check privileges  │
│    and limits        │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ 4. Attempt increase  │
│    (Setrlimit/       │
│     SetMaxStdio)     │
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│ 5. Return new limits │
│    or error          │
└──────────────────────┘
```

### Platform Behavior

| Aspect | Unix/Linux/macOS | Windows |
|--------|------------------|---------|
| **Implementation** | `syscall.Rlimit` | `maxstdio` (C runtime) |
| **Default Limit** | 1024-4096 | 512 |
| **Hard Limit** | 4096-unlimited | 8192 (fixed) |
| **Privilege Required** | Root for hard limit increase | None (within 8192) |
| **Decrease Allowed** | No | No |
| **Thread Safety** | Yes (kernel-level) | Yes (CRT-level) |

---

## Performance

### Benchmarks

Based on actual benchmark results on AMD Ryzen 9 7900X3D:

| Operation | Performance | Allocations | Notes |
|-----------|-------------|-------------|-------|
| **Query** | 172.6 ns/op | 0 B/op, 0 allocs/op | Single syscall |
| **Query (parallel)** | 201.0 ns/op | 0 B/op, 0 allocs/op | Thread-safe |
| **With error check** | 169.9 ns/op | 0 B/op, 0 allocs/op | Full validation |
| **Sequential calls** | 342.7 ns/op | 0 B/op, 0 allocs/op | 2 queries |
| **Throughput** | ~6M queries/sec | - | Single thread |

Run benchmarks:
```bash
go test -bench . -benchmem
```

### Memory Usage

- **Package Size**: ~10 KB compiled
- **Runtime Memory**: 0 bytes (no state maintained)
- **Heap Allocations**: 0 per call
- **System Memory**: Managed by OS kernel

### Scalability

- **Concurrent Writers**: Tested with 100 concurrent goroutines
- **Test Coverage**: 38 specs with concurrency and performance tests
- **Zero Race Conditions**: All tests pass with `-race` detector
- **Thread Safety**: Natural thread-safety via kernel-level synchronization

Performance is not a concern as the function is typically called once during application startup.

---

## Use Cases

This package is essential for applications managing many concurrent I/O operations.

### 1. High-Traffic Web Servers

**Scenario**: Web servers handling thousands of concurrent connections need sufficient file descriptors.

**Problem**: Each HTTP connection requires a file descriptor. Default system limits (1024) are insufficient for high-traffic servers.

**Solution**: Increase limits at startup to handle expected concurrent connections plus overhead for logs, databases, and temporary files.

**Benefits**: Prevents "too many open files" errors, enables proper scaling, predictable performance under load.

### 2. Database Connection Pools

**Scenario**: Applications maintaining large connection pools to databases, caches, and message queues.

**Problem**: Each connection consumes a file descriptor. Large pools quickly exhaust default limits.

**Solution**: Calculate required descriptors (pools + overhead) and set appropriate limits before initializing pools.

**Benefits**: Allows larger connection pools, better resource utilization, prevents connection failures.

### 3. File Processing Pipelines

**Scenario**: Batch processing applications opening many files simultaneously for parallel processing.

**Problem**: Opening hundreds or thousands of files concurrently for reading/writing exceeds default limits.

**Solution**: Check available descriptors before processing, increase if needed, maintain safety margin.

**Benefits**: Enables parallel processing, improves throughput, graceful handling of large batches.

### 4. Microservices with Multiple Backends

**Scenario**: Microservices connecting to multiple databases, caches, queues, and downstream services.

**Problem**: Sum of all persistent connections (DB pools, cache connections, queue consumers, HTTP clients) exceeds limits.

**Solution**: Calculate total connections needed across all services and set limits accordingly.

**Benefits**: Stable operation under load, predictable resource usage, prevents cascade failures.

### 5. Network Proxies and Load Balancers

**Scenario**: Proxies routing traffic between clients and backends, potentially doubling file descriptor usage.

**Problem**: Each proxied connection uses two descriptors (client + backend), rapidly exhausting limits.

**Solution**: Calculate maximum concurrent proxied connections and set limits to 2× plus overhead.

**Benefits**: High concurrent connection capacity, stable proxy operation, predictable scaling limits.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/ioutils/fileDescriptor
```

### Basic Query

Query current file descriptor limits without modification:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func main() {
    // Query without modification (newValue = 0)
    current, max, err := fileDescriptor.SystemFileDescriptor(0)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Current (soft) limit: %d\n", current)
    fmt.Printf("Maximum (hard) limit: %d\n", max)
}
```

### Increase Limit

Attempt to increase file descriptor limit:

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func main() {
    // Attempt to increase to 4096
    desired := 4096
    current, max, err := fileDescriptor.SystemFileDescriptor(desired)
    if err != nil {
        log.Fatalf("Cannot increase limit: %v (may need elevated privileges)", err)
    }
    
    fmt.Printf("Limit set to: %d (max: %d)\n", current, max)
}
```

### Server Initialization

Initialize server with required file descriptor limits:

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

const (
    MinRequiredFDs = 4096
    PreferredFDs   = 16384
)

func initializeServer() error {
    // Check current limits
    current, max, err := fileDescriptor.SystemFileDescriptor(0)
    if err != nil {
        return fmt.Errorf("cannot check file descriptor limits: %w", err)
    }
    
    fmt.Printf("Initial limits - Current: %d, Max: %d\n", current, max)
    
    // Verify minimum requirement
    if current < MinRequiredFDs {
        return fmt.Errorf("insufficient file descriptors: have %d, need %d", current, MinRequiredFDs)
    }
    
    // Try to reach preferred limit
    if current < PreferredFDs && max >= PreferredFDs {
        newCurrent, newMax, err := fileDescriptor.SystemFileDescriptor(PreferredFDs)
        if err != nil {
            log.Printf("Warning: Could not set preferred limit: %v", err)
        } else {
            fmt.Printf("Increased limits - Current: %d, Max: %d\n", newCurrent, newMax)
        }
    }
    
    return nil
}

func main() {
    if err := initializeServer(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Server initialized successfully")
    // Start server...
}
```

**Complete examples** available in [example_test.go](example_test.go).

---

## Best Practices

### 1. Initialize Early

Set limits during application startup before opening connections:

```go
func main() {
    // Set limits FIRST
    fileDescriptor.SystemFileDescriptor(8192)
    // Then start application
    startServer()
}
```

### 2. Handle Permissions Gracefully

Don't fail hard if limit increase is denied - continue with available limits:

```go
current, max, err := fileDescriptor.SystemFileDescriptor(16384)
if err != nil {
    log.Printf("Warning: Cannot increase limit: %v", err)
    log.Printf("Continuing with %d FDs (max: %d)", current, max)
}
```

### 3. Reserve Safety Margin

Never use all available file descriptors - reserve for logs, temporary files:

```go
current, _, _ := fileDescriptor.SystemFileDescriptor(0)
maxConnections := current - 200  // Reserve 200 for overhead
```

### 4. Verify Before Proceeding

Check limits meet requirements before starting services:

```go
const RequiredFDs = 4096
current, _, _ := fileDescriptor.SystemFileDescriptor(RequiredFDs)
if current < RequiredFDs {
    return fmt.Errorf("insufficient FDs: need %d, have %d", RequiredFDs, current)
}
```

### 5. Platform-Aware Code

Adjust expectations based on platform capabilities:

```go
import "runtime"

targetFDs := 65536
if runtime.GOOS == "windows" {
    targetFDs = 8192  // Windows maximum
}
fileDescriptor.SystemFileDescriptor(targetFDs)
```

### 6. Document Requirements

State file descriptor requirements clearly in documentation:

```go
// Server requires minimum 4096 file descriptors
// Recommended: 16384 for high-traffic deployments
// Unix: ulimit -n 16384 && ./server
// Windows: Automatically limited to 8192
```

### 7. Monitor and Log

Track limit modifications for debugging:

```go
original, _, _ := fileDescriptor.SystemFileDescriptor(0)
current, max, _ := fileDescriptor.SystemFileDescriptor(8192)
if current != original {
    log.Printf("FD limit: %d → %d (max: %d)", original, current, max)
}
```

### Testing

Comprehensive test suite with 38 specs covering:
- Basic functionality (15 specs)
- Platform-specific behavior (8 specs)
- Concurrency (9 specs)
- Performance (6 specs)

**Coverage**: 85.7% | **Tests**: 38 specs | **Race-Free**: ✅

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## API Reference

### SystemFileDescriptor Function

```go
func SystemFileDescriptor(newValue int) (current int, max int, err error)
```

The unified function for querying and modifying file descriptor limits.

**Parameters:**
- `newValue int`: Desired new limit
  - `<= 0`: Query mode (no modification)
  - `> 0 and <= current`: Query mode (already at or above)
  - `> current`: Attempt to increase to `newValue`

**Returns:**
- `current int`: Current (soft) file descriptor limit
- `max int`: Maximum (hard) file descriptor limit
- `err error`: Error if operation fails (nil on success)

**Behavior:**

| newValue | Action | Privileges Required |
|----------|--------|---------------------|
| `<= 0` | Query only | None |
| `<= current` | Query only | None |
| `> current && <= hard` | Increase soft limit | Usually none (Unix) / None (Windows) |
| `> hard` | Increase hard limit | Root/admin (Unix) / Auto-cap at 8192 (Windows) |

### Configuration

No configuration required. The package uses build tags to automatically select the appropriate platform implementation.

**Platform Detection:**
- Unix/Linux/macOS: Uses `syscall.RLIMIT_NOFILE`
- Windows: Uses `maxstdio` C runtime functions

**Thread Safety:**
- All operations are thread-safe
- Kernel-level synchronization (Unix)
- C runtime thread-safety (Windows)

### Error Handling

**Error Types:**

1. **Permission Denied**
   - Cause: Attempting to exceed hard limit without privileges
   - Solution: Run with elevated privileges or reduce requested limit
   - Unix: `sudo` or adjust `/etc/security/limits.conf`

2. **System Error**
   - Cause: Underlying syscall failure (rare)
   - Solution: Check system logs, verify kernel configuration

3. **Platform Limit**
   - Cause: Exceeding platform maximum (Windows: 8192)
   - Solution: Auto-capped, no action needed

**Error Handling Pattern:**

```go
current, max, err := fileDescriptor.SystemFileDescriptor(desired)
if err != nil {
    // Check if it's a permission issue
    if current < desired && max >= desired {
        log.Println("Insufficient privileges to increase limit")
    } else if max < desired {
        log.Printf("System maximum (%d) below requested (%d)", max, desired)
    }
    // Fallback: use current limit
}
```

**Platform-Specific Notes:**

**Unix/Linux/macOS:**
- Check current limits: `ulimit -n` (soft), `ulimit -Hn` (hard)
- Persistent config: `/etc/security/limits.conf`
- systemd services: `LimitNOFILE=` in unit file

**Windows:**
- Default: 512, Maximum: 8192
- No persistent configuration
- No privileges required within 8192 limit

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**AI Usage Policy:**
- ✅ **AI may assist** with testing, documentation, and bug resolution
- ❌ **AI must NOT** generate package implementation code
- All core functionality must be human-designed and validated

**Code Contributions:**
- All contributions must pass tests: `go test ./...`
- Maintain test coverage ≥85%
- Follow existing code style and patterns
- Add GoDoc comments for all public elements

**Documentation:**
- Update README.md for new features
- Add practical examples for common use cases
- Keep TESTING.md synchronized with test changes
- Use clear, concise English

**Testing:**
- Write tests for all new features
- Test on multiple platforms when possible
- Handle platform-specific behavior appropriately
- Document privilege requirements

**Pull Requests:**
- Provide clear description of changes
- Reference related issues
- Include test results and coverage
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for project-wide guidelines.

---

## Resources

**Documentation:**
- [GoDoc Reference](https://pkg.go.dev/github.com/nabbar/golib/ioutils/fileDescriptor) - Complete API documentation with examples
- [Testing Guide](TESTING.md) - Comprehensive testing documentation with 38 specs
- [Unix getrlimit manual](https://man7.org/linux/man-pages/man2/getrlimit.2.html) - System call documentation for Unix/Linux/macOS
- [Windows SetMaxStdio documentation](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio) - C runtime function documentation for Windows

**Related Packages:**
- [maxstdio](../maxstdio) - Windows C runtime limit management (internal dependency)
- [bufferReadCloser](../bufferReadCloser) - I/O wrappers for file operations
- [ioutils](../) - Parent package with additional I/O utilities

**Community:**
- [GitHub Repository](https://github.com/nabbar/golib) - Source code and issue tracking
- [GitHub Issues](https://github.com/nabbar/golib/issues) - Bug reports and feature requests
- [Contributing Guide](../../CONTRIBUTING.md) - Project-wide contribution guidelines

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License © Nicolas JUHEL

All source files in this package are licensed under the MIT License. See [LICENSE](../../../../LICENSE) file for full text or individual files for the full license header.

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Test Coverage**: 85.7% (38 specs, 9 benchmarks)  
**Maintained By**: fileDescriptor Package Contributors
