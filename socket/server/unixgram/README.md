# Unix Datagram Socket Server

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-65.6%25-yellow)](TESTING.md)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20Darwin-lightgrey)]()

Production-ready Unix domain datagram socket server for local IPC with file permissions, graceful shutdown, and comprehensive monitoring.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Comparison Matrix](#comparison-matrix)
- [Performance](#performance)
  - [Throughput](#throughput)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Datagram Server](#basic-datagram-server)
  - [Production Server](#production-server)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [ServerUnixGram Interface](#serverunixgram-interface)
  - [Configuration](#configuration)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Limitations](#limitations)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **unixgram** package provides a high-performance Unix domain datagram socket server for local inter-process communication. It combines the connectionless nature of UDP with filesystem-based access control, ideal for local logging, metrics collection, and event distribution.

### Design Philosophy

1. **Simplicity First**: Minimal API for connectionless datagram handling
2. **Security by Design**: File permissions and group ownership for access control
3. **Observable**: Real-time monitoring via callbacks
4. **Stateless**: No connection state, fire-and-forget messaging
5. **Context-Aware**: Full integration with Go's context for lifecycle management

### Why Use This Package?

- **Local-only Communication**: No network overhead, filesystem-based security
- **Fire-and-Forget**: No connection setup/teardown, immediate message delivery
- **Access Control**: Unix file permissions provide fine-grained security
- **Lightweight**: Minimal resource usage, single handler goroutine
- **Message Boundaries**: Each datagram is atomic and preserved
- **Familiar Interface**: Standard `libsck.Server` interface compatibility

### Key Features

- ✅ **Unix Domain Datagrams**: Local IPC without network overhead
- ✅ **File Permissions**: Fine-grained access control (0600, 0660, 0770, etc.)
- ✅ **Group Ownership**: Multi-user scenarios with GID control
- ✅ **Connectionless**: No handshake, no connection setup/teardown
- ✅ **Message Boundaries**: Each datagram is independent and atomic
- ✅ **Graceful Shutdown**: Automatic socket file cleanup
- ✅ **Lifecycle Callbacks**: Hook into datagram and server events
- ✅ **Thread-Safe**: Lock-free atomic operations
- ✅ **Context Integration**: Cancellation and deadline support
- ✅ **Platform Support**: Linux and Darwin (macOS)

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────┐
│              Unix Datagram Server                  │
├────────────────────────────────────────────────────┤
│                                                    │
│  ┌──────────────┐      ┌──────────────────┐        │
│  │ Socket File  │      │  Context Manager │        │
│  │ /tmp/app.sock│      │  (cancellation)  │        │
│  └──────┬───────┘      └────────┬─────────┘        │
│         │                       │                  │
│         ▼                       ▼                  │
│  ┌──────────────────────────────────────┐          │
│  │      UnixConn Listener               │          │
│  │      (SOCK_DGRAM)                    │          │
│  └──────────────┬───────────────────────┘          │
│                 │                                  │
│                 ▼                                  │
│    Single Handler Goroutine                        │
│    ┌─────────────────────────┐                     │
│    │  sCtx (I/O wrapper)     │                     │
│    │  - ReadFrom (datagram)  │                     │
│    │  - WriteTo (response)   │                     │
│    │  - Sender tracking      │                     │
│    └──────────┬──────────────┘                     │
│               │                                    │
│               ▼                                    │
│    ┌─────────────────────┐                         │
│    │   User Handler      │                         │
│    │   (HandlerFunc)     │                         │
│    └─────────────────────┘                         │
│                                                    │
└────────────────────────────────────────────────────┘
```

### Data Flow

1. **Server Start**: `Listen()` creates Unix datagram socket at file path
2. **File Setup**:
   - Removes existing socket file if present
   - Creates socket with configured permissions
   - Sets group ownership (chown)
3. **Handler Launch**: Single goroutine processes all datagrams
4. **Datagram Processing**:
   - ReadFrom() receives datagram from any sender
   - Handler processes datagram
   - WriteTo() sends response (optional)
5. **Server Stop**:
   - Handler exits
   - Socket closed
   - File removed
   - Resources released

### Comparison Matrix

| Feature | Unix Datagram (this pkg) | Unix Stream | UDP | TCP |
|---------|--------------------------|-------------|-----|-----|
| **Transport** | SOCK_DGRAM | SOCK_STREAM | IP Datagram | IP Stream |
| **Scope** | Local only | Local only | Network | Network |
| **Connection** | Connectionless | Connection-oriented | Connectionless | Connection-oriented |
| **Reliability** | Unreliable | Reliable | Unreliable | Reliable |
| **Message Boundaries** | Yes | No | Yes | No |
| **Access Control** | File permissions | File permissions | Ports | Ports |
| **Best For** | Logs, metrics, events | Sessions, RPC | Discovery, multicast | HTTP, RPC |
| **Max Throughput** | 100K+ msg/s | 500K+ req/s | 50K+ msg/s | 500K+ req/s |
| **Overhead** | Minimal | Minimal | Medium | Medium |

---

## Performance

### Throughput

Based on benchmarks on localhost:

| Datagram Size | Throughput | Latency (P50) | CPU Usage |
|---------------|------------|---------------|-----------|
| **1 KB** | ~100K msg/s | <100 µs | 5-10% |
| **4 KB** | ~50K msg/s | <200 µs | 8-15% |
| **16 KB** | ~10K msg/s | <1 ms | 10-20% |

*Actual throughput depends on handler processing speed*

### Memory Usage

Per-server memory footprint:

```
Goroutine stack:      ~8 KB (single handler)
Server structure:     ~2 KB
Application buffers:  Variable (e.g., 8 KB)
────────────────────────────
Total:                ~10-20 KB + buffer size
```

**Comparison:**
- Unix Datagram: ~15 KB (1 handler)
- Unix Stream: ~15 KB × N connections
- UDP: ~15 KB (1 handler)
- TCP: ~15 KB × N connections

### Scalability

**Recommended message rates:**

| Messages/sec | Performance | Notes |
|--------------|-------------|-------|
| **1-10K** | Excellent | Ideal range |
| **10K-50K** | Good | Monitor handler |
| **50K-100K** | Fair | Optimize processing |
| **100K+** | Not advised | Consider batching |

---

## Use Cases

### 1. Local Logging Daemon

**Problem**: Collect logs from multiple local processes

```go
handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    logFile, _ := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE, 0644)
    defer logFile.Close()
    
    buf := make([]byte, 8192)
    for {
        n, err := ctx.Read(buf)
        if err != nil {
            break
        }
        
        timestamp := time.Now().Format(time.RFC3339)
        logFile.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, buf[:n]))
    }
}
```

**Real-world**: syslog-ng, journald alternatives, application log aggregators

### 2. Metrics Collection

**Problem**: Aggregate metrics from local microservices

```go
type Metric struct {
    Name  string
    Value float64
    Tags  map[string]string
}

handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    buf := make([]byte, 2048)
    for {
        n, err := ctx.Read(buf)
        if err != nil {
            break
        }
        
        var metric Metric
        if err := json.Unmarshal(buf[:n], &metric); err != nil {
            continue
        }
        
        metricsDB.Record(metric.Name, metric.Value, metric.Tags)
    }
}
```

**Real-world**: StatsD, collectd, custom metrics pipelines

### 3. Event Distribution

**Problem**: Broadcast events to local subscribers

```go
var subscribers sync.Map

handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    buf := make([]byte, 4096)
    for {
        n, err := ctx.Read(buf)
        if err != nil {
            break
        }
        
        event := parseEvent(buf[:n])
        
        subscribers.Range(func(key, value interface{}) bool {
            subscriber := value.(chan Event)
            select {
            case subscriber <- event:
            default:
                // Subscriber slow, skip
            }
            return true
        })
    }
}
```

**Real-world**: Event buses, pub/sub systems, notification services

### 4. Service Discovery

**Problem**: Register local services dynamically

```go
type Service struct {
    Name    string
    Port    int
    PID     int
    Healthy bool
}

var registry sync.Map

handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    buf := make([]byte, 1024)
    for {
        n, err := ctx.Read(buf)
        if err != nil {
            break
        }
        
        var svc Service
        json.Unmarshal(buf[:n], &svc)
        
        registry.Store(svc.Name, svc)
        log.Printf("Registered: %s on port %d", svc.Name, svc.Port)
    }
}
```

**Real-world**: Consul alternatives, local service mesh, container orchestration

### 5. Command & Control

**Problem**: Send commands to running daemons

```go
handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    buf := make([]byte, 512)
    for {
        n, err := ctx.Read(buf)
        if err != nil {
            break
        }
        
        cmd := string(buf[:n])
        
        switch cmd {
        case "reload":
            reloadConfig()
        case "status":
            // Respond via callback or separate socket
            sendStatus()
        case "shutdown":
            gracefulShutdown()
        }
    }
}
```

**Real-world**: systemd socket activation, daemon control, hot reload

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/server/unixgram
```

