# Socket Server Factory

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)](TESTING.md)

Platform-aware socket server factory providing unified creation API for TCP, UDP, and Unix domain socket servers with automatic protocol delegation.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Factory Pattern](#factory-pattern)
  - [Platform-Specific Implementations](#platform-specific-implementations)
  - [Delegation Flow](#delegation-flow)
- [Performance](#performance)
  - [Factory Overhead](#factory-overhead)
  - [Protocol Comparison](#protocol-comparison)
- [Subpackages](#subpackages)
  - [tcp](#tcp)
  - [udp](#udp)
  - [unix](#unix)
  - [unixgram](#unixgram)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Examples](#basic-examples)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Factory Function](#factory-function)
  - [Configuration](#configuration)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **server** package provides a unified factory for creating socket servers across different network protocols. It automatically selects and instantiates the appropriate protocol-specific implementation based on configuration, providing a consistent API through the `socket.Server` interface.

### Design Philosophy

1. **Single Entry Point**: One `New()` function for all protocol types
2. **Platform Awareness**: Automatic protocol availability based on OS
3. **Zero Overhead**: Direct delegation without wrapping layers
4. **Type Safety**: Configuration-based creation with validation
5. **Consistent API**: All servers implement the same interface

### Key Features

- ✅ **Unified Factory**: Single entry point for TCP, UDP, Unix socket creation
- ✅ **Platform-Aware**: Automatic Unix socket support detection (Linux/Darwin)
- ✅ **Protocol Validation**: Returns error for unsupported protocols
- ✅ **Zero Dependencies**: Only delegates to protocol-specific packages
- ✅ **Minimal Overhead**: Single switch statement, no wrapping
- ✅ **Thread-Safe**: All created servers safe for concurrent use

---

## Architecture

### Factory Pattern

The server package implements the Factory Method design pattern:

```
           ┌─────────────────────────┐
           │   server.New(cfg)       │
           │   (Factory Function)    │
           └───────────┬─────────────┘
                       │
          ┌────────────┼────────────┬────────────┐
          │            │            │            │
          ▼            ▼            ▼            ▼
    ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
    │   TCP   │  │   UDP   │  │  Unix   │  │UnixGram │
    │ Server  │  │ Server  │  │ Server  │  │ Server  │
    └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘
         │            │            │            │
         └────────────┴────────────┴────────────┘
                      │
            ┌─────────▼─────────┐
            │  socket.Server    │
            │   (Interface)     │
            └───────────────────┘
```

**Benefits:**
- Simplified server creation
- Consistent error handling
- Protocol abstraction
- Easy protocol switching

### Platform-Specific Implementations

Build constraints determine protocol availability:

```
┌─────────────────────┬──────────────────┬─────────────────────┐
│  File               │  Build Tag       │  Protocols          │
├─────────────────────┼──────────────────┼─────────────────────┤
│  interface_linux.go │  linux           │  TCP, UDP, Unix, *  │
│  interface_darwin.go│  darwin          │  TCP, UDP, Unix, *  │
│  interface_other.go │  !linux&&!darwin │  TCP, UDP only      │
└─────────────────────┴──────────────────┴─────────────────────┘

* Unix includes Unix and UnixGram protocols
```

**Why Platform-Specific?**
- Unix domain sockets not available on Windows
- Compilation errors prevented on unsupported platforms
- Clear error messages for unsupported protocols

### Delegation Flow

```
1. User calls server.New(upd, handler, cfg)
   │
2. Factory validates cfg.Network
   │
3. Switch on protocol type
   │
   ├─ NetworkTCP* → tcp.New(upd, handler, cfg)
   ├─ NetworkUDP* → udp.New(upd, handler, cfg)
   ├─ NetworkUnix → unix.New(upd, handler, cfg)
   ├─ NetworkUnixGram → unixgram.New(upd, handler, cfg)
   └─ Other → ErrInvalidProtocol
   │
4. Return socket.Server implementation
```

**Delegation Properties:**
- Zero-copy: No wrapping, direct return
- Atomic: Single function call
- Type-safe: Compile-time protocol validation

---

## Performance

### Factory Overhead

Based on comprehensive benchmarks (100 samples):

| Operation | Median | Mean | Max |
|-----------|--------|------|-----|
| **TCP Creation** | <1ms | <1ms | <10ms |
| **UDP Creation** | <1ms | <1ms | <10ms |
| **Unix Creation** | <1ms | <1ms | <10ms |
| **Concurrent Creation (10)** | 200µs | 200µs | 400µs |

**Overhead Analysis:**
- Factory adds ~1µs (single switch statement)
- Total time dominated by protocol-specific initialization
- No measurable performance difference vs. direct package use

### Protocol Comparison

| Protocol | Throughput | Latency | Best For |
|----------|-----------|---------|----------|
| **TCP** | High | Low | Network IPC, reliability |
| **UDP** | Very High | Very Low | Datagrams, speed |
| **Unix** | Highest | Lowest | Local IPC, performance |
| **UnixGram** | Highest | Lowest | Local datagrams |

---

## Subpackages

### tcp

**Purpose**: TCP server implementation with TLS support.

**Key Features**:
- Connection-oriented, reliable
- TLS/SSL encryption support
- Graceful shutdown
- Idle timeout management
- Thread-safe connection handling

**Performance**: ~500K req/sec (localhost echo)

**Documentation**: [tcp/README.md](tcp/README.md)

---

### udp

**Purpose**: UDP server implementation for datagram protocols.

**Key Features**:
- Connectionless, fast
- Datagram handling
- Broadcast/multicast support
- Low-latency operations

**Performance**: Very high throughput for small packets

**Documentation**: [udp/README.md](udp/README.md)

---

### unix

**Purpose**: Unix domain socket server (connection-oriented).

**Key Features**:
- Local IPC only (same host)
- File system permissions
- Higher throughput than TCP loopback
- Lower latency than TCP

**Performance**: ~1M req/sec (localhost echo)

**Platforms**: Linux, Darwin/macOS only

**Documentation**: [unix/README.md](unix/README.md)

---

### unixgram

**Purpose**: Unix domain datagram socket server.

**Key Features**:
- Connectionless local IPC
- Fast datagram delivery
- File system permissions
- Low overhead

**Performance**: Very high throughput for local communication

**Platforms**: Linux, Darwin/macOS only

**Documentation**: [unixgram/README.md](unixgram/README.md)

---

## Use Cases

### 1. Cross-Platform Network Service

**Problem**: Build a service that works on all platforms.

**Solution**: Use TCP (available everywhere).

```go
cfg := config.Server{
    Network: protocol.NetworkTCP,
    Address: ":8080",
}

srv, err := server.New(nil, handler, cfg)
```

### 2. High-Performance Local IPC

**Problem**: Fast communication between processes on same host.

**Solution**: Use Unix sockets (when available).

```go
cfg := config.Server{
    Network:   protocol.NetworkUnix,
    Address:   "/tmp/app.sock",
    PermFile:  perm.Perm(0660),
    GroupPerm: -1,
}

srv, err := server.New(nil, handler, cfg)
if err == config.ErrInvalidProtocol {
    // Fall back to TCP on unsupported platforms
    cfg.Network = protocol.NetworkTCP
    cfg.Address = ":8080"
    srv, err = server.New(nil, handler, cfg)
}
```

### 3. Real-Time Metrics Collection

**Problem**: Collect metrics with minimal overhead.

**Solution**: Use UDP for fast, fire-and-forget delivery.

```go
cfg := config.Server{
    Network: protocol.NetworkUDP,
    Address: ":9000",
}

srv, err := server.New(nil, handler, cfg)
```

### 4. Secure Web Service

**Problem**: HTTPS server with TLS.

**Solution**: Use TCP with TLS configuration.

```go
cfg := config.Server{
    Network: protocol.NetworkTCP,
    Address: ":443",
    TLS: config.TLS{
        Enable: true,
        Config: tlsConfig,
    },
}

srv, err := server.New(nil, handler, cfg)
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/server
```

### Basic Examples

**TCP Server:**

```go
import (
    "context"
    "github.com/nabbar/golib/network/protocol"
    "github.com/nabbar/golib/socket"
    "github.com/nabbar/golib/socket/config"
    "github.com/nabbar/golib/socket/server"
)

func main() {
    handler := func(c socket.Context) {
        defer c.Close()
        // Handle connection
    }

    cfg := config.Server{
        Network: protocol.NetworkTCP,
        Address: ":8080",
    }

    srv, err := server.New(nil, handler, cfg)
    if err != nil {
        panic(err)
    }

    if err := srv.Listen(context.Background()); err != nil {
        panic(err)
    }
}
```

**UDP Server:**

```go
cfg := config.Server{
    Network: protocol.NetworkUDP,
    Address: ":9000",
}

srv, err := server.New(nil, handler, cfg)
```

**Unix Socket Server (with fallback):**

```go
import "github.com/nabbar/golib/file/perm"

cfg := config.Server{
    Network:   protocol.NetworkUnix,
    Address:   "/tmp/app.sock",
    PermFile:  perm.Perm(0660),
    GroupPerm: -1,
}

srv, err := server.New(nil, handler, cfg)
if err == config.ErrInvalidProtocol {
    // Platform doesn't support Unix sockets
    cfg.Network = protocol.NetworkTCP
    cfg.Address = ":8080"
    srv, err = server.New(nil, handler, cfg)
}
```

**With Connection Configuration:**

```go
import "net"

upd := func(c net.Conn) {
    if tcpConn, ok := c.(*net.TCPConn); ok {
        tcpConn.SetKeepAlive(true)
        tcpConn.SetKeepAlivePeriod(30 * time.Second)
    }
}

srv, err := server.New(upd, handler, cfg)
```

---

## Best Practices

### ✅ DO

**Choose Protocol for Use Case:**
```go
// Local IPC → Unix (fastest)
cfg.Network = protocol.NetworkUnix

// Network service → TCP (reliable)
cfg.Network = protocol.NetworkTCP

// Metrics/logging → UDP (fast)
cfg.Network = protocol.NetworkUDP
```

**Handle Platform Limitations:**
```go
srv, err := server.New(nil, handler, cfg)
if err == config.ErrInvalidProtocol {
    // Implement fallback strategy
}
```

**Resource Cleanup:**
```go
srv, err := server.New(nil, handler, cfg)
if err != nil {
    return err
}
defer srv.Close()
```

**Graceful Shutdown:**
```go
ctx, cancel := context.WithTimeout(
    context.Background(), 30*time.Second)
defer cancel()

if err := srv.Shutdown(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

### ❌ DON'T

**Don't Ignore Errors:**
```go
// ❌ BAD
srv, _ := server.New(nil, handler, cfg)

// ✅ GOOD
srv, err := server.New(nil, handler, cfg)
if err != nil {
    return fmt.Errorf("server creation failed: %w", err)
}
```

**Don't Assume Unix Support:**
```go
// ❌ BAD: Will fail on Windows
srv, _ := server.New(nil, handler, config.Server{
    Network: protocol.NetworkUnix,
    Address: "/tmp/app.sock",
})

// ✅ GOOD: Check and fallback
if err == config.ErrInvalidProtocol {
    // Use TCP as fallback
}
```

**Don't Mix Protocol-Specific Options:**
```go
// ❌ BAD: TLS on Unix socket (ignored)
cfg := config.Server{
    Network: protocol.NetworkUnix,
    TLS: config.TLS{Enable: true}, // Has no effect
}

// ✅ GOOD: TLS only for TCP
if cfg.Network.IsTCP() {
    cfg.TLS.Enable = true
}
```

---

## API Reference

### Factory Function

```go
func New(upd socket.UpdateConn, handler socket.HandlerFunc, cfg config.Server) (socket.Server, error)
```

**Parameters:**
- `upd`: Optional connection configuration callback (can be nil)
- `handler`: Required connection/datagram handler function
- `cfg`: Server configuration including protocol and address

**Returns:**
- `socket.Server`: Server instance implementing standard interface
- `error`: `config.ErrInvalidProtocol` if protocol unsupported

**Example:**
```go
srv, err := server.New(nil, handler, config.Server{
    Network: protocol.NetworkTCP,
    Address: ":8080",
})
```

### Configuration

```go
type config.Server struct {
    Network        protocol.Protocol // Required: NetworkTCP, NetworkUDP, etc.
    Address        string            // Required: ":port" or "/path/to/socket"
    PermFile       perm.Perm         // Unix only: file permissions
    GroupPerm      int               // Unix only: group ID
    ConIdleTimeout time.Duration     // Optional: idle timeout
    TLS            config.TLS        // TCP only: TLS configuration
}
```

**Protocol Values:**
- `protocol.NetworkTCP`, `NetworkTCP4`, `NetworkTCP6`
- `protocol.NetworkUDP`, `NetworkUDP4`, `NetworkUDP6`
- `protocol.NetworkUnix` (Linux/Darwin only)
- `protocol.NetworkUnixGram` (Linux/Darwin only)

### Error Codes

```go
config.ErrInvalidProtocol  // Protocol not supported or invalid
```

**Handling:**
```go
if err == config.ErrInvalidProtocol {
    // Implement fallback or return user-friendly error
}
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain 100% code coverage
   - Pass all tests including race detector
   - Use `gofmt` and `golint`

2. **AI Usage Policy**
   - ❌ **Do NOT use AI** for implementing package functionality
   - ✅ **AI may assist** with tests, documentation, debugging
   - All AI-assisted contributions must be reviewed by humans

3. **Testing**
   - Add tests for new features
   - Use Ginkgo v2 / Gomega
   - Ensure zero race conditions
   - Maintain coverage at 100%

4. **Documentation**
   - Update GoDoc comments
   - Add examples for new features
   - Update README.md and TESTING.md
   - Follow existing documentation structure

5. **Pull Request Process**
   - Fork the repository
   - Create a feature branch
   - Write clear commit messages
   - Ensure all tests pass
   - Update documentation
   - Submit PR with description

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **100.0% test coverage**
- ✅ **Zero race conditions** detected
- ✅ **Thread-safe** operations
- ✅ **Minimal overhead** (<1µs factory cost)
- ✅ **Platform-aware** compilation
- ✅ **41 comprehensive test specs**

### Future Enhancements (Non-urgent)

**Features:**
1. Protocol auto-detection from address format
2. Configuration presets for common scenarios
3. Server pool management (multiple servers)
4. Health check endpoints integration

**Performance:**
1. Compile-time protocol selection (build tags)
2. Protocol-specific fast paths
3. Configuration validation caching

**Monitoring:**
1. Factory metrics (creation rate, errors)
2. Protocol usage statistics
3. OpenTelemetry integration

These are **optional improvements** and not required for production use. The current implementation is stable and feature-complete.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Internal Documentation
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/server) - Complete API documentation
- [TESTING.md](TESTING.md) - Test suite documentation
- [doc.go](doc.go) - Detailed package documentation

### Subpackage Documentation
- [tcp/README.md](tcp/README.md) - TCP server implementation
- [udp/README.md](udp/README.md) - UDP server implementation
- [unix/README.md](unix/README.md) - Unix socket server (Linux/Darwin)
- [unixgram/README.md](unixgram/README.md) - Unix datagram server (Linux/Darwin)

### Related Packages
- [github.com/nabbar/golib/socket](../README.md) - Base interfaces and types
- [github.com/nabbar/golib/socket/config](../config/README.md) - Configuration structures
- [github.com/nabbar/golib/network/protocol](../../network/protocol/README.md) - Protocol constants

### External References
- [Go net package](https://pkg.go.dev/net) - Standard library networking
- [Unix Domain Sockets](https://en.wikipedia.org/wiki/Unix_domain_socket) - Background info
- [TCP/IP Guide](https://www.rfc-editor.org/rfc/rfc793) - TCP specification

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/server`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
