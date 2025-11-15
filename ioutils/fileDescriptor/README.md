# FileDescriptor Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-85.7%25-brightgreen)]()

Cross-platform file descriptor limit management for Go applications with automatic platform detection and safe limit modification.

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
- [Platform Support](#platform-support)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides a unified, cross-platform API for managing file descriptor limits (the maximum number of open files or I/O resources) in Go applications. It abstracts platform differences between Unix/Linux, macOS, and Windows, offering a simple interface for both querying current limits and safely increasing them when needed.

### Design Philosophy

1. **Platform Abstraction**: Single API works across Unix/Linux, macOS, and Windows
2. **Safety First**: Never decreases limits, respects system constraints
3. **Graceful Degradation**: Handles permission errors without panic
4. **Minimal Interface**: One function does everything
5. **Zero Dependencies**: Only standard library and platform-specific syscalls

---

## Key Features

- **Cross-Platform**: Unified API for Unix/Linux, macOS, and Windows
- **Simple Interface**: Single function call for all operations
- **Safe Operations**: Respects system hard limits, never decreases existing limits
- **Permission Aware**: Gracefully handles privilege requirements
- **Zero Overhead**: No runtime cost after initial setup
- **Production Ready**: 85.7% test coverage, 23 specs
- **Well Documented**: Comprehensive GoDoc and examples

---

## Installation

```bash
go get github.com/nabbar/golib/ioutils/fileDescriptor
```

---

## Architecture

### Package Structure

The package uses build tags to provide platform-specific implementations:

```
fileDescriptor/
├── fileDescriptor.go       # Public API and package documentation
├── fileDescriptor_ok.go    # Unix/Linux/macOS implementation (!windows)
└── fileDescriptor_ko.go    # Windows implementation (windows)
```

### Platform-Specific Implementations

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

### Operation Flow

```
User calls SystemFileDescriptor(newValue)
       ↓
1. Query current limits (platform-specific)
       ↓
2. newValue <= 0 or newValue <= current?
   ├─ Yes → Return current limits (no modification)
   └─ No  → Continue to step 3
       ↓
3. newValue > hard limit?
   ├─ Unix: May require root privileges
   └─ Windows: Cap at 8192
       ↓
4. Attempt to increase limit (syscall/maxstdio)
       ↓
5. Return new limits or error
```

### Platform Behavior Comparison

| Aspect | Unix/Linux/macOS | Windows |
|--------|------------------|---------|
| **Implementation** | `syscall.Rlimit` | `maxstdio` (C runtime) |
| **Default Limit** | 1024-4096 | 512 |
| **Hard Limit** | 4096-unlimited | 8192 (fixed) |
| **Privilege Required** | Root for hard limit increase | None (within 8192) |
| **Decrease Allowed** | No | No |
| **Thread Safety** | Yes (kernel-level) | Yes (CRT-level) |

---

## Quick Start

### Query Current Limits

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
    // Output example (Linux):
    // Current (soft) limit: 1024
    // Maximum (hard) limit: 65536
}
```

### Increase Limit

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func main() {
    // Attempt to increase to 4096
    newLimit := 4096
    current, max, err := fileDescriptor.SystemFileDescriptor(newLimit)
    if err != nil {
        log.Fatalf("Cannot increase limit: %v (may need elevated privileges)", err)
    }
    
    fmt.Printf("Limit increased to: %d (max: %d)\n", current, max)
}
```

### Initialize Server with Required Limits

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

const RequiredFDs = 8192