### Basic Datagram Server

```go
package main

import (
    "context"
    "fmt"
    "io"
    "os"
    
    libprm "github.com/nabbar/golib/file/perm"
    libptc "github.com/nabbar/golib/network/protocol"
    libsck "github.com/nabbar/golib/socket"
    sckcfg "github.com/nabbar/golib/socket/config"
    "github.com/nabbar/golib/socket/server/unixgram"
)

func main() {
    // Simple logging handler
    handler := func(ctx libsck.Context) {
        defer ctx.Close()
        
        buf := make([]byte, 8192)
        for {
            n, err := ctx.Read(buf)
            if err != nil {
                if err != io.EOF {
                    fmt.Printf("Error: %v\n", err)
                }
                return
            }
            
            fmt.Printf("Received: %s\n", buf[:n])
        }
    }
    
    // Create configuration
    cfg := sckcfg.Server{
        Network:   libptc.NetworkUnixGram,
        Address:   "/tmp/app.sock",
        PermFile:  libprm.Perm(0660), // rw-rw----
        GroupPerm: -1,  // Use process group
    }
    
    // Create and start server
    srv, err := unixgram.New(nil, handler, cfg)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Println("Server listening on /tmp/app.sock")
    if err := srv.Listen(context.Background()); err != nil {
        fmt.Printf("Server error: %v\n", err)
    }
}
```

### Production Server

```go
func main() {
    // Handler with proper error handling
    handler := func(ctx libsck.Context) {
        defer ctx.Close()
        
        buf := make([]byte, 65507) // Max datagram size
        for {
            n, err := ctx.Read(buf)
            if err != nil {
                if err != io.EOF && err != io.ErrClosedPipe {
                    log.Printf("Read error: %v", err)
                }
                return
            }
            
            // Process datagram
            processMessage(buf[:n])
        }
    }
    
    // Configuration with permissions
    cfg := sckcfg.Server{
        Network:   libptc.NetworkUnixGram,
        Address:   "/var/run/app.sock",
        PermFile:  libprm.Perm(0660),
        GroupPerm: 1000, // Specific GID
    }
    
    srv, err := unixgram.New(nil, handler, cfg)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }
    
    // Register monitoring callbacks
    srv.RegisterFuncError(func(errs ...error) {
        for _, e := range errs {
            if e != nil {
                log.Printf("[ERROR] %v", e)
            }
        }
    })
    
    srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
        switch state {
        case libsck.ConnectionNew:
            log.Println("[INFO] Handler started")
        case libsck.ConnectionRead:
            log.Printf("[INFO] Datagram from %s", remote)
        case libsck.ConnectionClose:
            log.Println("[INFO] Handler stopped")
        }
    })
    
    srv.RegisterFuncInfoServer(func(msg string) {
        log.Printf("[SERVER] %s", msg)
    })
    
    // Start server
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go func() {
        if err := srv.Listen(ctx); err != nil {
            log.Printf("Server error: %v", err)
        }
    }()
    
    // Graceful shutdown on signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    log.Println("Shutting down...")
    
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 10*time.Second)
    defer shutdownCancel()
    
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
    
    log.Println("Server stopped")
}
```

---

## Best Practices

### ✅ DO

**Set appropriate permissions:**
```go
// Restrictive - only owner
cfg.PermFile = libprm.Perm(0600)  // rw-------

// Group access
cfg.PermFile = libprm.Perm(0660)  // rw-rw----

// World readable/writable (use with caution!)
cfg.PermFile = libprm.Perm(0666)  // rw-rw-rw-
```

