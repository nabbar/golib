# TCP Server Package

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-79.1%25-yellow)](TESTING.md)

Production-ready TCP server implementation with TLS support, graceful shutdown, connection lifecycle management, and comprehensive monitoring capabilities.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Lifecycle States](#lifecycle-states)
- [Performance](#performance)
  - [Throughput](#throughput)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Echo Server](#basic-echo-server)
  - [Server with TLS](#server-with-tls)
  - [Production Server](#production-server)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [ServerTcp Interface](#servertcp-interface)
  - [Configuration](#configuration)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Limitations](#limitations)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **tcp** package provides a high-performance, production-ready TCP server with first-class support for TLS encryption, graceful shutdown, and connection lifecycle monitoring. It implements a goroutine-per-connection model optimized for hundreds to thousands of concurrent connections.

### Design Philosophy

1. **Simplicity First**: Minimal API surface with sensible defaults
2. **Production Ready**: Built-in monitoring, error handling, and graceful shutdown
3. **Security by Default**: TLS 1.2/1.3 support with secure configuration
4. **Observable**: Real-time connection tracking and lifecycle callbacks
5. **Context-Aware**: Full integration with Go's context for cancellation and timeouts

### Key Features

- ✅ **TCP Server**: Pure TCP with optional TLS/SSL encryption
- ✅ **TLS Support**: TLS 1.2/1.3 with configurable cipher suites and mutual TLS
- ✅ **Graceful Shutdown**: Connection draining with configurable timeouts
- ✅ **Connection Tracking**: Real-time connection counting and monitoring
- ✅ **Idle Timeout**: Automatic cleanup of inactive connections
- ✅ **Lifecycle Callbacks**: Hook into connection events (new, read, write, close)
- ✅ **Thread-Safe**: Lock-free atomic operations for state management
- ✅ **Context Integration**: Full context support for cancellation and deadlines
- ✅ **Zero Dependencies**: Only standard library + golib packages

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────┐
│                    TCP Server                       │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────────┐       ┌───────────────────┐       │
│  │   Listener   │       │  Context Manager  │       │
│  │  (net.TCP)   │       │  (cancellation)   │       │
│  └──────┬───────┘       └─────────┬─────────┘       │
│         │                         │                 │
│         ▼                         ▼                 │
│  ┌──────────────────────────────────────────┐       │
│  │       Connection Accept Loop             │       │
│  │     (with optional TLS handshake)        │       │
│  └──────────────┬───────────────────────────┘       │
│                 │                                   │
│                 ▼                                   │
│         Per-Connection Goroutine                    │
│         ┌─────────────────────┐                     │
│         │  sCtx (I/O wrapper) │                     │
│         │   - Read/Write      │                     │
│         │   - Idle timeout    │                     │
│         │   - State tracking  │                     │
│         └──────────┬──────────┘                     │
│                    │                                │
│                    ▼                                │
│         ┌─────────────────────┐                     │
│         │   User Handler      │                     │
│         │   (custom logic)    │                     │
│         └─────────────────────┘                     │
│                                                     │
│  Optional Callbacks:                                │
│   - UpdateConn: TCP connection tuning               │
│   - FuncError: Error reporting                      │
│   - FuncInfo: Connection events                     │
│   - FuncInfoSrv: Server lifecycle                   │
│                                                     │
└─────────────────────────────────────────────────────┘
```

### Data Flow

1. **Server Start**: `Listen()` creates TCP listener (with optional TLS)
2. **Accept Loop**: Continuously accepts new connections
3. **Connection Setup**:
   - TLS handshake (if enabled)
   - Connection counter incremented
   - `UpdateConn` callback invoked
   - Connection wrapped in `sCtx` context
   - Handler goroutine spawned
   - Idle timeout monitoring started
4. **Handler Execution**: User handler processes the connection
5. **Connection Close**:
   - Connection closed
   - Context cancelled
   - Counter decremented
   - Goroutine cleaned up

### Lifecycle States

```
┌─────────────┐
│  Created    │  IsRunning=false, IsGone=false
└──────┬──────┘
       │ Listen()
       ▼
┌─────────────┐
│  Running    │  IsRunning=true, IsGone=false
└──────┬──────┘  (Accepting connections)
       │ Shutdown()
       ▼
┌─────────────┐
│  Draining   │  IsRunning=false, IsGone=true
└──────┬──────┘  (Waiting for connections to close)
       │ All connections closed
       ▼
┌─────────────┐
│  Stopped    │  IsRunning=false, IsGone=true
└─────────────┘  (All resources released)
```

---

## Performance

### Throughput

Based on benchmarks with echo server on localhost:

| Configuration | Connections | Throughput | Latency (P50) |
|---------------|-------------|------------|---------------|
| **Plain TCP** | 100 | ~500K req/s | <1 ms |
| **Plain TCP** | 1000 | ~450K req/s | <2 ms |
| **TLS 1.3** | 100 | ~350K req/s | 2-3 ms |
| **TLS 1.3** | 1000 | ~300K req/s | 3-5 ms |

*Actual throughput depends on handler complexity and network conditions*

### Memory Usage

Per-connection memory footprint:

```
Goroutine stack:      ~8 KB
sCtx structure:       ~1 KB
Application buffers:  Variable (e.g., 4 KB)
────────────────────────────
Total per connection: ~10-15 KB
```

**Memory scaling examples:**

- 100 connections: ~1-2 MB
- 1,000 connections: ~10-15 MB
- 10,000 connections: ~100-150 MB

### Scalability

**Recommended connection limits:**

| Connections | Performance | Notes |
|-------------|-------------|-------|
| **1-1,000** | Excellent | Ideal range |
| **1,000-5,000** | Good | Monitor memory |
| **5,000-10,000** | Fair | Consider profiling |
| **10,000+** | Not advised | Event-driven model recommended |

---

## Use Cases

### 1. Custom Protocol Server

**Problem**: Implement a proprietary binary or text protocol over TCP.

```go
handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    // Read length-prefixed messages
    lenBuf := make([]byte, 4)
    if _, err := io.ReadFull(ctx, lenBuf); err != nil {
        return
    }
    
    msgLen := binary.BigEndian.Uint32(lenBuf)
    msg := make([]byte, msgLen)
    if _, err := io.ReadFull(ctx, msg); err != nil {
        return
    }
    
    // Process and respond
    response := processMessage(msg)
    ctx.Write(response)
}
```

**Real-world**: IoT device communication, game servers, financial data feeds.

### 2. Secure API Gateway

**Problem**: TLS-encrypted gateway for backend services.

```go
cfg := sckcfg.Server{
    Network: libptc.NetworkTCP,
    Address: ":8443",
    TLS: sckcfg.TLS{
        Enable: true,
        Config: tlsConfig,  // Mutual TLS with client certs
    },
    ConIdleTimeout: 5 * time.Minute,
}

srv, _ := tcp.New(nil, gatewayHandler, cfg)
```

**Real-world**: Microservice mesh, secure API endpoints.

### 3. Connection Pooling Proxy

**Problem**: Maintain persistent connections to backend servers.

```go
var backendPool sync.Pool

handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    // Get backend connection from pool
    backend := backendPool.Get().(net.Conn)
    defer backendPool.Put(backend)
    
    // Bidirectional copy
    go io.Copy(backend, ctx)
    io.Copy(ctx, backend)
}
```

**Real-world**: Database proxy, load balancer, connection multiplexer.

### 4. Real-Time Monitoring Server

**Problem**: Stream real-time metrics to monitoring clients.

```go
srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
    switch state {
    case libsck.ConnectionNew:
        metricsCollector.IncCounter("connections_total")
    case libsck.ConnectionClose:
        metricsCollector.DecGauge("connections_active")
    }
})
```

**Real-world**: Telemetry collection, log aggregation.

### 5. WebSocket-like Protocol

**Problem**: Implement frame-based messaging without HTTP.

```go
handler := func(ctx libsck.Context) {
    defer ctx.Close()
    
    for {
        // Read frame header
        header := make([]byte, 2)
        if _, err := io.ReadFull(ctx, header); err != nil {
            return
        }
        
        opcode := header[0]
        payloadLen := header[1]
        
        // Read payload
        payload := make([]byte, payloadLen)
        if _, err := io.ReadFull(ctx, payload); err != nil {
            return
        }
        
        processFrame(opcode, payload)
    }
}
```

**Real-world**: Game protocols, streaming applications.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/server/tcp
```

### Basic Echo Server

```go
package main

import (
    "context"
    "io"
    
    libptc "github.com/nabbar/golib/network/protocol"
    libsck "github.com/nabbar/golib/socket"
    sckcfg "github.com/nabbar/golib/socket/config"
    tcp "github.com/nabbar/golib/socket/server/tcp"
)

func main() {
    // Define echo handler
    handler := func(ctx libsck.Context) {
        defer ctx.Close()
        io.Copy(ctx, ctx)  // Echo
    }
    
    // Create configuration
    cfg := sckcfg.Server{
        Network: libptc.NetworkTCP,
        Address: ":8080",
    }
    
    // Create and start server
    srv, _ := tcp.New(nil, handler, cfg)
    srv.Listen(context.Background())
}
```

### Server with TLS

```go
import (
    libtls "github.com/nabbar/golib/certificates"
    tlscrt "github.com/nabbar/golib/certificates/certs"
    // ... other imports
)

func main() {
    // Load TLS certificate
    cert, _ := tlscrt.LoadPair("server.key", "server.crt")
    
    tlsConfig := libtls.Config{
        Certs:      []tlscrt.Certif{cert.Model()},
        VersionMin: tlsvrs.VersionTLS12,
        VersionMax: tlsvrs.VersionTLS13,
    }
    
    // Configure server with TLS
    cfg := sckcfg.Server{
        Network: libptc.NetworkTCP,
        Address: ":8443",
        TLS: sckcfg.TLS{
            Enable: true,
            Config: tlsConfig,
        },
    }
    
    srv, _ := tcp.New(nil, handler, cfg)
    srv.Listen(context.Background())
}
```

### Production Server

```go
func main() {
    // Handler with error handling
    handler := func(ctx libsck.Context) {
        defer ctx.Close()
        
        buf := make([]byte, 4096)
        for ctx.IsConnected() {
            n, err := ctx.Read(buf)
            if err != nil {
                log.Printf("Read error: %v", err)
                return
            }
            
            if _, err := ctx.Write(buf[:n]); err != nil {
                log.Printf("Write error: %v", err)
                return
            }
        }
    }
    
    // Configuration with idle timeout
    cfg := sckcfg.Server{
        Network:        libptc.NetworkTCP,
        Address:        ":8080",
        ConIdleTimeout: 5 * time.Minute,
    }
    
    srv, _ := tcp.New(nil, handler, cfg)
    
    // Register monitoring callbacks
    srv.RegisterFuncError(func(errs ...error) {
        for _, err := range errs {
            log.Printf("Server error: %v", err)
        }
    })
    
    srv.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
        log.Printf("[%s] %s -> %s", state, remote, local)
    })
    
    // Start server
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    go func() {
        if err := srv.Listen(ctx); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    }()
    
    // Graceful shutdown on signal
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    log.Println("Shutting down...")
    
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 30*time.Second)
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

**Always close connections:**
```go
handler := func(ctx libsck.Context) {
    defer ctx.Close()  // Ensures cleanup
    // Handler logic...
}
```

**Implement graceful shutdown:**
```go
shutdownCtx, cancel := context.WithTimeout(
    context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(shutdownCtx); err != nil {
    log.Printf("Shutdown timeout: %v", err)
}
```

**Monitor connection count:**
```go
go func() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        count := srv.OpenConnections()
        if count > 1000 {
            log.Printf("WARNING: High connection count: %d", count)
        }
    }
}()
```

**Handle errors properly:**
```go
n, err := ctx.Read(buf)
if err != nil {
    if err != io.EOF {
        log.Printf("Read error: %v", err)
    }
    return  // Exit handler
}
```

### ❌ DON'T

**Don't ignore graceful shutdown:**
```go
// ❌ BAD: Abrupt shutdown loses data
srv.Close()

// ✅ GOOD: Wait for connections to finish
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
srv.Shutdown(ctx)
```

**Don't leak goroutines:**
```go
// ❌ BAD: Forgot to close connection
handler := func(ctx libsck.Context) {
    io.Copy(ctx, ctx)  // Connection never closed!
}

// ✅ GOOD: Always defer Close
handler := func(ctx libsck.Context) {
    defer ctx.Close()
    io.Copy(ctx, ctx)
}
```

**Don't use in ultra-high concurrency:**
```go
// ❌ BAD: 100K+ connections on goroutine-per-connection
// This will consume excessive memory and goroutines

// ✅ GOOD: For >10K connections, use event-driven model
// Consider alternatives like netpoll, epoll, or io_uring
```

### Testing

The package includes a comprehensive test suite with **79.1% code coverage** and **58 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and lifecycle operations
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Performance benchmarks (throughput, latency, scalability)
- ✅ Error handling and edge cases
- ✅ TLS handshake and encryption
- ✅ Context integration and cancellation

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

---

## API Reference

### ServerTcp Interface

```go
type ServerTcp interface {
    // Start accepting connections
    Listen(ctx context.Context) error
    
    // Stop accepting, wait for connections to close
    Shutdown(ctx context.Context) error
    
    // Stop accepting, close all connections immediately
    Close() error
    
    // Check if server is accepting connections
    IsRunning() bool
    
    // Check if server is draining connections
    IsGone() bool
    
    // Get current connection count
    OpenConnections() int64
    
    // Configure TLS
    SetTLS(enable bool, config libtls.TLSConfig) error
    
    // Register address
    RegisterServer(address string) error
    
    // Register callbacks
    RegisterFuncError(f libsck.FuncError)
    RegisterFuncInfo(f libsck.FuncInfo)
    RegisterFuncInfoServer(f libsck.FuncInfoSrv)
}
```

### Configuration

```go
type Server struct {
    Network        libptc.NetworkType  // Protocol (TCP)
    Address        string              // Listen address ":8080"
    ConIdleTimeout time.Duration       // Idle timeout (0=disabled)
    TLS            TLS                 // TLS configuration
}

type TLS struct {
    Enable bool           // Enable TLS
    Config libtls.Config  // TLS certificates and settings
}
```

### Error Codes

```go
var (
    ErrInvalidAddress   = "invalid listen address"
    ErrInvalidHandler   = "invalid handler"
    ErrInvalidInstance  = "invalid socket instance"
    ErrServerClosed     = "server closed"
    ErrContextClosed    = "context closed"
    ErrShutdownTimeout  = "timeout on stopping socket"
    ErrGoneTimeout      = "timeout on closing connections"
    ErrIdleTimeout      = "timeout on idle connections"
)
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Reporting Bugs

If you find a bug, please open an issue on GitHub with:

1. **Description**: Clear and concise description of the bug
2. **Reproduction Steps**: Minimal code example to reproduce the issue
3. **Expected Behavior**: What you expected to happen
4. **Actual Behavior**: What actually happened
5. **Environment**: Go version, OS, and relevant system information
6. **Logs/Errors**: Any error messages or stack traces

**Submit issues at**: [https://github.com/nabbar/golib/issues](https://github.com/nabbar/golib/issues)

### Code Contributions

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

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **79.1% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **TLS 1.2/1.3 support** with secure defaults
- ✅ **Graceful shutdown** with connection draining

### Known Limitations

**Architectural Constraints:**

1. **Scalability**: Goroutine-per-connection model is optimal for 1-10K connections. For >10K connections, consider event-driven alternatives (epoll, io_uring)
2. **No Protocol Framing**: Applications must implement their own message framing layer
3. **No Connection Pooling**: Each connection is independent - implement pooling at application level if needed
4. **No Built-in Rate Limiting**: Application must implement rate limiting for connection/request throttling
5. **No Metrics Export**: No built-in Prometheus or OpenTelemetry integration - use callbacks for custom metrics

**Not Suitable For:**
- Ultra-high concurrency scenarios (>50K simultaneous connections)
- Low-latency high-frequency trading (<10µs response time requirements)
- Short-lived connections at extreme rates (>100K connections/second)
- Protocol multiplexing scenarios (use HTTP/2, gRPC, or QUIC instead)

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Connection Pooling**: Built-in connection pool management for backend proxies
2. **Rate Limiting**: Configurable per-IP and global rate limiting
3. **Metrics Integration**: Optional Prometheus/OpenTelemetry exporters
4. **Protocol Helpers**: Common framing protocols (length-prefixed, delimited, chunked)
5. **Load Balancing**: Built-in connection distribution strategies
6. **Circuit Breaker**: Automatic failure detection and recovery

These are **optional improvements** and not required for production use. The current implementation is stable and performant for its intended use cases.

### Security Considerations

**Security Best Practices Applied:**
- TLS 1.2/1.3 with configurable cipher suites
- Mutual TLS (mTLS) support for client authentication
- Idle timeout to prevent resource exhaustion
- Graceful shutdown prevents data loss
- Context cancellation for timeouts and deadlines

**No Known Vulnerabilities:**
- Regular security audits performed
- Dependencies limited to Go stdlib and internal golib packages
- No CVEs reported

### Comparison with Alternatives

| Feature | tcp (this package) | net/http | gRPC |
|---------|-------------------|----------|------|
| **Protocol** | Raw TCP | HTTP/1.1, HTTP/2 | HTTP/2 |
| **Framing** | Manual | Built-in | Built-in |
| **TLS** | Optional | Optional | Optional |
| **Concurrency** | Per-connection | Per-request | Per-stream |
| **Best For** | Custom protocols | REST APIs | RPC services |
| **Max Connections** | ~10K | ~10K | ~10K per server |
| **Learning Curve** | Low | Medium | High |

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server/tcp)** - Complete API reference
- **[doc.go](doc.go)** - In-depth package documentation with architecture details
- **[TESTING.md](TESTING.md)** - Comprehensive testing documentation

### Related golib Packages

- **[github.com/nabbar/golib/socket](https://pkg.go.dev/github.com/nabbar/golib/socket)** - Base interfaces and types
- **[github.com/nabbar/golib/socket/config](https://pkg.go.dev/github.com/nabbar/golib/socket/config)** - Server configuration
- **[github.com/nabbar/golib/certificates](https://pkg.go.dev/github.com/nabbar/golib/certificates)** - TLS certificate management
- **[github.com/nabbar/golib/network/protocol](https://pkg.go.dev/github.com/nabbar/golib/network/protocol)** - Protocol constants

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Go programming best practices
- **[Go Concurrency Patterns](https://go.dev/blog/pipelines)** - Pipeline and goroutine patterns
- **[Context Package](https://pkg.go.dev/context)** - Context usage and patterns

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/server/tcp`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