func main() {
    // Ensure minimum file descriptors at startup
    current, max, err := fileDescriptor.SystemFileDescriptor(RequiredFDs)
    if err != nil {
        log.Fatalf("Cannot set required FD limit: %v", err)
    }
    
    if current < RequiredFDs {
        log.Fatalf("Insufficient FDs: have %d, need %d (max: %d)", current, RequiredFDs, max)
    }
    
    fmt.Printf("Server starting with %d file descriptors available\n", current)
    // Start server...
}
```

---

## Performance

### Operation Metrics

The package adds **minimal overhead** to your application:

| Operation | Time | Overhead | Notes |
|-----------|------|----------|-------|
| Query (`newValue=0`) | ~500 ns | Negligible | Single syscall |
| Increase limit | ~1-5 µs | One-time cost | Syscall + validation |
| Subsequent calls | 0 | Zero overhead | Limits persist process-wide |

*Measured on Linux AMD64, Go 1.21*

### Performance Characteristics

- **No Runtime Overhead**: After setting limits, zero performance cost
- **One-Time Setup**: Typically called once during application startup
- **Fast Queries**: Sub-microsecond limit queries
- **Cached by OS**: System maintains limits, no memory overhead
- **Thread-Safe**: Kernel-level synchronization (no app-level locks needed)

### Benchmark Results

```go
BenchmarkQuery-8          2000000    537 ns/op     0 B/op    0 allocs/op
BenchmarkIncrease-8       500000     3421 ns/op    0 B/op    0 allocs/op
```

### Memory Usage

- **Package Size**: ~10 KB compiled
- **Runtime Memory**: 0 bytes (no state maintained)
- **Allocations**: 0 per call
- **System Memory**: Managed by OS kernel

---

## Use Cases

This package is essential for applications that need to manage many concurrent I/O operations:

### High-Traffic Web Servers

Web servers handling thousands of concurrent connections:

```go
package main