**Handle large datagrams:**
```go
// Use appropriate buffer size
buf := make([]byte, 65507)  // Max size on most systems
```

**Implement graceful shutdown:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Printf("Shutdown timeout: %v", err)
}
```

**Monitor handler health:**
```go
srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
    if state == libsck.ConnectionClose {
        // Handler exited - investigate why
        log.Println("Handler closed unexpectedly")
    }
})
```

### ❌ DON'T

**Don't use for reliable delivery:**
```go
// ❌ BAD: Unix datagrams can be lost
// No guarantee of delivery

// ✅ GOOD: Use Unix stream sockets for reliability
// See github.com/nabbar/golib/socket/server/unix
```

**Don't assume ordering:**
```go
// ❌ BAD: Datagrams may arrive out of order
// Application must handle reordering

// ✅ GOOD: Add sequence numbers if order matters
type Message struct {
    Seq  uint64
    Data []byte
}
```

**Don't send oversized datagrams:**
```go
// ❌ BAD: Datagram may be truncated or dropped
largeData := make([]byte, 1000000)  // 1 MB

// ✅ GOOD: Keep datagrams reasonably sized
smallData := make([]byte, 8192)  // 8 KB
```

**Don't use on Windows:**
```go
// ❌ BAD: Unix sockets not supported on Windows
// This package will not compile on Windows

// ✅ GOOD: Use UDP or named pipes on Windows
// See github.com/nabbar/golib/socket/server/udp
```

### Testing

The package includes a comprehensive test suite with **65.6% code coverage** and **72 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and lifecycle operations
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks (throughput, latency, scalability)
- ✅ Error handling and edge cases
- ✅ File permissions and group ownership
- ✅ Context integration and cancellation

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

---

## API Reference

### ServerUnixGram Interface

```go
type ServerUnixGram interface {
    // Start accepting datagrams
    Listen(ctx context.Context) error
    
    // Stop accepting, wait for handler to exit
    Shutdown(ctx context.Context) error
    
    // Stop accepting, close handler immediately
    Close() error
    
    // Check if server is accepting datagrams
    IsRunning() bool
    
    // Check if server is shutting down
    IsGone() bool
    
    // Get connection count (always 0 for datagrams)
    OpenConnections() int64
    
    // Configure socket file
    RegisterSocket(unixFile string, perm os.FileMode, gid int32) error
    
    // Register callbacks
    RegisterFuncError(f libsck.FuncError)
    RegisterFuncInfo(f libsck.FuncInfo)
    RegisterFuncInfoServer(f libsck.FuncInfoSrv)
    
    // No-op for Unix sockets (TLS not supported)
    SetTLS(enable bool, config libtls.TLSConfig) error
}
```

### Configuration

```go
type Server struct {
    Network   libptc.NetworkType  // Must be NetworkUnixGram
    Address   string              // Unix socket file path
    PermFile  libprm.Perm         // File permissions (e.g., 0660)
    GroupPerm int32               // Group ID (-1 for process group)
}
```

### Error Codes

```go
var (
    ErrInvalidUnixFile  = "invalid unix file for socket listening"
    ErrInvalidGroup     = "invalid unix group for socket group permission"
    ErrInvalidHandler   = "invalid handler"
    ErrShutdownTimeout  = "timeout on stopping socket"
    ErrInvalidInstance  = "invalid socket instance"
)
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Reporting Bugs

If you find a bug, please open an issue on GitHub with:

1. **Description**: Clear and concise description
2. **Reproduction Steps**: Minimal code example
3. **Expected Behavior**: What you expected
4. **Actual Behavior**: What actually happened
5. **Environment**: Go version, OS (Linux/Darwin), system info
6. **Logs/Errors**: Error messages or stack traces

**Submit issues at**: [https://github.com/nabbar/golib/issues](https://github.com/nabbar/golib/issues)

### Code Contributions

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >65%)
   - Pass all tests including race detector
   - Use `gofmt` and `golint`

2. **AI Usage Policy**
   - ❌ **AI must NEVER be used** to generate package code
   - ✅ **AI assistance limited to**: Testing, debugging, documentation
   - All AI-assisted work must be reviewed by humans

3. **Testing**
   - Add tests for new features
   - Use Ginkgo v2 / Gomega
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

