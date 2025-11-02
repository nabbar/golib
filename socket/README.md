# Socket Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Unified, high-performance socket library for Go with thread-safe client and server implementations across TCP, UDP, and Unix domain sockets, featuring TLS support, graceful shutdown, and comprehensive connection management.

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
  - [client - Socket Clients](#client-subpackage)
  - [server - Socket Servers](#server-subpackage)
  - [config - Configuration Builders](#config-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This library provides production-ready socket communication for Go applications with a unified interface across multiple transport protocols. It emphasizes thread safety, protocol abstraction, and flexible callback mechanisms while supporting both client and server implementations for TCP, UDP, and Unix domain sockets.

### Design Philosophy

1. **Unified Interface**: Single `socket.Client` and `socket.Server` interfaces for all protocols
2. **Thread-Safe**: Atomic operations and proper synchronization prevent race conditions
3. **Protocol Agnostic**: Factory pattern abstracts protocol-specific implementations
4. **Context-Aware**: All operations support context for timeouts and cancellation
5. **Callback-Driven**: Asynchronous error and state notifications for monitoring

---

## Key Features

- **Multiple Protocols**:
  - **TCP**: Connection-oriented, reliable, ordered, with optional TLS encryption
  - **UDP**: Connectionless, fast, datagram-based for low-latency scenarios
  - **Unix**: IPC via filesystem sockets (stream mode, connection-oriented)
  - **Unixgram**: IPC via filesystem sockets (datagram mode, connectionless)
- **Thread-Safe Operations**: Atomic state management (`atomic.Bool`, `atomic.Int64`, `atomic.Map`)
- **TLS Support**: Built-in encryption for TCP with certificate validation
- **Callback System**:
  - **Error Callbacks**: Asynchronous error notifications
  - **State Callbacks**: Connection lifecycle tracking
  - **Server Callbacks**: Server lifecycle and listening state notifications
- **Graceful Shutdown**: Context-aware connection draining for servers
- **Connection Management**: Per-client goroutines, connection tracking, lifecycle hooks
- **Platform Support**:
  - Linux: All protocols (TCP, UDP, Unix, Unixgram)
  - Darwin/macOS: All protocols
  - Windows/Other: TCP and UDP only
- **Standard Interfaces**: Implements `io.Reader`, `io.Writer`, `io.Closer`
- **One-Shot Operations**: Convenient `Once()` method for request/response patterns

---

## Installation

```bash
go get github.com/nabbar/golib/socket
```

---

## Architecture

### Package Structure

The package is organized into client and server subpackages with protocol-specific implementations:

```
socket/
├── interface.go            # Core interfaces (Client, Server, Reader, Writer)
├── io.go                   # I/O abstractions with state tracking
├── client/                 # Client implementations
│   ├── tcp/               # TCP client (with TLS)
│   ├── udp/               # UDP client
│   ├── unix/              # Unix domain socket client (stream)
│   └── unixgram/          # Unix domain socket client (datagram)
├── server/                # Server implementations
│   ├── tcp/               # TCP server (with TLS)
│   ├── udp/               # UDP server
│   ├── unix/              # Unix domain socket server (stream)
│   └── unixgram/          # Unix domain socket server (datagram)
└── config/                # Configuration builders
    ├── client.go          # Client configuration builder
    └── server.go          # Server configuration builder
```

### Component Overview

```
┌────────────────────────────────────────────────────────┐
│                    Socket Package                      │
│            Unified Client & Server Interfaces          │
└──────────────┬───────────────┬─────────────────────────┘
               │               │
      ┌────────▼────────┐  ┌───▼────────────┐
      │     Client      │  │     Server     │
      │                 │  │                │
      │  TCP, UDP       │  │  TCP, UDP      │
      │  Unix, Unixgram │  │  Unix, Unixgram│
      └────────┬────────┘  └────┬───────────┘
               │                │
       ┌───────▼────────────────▼──────────┐
       │       Protocol Implementations    │
       │                                   │
       │  ┌─────┐  ┌─────┐  ┌──────┐       │
       │  │ TCP │  │ UDP │  │ Unix │       │
       │  └─────┘  └─────┘  └──────┘       │
       └───────────────────────────────────┘
```

### Transport Comparison

| Protocol | Type | Scope | Reliable | Ordered | TLS | Best For |
|----------|------|-------|----------|---------|-----|----------|
| **TCP** | Connection | Network | ✅ | ✅ | ✅ | Web servers, APIs, databases |
| **UDP** | Datagram | Network | ❌ | ❌ | ❌ | Real-time, gaming, streaming |
| **Unix** | Connection | Local | ✅ | ✅ | ❌ | Container IPC, microservices |
| **Unixgram** | Datagram | Local | ❌ | ❌ | ❌ | Fast IPC, notifications |

### Connection Models

**Connection-Oriented (TCP, Unix)**
- Persistent client-server connections
- Per-connection handler goroutines (servers)
- Connection state tracking
- Graceful connection draining on shutdown
- Best for: Long-lived sessions, stateful protocols

**Connectionless (UDP, Unixgram)**
- Stateless datagram processing
- Single handler for all datagrams (servers)
- No connection tracking (returns 1/0 for running/stopped)
- Immediate shutdown (no draining)
- Best for: Request-response, fire-and-forget, real-time data

---

## Performance

### Memory Efficiency

All implementations maintain **O(1) memory per connection** or **O(1) total for datagram**:

- **TCP/Unix Servers**: One goroutine per client (~2KB stack + buffers)
- **UDP/Unixgram Servers**: Single handler goroutine
- **Clients**: Constant memory usage regardless of data size
- **Buffer Management**: `bufio.Reader/Writer` with configurable sizes (default 32KB)
- **Zero Allocations**: Atomic state storage prevents unnecessary allocations

### Thread Safety

All operations are thread-safe through:

- **Atomic Operations**: `atomic.Bool`, `atomic.Int64`, `atomic.Map` for state
- **Mutex Protection**: `sync.Mutex` for callback registration where needed
- **Goroutine Synchronization**: Proper lifecycle management with `context.Context`
- **Verified**: Tested with `go test -race` (zero data races across 502 specs)

### Throughput Benchmarks

| Component | Protocol | Operation | Throughput | Latency | Notes |
|-----------|----------|-----------|------------|---------|-------|
| **Client** | TCP | Send/Receive | ~1.1-1.2 GB/s | <1ms | Localhost |
| **Client** | TCP+TLS | Send/Receive | ~750-800 MB/s | <2ms | AES-128-GCM |
| **Client** | UDP | Datagram | ~900 MB/s | <0.5ms | 1472-byte packets |
| **Client** | Unix | Send/Receive | ~1.7-1.8 GB/s | <0.5ms | Kernel-only |
| **Client** | Unixgram | Datagram | ~1.5 GB/s | <0.2ms | Fastest IPC |
| **Server** | TCP | Stream | ~800 MB/s | <1ms | 10,000+ connections |
| **Server** | UDP | Datagram | ~900 MB/s | <0.5ms | Stateless |
| **Server** | Unix | Stream | ~1.2 GB/s | <0.5ms | IPC optimized |
| **Server** | Unixgram | Datagram | ~1.5 GB/s | <0.2ms | Fastest |

*Measured on AMD64, Linux 5.x, Go 1.21+*

### Capacity

**TCP Servers**: 10,000+ concurrent connections (tested up to 50,000 on high-memory systems)  
**Unix Servers**: 1,000+ concurrent connections (tested up to 10,000 on Linux)  
**UDP/Unixgram Servers**: Unlimited concurrent senders (stateless operation)

### Protocol Selection Guide

```
Reliability:    TCP > Unix > UDP ≈ Unixgram
Speed:          Unixgram > Unix > UDP > TCP
Overhead:       UDP/Unixgram < Unix < TCP

Network Communication:
├─ Reliable & Secure → TCP (with TLS)
├─ Reliable, Fast → TCP
└─ Low Latency → UDP

Local IPC:
├─ Reliable → Unix (Docker, databases)
└─ Fast → Unixgram (metrics, events)
```

---

## Use Cases

This library is designed for scenarios requiring reliable socket communication:

**Web Applications**
- HTTP/HTTPS servers with TLS support
- WebSocket backends with connection management
- REST API endpoints with graceful shutdown
- Long-polling connections with timeout control

**Microservices Communication**
- TCP with TLS for secure inter-service communication
- Unix sockets for same-host, high-performance IPC
- Context-aware operations for service mesh integration
- Callback-driven monitoring and observability

**Real-Time Systems**
- UDP-based game servers with low latency
- Video/audio streaming with datagram delivery
- Sensor data collection with minimal overhead
- Time-series data ingestion at high frequency

**Database Proxies & Middleware**
- Protocol translation layers with connection pooling
- Query routing with connection state tracking
- Cache layers with persistent connections
- Load balancing with graceful connection draining

**Local IPC**
- Docker daemon communication via Unix sockets
- Database connections (PostgreSQL, MySQL Unix sockets)
- System daemon control (systemd, dbus)
- High-speed inter-process messaging

**Network Daemons**
- Custom protocol servers with TLS
- Tunneling and proxying with connection tracking
- Network monitoring agents with callback notifications
- Log aggregation services with datagram support

---

## Quick Start

### TCP Echo Server

Simple TCP server with graceful shutdown:

```go
package main

import (
    "context"
    "io"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
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
    
    // Start server in goroutine
    ctx, cancel := context.WithCancel(context.Background())
    go func() {
        if err := srv.Listen(ctx); err != nil {
            log.Println("Server stopped:", err)
        }
    }()
    
    log.Println("Listening on :8080")
    
    // Wait for shutdown signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down...")
    cancel()
    
    // Graceful shutdown with timeout
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

### TCP Client

Simple TCP client with error handling:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/nabbar/golib/network/protocol"
    "github.com/nabbar/golib/socket/client"
)

func main() {
    // Create TCP client
    cli, err := client.New(protocol.NetworkTCP, "localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()
    
    // Connect with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := cli.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Write data
    _, err = cli.Write([]byte("Hello, server!\n"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Read response
    buf := make([]byte, 4096)
    n, err := cli.Read(buf)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", buf[:n])
}
```

### UDP Server & Client

Connectionless datagram communication:

```go
package main

import (
    "context"
    "log"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/udp"
    "github.com/nabbar/golib/socket/client/udp"
)

// Server
func runServer() {
    handler := func(r socket.Reader, w socket.Writer) {
        defer r.Close()
        defer w.Close()
        
        buf := make([]byte, 65507) // Max UDP datagram
        for {
            n, err := r.Read(buf)
            if err != nil {
                break
            }
            log.Printf("Received: %s", buf[:n])
            w.Write(buf[:n]) // Reply
        }
    }
    
    srv := udp.New(nil, handler)
    srv.RegisterServer(":8125")
    srv.Listen(context.Background())
}

// Client
func sendMetric() {
    cli, _ := udp.New("localhost:8125")
    defer cli.Close()
    
    cli.Connect(context.Background())
    cli.Write([]byte("myapp.requests:1|c"))
}
```

### Unix Socket IPC

High-performance local communication:

```go
package main

import (
    "context"
    "io"
    "log"
    "os"
    
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/server/unix"
    "github.com/nabbar/golib/socket/client/unix"
)

// Server
func runServer() {
    handler := func(r socket.Reader, w socket.Writer) {
        defer r.Close()
        defer w.Close()
        io.Copy(w, r)
    }
    
    srv := unix.New(nil, handler)
    srv.RegisterSocket("/tmp/app.sock", 0600, -1)
    
    defer os.Remove("/tmp/app.sock")
    
    log.Println("Unix socket: /tmp/app.sock")
    srv.Listen(context.Background())
}

// Client
func connect() {
    cli := unix.New("/tmp/app.sock")
    if cli == nil {
        log.Fatal("Unix sockets not available on this platform")
    }
    defer cli.Close()
    
    cli.Connect(context.Background())
    cli.Write([]byte("Hello via Unix socket"))
}
```

### Connection Monitoring

Track connections and errors:

```go
package main

import (
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
            log.Printf("Active: %d connections", srv.OpenConnections())
            time.Sleep(5 * time.Second)
        }
    }()
    
    srv.Listen(context.Background())
}
```

For more examples, see the [client](client/README.md) and [server](server/README.md) documentation.

---

## Subpackages

### `client` Subpackage

Thread-safe, multi-protocol socket client implementations with unified interface.

**Features**
- TCP, UDP, Unix, and Unixgram support
- Optional TLS encryption for TCP
- Atomic state management (`atomic.Map`)
- Callback mechanisms for errors and state changes
- One-shot `Once()` method for request/response patterns
- Platform-aware (Unix sockets only on Linux/Darwin)

**Quick Example**
```go
import "github.com/nabbar/golib/socket/client/tcp"

cli, _ := tcp.New("localhost:8080")
defer cli.Close()

cli.Connect(context.Background())
cli.Write([]byte("data"))
```

**Characteristics**
- Total Specs: 324 (all passing)
- Coverage: 74.2%
- Execution Time: ~112s (without race), ~180s (with race)
- Thread Safety: ✅ Zero data races

**Subpackages**
- **tcp**: Connection-oriented with TLS support (119 specs, 74.0% coverage)
- **udp**: Connectionless datagram (73 specs, 73.7% coverage)
- **unix**: IPC stream socket (67 specs, 76.3% coverage)
- **unixgram**: IPC datagram socket (65 specs, 76.8% coverage)

See [client/README.md](client/README.md) for comprehensive documentation.

---

### `server` Subpackage

High-performance, protocol-agnostic socket server implementations with graceful shutdown.

**Features**
- TCP, UDP, Unix, and Unixgram support
- Per-connection handler goroutines (TCP, Unix)
- Optional TLS encryption for TCP
- Graceful shutdown with connection draining
- Connection tracking (`OpenConnections()`)
- File permissions for Unix sockets
- Callback system for monitoring

**Quick Example**
```go
import "github.com/nabbar/golib/socket/server/tcp"

srv := tcp.New(nil, handler)
srv.RegisterServer(":8080")
srv.Listen(context.Background())
```

**Characteristics**
- Total Specs: 178 (all passing)
- Coverage: ≥70% (71.2%-84.6% by subpackage)
- Execution Time: ~34s (without race), ~39s (with race)
- Thread Safety: ✅ Zero data races

**Subpackages**
- **tcp**: Connection-oriented with TLS (117 specs, 84.6% coverage, ~28s)
- **udp**: Connectionless datagram (18 specs, 72.0% coverage, ~1.4s)
- **unix**: IPC stream socket (23 specs, 73.8% coverage, ~2s)
- **unixgram**: IPC datagram socket (20 specs, 71.2% coverage, ~2.4s)

See [server/README.md](server/README.md) for comprehensive documentation.

---

### `config` Subpackage

Configuration builders for creating clients and servers with fluent API.

**Features**
- Builder pattern for configuration
- Type-safe protocol selection
- TLS configuration support
- Callback registration
- Validation and error handling

**Quick Example**
```go
import "github.com/nabbar/golib/socket/config"

// Server configuration
cfg := config.NewServer().
    Network(config.NetworkTCP).
    Address(":8080").
    Handler(myHandler).
    TLS(true, tlsConfig)

server, err := cfg.Build(ctx)
if err != nil {
    log.Fatal(err)
}

server.Listen(ctx)
```

**Characteristics**
- Coverage: 0.0% (configuration only, no business logic to test)
- Thread Safety: ✅ Atomic operations

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/config) for complete API.

---

## Best Practices

**Always Check Errors**
```go
// ✅ Good
cli, err := client.New(protocol.NetworkTCP, "localhost:8080")
if err != nil {
    return fmt.Errorf("create client: %w", err)
}
defer cli.Close()

// ❌ Bad: Ignoring errors
cli, _ := client.New(protocol.NetworkTCP, "localhost:8080")
```

**Use Context for Timeouts**
```go
// ✅ Good: Timeout protection
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := cli.Connect(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return fmt.Errorf("connection timeout")
    }
    return err
}

// ❌ Bad: No timeout
cli.Connect(context.Background()) // May hang forever
```

**Close All Resources**
```go
// ✅ Good: Defer cleanup
srv := tcp.New(nil, handler)
srv.RegisterServer(":8080")

ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go srv.Listen(ctx)

// ... later ...
shutdownCtx, shutdownCancel := context.WithTimeout(
    context.Background(), 
    25*time.Second,
)
defer shutdownCancel()
srv.Shutdown(shutdownCtx)

// ❌ Bad: Immediate close
srv.Close() // Drops active connections
```

**Handle Callbacks Safely**
```go
// ✅ Good: Non-blocking, error handling
srv.RegisterFuncError(func(err error) {
    log.Printf("socket error: %v", err)
    // Optionally: send to error channel, metrics, etc.
})

// ❌ Bad: Blocking operations in callbacks
srv.RegisterFuncError(func(err error) {
    // Don't do this!
    reconnect() // Blocks callback goroutine
    sendEmail() // Slow operation
    panic("error") // Crashes callback goroutine
})
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

**Graceful Shutdown**
```go
// ✅ Good: Wait for connections to drain
shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
defer cancel()

if err := srv.Shutdown(shutdownCtx); err != nil {
    log.Println("Forced shutdown:", err)
}

// Monitor during shutdown
for srv.OpenConnections() > 0 {
    log.Printf("Waiting for %d connections...", srv.OpenConnections())
    time.Sleep(1 * time.Second)
}
```

---

## Testing

**Test Suite**: 502 specs across all subpackages using Ginkgo v2 and Gomega

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Specific subpackage
go test -v ./client/tcp/
go test -v ./server/tcp/
```

**Test Results Summary**

| Component | Subpackage | Specs | Coverage | Duration | Duration (race) |
|-----------|------------|-------|----------|----------|-----------------|
| **Client** | TCP | 119 | 74.0% | ~88.3s | ~90s |
| **Client** | UDP | 73 | 73.7% | ~8.1s | ~9s |
| **Client** | Unix | 67 | 76.3% | ~13.2s | ~14s |
| **Client** | Unixgram | 65 | 76.8% | ~2.9s | ~4s |
| **Server** | TCP | 117 | 84.6% | ~27.7s | ~28.3s |
| **Server** | UDP | 18 | 72.0% | ~1.4s | ~2.5s |
| **Server** | Unix | 23 | 73.8% | ~2.0s | ~3s |
| **Server** | Unixgram | 20 | 71.2% | ~2.4s | ~3.4s |
| **Total** | **All** | **502** | **≥70%** | **~146s** | **~219s** |

**Coverage Areas**
- Connection lifecycle (connect, read, write, close)
- Graceful shutdown with connection draining
- Concurrent client/server operations
- TLS configuration and encryption
- Error handling and edge cases
- Callback mechanisms (error, state, server lifecycle)
- Thread safety validation (atomic operations)
- Platform-specific implementations

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Goroutine synchronization with `context.Context`
- ✅ Atomic operation correctness (`atomic.Bool`, `atomic.Int64`, `atomic.Map`)
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
- Include usage examples in GoDoc comments

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Test graceful shutdown scenarios
- Test platform-specific code on target platforms
- Add comments explaining complex test scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results (with and without `-race`)
- Update documentation
- Verify all quality checks pass

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Protocol Support**
- SCTP (Stream Control Transmission Protocol)
- QUIC (UDP-based multiplexed transport)
- WebSocket server/client wrappers
- HTTP/2 and HTTP/3 integration

**Features**
- Connection pooling with lifecycle management
- Automatic reconnection with exponential backoff
- Rate limiting per client/connection
- Prometheus metrics integration
- OpenTelemetry tracing
- Dynamic TLS certificate loading
- Unix socket authentication (SO_PEERCRED)
- Priority queues for connections
- Circuit breaker pattern integration

**Performance**
- io_uring support (Linux)
- Zero-copy networking (sendfile)
- Connection multiplexing
- Buffer pooling with sync.Pool
- Batch operations for UDP

**Monitoring**
- Built-in health checks
- Connection statistics and analytics
- Bandwidth monitoring per connection
- Client IP whitelisting/blacklisting
- Request/response correlation IDs

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: 
  - [Socket Package](https://pkg.go.dev/github.com/nabbar/golib/socket)
  - [Client Package](https://pkg.go.dev/github.com/nabbar/golib/socket/client)
  - [Server Package](https://pkg.go.dev/github.com/nabbar/golib/socket/server)
  - [Config Package](https://pkg.go.dev/github.com/nabbar/golib/socket/config)
- **Subpackage Documentation**:
  - [Client README](client/README.md)
  - [Server README](server/README.md)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../CONTRIBUTING.md)
