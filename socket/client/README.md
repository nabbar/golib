# Socket Client Factory

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-81.2%25-brightgreen)](TESTING.md)

Platform-aware factory for creating socket clients across different network protocols (TCP, UDP, UNIX) with unified interface, TLS support, and automatic protocol selection.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Factory Pattern](#factory-pattern)
  - [Platform Support](#platform-support)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [TCP Client](#tcp-client)
  - [TCP with TLS](#tcp-with-tls)
  - [UDP Client](#udp-client)
  - [UNIX Socket Client](#unix-socket-client)
  - [Error Handling](#error-handling)
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

The **client** package provides a unified factory for creating socket clients across different network protocols. It automatically selects the appropriate protocol-specific implementation based on configuration while providing a consistent API through the `github.com/nabbar/golib/socket.Client` interface.

### Design Philosophy

1. **Simplicity First**: Single entry point (New) for all protocol types
2. **Platform Awareness**: Automatic protocol availability based on OS
3. **Type Safety**: Configuration-based client creation with validation
4. **Consistent API**: All clients implement socket.Client interface
5. **Zero Overhead**: Factory only adds a single switch statement

### Key Features

- ✅ **Unified Factory**: Single New() function for all protocols
- ✅ **Platform-Aware**: Automatic Unix socket support detection
- ✅ **Type-Safe**: Uses config.Client struct for configuration
- ✅ **Protocol Validation**: Returns error for unsupported protocols
- ✅ **TLS Support**: Transparent TLS configuration for TCP clients
- ✅ **Zero Dependencies**: Only delegates to sub-packages
- ✅ **Minimal Overhead**: Direct delegation without wrapping
- ✅ **Panic Recovery**: Automatic recovery with detailed logging

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                  client.New(cfg, def)                   │
│                   (Factory Function)                    │
└───────────────────────────┬─────────────────────────────┘
                            │
        ┌─────────────┬─────┴───────┬───────────┐
        │             │             │           │
        ▼             ▼             ▼           ▼
 ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
 │   TCP    │  │   UDP    │  │   Unix   │  │ UnixGram │
 │  Client  │  │  Client  │  │  Client  │  │  Client  │
 └──────────┘  └──────────┘  └──────────┘  └──────────┘
      │             │             │             │
      └─────────────┴──────┬──────┴─────────────┘
                           │
                 ┌─────────▼─────────┐
                 │  socket.Client    │
                 │    (Interface)    │
                 └───────────────────┘
```

### Factory Pattern

The package implements the Factory Method pattern:

**Protocol Selection Logic:**

```
New(cfg, def) → cfg.Network.IsTCP()    → tcp.New()
             → cfg.Network.IsUDP()    → udp.New()
             → cfg.Network.IsUnix()   → unix.New() (Linux/Darwin)
             → cfg.Network.IsUnixGram() → unixgram.New() (Linux/Darwin)
             → Unknown                → ErrInvalidProtocol
```

**Advantages:**
- Single import for all protocols
- Consistent API across protocols
- Easy protocol switching via configuration
- Centralized error handling

**Trade-offs:**
- Slight indirection overhead (negligible)
- Less explicit about protocol used

### Platform Support

| Platform | TCP | UDP | Unix | UnixGram |
|----------|-----|-----|------|----------|
| **Linux** | ✅ | ✅ | ✅ | ✅ |
| **Darwin/macOS** | ✅ | ✅ | ✅ | ✅ |
| **Windows** | ✅ | ✅ | ❌ | ❌ |
| **Other** | ✅ | ✅ | ❌ | ❌ |

---

## Performance

### Benchmarks

Factory overhead is negligible:

| Operation | Time | Overhead |
|-----------|------|----------|
| **Factory Call** | <1µs | Single switch + function call |
| **TCP Creation** | ~50µs | Dominated by protocol implementation |
| **UDP Creation** | ~40µs | Dominated by protocol implementation |
| **Unix Creation** | ~35µs | Dominated by protocol implementation |

**Conclusion**: Factory adds <1% overhead compared to direct protocol package usage.

### Memory Usage

```
Base overhead:        ~0 bytes (no state stored)
Per client:           Same as direct protocol usage
Factory function:     Stack-only allocation
```

**No heap allocations** - factory is allocation-free.

### Scalability

- **Concurrent Factory Calls**: Thread-safe, tested with 100 concurrent goroutines
- **Client Independence**: Each client is fully independent
- **Zero Shared State**: No contention between clients

---

## Use Cases

### 1. Multi-Protocol Application

**Problem**: Application needs to support multiple protocols based on configuration.

```go
// Configuration-driven protocol selection
cfg := loadConfig()  // Returns config.Client

cli, err := client.New(cfg, nil)
if err != nil {
    log.Fatal(err)
}
defer cli.Close()

// Same code works for TCP, UDP, Unix
ctx := context.Background()
cli.Connect(ctx)
cli.Write([]byte("data"))
```

**Real-world**: Used in microservices that communicate via TCP over network or Unix sockets locally.

### 2. Platform-Specific Optimization

**Problem**: Use Unix sockets on Linux/Darwin, fall back to TCP on Windows.

```go
// Try Unix first (fastest)
cfg := config.Client{
    Network: protocol.NetworkUnix,
    Address: "/tmp/app.sock",
}

cli, err := client.New(cfg, nil)
if err == config.ErrInvalidProtocol {
    // Fall back to TCP on unsupported platforms
    cfg.Network = protocol.NetworkTCP
    cfg.Address = "localhost:8080"
    cli, err = client.New(cfg, nil)
}
```

### 3. TLS Configuration Management

**Problem**: Centralized TLS configuration for TCP clients.

```go
// Shared TLS config
tlsCfg := loadTLSConfig()

// Create multiple TCP clients with same TLS
for _, addr := range servers {
    cfg := config.Client{
        Network: protocol.NetworkTCP,
        Address: addr,
        TLS: config.ClientTLS{
            Enabled:    true,
            ServerName: extractHost(addr),
        },
    }
    
    cli, _ := client.New(cfg, tlsCfg)
    clients = append(clients, cli)
}
```

### 4. Testing with Protocol Mocking

**Problem**: Test application with different protocols without changing code.

```go
// Production: Unix socket
prodCfg := config.Client{
    Network: protocol.NetworkUnix,
    Address: "/var/run/app.sock",
}

// Test: TCP socket for easier testing
testCfg := config.Client{
    Network: protocol.NetworkTCP,
    Address: "localhost:" + strconv.Itoa(testPort),
}

// Same application code
cfg := selectConfig(isTest)
cli, _ := client.New(cfg, nil)
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/client
```

### TCP Client

Simple TCP connection:

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
    
    // Create client using factory
    cli, err := sckclt.New(cfg, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()
    
    // Connect and communicate
    ctx := context.Background()
    if err := cli.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    cli.Write([]byte("Hello, server!"))
}
```

### TCP with TLS

Secure TCP connection:

```go
import (
    libtls "github.com/nabbar/golib/certificates"
)

func main() {
    // Configure TLS
    tlsCfg := libtls.NewTLSConfig()
    // ... configure certificates ...
    
    cfg := sckcfg.Client{
        Network: libptc.NetworkTCP,
        Address: "secure.example.com:443",
        TLS: sckcfg.ClientTLS{
            Enabled:    true,
            ServerName: "secure.example.com",
        },
    }
    
    cli, err := sckclt.New(cfg, tlsCfg)
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()
    
    // Encrypted communication
    ctx := context.Background()
    cli.Connect(ctx)
    cli.Write([]byte("Secure data"))
}
```

### UDP Client

Connectionless datagram communication:

```go
func main() {
    cfg := sckcfg.Client{
        Network: libptc.NetworkUDP,
        Address: "localhost:9000",
    }
    
    cli, err := sckclt.New(cfg, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()
    
    ctx := context.Background()
    cli.Connect(ctx)
    
    // Send datagram
    cli.Write([]byte("metric:value|type"))
}
```

### UNIX Socket Client

High-performance local IPC (Linux/Darwin only):

```go
func main() {
    cfg := sckcfg.Client{
        Network: libptc.NetworkUnix,
        Address: "/tmp/app.sock",
    }
    
    cli, err := sckclt.New(cfg, nil)
    if err != nil {
        if err == sckcfg.ErrInvalidProtocol {
            log.Fatal("Unix sockets not supported on this platform")
        }
        log.Fatal(err)
    }
    defer cli.Close()
    
    ctx := context.Background()
    cli.Connect(ctx)
    cli.Write([]byte("command"))
}
```

### Error Handling

Proper error handling patterns:

```go
func main() {
    cfg := sckcfg.Client{
        Network: libptc.NetworkUnix,
        Address: "/tmp/app.sock",
    }
    
    cli, err := sckclt.New(cfg, nil)
    if err != nil {
        if err == sckcfg.ErrInvalidProtocol {
            // Protocol not supported on this platform
            // Fall back to TCP
            cfg.Network = libptc.NetworkTCP
            cfg.Address = "localhost:8080"
            cli, err = sckclt.New(cfg, nil)
            if err != nil {
                log.Fatal(err)
            }
        } else {
            log.Fatal(err)
        }
    }
    defer cli.Close()
    
    // Use client...
}
```

---

## Best Practices

### ✅ DO

**Use Factory for Protocol Abstraction:**
```go
// ✅ Good: Configuration-driven
cfg := loadConfig()
cli, err := client.New(cfg, nil)
if err != nil {
    return err
}
defer cli.Close()
```

**Handle Platform-Specific Protocols:**
```go
// ✅ Good: Check for platform support
cli, err := client.New(cfg, nil)
if err == config.ErrInvalidProtocol {
    // Handle unsupported protocol
    cfg.Network = protocol.NetworkTCP
    cli, err = client.New(cfg, nil)
}
```

**Resource Management:**
```go
// ✅ Good: Always cleanup
cli, err := client.New(cfg, nil)
if err != nil {
    return err
}
defer cli.Close()  // Ensure cleanup
```

**Context Usage:**
```go
// ✅ Good: Use context for timeouts
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := cli.Connect(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return fmt.Errorf("connection timeout")
    }
    return err
}
```

### ❌ DON'T

**Don't ignore protocol errors:**
```go
// ❌ BAD: Ignoring errors
cli, _ := client.New(cfg, nil)
cli.Connect(ctx)  // May panic on nil client

// ✅ GOOD: Check errors
cli, err := client.New(cfg, nil)
if err != nil {
    return err
}
```

**Don't assume protocol availability:**
```go
// ❌ BAD: Assume Unix sockets available
cfg := config.Client{
    Network: protocol.NetworkUnix,
    Address: "/tmp/app.sock",
}
cli, _ := client.New(cfg, nil)  // Returns error on Windows

// ✅ GOOD: Check platform support
cli, err := client.New(cfg, nil)
if err == config.ErrInvalidProtocol {
    // Fall back to TCP
}
```

**Don't create clients without cleanup:**
```go
// ❌ BAD: No cleanup
client.New(cfg, nil)

// ✅ GOOD: Always defer Close
cli, err := client.New(cfg, nil)
if err != nil {
    return err
}
defer cli.Close()
```

---

## API Reference

### Factory Function

```go
func New(cfg sckcfg.Client, def libtls.TLSConfig) (libsck.Client, error)
```

**Parameters:**
- `cfg`: Client configuration (network type, address, TLS settings)
- `def`: Default TLS configuration (optional, can be nil)

**Returns:**
- `libsck.Client`: Client instance implementing socket.Client interface
- `error`: Error if configuration is invalid or protocol unsupported

**Behavior:**
1. Validates configuration (Validate())
2. Switches on cfg.Network type
3. Delegates to appropriate protocol package
4. Returns configured client or error

**Panic Recovery:**
All panics are caught and logged via RecoveryCaller.

### Configuration

```go
type Client struct {
    Network NetworkProtocol  // Required: TCP, UDP, Unix, UnixGram
    Address string           // Required: "host:port" or "/path/to/socket"
    TLS     ClientTLS        // Optional: TLS configuration (TCP only)
}

type ClientTLS struct {
    Enabled    bool      // Enable TLS
    Config     TLSConfig // TLS certificates and settings
    ServerName string    // Server name for verification
}
```

**Validation Rules:**
- Network must be valid protocol constant
- Address must be non-empty
- Unix sockets only on Linux/Darwin

### Error Codes

```go
var (
    ErrInvalidProtocol = errors.New("invalid protocol")
)
```

**Error Scenarios:**

| Error | Cause | Action |
|-------|-------|--------|
| `ErrInvalidProtocol` | Protocol not supported on platform | Fall back to supported protocol |
| `ErrInvalidInstance` | Configuration validation failed | Check cfg.Network and cfg.Address |
| Protocol-specific errors | From underlying implementation | See protocol package documentation |

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
   - Test platform-specific code on target platforms

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

- ✅ **81.2% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Panic recovery** in all critical paths
- ✅ **Memory-safe** with proper resource cleanup

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Protocol Auto-Detection**: Automatically detect best protocol for given address
2. **Connection Pooling**: Factory-managed connection pools per protocol
3. **Metrics Integration**: Optional Prometheus metrics for factory usage
4. **Configuration Validation**: Enhanced validation with detailed error messages

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/client)** - Complete API reference with function signatures, method descriptions, and runnable examples.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, protocol selection logic, and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (81.2%), and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/socket](https://pkg.go.dev/github.com/nabbar/golib/socket)** - Base socket interfaces and types. Defines the Client interface implemented by all protocol clients.

- **[github.com/nabbar/golib/socket/config](https://pkg.go.dev/github.com/nabbar/golib/socket/config)** - Configuration structures for clients and servers. Provides Client struct used by the factory.

- **[github.com/nabbar/golib/socket/client/tcp](https://pkg.go.dev/github.com/nabbar/golib/socket/client/tcp)** - TCP client implementation with TLS support.

- **[github.com/nabbar/golib/socket/client/udp](https://pkg.go.dev/github.com/nabbar/golib/socket/client/udp)** - UDP client implementation for connectionless communication.

- **[github.com/nabbar/golib/socket/client/unix](https://pkg.go.dev/github.com/nabbar/golib/socket/client/unix)** - Unix domain socket client for high-performance local IPC (Linux/Darwin).

- **[github.com/nabbar/golib/socket/client/unixgram](https://pkg.go.dev/github.com/nabbar/golib/socket/client/unixgram)** - Unix datagram socket client for connectionless local IPC (Linux/Darwin).

- **[github.com/nabbar/golib/network/protocol](https://pkg.go.dev/github.com/nabbar/golib/network/protocol)** - Network protocol constants and utilities.

- **[github.com/nabbar/golib/certificates](https://pkg.go.dev/github.com/nabbar/golib/certificates)** - TLS configuration and certificate management.

### External References

- **[Go net Package](https://pkg.go.dev/net)** - Standard library networking primitives used by all protocol implementations.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interface design and error handling.

- **[Factory Method Pattern](https://refactoring.guru/design-patterns/factory-method)** - Design pattern documentation explaining the factory pattern used by this package.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/client`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