import (
    "log"
    "net/http"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func main() {
    // Increase limit for handling many concurrent connections
    const ServerFDs = 16384
    current, _, err := fileDescriptor.SystemFileDescriptor(ServerFDs)
    if err != nil || current < ServerFDs {
        log.Fatalf("Cannot set required FD limit for web server")
    }
    
    log.Printf("Server starting with %d file descriptors\n", current)
    http.ListenAndServe(":8080", nil)
}
```

**Why**: Each HTTP connection requires a file descriptor. High-traffic servers need many FDs.

### Database Connection Pools

Applications with large database connection pools:

```go
package main

import (
    "database/sql"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func initDatabase() (*sql.DB, error) {
    // Reserve FDs for DB connections + overhead
    poolSize := 100
    requiredFDs := poolSize + 500 // Pool + logs + other files
    
    current, _, err := fileDescriptor.SystemFileDescriptor(requiredFDs)
    if err != nil {
        return nil, err
    }
    
    if current < requiredFDs {
        log.Printf("Warning: Only %d FDs available, may limit connections", current)
    }
    
    db, _ := sql.Open("postgres", "connstring")
    db.SetMaxOpenConns(min(poolSize, current-500))
    return db, nil
}
```

**Why**: Each database connection consumes a file descriptor. Large pools need adequate limits.

### File Processing Pipelines

Batch processing many files simultaneously:

```go
package main

import (
    "fmt"
    "os"
    "sync"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func processFiles(files []string) error {
    // Check if we have enough FDs for parallel processing
    required := len(files) + 100 // Files + safety margin
    current, _, _ := fileDescriptor.SystemFileDescriptor(required)
    
    if current < required {
        return fmt.Errorf("need %d FDs to process %d files, have %d", 
            required, len(files), current)
    }
    
    var wg sync.WaitGroup
    for _, path := range files {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            f, _ := os.Open(p)
            defer f.Close()
            // Process file...
        }(path)
    }
    wg.Wait()
    return nil
}
```

**Why**: Opening many files concurrently requires sufficient file descriptor limit.

### Microservices with Many Connections

Services connecting to multiple backends:

```go
package main

import (
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

type Microservice struct {
    databases []Database
    caches    []Cache
    queues    []Queue
}

func (m *Microservice) Initialize() error {
    // Calculate total FDs needed
    totalConnections := 
        len(m.databases) * 20 +  // DB pools
        len(m.caches) * 10 +     // Cache connections
        len(m.queues) * 5 +      // Queue connections
        1000                     // Incoming HTTP requests
    
    current, max, err := fileDescriptor.SystemFileDescriptor(totalConnections)
    if err != nil {
        return err
    }
    
    log.Printf("Microservice initialized with %d/%d FDs\n", current, max)
    return nil
}
```

**Why**: Microservices often maintain many persistent connections to various services.

### Network Proxies and Load Balancers

Proxies routing traffic between clients and servers:

```go
package main

import (
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func startProxy() error {
    // Each proxied connection uses 2 FDs (client + backend)
    maxConnections := 10000
    requiredFDs := maxConnections * 2 + 1000 // Connections + overhead
    
    current, max, err := fileDescriptor.SystemFileDescriptor(requiredFDs)
    if err != nil {
        log.Printf("Warning: cannot increase FD limit: %v", err)
    }
    
    actualMax := min(current/2-500, maxConnections)
    log.Printf("Proxy starting: max %d concurrent connections (FDs: %d/%d)\n", 
        actualMax, current, max)
    
    // Start proxy with actualMax limit...
    return nil
}
```

**Why**: Proxies double file descriptor usage (one FD for client, one for backend).

---

## API Reference

### SystemFileDescriptor

```go
func SystemFileDescriptor(newValue int) (current int, max int, err error)
```

The single function that does everything - queries and optionally modifies file descriptor limits.

**Parameters:**
- `newValue int`: Desired new limit
  - `<= 0`: Query mode (no modification)
  - `> 0 and <= current`: Query mode (already at or above requested value)
  - `> current`: Attempt to increase to `newValue`

**Returns:**
- `current int`: Current (soft) file descriptor limit
- `max int`: Maximum (hard) file descriptor limit  
- `err error`: Error if the operation fails (nil on success)

**Behavior Summary:**

| newValue | Action | Privileges Required |
|----------|--------|---------------------|
| `<= 0` | Query only | None |
| `<= current` | Query only | None |
| `> current && <= hard` | Increase soft limit | Usually none (Unix) / None (Windows) |
| `> hard` | Increase hard limit | Root/admin (Unix) / Capped at 8192 (Windows) |

**Error Conditions:**

- **Permission Denied**: Attempting to increase beyond allowed limit without privileges
- **System Error**: Underlying syscall failure (rare)
- **Platform Limit**: On Windows, automatically caps at 8192

**Examples:**

```go
// Query current limits
current, max, _ := fileDescriptor.SystemFileDescriptor(0)

// Increase to 4096 (may fail without privileges)
current, max, err := fileDescriptor.SystemFileDescriptor(4096)
if err != nil {
    log.Printf("Cannot increase: %v", err)
}

// Safe increase with fallback
desired := 16384
current, max, err := fileDescriptor.SystemFileDescriptor(desired)
if err != nil || current < desired {
    log.Printf("Using %d FDs instead of %d", current, desired)
}
```

---

## Usage Examples

### Example 1: Check Before Opening Many Files

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func main() {
    filesNeeded := 2000
    
    current, max, err := fileDescriptor.SystemFileDescriptor(0)
    if err != nil {
        log.Fatal(err)
    }
    
    if current < filesNeeded {
        fmt.Printf("Need %d file descriptors, but only %d available\n", filesNeeded, current)
        fmt.Printf("Attempting to increase limit...\n")
        
        newCurrent, newMax, err := fileDescriptor.SystemFileDescriptor(filesNeeded)
        if err != nil {
            log.Fatalf("Cannot increase limit: %v (may need elevated privileges)", err)
        }
        
        fmt.Printf("Successfully increased limit to %d (max: %d)\n", newCurrent, newMax)
    } else {
        fmt.Printf("Current limit (%d) is sufficient\n", current)
    }
}
```

### Example 2: Initialize High-Performance Server

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

const (
    MinRequiredFDs = 4096
    PreferredFDs   = 65536
)

func initializeServer() error {
    current, max, err := fileDescriptor.SystemFileDescriptor(0)
    if err != nil {
        return fmt.Errorf("cannot check file descriptor limits: %w", err)
    }
    
    fmt.Printf("Initial limits - Current: %d, Max: %d\n", current, max)
    
    // Check minimum requirement
    if current < MinRequiredFDs {
        return fmt.Errorf("insufficient file descriptors: have %d, need at least %d", current, MinRequiredFDs)
    }
    
    // Try to set preferred limit
    if current < PreferredFDs && max >= PreferredFDs {
        fmt.Printf("Attempting to increase limit to %d...\n", PreferredFDs)
        
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
}
```

### Example 3: Safe Limit Increase with Fallback

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func ensureMinimumFDs(required int) (int, error) {
    current, max, err := fileDescriptor.SystemFileDescriptor(0)
    if err != nil {
        return 0, err
    }
    
    if current >= required {
        return current, nil
    }
    
    // Try to increase to required
    newCurrent, _, err := fileDescriptor.SystemFileDescriptor(required)
    if err != nil {
        // If we can't increase, check if max is sufficient
        if max >= required {
            return 0, fmt.Errorf("current limit is %d but need %d (max is %d, may need elevated privileges)", current, required, max)
        }
        return 0, fmt.Errorf("system maximum (%d) is below required (%d)", max, required)
    }
    
    return newCurrent, nil
}

func main() {
    limit, err := ensureMinimumFDs(2048)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("File descriptor limit ensured: %d\n", limit)
}
```

### Example 4: Graceful Degradation

```go
package main

import (
    "fmt"
    "log"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

type ServerConfig struct {
    TargetFDs      int
    MinimumFDs     int
    MaxConnections int
}

func configureServer(cfg *ServerConfig) error {
    current, _, err := fileDescriptor.SystemFileDescriptor(0)
    if err != nil {
        return err
    }
    
    // Try to reach target
    if current < cfg.TargetFDs {
        newCurrent, _, err := fileDescriptor.SystemFileDescriptor(cfg.TargetFDs)
        if err == nil {
            current = newCurrent
            log.Printf("Increased FD limit to %d", current)
        } else {
            log.Printf("Could not increase to target %d: %v", cfg.TargetFDs, err)
        }
    }
    
    // Check minimum
    if current < cfg.MinimumFDs {
        return fmt.Errorf("insufficient file descriptors: %d (need at least %d)", current, cfg.MinimumFDs)
    }
    
    // Adjust max connections based on available FDs
    // Reserve some FDs for other purposes (logs, databases, etc.)
    availableForConns := current - 100
    if availableForConns < cfg.MaxConnections {
        log.Printf("Reducing max connections from %d to %d due to FD limits", cfg.MaxConnections, availableForConns)
        cfg.MaxConnections = availableForConns
    }
    
    log.Printf("Server configured with %d FDs, max %d connections", current, cfg.MaxConnections)
    return nil
}

func main() {
    cfg := &ServerConfig{
        TargetFDs:      8192,
        MinimumFDs:     1024,
        MaxConnections: 5000,
    }
    
    if err := configureServer(cfg); err != nil {
        log.Fatal(err)
    }
}
```

### Example 5: Monitoring and Alerting

```go
package main

import (
    "fmt"
    "time"
    "github.com/nabbar/golib/ioutils/fileDescriptor"
)

func monitorFileDescriptors(interval time.Duration, warningThreshold float64) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    for range ticker.C {
        current, max, err := fileDescriptor.SystemFileDescriptor(0)
        if err != nil {
            fmt.Printf("Error checking FD limits: %v\n", err)
            continue
        }
        
        // Calculate usage percentage (rough estimate)
        // Note: current is the limit, not the actual usage
        usage := float64(current) / float64(max)
        
        fmt.Printf("FD Limits - Current: %d, Max: %d (%.1f%% of max)\n", current, max, usage*100)
        
        if usage >= warningThreshold {
            fmt.Printf("WARNING: File descriptor limit is high (%.1f%%)\n", usage*100)
            // Send alert, log, etc.
        }
    }
}

func main() {
    fmt.Println("Monitoring file descriptor limits...")
    monitorFileDescriptors(30*time.Second, 0.8) // Check every 30s, warn at 80%
}
```

---

## Platform Support

### Unix/Linux/macOS

**Implementation**: `syscall.Getrlimit` / `syscall.Setrlimit` with `RLIMIT_NOFILE`

**Typical Defaults:**
- Soft limit: 1024-4096 (distribution-dependent)
- Hard limit: 4096-unlimited (often 65536 or unlimited)

**Privilege Requirements:**
- Increase soft limit (≤ hard): No privileges needed
- Increase hard limit: Root (superuser) privileges required

**System Commands:**

```bash
# Check current limits
ulimit -n           # Soft limit (current)
ulimit -Hn          # Hard limit (maximum)

# Temporarily increase (current shell only)
ulimit -n 8192

# View process limits
cat /proc/<pid>/limits | grep "open files"

# Check system-wide max
cat /proc/sys/fs/file-max
```

**Persistent Configuration:**

Edit `/etc/security/limits.conf` (requires root):
```
*     soft  nofile  4096
*     hard  nofile  65536
root  soft  nofile  8192
root  hard  nofile  unlimited
```

**Distribution-Specific:**
- **Ubuntu/Debian**: Also check `/etc/pam.d/common-session`
- **RHEL/CentOS**: Also check `/etc/security/limits.d/*.conf`
- **systemd services**: Set `LimitNOFILE=` in service unit file

### Windows

**Implementation**: `maxstdio.GetMaxStdio` / `maxstdio.SetMaxStdio` (C runtime)

**Limits:**
- Default: 512 file descriptors
- Maximum: 8192 file descriptors (hard limit, cannot exceed)
- No privileges required to increase (within 8192)

**Important Differences:**
- Windows limits are per-process C runtime limits, not OS-level kernel limits
- Much lower than Unix systems (8192 vs potentially unlimited)
- Applies to C stdio functions and file handles opened with them
- Go's os.Open uses Windows API directly, but may still be affected

**Registry (advanced):**
```
HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Control\Session Manager\SubSystems
```
Modify "Windows" value (requires reboot, not recommended)

**Recommendation**: Design applications to work within 8192 FD limit on Windows.

---

## Testing

The package has **85.7% test coverage** with 23 comprehensive specs using Ginkgo v2 and Gomega.

### Run Tests

```bash
# Standard go test
go test -v -cover .

# With Ginkgo CLI (recommended)
go install github.com/onsi/ginkgo/v2/ginkgo@latest
ginkgo -v -cover
```

### Test Statistics

| Metric | Value |
|--------|-------|
| Total Specs | 23 |
| Passed | 20 |
| Skipped | 3 (permission/state dependent) |
| Coverage | 85.7% |
| Execution Time | ~2ms |

### Coverage Breakdown

| Component | File | Coverage | Notes |
|-----------|------|----------|-------|
| Public API | `fileDescriptor.go` | 100.0% | Fully tested |
| Unix impl | `fileDescriptor_ok.go` | 89.5% | Some branches need root |
| Windows impl | `fileDescriptor_ko.go` | N/A | Platform-specific |

**Why 3 tests are skipped**: Some tests require specific system states (already at max) or elevated privileges. This is expected and normal.

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Best Practices

### 1. Initialize Early in main()

Set limits before creating connections or opening files:

```go
func main() {
    // First: Set file descriptor limits
    if _, _, err := fileDescriptor.SystemFileDescriptor(8192); err != nil {
        log.Fatalf("Cannot set FD limit: %v", err)
    }
    
    // Then: Start application
    startServer()
}
```

### 2. Handle Permission Errors Gracefully

Don't fail hard if limit increase is denied:

```go
current, max, err := fileDescriptor.SystemFileDescriptor(16384)
if err != nil {
    log.Printf("Warning: Cannot increase FD limit: %v", err)
    log.Printf("Continuing with %d FDs (max: %d)", current, max)
}
```

### 3. Reserve Safety Margin

Don't use all available FDs:

```go
current, _, _ := fileDescriptor.SystemFileDescriptor(0)
maxConnections := current - 200  // Reserve for logs, DB, etc.
```

### 4. Check Before Requiring

Verify limits meet requirements before proceeding:

```go
const RequiredFDs = 4096

current, _, _ := fileDescriptor.SystemFileDescriptor(RequiredFDs)
if current < RequiredFDs {
    return fmt.Errorf("insufficient FDs: need %d, have %d", RequiredFDs, current)
}
```

### 5. Platform-Aware Limits

Adjust expectations based on platform:

```go
import "runtime"

targetFDs := 65536
if runtime.GOOS == "windows" {
    targetFDs = 8192  // Windows maximum
}
fileDescriptor.SystemFileDescriptor(targetFDs)
```

### 6. Document Requirements

Clearly state FD requirements in documentation:

```go
// Server requires at least 4096 file descriptors.
// Recommended: 16384 for high-traffic deployments.
// Run with: ulimit -n 16384 && ./server (Unix)
const (
    MinimumFDs    = 4096
    RecommendedFDs = 16384
)
```

### 7. Log Limit Changes

Track limit modifications for debugging:

```go
original, _, _ := fileDescriptor.SystemFileDescriptor(0)
current, max, err := fileDescriptor.SystemFileDescriptor(8192)
if err == nil && current != original {
    log.Printf("Increased FD limit: %d → %d (max: %d)", original, current, max)
}
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- **Do not use AI** to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass tests: `go test ./...`
- Maintain test coverage (≥85%)
- Follow existing code style and patterns
- Add GoDoc comments for all public elements

**Documentation**
- Update README.md for new features
- Add practical examples for common use cases
- Keep TESTING.md synchronized with test changes
- Use clear, concise English

**Testing**
- Write tests for all new features
- Test on multiple platforms when possible
- Handle platform-specific behavior appropriately
- Document privilege requirements

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
- Auto-adjustment: Detect optimal FD limit based on application needs
- Monitoring: Built-in FD usage tracking and alerts
- Cross-process: Coordinate limits across multiple processes
- Context support: Context-aware limit changes

**Platform Support**
- BSD variants: Explicit support and testing
- Solaris: Native support
- Plan 9: If feasible

**Developer Experience**
- CLI tool: `fdlimit check` and `fdlimit set` commands
- Metrics export: Prometheus/OpenMetrics format
- Validation: Pre-flight checks for deployment environments

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
- [GoDoc Reference](https://pkg.go.dev/github.com/nabbar/golib/ioutils/fileDescriptor)
- [Testing Guide](TESTING.md)
- [Unix getrlimit man page](https://man7.org/linux/man-pages/man2/getrlimit.2.html)
- [Windows SetMaxStdio](https://learn.microsoft.com/en-us/cpp/c-runtime-library/reference/setmaxstdio)

**Related Packages**
- [maxstdio](../maxstdio) - Windows C runtime limit management (used internally)
- [bufferReadCloser](../bufferReadCloser) - I/O wrappers with close support
- [ioutils](../) - Parent package with additional I/O utilities

**Community**
- [GitHub Issues](https://github.com/nabbar/golib/issues)
- [Contributing Guide](../../CONTRIBUTING.md)

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: fileDescriptor Package Contributors
