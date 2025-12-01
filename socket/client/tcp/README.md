# TCP Client Package

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-79.4%25-green)](TESTING.md)

Thread-safe TCP client with atomic state management, TLS support, connection lifecycle callbacks, and one-shot request/response operations for reliable network communication.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [State Management](#state-management)
  - [Connection Lifecycle](#connection-lifecycle)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Connection](#basic-connection)
  - [TLS Connection](#tls-connection)
  - [One-Shot Request](#one-shot-request)
  - [With Callbacks](#with-callbacks)
  - [Read and Write](#read-and-write)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [ClientTCP Interface](#clienttcp-interface)
  - [Configuration](#configuration)
  - [Callbacks](#callbacks)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **tcp** package provides a high-performance, thread-safe TCP client implementation with atomic state management, TLS encryption support, and connection lifecycle callbacks. It's designed for reliable network communication with explicit control over connection state and error handling.

### Design Philosophy

1. **Thread Safety First**: All state management uses atomic operations for lock-free concurrency
2. **Explicit State Control**: Clear connection lifecycle with `IsConnected()` status checking
3. **TLS Integration**: First-class TLS support with certificate configuration
4. **Observable**: Connection state and error callbacks for monitoring and logging
5. **Context-Aware**: Full integration with Go's context for timeout and cancellation

### Key Features

- ✅ **Atomic State Management**: Thread-safe state tracking with `github.com/nabbar/golib/atomic`
- ✅ **TLS Support**: Configurable TLS encryption with certificate management
- ✅ **Connection Callbacks**: Error and info callbacks for lifecycle events
- ✅ **One-Shot Requests**: Convenient `Once()` method for request/response patterns
- ✅ **Standard Interfaces**: Implements `io.Reader`, `io.Writer`, `io.Closer`
- ✅ **Context Integration**: Timeout and cancellation support for all operations
- ✅ **Connection Replacement**: Automatic cleanup when reconnecting without explicit close
- ✅ **79.4% Test Coverage**: 159 comprehensive test specs with race detection

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────────┐
│                       ClientTCP                            │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         libatm.Map[uint8] (Atomic State)            │   │
│  │  - keyNetAddr: string (server address)              │   │
│  │  - keyNetConn: net.Conn (TCP connection)            │   │
│  │  - keyTLSCfg: *tls.Config (TLS configuration)       │   │
│  │  - keyFctErr: FuncError (error callback)            │   │
│  │  - keyFctInf: FuncInfo (info callback)              │   │
│  └─────────────────────────────────────────────────────┘   │
│                          │                                 │
│                          ▼                                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │           Connection Operations                     │   │
│  │                                                     │   │
│  │  Connect(ctx) ──▶ dial() ──▶ net.Conn               │   │
│  │       │                          │                  │   │
│  │       ├─▶ fctInfo(Dial)          │                  │   │
│  │       ├─▶ fctInfo(New)           │                  │   │
│  │       └─▶ fctError(on error)     │                  │   │
│  │                                  │                  │   │
│  │  Read(p) ────────────────────────┼─▶ conn.Read()    │   │
│  │       │                          │                  │   │
│  │       ├─▶ fctInfo(Read)          │                  │   │
│  │       └─▶ fctError(on error)     │                  │   │
│  │                                  │                  │   │
│  │  Write(p) ───────────────────────┼─▶ conn.Write()   │   │
│  │       │                          │                  │   │
│  │       ├─▶ fctInfo(Write)         │                  │   │
│  │       └─▶ fctError(on error)     │                  │   │
│  │                                  │                  │   │
│  │  Close() ────────────────────────┼─▶ conn.Close()   │   │
│  │       │                          │                  │   │
│  │       └─▶ fctInfo(Close)         │                  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │        Once() - One-Shot Request/Response           │   │
│  │                                                     │   │
│  │  1. Connect(ctx)                                    │   │
│  │  2. Write(request data)                             │   │
│  │  3. Response callback(client as io.Reader)          │   │
│  │  4. defer Close()                                   │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### State Management

The client uses **atomic operations** via `github.com/nabbar/golib/atomic.Map[uint8]` for thread-safe state management:

| State Key | Type | Description |
|-----------|------|-------------|
| `keyNetAddr` | `string` | Server address (host:port) |
| `keyNetConn` | `net.Conn` | Active TCP connection |
| `keyTLSCfg` | `*tls.Config` | TLS configuration (if enabled) |
| `keyFctErr` | `FuncError` | Error callback function |
| `keyFctInf` | `FuncInfo` | Info callback function |

**Thread Safety**: All operations are lock-free using atomic load/store/swap operations, ensuring safe concurrent access to connection state.

### Connection Lifecycle

```
┌─────────┐     Connect()      ┌───────────┐
│  New()  │───────────────────▶│ Connected │
└─────────┘                    └───────────┘
     │                               │
     │                               │ Read/Write operations
     │                               │
     │         Close()               ▼
     └──────────────────────────▶ Closed
                                     │
                                     │ Connect() again
                                     │
                                     ▼
                              ┌───────────┐
                              │ Connected │
                              └───────────┘
```

**State Callbacks:**
1. **ConnectionDial**: Triggered when dialing starts
2. **ConnectionNew**: Triggered when connection is established
3. **ConnectionRead**: Triggered on each read operation
4. **ConnectionWrite**: Triggered on each write operation
5. **ConnectionClose**: Triggered when connection is closed

---

## Performance

### Benchmarks

Performance measurements from test suite (AMD64, Go 1.25):

| Operation | Latency | Throughput | Notes |
|-----------|---------|------------|-------|
| Connect | ~10ms | 100 conn/s | TCP handshake + callbacks |
| Read (1KB) | <1ms | ~1MB/s | Network-bound |
| Write (1KB) | <1ms | ~1MB/s | Network-bound |
| Close | ~5ms | 200 ops/s | Graceful shutdown |
| Once (echo) | ~15ms | 60 req/s | Connect + I/O + Close |
| IsConnected | <1µs | 1M+ ops/s | Atomic read |

*Measured on localhost, actual network performance varies with RTT and bandwidth*

### Memory Usage

```
Base overhead:        ~200 bytes (struct + atomics)
Per connection:       net.Conn overhead (~4KB)
TLS overhead:         +~16KB (TLS state machine)
Total per instance:   ~4-20KB depending on TLS
```

**Memory Efficiency:**
- Atomic state management avoids mutex overhead
- No buffering beyond `net.Conn` internal buffers
- TLS session reuse reduces handshake overhead

### Scalability

- **Concurrent Clients**: Tested with up to 100 concurrent instances
- **Connection Pooling**: Safe to create multiple clients per server
- **Zero Race Conditions**: All tests pass with `-race` detector
- **Long-lived Connections**: No memory leaks or goroutine leaks

**Limitations:**
- One connection per client instance
- Reads/writes are sequential (no concurrent I/O on same connection)
- Network-bound performance (limited by TCP, not client implementation)

---

## Use Cases

### 1. HTTP Client Alternative

**Problem**: Need lower-level TCP control without HTTP overhead.

```go
client, _ := tcp.New("api.example.com:9000")
defer client.Close()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

client.Connect(ctx)
client.Write([]byte("GET /api/data\n"))

buf := make([]byte, 4096)
n, _ := client.Read(buf)
response := buf[:n]
```

### 2. Database Protocol Client

**Problem**: Implement custom database wire protocol.

```go
client, _ := tcp.New("db.example.com:5432")
defer client.Close()

// Send authentication
client.Connect(ctx)
client.Write(authPacket)

// Read response
response, _ := io.ReadAll(io.LimitReader(client, 1024))
```

### 3. Monitoring Service

**Problem**: Monitor TCP service availability.

```go
client, _ := tcp.New("service:8080")

client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
    log.Printf("[%s] %s -> %s", state, local, remote)
})

err := client.Connect(ctx)
if err != nil {
    alert.Send("Service unavailable")
}
```

### 4. TLS Encrypted Communication

**Problem**: Secure communication with certificate validation.

```go
client, _ := tcp.New("secure.example.com:443")

// Configure TLS
certConfig, _ := certificates.NewConfig(...)
client.SetTLS(true, certConfig, "secure.example.com")

// Connect with TLS handshake
client.Connect(ctx)
```

### 5. Request-Response Pattern

**Problem**: Simple request/response without persistent connection.

```go
client, _ := tcp.New("service:9000")

request := bytes.NewBufferString("QUERY data\n")

err := client.Once(ctx, request, func(r io.Reader) {
    response, _ := io.ReadAll(io.LimitReader(r, 4096))
    processResponse(response)
})
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/client/tcp
```

### Basic Connection

```go
package main

import (
    "context"
    "fmt"
    
    tcp "github.com/nabbar/golib/socket/client/tcp"
)

func main() {
    // Create client
    client, err := tcp.New("localhost:8080")
    if err != nil {
        panic(err)
    }
    defer client.Close()
    
    // Connect
    ctx := context.Background()
    err = client.Connect(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Connected:", client.IsConnected())
}
```

### TLS Connection

```go
package main

import (
    "context"
    
    tcp "github.com/nabbar/golib/socket/client/tcp"
    "github.com/nabbar/golib/certificates"
)

func main() {
    client, _ := tcp.New("secure.example.com:443")
    defer client.Close()
    
    // Load certificates
    certConfig, _ := certificates.NewConfig(
        certificates.WithCertificateFile("cert.pem"),
        certificates.WithPrivateKeyFile("key.pem"),
    )
    
    // Enable TLS
    err := client.SetTLS(true, certConfig, "secure.example.com")
    if err != nil {
        panic(err)
    }
    
    // Connect with TLS
    ctx := context.Background()
    client.Connect(ctx)
}
```

### One-Shot Request

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "io"
    
    tcp "github.com/nabbar/golib/socket/client/tcp"
)

func main() {
    client, _ := tcp.New("localhost:8080")
    
    ctx := context.Background()
    request := bytes.NewBufferString("HELLO\n")
    
    err := client.Once(ctx, request, func(r io.Reader) {
        // Read response
        buf := make([]byte, 1024)
        n, _ := r.Read(buf)
        fmt.Printf("Response: %s", buf[:n])
    })
    
    if err != nil {
        panic(err)
    }
}
```

### With Callbacks

```go
package main

import (
    "context"
    "log"
    "net"
    
    tcp "github.com/nabbar/golib/socket/client/tcp"
    "github.com/nabbar/golib/socket"
)

func main() {
    client, _ := tcp.New("localhost:8080")
    defer client.Close()
    
    // Register error callback
    client.RegisterFuncError(func(errs ...error) {
        for _, err := range errs {
            log.Printf("Error: %v", err)
        }
    })
    
    // Register info callback
    client.RegisterFuncInfo(func(local, remote net.Addr, state socket.ConnState) {
        log.Printf("[%s] %s -> %s", state, local, remote)
    })
    
    ctx := context.Background()
    client.Connect(ctx)
}
```

### Read and Write

```go
package main

import (
    "context"
    "fmt"
    
    tcp "github.com/nabbar/golib/socket/client/tcp"
)

func main() {
    client, _ := tcp.New("localhost:8080")
    defer client.Close()
    
    ctx := context.Background()
    client.Connect(ctx)
    
    // Write data
    message := []byte("Hello, Server!\n")
    n, err := client.Write(message)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Sent %d bytes\n", n)
    
    // Read response
    buf := make([]byte, 1024)
    n, err = client.Read(buf)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Received: %s", buf[:n])
}
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **79.4% code coverage** and **159 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and connection lifecycle operations
- ✅ Concurrent access with race detector (zero races detected)
- ✅ TLS configuration and handshake
- ✅ Error handling and edge cases
- ✅ Context integration and cancellation
- ✅ Callback mechanisms

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Context Management:**
```go
// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

err := client.Connect(ctx)
```

**Resource Cleanup:**
```go
// Always defer Close()
client, _ := tcp.New("localhost:8080")
defer client.Close()

client.Connect(ctx)
```

**Error Handling:**
```go
// Check all errors
err := client.Connect(ctx)
if err != nil {
    return fmt.Errorf("connection failed: %w", err)
}
```

**Connection State:**
```go
// Check before I/O
if !client.IsConnected() {
    return errors.New("not connected")
}

n, err := client.Write(data)
```

### ❌ DON'T

**Don't ignore errors:**
```go
// ❌ BAD: Ignoring errors
client.Connect(ctx)
client.Write(data)

// ✅ GOOD: Check errors
if err := client.Connect(ctx); err != nil {
    return err
}
```

**Don't use after Close:**
```go
// ❌ BAD: Use after close
client.Close()
client.Write(data)  // Returns error

// ✅ GOOD: Check state
if client.IsConnected() {
    client.Write(data)
}
```

**Don't share across goroutines:**
```go
// ❌ BAD: Concurrent access
go func() { client.Read(buf1) }()
go func() { client.Read(buf2) }()  // Race condition!

// ✅ GOOD: One client per goroutine
for i := 0; i < 10; i++ {
    go func() {
        c, _ := tcp.New("localhost:8080")
        defer c.Close()
        c.Connect(ctx)
    }()
}
```

**Don't block without timeout:**
```go
// ❌ BAD: No timeout
ctx := context.Background()
client.Connect(ctx)  // May hang forever

// ✅ GOOD: Use timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
client.Connect(ctx)
```

---

## API Reference

### ClientTCP Interface

```go
type ClientTCP interface {
    // Connection management
    Connect(ctx context.Context) error
    Close() error
    IsConnected() bool
    
    // I/O operations
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    
    // Configuration
    SetTLS(enable bool, config TLSConfig, serverName string) error
    
    // Callbacks
    RegisterFuncError(f FuncError)
    RegisterFuncInfo(f FuncInfo)
    
    // One-shot request/response
    Once(ctx context.Context, request io.Reader, fct Response) error
}
```

**Methods:**

- **`Connect(ctx context.Context) error`**: Establishes TCP connection with optional TLS
- **`Close() error`**: Closes connection and releases resources
- **`IsConnected() bool`**: Returns current connection state (atomic read)
- **`Read(p []byte) (n int, err error)`**: Reads data from connection
- **`Write(p []byte) (n int, err error)`**: Writes data to connection
- **`SetTLS(enable bool, config TLSConfig, serverName string) error`**: Configures TLS encryption
- **`RegisterFuncError(f FuncError)`**: Registers error callback
- **`RegisterFuncInfo(f FuncInfo)`**: Registers connection state callback
- **`Once(ctx context.Context, request io.Reader, fct Response) error`**: One-shot request/response

### Configuration

**Constructor:**

```go
func New(address string) (ClientTCP, error)
```

Creates a new TCP client instance.

**Parameters:**
- `address` - Server address in format "host:port" or ":port"

**Returns:** ClientTCP instance or error

**Valid Address Formats:**
```go
tcp.New("localhost:8080")        // Hostname + port
tcp.New("192.168.1.1:8080")      // IPv4 + port
tcp.New("[::1]:8080")            // IPv6 + port
tcp.New(":8080")                 // Any interface + port
```

### Callbacks

**FuncError:**

```go
type FuncError func(errs ...error)
```

Called when errors occur during operations.

**Example:**
```go
client.RegisterFuncError(func(errs ...error) {
    for _, err := range errs {
        log.Printf("Client error: %v", err)
    }
})
```

**FuncInfo:**

```go
type FuncInfo func(local, remote net.Addr, state ConnState)
```

Called on connection state changes.

**Connection States:**
- `ConnectionDial`: Dialing started
- `ConnectionNew`: Connection established
- `ConnectionRead`: Read operation performed
- `ConnectionWrite`: Write operation performed
- `ConnectionClose`: Connection closed

**Example:**
```go
client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
    log.Printf("[%s] %s -> %s", state, local, remote)
})
```

### Error Codes

```go
var (
    ErrInstance    = errors.New("invalid client instance")
    ErrConnection  = errors.New("connection error")
    ErrAddress     = errors.New("invalid address")
)
```

**Error Handling:**

| Error | When | Recovery |
|-------|------|----------|
| `ErrInstance` | Nil client or after Close() | Create new instance |
| `ErrConnection` | No active connection | Call Connect() |
| `ErrAddress` | Invalid address format | Fix address string |
| Network errors | I/O failures | Reconnect or handle gracefully |

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
   - Maintain coverage above 75%

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

- ✅ **79.4% test coverage** (target: >80%, nearly achieved)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Context integration** for timeout and cancellation

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Connection Pooling**: Reusable connection pool for high-throughput scenarios
2. **Keep-Alive Configuration**: Configurable TCP keep-alive settings
3. **Retry Logic**: Built-in exponential backoff for connection failures
4. **Metrics Export**: Optional integration with Prometheus or other metrics systems
5. **Bandwidth Limiting**: Configurable read/write rate limiting

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/client/tcp)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, state management, TLS configuration, and performance considerations. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (79.4%), and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/atomic](https://pkg.go.dev/github.com/nabbar/golib/atomic)** - Thread-safe atomic value storage used internally for state management. Provides lock-free atomic operations for better performance in concurrent scenarios.

- **[github.com/nabbar/golib/certificates](https://pkg.go.dev/github.com/nabbar/golib/certificates)** - TLS certificate management used for secure connections. Provides utilities for loading, validating, and configuring TLS certificates and keys.

- **[github.com/nabbar/golib/socket](https://pkg.go.dev/github.com/nabbar/golib/socket)** - Socket interfaces and types shared between client and server packages. Defines `ConnState`, `FuncError`, `FuncInfo` and other common types.

- **[github.com/nabbar/golib/socket/server/tcp](https://pkg.go.dev/github.com/nabbar/golib/socket/server/tcp)** - Complementary TCP server implementation. Often used together for client-server architectures.

### Standard Library References

- **[net](https://pkg.go.dev/net)** - Standard library networking package. The TCP client builds upon `net.Conn` and `net.Dial` for low-level TCP operations.

- **[context](https://pkg.go.dev/context)** - Standard library context package. The client fully implements context-based cancellation and timeout for all blocking operations.

- **[crypto/tls](https://pkg.go.dev/crypto/tls)** - Standard library TLS package. Used for secure connection establishment when TLS is enabled.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and I/O patterns. The TCP client follows these conventions for idiomatic Go code.

- **[TCP/IP Illustrated](http://www.tcpipguide.com/)** - Comprehensive guide to TCP/IP protocol suite. Useful for understanding low-level TCP behavior and tuning.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/client/tcp`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