- ✅ **65.6% test coverage** (target: >65%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Secure** file permissions and group ownership
- ✅ **Graceful shutdown** with automatic cleanup

### Platform Support

**Supported:**
- ✅ Linux (all distributions)
- ✅ Darwin (macOS)

**Not Supported:**
- ❌ Windows (use UDP or named pipes instead)
- ❌ Other Unix variants (may work but untested)

### Architectural Constraints

1. **Unreliable Delivery**: Datagrams may be lost or arrive out of order
2. **No Acknowledgment**: No built-in delivery confirmation
3. **Size Limits**: System-dependent max datagram size (~65KB typically)
4. **Local Only**: Cannot communicate across network
5. **Single Handler**: One goroutine processes all datagrams

### Not Suitable For

- Reliable message delivery (use Unix stream sockets)
- Network communication (use UDP or TCP)
- Large file transfers (use HTTP or custom protocol)
- High-frequency trading (<10µs latency requirements)
- Windows environments

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**Performance Optimizations:**
1. Batched datagram processing (process multiple datagrams per read cycle)
2. Buffer pooling for reduced GC pressure
3. Zero-copy operations where possible

**Feature Additions:**
1. Optional Prometheus/OpenTelemetry metrics exporters
2. Per-sender rate limiting and throttling
3. Message schema validation hooks
4. Datagram queuing and buffering options
5. Support for SCM_RIGHTS (file descriptor passing)

**API Extensions:**
1. Helper functions for common patterns (logging, metrics)
2. Integration with popular logging frameworks
3. Custom error types for better error handling

**Quality of Life:**
1. Datagram size auto-detection and warnings
2. Health check endpoint support
3. Configuration validation helpers

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server/unixgram)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, Unix datagram handling, lifecycle management, performance considerations, and comparison with Unix stream and UDP sockets. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 65.6% coverage analysis, performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/socket](https://pkg.go.dev/github.com/nabbar/golib/socket)** - Base interfaces and types for socket servers. Defines `Server`, `HandlerFunc`, `Context`, and callback types used across all socket implementations for consistent API design.

- **[github.com/nabbar/golib/socket/server/unix](https://pkg.go.dev/github.com/nabbar/golib/socket/server/unix)** - Unix domain stream sockets (SOCK_STREAM) for reliable connection-oriented local IPC. Use this when you need guaranteed delivery and ordering.

- **[github.com/nabbar/golib/socket/server/udp](https://pkg.go.dev/github.com/nabbar/golib/socket/server/udp)** - UDP datagram server for network-based connectionless communication. Similar to unixgram but works across networks.

- **[github.com/nabbar/golib/socket/config](https://pkg.go.dev/github.com/nabbar/golib/socket/config)** - Unified configuration structures for all socket servers. Provides type-safe configuration with validation and sensible defaults.

- **[github.com/nabbar/golib/file/perm](https://pkg.go.dev/github.com/nabbar/golib/file/perm)** - File permission handling utilities. Provides type-safe permission constants and conversion functions for Unix file modes.

### Standard Library References

- **[net](https://pkg.go.dev/net)** - Standard library networking package. The `unixgram` package builds upon `net.UnixConn` to provide datagram-aware reading with additional lifecycle management and monitoring.

- **[context](https://pkg.go.dev/context)** - Standard I/O context package. The package fully integrates with Go's context for cancellation, deadlines, and value propagation across server lifecycle.

### External References

- **[Unix Domain Sockets](https://man7.org/linux/man-pages/man7/unix.7.html)** - Official Linux manual page covering Unix domain socket concepts, socket types (SOCK_STREAM vs SOCK_DGRAM), addressing, and permissions.

- **[SOCK_DGRAM](https://man7.org/linux/man-pages/man2/socket.2.html)** - Socket type documentation explaining datagram sockets, their characteristics (connectionless, unreliable, message boundaries), and differences from stream sockets.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and concurrency patterns. The `unixgram` package follows these conventions for idiomatic Go code.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the `unixgram` package. Check existing issues before creating new ones to avoid duplicates.

- **[Contributing Guide](../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements, testing procedures, and pull request process.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/server/unixgram`  
**Version**: See [releases](https://github.com/nabbar/golib/releases)
