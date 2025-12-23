# Socket Package

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)](TESTING.md)

Unified, production-ready socket communication framework for TCP, UDP, and Unix domain sockets with optional TLS encryption, comprehensive configuration, and platform-aware protocol selection.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Package Structure](#package-structure)
  - [Component Diagram](#component-diagram)
  - [Protocol Selection](#protocol-selection)
- [Performance](#performance)
  - [Protocol Comparison](#protocol-comparison)
  - [Benchmarks](#benchmarks)
  - [Scalability](#scalability)
- [Subpackages](#subpackages)
  - [client](#client)
  - [server](#server)
  - [config](#config)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic TCP Server](#basic-tcp-server)
  - [Basic TCP Client](#basic-tcp-client)
  - [Unix Socket Server](#unix-socket-server)
  - [UDP Server](#udp-server)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Core Interfaces](#core-interfaces)
  - [Connection States](#connection-states)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **socket** package provides a comprehensive, production-ready framework for network socket communication in Go. It offers unified interfaces for both client and server implementations across multiple protocols (TCP, UDP, Unix domain sockets) with optional TLS encryption, automatic platform detection, and comprehensive error handling.

This package serves as the foundation for all socket-based communication in golib, providing platform-aware abstractions that work seamlessly across different network protocols and operating systems.

### Design Philosophy

1. **Unified Interface**: All socket types implement common interfaces (Server, Client, Context)
2. **Platform Awareness**: Automatic protocol availability based on operating system
3. **Type Safety**: Configuration-driven construction with compile-time validation
4. **Performance First**: Zero-copy operations and minimal allocations where possible
5. **Production Ready**: Built-in error handling, logging, and monitoring capabilities
6. **Concurrent by Design**: Thread-safe operations with atomic state management
7. **Standard Compliance**: Implements io.Reader, io.Writer, io.Closer, context.Context

### Key Features

- ✅ **Multiple Protocols**: TCP, UDP, Unix domain sockets, Unix datagrams
- ✅ **TLS/SSL Support**: Optional encryption for TCP connections
- ✅ **Platform-Aware**: Automatic Unix socket support on Linux/Darwin
- ✅ **Unified API**: Consistent interface across all protocols
- ✅ **Configuration Builders**: Type-safe configuration with validation
- ✅ **Connection Monitoring**: State tracking and event callbacks
- ✅ **Error Handling**: Comprehensive error propagation and filtering
- ✅ **Context Integration**: Full support for Go's context.Context
- ✅ **Resource Management**: Automatic cleanup and graceful shutdown
- ✅ **High Performance**: Optimized for concurrent, high-throughput scenarios

---

## Architecture

### Package Structure

```
socket/                           # Core interfaces and types
├── interface.go                  # Server, Client interfaces
├── context.go                    # Context interface
├── doc.go                        # Package documentation
├── example_test.go               # Example tests
├── socket_test.go                # Unit tests
│
├── config/                       # Configuration and validation
│   ├── client.go                 # Client configuration
│   ├── server.go                 # Server configuration
│   └── validator.go              # Validation logic
│
├── client/                       # Client factory and implementations
│   ├── interface.go              # Factory (New)
│   ├── tcp/                      # TCP client
│   ├── udp/                      # UDP client
│   ├── unix/                     # Unix socket client (Linux/Darwin)
│   └── unixgram/                 # Unix datagram client (Linux/Darwin)
│
└── server/                       # Server factory and implementations
    ├── interface.go              # Factory (New)
    ├── tcp/                      # TCP server
    ├── udp/                      # UDP server
    ├── unix/                     # Unix socket server (Linux/Darwin)
    └── unixgram/                 # Unix datagram server (Linux/Darwin)
```

### Component Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                       socket Package                               │
│                  (Core Interfaces & Types)                         │
├────────────────────────────────────────────────────────────────────┤
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │             Core Interfaces                              │      │
│  │  • Server   - Server operations                          │      │
│  │  • Client   - Client operations                          │      │
│  │  • Context  - Connection context                         │      │
│  └──────────────────────────────────────────────────────────┘      │
│                                                                    │
│  ┌──────────────────────────────────────────────────────────┐      │
│  │             Core Types                                   │      │
│  │  • ConnState       - Connection state tracking           │      │
│  │  • HandlerFunc     - Request handler                     │      │
│  │  • FuncError       - Error callback                      │      │
│  │  • FuncInfo        - Connection info callback            │      │
│  └──────────────────────────────────────────────────────────┘      │
│                                                                    │
└─────┬──────────────────────────────────────────────────┬───────────┘
      │                                                  │
      ▼                                                  ▼
┌───────────────────────┐                   ┌───────────────────────┐
│   client Package      │                   │   server Package      │
│   (Client Factory)    │                   │   (Server Factory)    │
├───────────────────────┤                   ├───────────────────────┤
│ • TCP Client          │                   │ • TCP Server          │
│ • UDP Client          │                   │ • UDP Server          │
│ • Unix Client         │                   │ • Unix Server         │
│ • UnixGram Client     │                   │ • UnixGram Server     │
└───────────────────────┘                   └───────────────────────┘
              │                                       │
              └───────────────┬───────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   config Package  │
                    │  (Configuration)  │
                    ├───────────────────┤
                    │ • Client Config   │
                    │ • Server Config   │
                    │ • Validation      │
                    └───────────────────┘
```

### Protocol Selection

The package automatically selects the appropriate implementation based on protocol type:

```
┌─────────────────────┬──────────────────┬─────────────────────┐
│  Protocol Value     │  Platform        │  Implementation     │
├─────────────────────┼──────────────────┼─────────────────────┤
│  NetworkTCP         │  All             │  tcp/*              │
│  NetworkTCP4        │  All             │  tcp/*              │
│  NetworkTCP6        │  All             │  tcp/*              │
│  NetworkUDP         │  All             │  udp/*              │
│  NetworkUDP4        │  All             │  udp/*              │
│  NetworkUDP6        │  All             │  udp/*              │
│  NetworkUnix        │  Linux/Darwin    │  unix/*             │
│  NetworkUnixGram    │  Linux/Darwin    │  unixgram/*         │
│  Other values       │  All             │  ErrInvalidProtocol │
└─────────────────────┴──────────────────┴─────────────────────┘
```

---

## Performance

### Protocol Comparison

| Protocol | Throughput | Latency | Best For | Platform |
|----------|-----------|---------|----------|----------|
| **TCP** | High | Low | Network IPC, reliability | All |
| **UDP** | Very High | Very Low | Datagrams, speed | All |
| **Unix** | Highest | Lowest | Local IPC, performance | Linux/Darwin |
| **UnixGram** | Highest | Lowest | Local datagrams | Linux/Darwin |

### Benchmarks

Based on actual test execution:

| Operation | Time | Notes |
|-----------|------|-------|
| **Factory Overhead** | <1µs | Negligible |
| **TCP Connection** | ~1-5ms | Network-dependent |
| **UDP Connection** | ~0ms | Connectionless |
| **Unix Connection** | ~35µs | Fastest |
| **Read/Write** | ~100µs | Protocol-dependent |
| **State Tracking** | <1µs | Atomic operations |

**Interface Operations:**

| Operation | Throughput | Latency |
|-----------|------------|---------|
| **ConnState.String()** | N/A | <10ns |
| **ErrorFilter()** | N/A | <50ns |
| **Context methods** | N/A | <100ns |

### Scalability

- **Concurrent Connections**: Tested with 1000+ concurrent connections
- **Factory Calls**: Thread-safe, tested with 100 concurrent goroutines
- **Memory per Connection**: ~32KB (configurable buffer)
- **Zero Race Conditions**: All tests pass with `-race` detector

---

## Subpackages

### client

**Purpose**: Unified factory for creating socket clients across protocols.

**Key Features**:
- Single entry point (New) for all protocols
- Platform-aware protocol selection
- TLS support for TCP
- Type-safe configuration

**Performance**: <1µs factory overhead

**Coverage**: 81.2%

**Documentation**: [client/README.md](client/README.md)

---

### server

**Purpose**: Unified factory for creating socket servers across protocols.

**Key Features**:
- Single entry point (New) for all protocols
- Platform-aware protocol selection
- TLS support for TCP servers
- Graceful shutdown
- Connection tracking

**Performance**: <1µs factory overhead

**Coverage**: 100.0%

**Documentation**: [server/README.md](server/README.md)

---

### config

**Purpose**: Configuration structures and validation for clients and servers.

**Key Features**:
- Declarative configuration API
- Comprehensive validation
- Platform compatibility checks
- TLS configuration support
- Unix socket permissions

**Performance**: <1ms validation

**Coverage**: 89.4%

**Documentation**: [config/README.md](config/README.md)

---

## Use Cases

### 1. TCP Server with TLS

**Problem**: Secure network service with encrypted connections.

```go
import (
    "github.com/nabbar/golib/socket/config"
    "github.com/nabbar/golib/socket/server"
)

cfg := config.Server{
    Network: protocol.NetworkTCP,
    Address: ":443",
    TLS: config.ServerTLS{
        Enabled: true,
    },
}

srv, err := server.New(nil, handleRequest, cfg)
if err != nil {
    log.Fatal(err)
}
defer srv.Close()

ctx := context.Background()
srv.Listen(ctx)
```

**Real-world**: HTTPS servers, secure APIs, encrypted microservices.

### 2. High-Performance Local IPC

**Problem**: Fast inter-process communication on same host.

```go
cfg := config.Server{
    Network:   protocol.NetworkUnix,
    Address:   "/tmp/app.sock",
    PermFile:  0660,
}

srv, err := server.New(nil, handleRequest, cfg)
if err == config.ErrInvalidProtocol {
    // Fall back to TCP on Windows
    cfg.Network = protocol.NetworkTCP
    cfg.Address = ":8080"
    srv, err = server.New(nil, handleRequest, cfg)
}
```

**Real-world**: Database connections, microservices on same host, container communication.

### 3. Real-Time Metrics Collection

**Problem**: Low-latency metrics and monitoring data collection.

```go
cfg := config.Server{
    Network: protocol.NetworkUDP,
    Address: ":9000",
}

srv, err := server.New(nil, handleMetrics, cfg)
```

**Real-world**: StatsD-like metrics, monitoring agents, real-time logging.

### 4. Multi-Protocol Service

**Problem**: Service accessible via multiple protocols simultaneously.

```go
// Network access via TCP
tcpSrv, _ := server.New(nil, handler, config.Server{
    Network: protocol.NetworkTCP,
    Address: ":8080",
})

// Local access via Unix socket
unixSrv, _ := server.New(nil, handler, config.Server{
    Network: protocol.NetworkUnix,
    Address: "/tmp/app.sock",
})

// Start both
go tcpSrv.Listen(ctx)
go unixSrv.Listen(ctx)
```

**Real-world**: Database servers, cache services, message brokers.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket
```

### Basic TCP Server

```go
package main

import (
    "context"
    "log"
    
    libsck "github.com/nabbar/golib/socket"
    libptc "github.com/nabbar/golib/network/protocol"
    sckcfg "github.com/nabbar/golib/socket/config"
    scksrv "github.com/nabbar/golib/socket/server"
)

func main() {
    // Define handler
    handler := func(ctx libsck.Context) {
        buf := make([]byte, 1024)
        n, _ := ctx.Read(buf)
        
        response := []byte("Echo: " + string(buf[:n]))
        ctx.Write(response)
    }
    
    // Create configuration
    cfg := sckcfg.Server{
        Network: libptc.NetworkTCP,
        Address: ":8080",
    }
    
    // Create server
    srv, err := scksrv.New(nil, handler, cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer srv.Close()
    
    // Start listening
    log.Println("Server listening on :8080")
    ctx := context.Background()
    if err := srv.Listen(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Basic TCP Client

```go
package main

import (
    "context"
    "log"
    
    libptc "github.com/nabbar/golib/network/protocol"
    sckcfg "github.com/nabbar/golib/socket/config"
    sckclt "github.com/nabbar/golib/socket/client"
)

func main() {
    // Create configuration
    cfg := sckcfg.Client{
        Network: libptc.NetworkTCP,
        Address: "localhost:8080",
    }
    
    // Create client
    cli, err := sckclt.New(cfg, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()
    
    // Connect
    ctx := context.Background()
    if err := cli.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Send message
    _, err = cli.Write([]byte("Hello, server!"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Read response
    buf := make([]byte, 1024)
    n, err := cli.Read(buf)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Response: %s", buf[:n])
}
```

### Unix Socket Server

Linux/Darwin only:

```go
cfg := sckcfg.Server{
    Network:   libptc.NetworkUnix,
    Address:   "/tmp/app.sock",
    PermFile:  0660,
    GroupPerm: -1,
}

srv, err := scksrv.New(nil, handler, cfg)
if err != nil {
    if err == sckcfg.ErrInvalidProtocol {
        log.Fatal("Unix sockets not supported on this platform")
    }
    log.Fatal(err)
}
defer srv.Close()

ctx := context.Background()
srv.Listen(ctx)
```

### UDP Server

```go
handler := func(ctx libsck.Context) {
    buf := make([]byte, 65536)
    n, _ := ctx.Read(buf)
    log.Printf("Received datagram from %s: %s", ctx.RemoteHost(), buf[:n])
}

cfg := sckcfg.Server{
    Network: libptc.NetworkUDP,
    Address: ":9000",
}

srv, _ := scksrv.New(nil, handler, cfg)
defer srv.Close()

ctx := context.Background()
srv.Listen(ctx)
```

---

## Best Practices

### ✅ DO

**Choose Appropriate Protocol:**
```go
// Network communication → TCP (reliable)
cfg.Network = protocol.NetworkTCP

// Local IPC → Unix sockets (fastest)
cfg.Network = protocol.NetworkUnix

// Datagrams → UDP (lowest latency)
cfg.Network = protocol.NetworkUDP
```

**Resource Management:**
```go
// Always close resources
srv, err := server.New(nil, handler, cfg)
if err != nil {
    return err
}
defer srv.Close()  // Ensure cleanup
```

**Context Usage:**
```go
// Use context for lifecycle control
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := srv.Listen(ctx)
```

**Error Handling:**
```go
// Filter expected errors
srv.RegisterFuncError(func(errs ...error) {
    for _, err := range errs {
        if err := socket.ErrorFilter(err); err != nil {
            log.Printf("Error: %v", err)
        }
    }
})
```

**Platform Compatibility:**
```go
// Handle platform-specific protocols
cli, err := client.New(cfg, nil)
if err == config.ErrInvalidProtocol {
    // Fall back to TCP on unsupported platforms
    cfg.Network = protocol.NetworkTCP
    cli, err = client.New(cfg, nil)
}
```

### ❌ DON'T

**Don't ignore protocol errors:**
```go
// ❌ BAD: Ignoring errors
cli, _ := client.New(cfg, nil)

// ✅ GOOD: Check errors
cli, err := client.New(cfg, nil)
if err != nil {
    return err
}
```

**Don't leave resources open:**
```go
// ❌ BAD: No cleanup
srv, _ := server.New(nil, handler, cfg)

// ✅ GOOD: Always defer Close
defer srv.Close()
```

**Don't assume platform support:**
```go
// ❌ BAD: Assuming Unix sockets available
cfg.Network = protocol.NetworkUnix
srv, _ := server.New(nil, handler, cfg)  // Fails on Windows

// ✅ GOOD: Check platform or handle errors
if err != nil && err == config.ErrInvalidProtocol {
    // Handle unsupported protocol
}
```

---

## API Reference

### Core Interfaces

**Server Interface:**
```go
type Server interface {
    io.Closer
    RegisterFuncError(f FuncError)
    RegisterFuncInfo(f FuncInfo)
    RegisterFuncInfoServer(f FuncInfoSrv)
    SetTLS(enable bool, config TLSConfig) error
    Listen(ctx context.Context) error
    Listener() (network NetworkProtocol, listener string, tls bool)
    Shutdown(ctx context.Context) error
    IsRunning() bool
    IsGone() bool
    OpenConnections() int64
}
```

**Client Interface:**
```go
type Client interface {
    io.ReadWriteCloser
    SetTLS(enable bool, config TLSConfig, serverName string) error
    RegisterFuncError(f FuncError)
    RegisterFuncInfo(f FuncInfo)
    Connect(ctx context.Context) error
    IsConnected() bool
    Once(ctx context.Context, request io.Reader, fct Response) error
}
```

**Context Interface:**
```go
type Context interface {
    context.Context  // Deadline, Done, Err, Value
    io.Reader        // Read from connection
    io.Writer        // Write to connection
    io.Closer        // Close connection
    IsConnected() bool
    RemoteHost() string
    LocalHost() string
}
```

### Connection States

```go
const (
    ConnectionDial       // Client dialing
    ConnectionNew        // New connection
    ConnectionRead       // Reading data
    ConnectionCloseRead  // Closing read
    ConnectionHandler    // Handler executing
    ConnectionWrite      // Writing data
    ConnectionCloseWrite // Closing write
    ConnectionClose      // Closing connection
)
```

### Error Handling

```go
// ErrorFilter removes expected network errors
func ErrorFilter(err error) error
```

**Usage:**
```go
if err := socket.ErrorFilter(err); err != nil {
    // Handle unexpected error
}
```

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
   - Ensure zero race conditions
   - Maintain coverage above 80%

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

- ✅ **100% test coverage** for root package (target: >80%)
- ✅ **90.3% average coverage** across all subpackages
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementations throughout
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**New Features:**
1. WebSocket support (ws/wss protocols)
2. QUIC protocol support (HTTP/3)
3. Connection pooling utilities
4. Load balancing abstractions

**Performance Optimizations:**
1. Zero-copy operations where possible
2. Memory pooling for buffers
3. SIMD optimizations for checksums
4. io_uring support on Linux

**Monitoring Enhancements:**
1. Prometheus metrics integration
2. OpenTelemetry tracing
3. Connection analytics
4. Performance profiling hooks

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket)** - Complete API reference with function signatures, method descriptions, and runnable examples.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, protocol selection guide, and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (100%), and guidelines for writing new tests.

### Subpackage Documentation

- **[client/README.md](client/README.md)** - Client factory documentation and protocol-specific client implementations
- **[server/README.md](server/README.md)** - Server factory documentation and protocol-specific server implementations
- **[config/README.md](config/README.md)** - Configuration structures, validation, and best practices

### Related golib Packages

- **[github.com/nabbar/golib/network/protocol](https://pkg.go.dev/github.com/nabbar/golib/network/protocol)** - Network protocol constants and utilities used throughout socket package.

- **[github.com/nabbar/golib/certificates](https://pkg.go.dev/github.com/nabbar/golib/certificates)** - TLS configuration and certificate management for secure TCP connections.

- **[github.com/nabbar/golib/ioutils/aggregator](https://pkg.go.dev/github.com/nabbar/golib/ioutils/aggregator)** - Thread-safe write aggregation, commonly used with socket servers for logging.

### External References

- **[Go net Package](https://pkg.go.dev/net)** - Standard library networking primitives underlying all socket implementations.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for concurrency, error handling, and interface design.

- **[Unix Network Programming](https://www.amazon.com/Unix-Network-Programming-Volume-Sockets/dp/0131411551)** - Classic reference on socket programming concepts and patterns.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
