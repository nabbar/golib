# Socket Server Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

High-performance, protocol-agnostic socket server library for Go with unified interface across TCP, UDP, and Unix domain sockets, featuring thread-safe operations, graceful shutdown, and comprehensive connection management.

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
  - [tcp - TCP Server](#tcp-subpackage)
  - [udp - UDP Server](#udp-subpackage)
  - [unix - Unix Domain Sockets (Stream)](#unix-subpackage)
  - [unixgram - Unix Domain Sockets (Datagram)](#unixgram-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This library provides production-ready socket server implementations for Go applications. It emphasizes connection management, thread safety, and protocol independence while supporting multiple transport mechanisms (TCP, UDP, Unix domain sockets in both stream and datagram modes).

### Design Philosophy

1. **Protocol Agnostic**: Unified `socket.Server` interface across all transport protocols
2. **Connection Management**: Automatic lifecycle handling with per-connection goroutines
3. **Thread-Safe**: Atomic operations (`atomic.Bool`, `atomic.Int64`) for concurrent access
4. **Context-Aware**: Graceful shutdown with connection draining via `context.Context`
5. **Event-Driven**: Callback system for monitoring, instrumentation, and error handling

---

## Key Features

- **Multiple Protocols**:
  - **TCP**: Connection-oriented, reliable, ordered delivery
  - **UDP**: Connectionless, fast, datagram-based
  - **Unix**: IPC via filesystem sockets (stream mode)
  - **Unixgram**: IPC via filesystem sockets (datagram mode)
- **Thread-Safe Operations**: Atomic operations (`atomic.Bool`, `atomic.Int64`), mutex protection
- **Connection Management**: Per-client goroutines, connection tracking, lifecycle callbacks
- **Graceful Shutdown**: Context-aware with connection draining
- **TLS Support**: Built-in for TCP (optional)
- **File Permissions**: Unix sockets with configurable permissions and group ownership
- **Half-Close Support**: Unix sockets with independent read/write shutdown
- **Standard Interfaces**: Implements `socket.Server` interface
- **Callback System**: Error reporting, connection events, server lifecycle

---

## Installation

```bash
go get github.com/nabbar/golib/socket/server
```

---

## Architecture

### Package Structure

The package is organized into four transport-specific subpackages:

```
socket/server/
├── tcp/                 # TCP server (connection-oriented, network)
├── udp/                 # UDP server (connectionless, network)
├── unix/                # Unix domain sockets (connection-oriented, IPC)
└── unixgram/            # Unix domain sockets (connectionless, IPC)
```

### Transport Comparison

```
┌────────────────────────────────────────────────────────────┐
│                  Transport Selection Matrix                │
└──────────────┬────────────┬────────────┬─────────┬─────────┘
               │            │            │         │
      ┌────────▼────┐   ┌───▼─────┐  ┌───▼────┐  ┌─▼───────┐
      │     TCP     │   │   UDP   │  │  Unix  │  │Unixgram │
      │             │   │         │  │        │  │         │
      │ Reliable    │   │ Fast    │  │ IPC    │  │ IPC     │
      │ Ordered     │   │ Low     │  │ Stream │  │ Dgram   │
      │ Network     │   │ Overhead│  │ Secure │  │ Fast    │
      └─────────────┘   └─────────┘  └────────┘  └─────────┘
```

| Protocol | Mode | Scope | Reliable | Ordered | Use Case |
|----------|------|-------|----------|---------|----------|
| **TCP** | Connection | Network | ✅ | ✅ | Web servers, APIs, databases |
| **UDP** | Datagram | Network | ❌ | ❌ | Real-time, gaming, streaming |
| **Unix** | Connection | Local | ✅ | ✅ | IPC, containers, microservices |
| **Unixgram** | Datagram | Local | ❌ | ❌ | Fast IPC, notifications |

### Connection vs Connectionless

**Connection-Oriented (TCP, Unix)**
- Persistent client connections
- Per-connection handler goroutines
- Connection state tracking
- Graceful connection draining
- Best for: Long-lived sessions, stateful protocols

**Connectionless (UDP, Unixgram)**
- Stateless datagram processing
- Single handler for all datagrams
- No connection tracking (returns 1 when running)
- Immediate shutdown (no draining)
- Best for: Request-response, fire-and-forget

---

## Performance

### Memory Efficiency

All servers maintain **O(1) memory per connection** or **O(1) for datagram**:

- **TCP/Unix**: One goroutine per client (~2KB stack + buffers)
- **UDP/Unixgram**: Single handler goroutine
- **Buffer Management**: `bufio.Reader/Writer` with controlled sizes
- **Streaming I/O**: Zero-copy direct passthrough where possible

### Thread Safety

All operations are thread-safe through:

- **Atomic Operations**: `atomic.Bool`, `atomic.Int64` for state and counters
- **Mutex Protection**: `sync.Mutex` for callback registration (where needed)
- **Goroutine Management**: Proper lifecycle with `context.Context`
- **No Data Races**: Verified with `go test -race`

### Throughput Benchmarks

| Server | Mode | Throughput | Connections | Memory/Conn | Notes |
|--------|------|------------|-------------|-------------|-------|
| TCP | Stream | ~800 MB/s | 10,000+ | ~4KB | Per-connection goroutine |
| UDP | Datagram | ~900 MB/s | N/A | ~8KB total | Single handler |
| Unix | Stream | ~1.2 GB/s | 1,000+ | ~4KB | IPC optimized |
| Unixgram | Datagram | ~1.5 GB/s | N/A | ~8KB total | Fastest IPC |

*Measured on AMD64, Go 1.21+, localhost/IPC*

### Connection Capacity

**TCP Server**: 10,000+ concurrent connections (tested up to 50,000 on high-memory systems)  
**Unix Server**: 1,000+ concurrent connections (tested up to 10,000 on Linux)  
**UDP/Unixgram**: Unlimited concurrent senders (stateless operation)

### Protocol Selection Guide

```
Reliability:  TCP > Unix > UDP > Unixgram
Speed:        Unixgram > Unix > UDP > TCP
Overhead:     UDP/Unixgram < Unix < TCP

Network Communication:
├─ Reliable → TCP (web servers, databases)
└─ Fast → UDP (streaming, gaming)

Local IPC:
├─ Reliable → Unix (Docker, microservices)
└─ Fast → Unixgram (metrics, notifications)
```

---

## Use Cases

This library is designed for scenarios requiring reliable socket server implementations:

**Web Applications**
- HTTP/HTTPS servers with TLS support
- WebSocket backends
- REST API endpoints
- Long-polling connections

**Microservices**
- Service-to-service communication via Unix sockets
- Container IPC (Docker, Kubernetes)
- Fast inter-process messaging
- Service discovery listeners

**Real-Time Systems**
- UDP-based game servers
- Video/audio streaming
- Sensor data collection
- Time-series data ingestion

**Database Proxies**
- Protocol translation layers
- Connection pooling
- Query routing
- Cache layers

**Network Daemons**
- Custom protocol servers
- Tunneling and proxying
- Network monitoring agents
- Log aggregation services

---

## Quick Start

### TCP Echo Server

Simple TCP server echoing received data:

```go
package main

import (
    "context"
    "io"
    "log"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/tcp"
)

func main() {
    // Handler processes each connection
    handler := func(r socket.Reader, w socket.Writer) {
        defer r.Close()
        defer w.Close()
        io.Copy(w, r) // Echo back
    }
    
    // Create server
    srv := tcp.New(nil, handler)
    srv.RegisterServer(":8080")
    
    // Start listening
    log.Println("Listening on :8080")
    if err := srv.Listen(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### UDP Datagram Server

Connectionless UDP server:

```go
package main

import (
    "context"
    "log"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/udp"
)

func main() {
    handler := func(r socket.Reader, w socket.Writer) {
        defer r.Close()
        defer w.Close()
        
        buf := make([]byte, 65507) // Max UDP datagram
        for {
            n, err := r.Read(buf)
            if err != nil {
                break
            }
            w.Write(buf[:n]) // Reply to sender
        }
    }
    
    srv := udp.New(nil, handler)
    srv.RegisterServer(":8080")
    
    log.Println("UDP server on :8080")
    if err := srv.Listen(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Unix Domain Socket (IPC)

Fast inter-process communication:

```go
package main

import (
    "context"
    "io"
    "log"
    "os"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/unix"
)

func main() {
    handler := func(r socket.Reader, w socket.Writer) {
        defer r.Close()
        defer w.Close()
        io.Copy(w, r)
    }
    
    srv := unix.New(nil, handler)
    srv.RegisterSocket("/tmp/app.sock", 0600, -1)
    
    defer os.Remove("/tmp/app.sock")
    
    log.Println("Unix socket: /tmp/app.sock")
    if err := srv.Listen(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

### Graceful Shutdown

Context-aware shutdown with connection draining:

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/tcp"
)

func main() {
    handler := func(r socket.Reader, w socket.Writer) {
        defer r.Close()
        defer w.Close()
        // Handle connection...
    }
    
    srv := tcp.New(nil, handler)
    srv.RegisterServer(":8080")
    
    // Start server in goroutine
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        if err := srv.Listen(ctx); err != nil {
            log.Println("Server stopped:", err)
        }
    }()
    
    // Wait for shutdown signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
    
    // Graceful shutdown (25s timeout)
    shutdownCtx, shutdownCancel := context.WithTimeout(
        context.Background(), 
        25*time.Second,
    )
    defer shutdownCancel()
    
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Println("Forced shutdown:", err)
    }
    
    log.Println("Server stopped")
}
```

### Connection Monitoring

Track connections and errors with callbacks:

```go
package main

import (
    "context"
    "log"
    "net"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/tcp"
)

func main() {
    handler := func(r socket.Reader, w socket.Writer) {
        // Handle connection...
    }
    
    srv := tcp.New(nil, handler)
    srv.RegisterServer(":8080")
    
    // Error callback
    srv.RegisterFuncError(func(err error) {
        log.Printf("ERROR: %v", err)
    })
    
    // Connection events
    srv.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
        log.Printf("Connection %s -> %s: %v", remote, local, state)
    })
    
    // Server lifecycle
    srv.RegisterFuncInfoServer(func(msg string) {
        log.Printf("SERVER: %s", msg)
    })
    
    // Monitor connections
    go func() {
        for {
            log.Printf("Active connections: %d", srv.OpenConnections())
            time.Sleep(5 * time.Second)
        }
    }()
    
    srv.Listen(context.Background())
}
```

---

## Subpackages

### `tcp` Subpackage

Connection-oriented TCP server with TLS support.

**Features**
- Persistent client connections
- Per-connection handler goroutines
- Connection tracking (OpenConnections)
- Optional TLS encryption
- Half-close support (CloseRead/CloseWrite)
- Graceful shutdown with draining

**API Example**

```go
import "github.com/nabbar/golib/socket/server/tcp"

srv := tcp.New(nil, handler)
srv.RegisterServer(":8080")

// Optional TLS
srv.SetTLS(true, tlsConfig)

// Start server
srv.Listen(ctx)
```

**Use Cases**
- HTTP/HTTPS servers
- Database connections
- Long-lived RPC
- Stateful protocols

**Characteristics**
- Memory: ~4KB per connection
- Connections: 10,000+ concurrent
- Throughput: ~800 MB/s
- Test Coverage: 84.6% (117 specs, ~28s)
- Thread Safety: ✅ Zero data races

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server/tcp) for complete API.

---

### `udp` Subpackage

Connectionless UDP server for fast datagram processing.

**Features**
- Stateless datagram handling
- Single handler for all packets
- Sender address tracking for replies
- No connection overhead
- OpenConnections returns 1/0 (running/stopped)

**API Example**

```go
import "github.com/nabbar/golib/socket/server/udp"

srv := udp.New(nil, handler)
srv.RegisterServer(":8080")

// Start server
srv.Listen(ctx)
```

**Use Cases**
- Real-time gaming
- Video/audio streaming
- DNS servers
- Broadcast/multicast

**Characteristics**
- Memory: ~8KB total (stateless)
- Throughput: ~900 MB/s
- Latency: <1ms
- Test Coverage: 72.0% (18 specs, ~1.4s)
- Thread Safety: ✅ Zero data races

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server/udp) for complete API.

---

### `unix` Subpackage

Connection-oriented Unix domain socket server for IPC.

**Features**
- Filesystem-based IPC
- Persistent connections (like TCP)
- File permissions and group ownership
- Half-close support
- Lower overhead than TCP (no network stack)
- Automatic socket file cleanup

**API Example**

```go
import "github.com/nabbar/golib/socket/server/unix"

srv := unix.New(nil, handler)
srv.RegisterSocket("/tmp/app.sock", 0600, -1)

// Start server
srv.Listen(ctx)
```

**Use Cases**
- Container IPC (Docker, Kubernetes)
- Microservice communication
- Database connections (PostgreSQL, MySQL)
- Process coordination

**Characteristics**
- Memory: ~4KB per connection
- Throughput: ~1.2 GB/s
- Latency: <0.5ms
- Platform: Linux only
- Test Coverage: 73.8% (23 specs, ~2s)
- Thread Safety: ✅ Zero data races

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server/unix) for complete API.

---

### `unixgram` Subpackage

Connectionless Unix domain datagram socket for fast IPC.

**Features**
- Filesystem-based IPC (like Unix)
- Datagram mode (like UDP)
- File permissions
- Fastest IPC option
- Stateless operation

**API Example**

```go
import "github.com/nabbar/golib/socket/server/unixgram"

srv := unixgram.New(nil, handler)
srv.RegisterSocket("/tmp/app.sock", 0600, -1)

// Start server
srv.Listen(ctx)
```

**Use Cases**
- Fast notifications
- Log collection
- Metrics gathering
- Event broadcasting

**Characteristics**
- Memory: ~8KB total (stateless)
- Throughput: ~1.5 GB/s (fastest)
- Latency: <0.2ms
- Platform: Linux only
- Test Coverage: 71.2% (20 specs, ~2.4s)
- Thread Safety: ✅ Zero data races

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server/unixgram) for complete API.

---

## Best Practices

**Always Close Resources**
```go
// ✅ Good: Defer cleanup
func handler(r socket.Reader, w socket.Writer) {
    defer r.Close()
    defer w.Close()
    
    // Process connection...
}

// ❌ Bad: Missing cleanup
func handlerBad(r socket.Reader, w socket.Writer) {
    // May leak resources
}
```

**Handle Errors Properly**
```go
// ✅ Good: Check all errors
if err := srv.RegisterServer(":8080"); err != nil {
    return fmt.Errorf("register: %w", err)
}

if err := srv.Listen(ctx); err != nil {
    return fmt.Errorf("listen: %w", err)
}

// ❌ Bad: Ignore errors
srv.RegisterServer(":8080")
srv.Listen(ctx)
```

**Use Context for Cancellation**
```go
// ✅ Good: Context-aware
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

go srv.Listen(ctx)

// Wait or cancel...
cancel()

// ❌ Bad: No cancellation mechanism
go srv.Listen(context.Background())
```

**Graceful Shutdown**
```go
// ✅ Good: Drain connections
shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
defer cancel()

// Shutdown waits for connections to close
if err := srv.Shutdown(shutdownCtx); err != nil {
    log.Println("Forced shutdown:", err)
}

// ❌ Bad: Immediate close
srv.Close() // Drops active connections
```

**Monitor Server State**
```go
// ✅ Good: Use callbacks
srv.RegisterFuncError(func(err error) {
    log.Printf("Error: %v", err)
})

srv.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
    log.Printf("%s: %s -> %s", state, remote, local)
})

// Check connection count
if srv.OpenConnections() > maxConnections {
    // Take action...
}
```

**Choose Right Protocol**
```go
// TCP: Reliable, ordered, connection-oriented
// Use for: Web servers, databases, long sessions
srv := tcp.New(nil, handler)

// UDP: Fast, unreliable, connectionless
// Use for: Real-time, gaming, streaming
srv := udp.New(nil, handler)

// Unix: Fast IPC, connection-oriented
// Use for: Container IPC, microservices
srv := unix.New(nil, handler)

// Unixgram: Fastest IPC, connectionless
// Use for: Notifications, metrics
srv := unixgram.New(nil, handler)
```

**Secure Unix Sockets**
```go
// ✅ Good: Restrict permissions
srv.RegisterSocket("/tmp/app.sock", 0600, -1) // Owner only

// For group access
srv.RegisterSocket("/tmp/app.sock", 0660, 1000) // Owner + group 1000

// ❌ Bad: World-accessible
srv.RegisterSocket("/tmp/app.sock", 0666, -1) // Anyone can connect
```

---

## Testing

**Test Suite**: 178 specs using Ginkgo v2 and Gomega

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Specific subpackage
go test -v ./tcp/
go test -v ./udp/
go test -v ./unix/
go test -v ./unixgram/
```

**Test Results**

| Subpackage | Specs | Coverage | Duration | Duration (race) |
|------------|-------|----------|----------|-----------------|
| TCP | 117 | 84.6% | ~28s | ~30s |
| UDP | 18 | 72.0% | ~1.4s | ~2.5s |
| Unix | 23 | 73.8% | ~2s | ~3s |
| Unixgram | 20 | 71.2% | ~2.4s | ~3.4s |
| **Total** | **178** | **≥70%** | **~34s** | **~39s** |

**Coverage Areas**
- Server lifecycle (start, stop, shutdown)
- Connection handling (accept, read, write, close)
- Graceful shutdown with connection draining
- Concurrent client operations
- TLS configuration (TCP)
- File permissions and ownership (Unix sockets)
- Callback registration and invocation
- Error conditions and edge cases
- Thread safety validation

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Goroutine synchronization with `context.Context`
- ✅ Atomic operation correctness (`atomic.Bool`, `atomic.Int64`)
- ✅ Connection counter accuracy
- ✅ Graceful shutdown behavior

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥70%)
- Follow existing code style and patterns
- Ensure thread safety with atomic operations

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes
- Document protocol-specific behavior
- Include ASCII diagrams for complex architectures

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Test graceful shutdown scenarios
- Add comments explaining complex test scenarios
- Use descriptive test names

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results (with and without `-race`)
- Update documentation
- Verify all quality checks pass

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Protocol Support**
- SCTP (Stream Control Transmission Protocol)
- QUIC (UDP-based multiplexed transport)
- WebSocket server wrapper
- HTTP/2 and HTTP/3 integration

**Features**
- Connection pooling and reuse
- Rate limiting per client
- Automatic reconnection handling
- Prometheus metrics integration
- OpenTelemetry tracing
- Dynamic TLS certificate loading
- Unix socket authentication (SO_PEERCRED)
- Priority queues for connections

**Performance**
- io_uring support (Linux)
- Zero-copy networking (sendfile)
- Connection multiplexing
- Load balancing across goroutines

**Monitoring**
- Built-in health checks
- Connection statistics
- Bandwidth monitoring
- Client IP whitelisting/blacklisting

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: 
  - [socket.Server Interface](https://pkg.go.dev/github.com/nabbar/golib/socket#Server)
  - [tcp Package](https://pkg.go.dev/github.com/nabbar/golib/socket/server/tcp)
  - [udp Package](https://pkg.go.dev/github.com/nabbar/golib/socket/server/udp)
  - [unix Package](https://pkg.go.dev/github.com/nabbar/golib/socket/server/unix)
  - [unixgram Package](https://pkg.go.dev/github.com/nabbar/golib/socket/server/unixgram)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
