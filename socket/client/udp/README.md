# UDP Client

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-77.0%25-brightgreen)](TESTING.md)

Thread-safe UDP client for connectionless datagram communication with context integration, callback support, and panic recovery.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [State Management](#state-management)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Example](#basic-example)
  - [With Callbacks](#with-callbacks)
  - [One-Shot Request](#one-shot-request)
  - [Context Timeout](#context-timeout)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [ClientUDP Interface](#clientudp-interface)
  - [Constructor](#constructor)
  - [Callbacks](#callbacks)
  - [Error Handling](#error-handling)
  - [Monitoring](#monitoring)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **UDP Client** package provides a production-ready implementation for UDP datagram communication in Go. Unlike TCP, UDP is connectionless and unreliable, making it suitable for applications where speed and low latency are more important than guaranteed delivery. This package wraps the complexity of UDP socket management while providing modern Go idioms like context support and callback mechanisms.

### Design Philosophy

1. **Connectionless by Nature**: Embraces UDP's stateless design while providing convenience methods
2. **Thread-Safe Operations**: All methods safe for concurrent use via atomic state management
3. **Context Integration**: First-class support for cancellation, timeouts, and deadlines
4. **Non-Blocking Callbacks**: Asynchronous event notifications without blocking I/O
5. **Production-Ready**: Comprehensive testing with 77% coverage, panic recovery, and race detection

### Key Features

- ✅ **Thread-Safe**: All operations safe for concurrent access without external synchronization
- ✅ **Context-Aware**: Full `context.Context` integration for cancellation and timeouts
- ✅ **Event Callbacks**: Asynchronous error and state change notifications
- ✅ **One-Shot Operations**: Convenient `Once()` method for fire-and-forget patterns
- ✅ **Panic Recovery**: Automatic recovery from callback panics with detailed logging
- ✅ **Zero External Dependencies**: Only Go standard library and internal golib packages
- ✅ **TLS No-Op**: `SetTLS()` is a documented no-op (UDP doesn't support TLS)
- ✅ **IPv4/IPv6 Support**: Works with both IP protocols

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    ClientUDP Interface                  │
├─────────────────────────────────────────────────────────┤
│  Connect() │ Write() │ Read() │ Close() │ Once()        │
│  RegisterFuncError() │ RegisterFuncInfo()               │
└──────────────┬──────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────┐
│          Internal Implementation (cli struct)           │
├─────────────────────────────────────────────────────────┤
│  • atomic.Map for thread-safe state storage             │
│  • net.UDPConn for socket operations                    │
│  • Async callback execution in goroutines               │
│  • Panic recovery with runner.RecoveryCaller            │
└──────────────┬──────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────┐
│               Go Standard Library                       │
│     net.DialUDP │ net.UDPConn │ context.Context         │
└─────────────────────────────────────────────────────────┘
```

### Data Flow

**Write Operation:**
```
Client.Write(data) → Check Connection → UDPConn.Write() → Trigger Callbacks
```

**Read Operation:**
```
Client.Read(buffer) → Check Connection → UDPConn.Read() → Trigger Callbacks
```

**Once Operation** (One-shot request/response):
```
Client.Once(ctx, req, rsp) → Connect → Write → [Read if rsp] → Close → Callbacks
```

### State Management

The client uses `atomic.Map` for lock-free state storage:

| Key | Type | Purpose |
|-----|------|---------|
| `keyConn` | `*net.UDPConn` | Active UDP socket |
| `keyAddr` | `*net.UDPAddr` | Remote endpoint address |
| `keyFctErr` | `FuncError` | Error callback function |
| `keyFctInfo` | `FuncInfo` | Info callback function |

All state transitions are atomic, ensuring thread-safe concurrent access without mutexes.

---

## Performance

### Benchmarks

Performance results from benchmark tests on a standard development machine:

| Operation | Median | Mean | Notes |
|-----------|--------|------|-------|
| **Client Creation** | <100µs | ~50µs | Memory allocation only |
| **Connect** | <300µs | ~150µs | UDP socket association |
| **Write (Small)** | <5ms | ~2ms | 13-byte datagram |
| **Write (Large)** | <10ms | ~5ms | 1400-byte datagram |
| **Throughput** | 100 msgs/<500ms | - | Sequential writes |
| **State Check** | <100µs | ~10µs | Atomic `IsConnected()` |
| **Close** | <5ms | ~2ms | Socket cleanup |
| **Full Cycle** | <20ms | ~10ms | Create-connect-write-close |

### Memory Usage

- **Base Client**: ~200 bytes per instance
- **Per Connection**: ~4KB (kernel UDP buffer)
- **Callback Storage**: Negligible (function pointers only)
- **No Allocations**: Write operations are allocation-free

### Scalability

- **Concurrent Clients**: Limited only by system file descriptors
- **Message Rate**: Suitable for high-frequency communication (1000+ msg/sec)
- **Thread Safety**: Zero contention on state checks (atomic operations)
- **CPU Usage**: <1% for typical workloads

---

## Use Cases

### 1. Service Discovery
**Pattern**: Broadcast/multicast queries for network services  
**Advantages**: Low latency, minimal overhead, no connection setup  
**Example**: Finding available servers on local network

### 2. Metrics Collection
**Pattern**: Send telemetry to monitoring systems (StatsD, InfluxDB)  
**Advantages**: Fire-and-forget, non-blocking, tolerates packet loss  
**Example**: Application performance metrics

### 3. Real-Time Gaming
**Pattern**: Player state synchronization and game events  
**Advantages**: Ultra-low latency, packet loss acceptable  
**Example**: Multiplayer position updates

### 4. IoT Communication
**Pattern**: Device-to-server communication with minimal overhead  
**Advantages**: Simple protocol, low power consumption  
**Example**: Sensor data transmission

### 5. DNS Queries
**Pattern**: Custom DNS resolution or proxying  
**Advantages**: Standard protocol support, one-shot request/response  
**Example**: Local DNS cache or filtering proxy

### 6. Syslog Client
**Pattern**: Remote logging without delivery guarantees  
**Advantages**: Non-blocking, asynchronous, simple  
**Example**: Application log aggregation

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/client/udp
```

### Basic Example

```go
package main

import (
    "context"
    "log"
    
    "github.com/nabbar/golib/socket/client/udp"
)

func main() {
    // Create client
    client, err := udp.New("localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // Connect (associates socket with remote address)
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Send data
    if _, err := client.Write([]byte("Hello, UDP!")); err != nil {
        log.Fatal(err)
    }
}
```

### With Callbacks

```go
client, _ := udp.New("localhost:8080")
defer client.Close()

// Register error callback
client.RegisterFuncError(func(errs ...error) {
    for _, err := range errs {
        log.Printf("Error: %v", err)
    }
})

// Register info callback
client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
    log.Printf("State: %s (%s -> %s)", state, local, remote)
})

client.Connect(context.Background())
client.Write([]byte("message"))
```

### One-Shot Request

```go
client, _ := udp.New("localhost:8080")

request := bytes.NewBufferString("query")
err := client.Once(context.Background(), request, func(reader io.Reader) {
    response, _ := io.ReadAll(reader)
    fmt.Printf("Response: %s\n", response)
})
// Socket automatically closed after Once()
```

### Context Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

client, _ := udp.New("localhost:8080")
defer client.Close()

if err := client.Connect(ctx); err != nil {
    log.Printf("Connect timeout: %v", err)
}
```

---

## Best Practices

1. **Always Close**: Use `defer client.Close()` to prevent socket leaks
2. **Set Timeouts**: Use context deadlines for all operations
3. **Handle Errors**: Check `Write()` errors (UDP can fail silently)
4. **Use Callbacks**: Prefer async error handling over polling
5. **Datagram Size**: Keep ≤1472 bytes to avoid IP fragmentation
6. **No Guarantees**: Never rely on delivery or ordering
7. **Thread Safety**: All methods are concurrent-safe
8. **Callback Sync**: Use mutexes when accessing shared state in callbacks

**Testing Best Practices:**
- Run tests with `-race` flag
- Test with realistic packet loss rates
- Use `Once()` for simple request/response patterns
- Monitor callback execution with timeouts

See [TESTING.md](TESTING.md) for comprehensive testing documentation.

---

## API Reference

### ClientUDP Interface

```go
type ClientUDP interface {
    // Connection lifecycle
    Connect(ctx context.Context) error
    IsConnected() bool
    Close() error
    
    // Data transfer
    Write(p []byte) (n int, err error)
    Read(p []byte) (n int, err error)
    Once(ctx context.Context, req io.Reader, rsp func(io.Reader)) error
    
    // Callbacks
    RegisterFuncError(f FuncError)
    RegisterFuncInfo(f FuncInfo)
    
    // TLS (no-op for UDP)
    SetTLS(enable bool, config *tls.Config, host string) error
}
```

### Constructor

**`func New(address string) (ClientUDP, error)`**

Creates a new UDP client for the specified remote address.

- **address**: Format `"host:port"` or `"ip:port"` (IPv4/IPv6)
- **Returns**: Client instance or `ErrAddress` if invalid

```go
client, err := udp.New("192.168.1.100:9000")
client, err := udp.New("[::1]:8080") // IPv6
```

### Callbacks

**FuncError Callback:**
```go
type FuncError func(errs ...error)
```
Invoked asynchronously when errors occur. Panics are automatically recovered and logged.

**FuncInfo Callback:**
```go
type FuncInfo func(local, remote net.Addr, state ConnState)
```
Invoked asynchronously on state changes:
- `ConnectionDial`: Socket created
- `ConnectionNew`: Socket associated with remote address
- `ConnectionRead`: Data received
- `ConnectionWrite`: Data sent
- `ConnectionClose`: Socket closed

### Error Handling

**Error Constants:**
- `ErrInstance`: Client instance is nil
- `ErrConnection`: No active connection
- `ErrAddress`: Invalid address format

All errors implement `error` interface and are comparable with `errors.Is()`.

### Monitoring

Monitor client health via callbacks and state checks:

```go
client.RegisterFuncError(func(errs ...error) {
    metrics.IncrementErrorCount()
})

client.RegisterFuncInfo(func(_, _ net.Addr, state libsck.ConnState) {
    metrics.RecordStateChange(state.String())
})

// Periodic health check
if !client.IsConnected() {
    client.Connect(context.Background())
}
```

---

## Contributing

This package is part of the [golib](https://github.com/nabbar/golib) project.

**How to Contribute:**
1. Fork the repository
2. Create a feature branch
3. Write tests for new features
4. Ensure all tests pass with `-race` flag
5. Submit a pull request

**Contribution Guidelines:**
- Follow Go best practices and idioms
- Maintain or improve code coverage (currently 77%)
- Add tests for all new features
- Update documentation as needed
- Run `go fmt` and `go vet` before committing

---

## Improvements & Security

**Potential Improvements:**
- Higher test coverage for `Once()` method (currently 61.9%)
- Additional benchmarks for concurrent scenarios
- IPv6-specific optimization paths
- Multicast support

**Security Considerations:**
- UDP doesn't provide encryption (use application-layer security)
- No built-in authentication (implement at protocol level)
- Susceptible to amplification attacks (validate packet sources)
- Rate limiting recommended for public-facing services

**Reporting Security Issues:**  
See [TESTING.md - Reporting Bugs & Vulnerabilities](TESTING.md#reporting-bugs--vulnerabilities)

---

## Resources

**Documentation:**
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/client/udp)
- [Testing Guide](TESTING.md)
- [Package Overview](doc.go)

**Related Packages:**
- [socket/server/udp](../../../server/udp/) - UDP server implementation
- [socket/client/tcp](../tcp/) - TCP client implementation
- [socket](../../) - Base socket interfaces

**Learning Resources:**
- [UDP RFC 768](https://tools.ietf.org/html/rfc768) - User Datagram Protocol specification
- [Go net package](https://golang.org/pkg/net/) - Standard library documentation
- [Effective Go](https://golang.org/doc/effective_go) - Go programming best practices

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and code generation under human supervision. All core functionality is human-designed, reviewed, and validated. Test coverage and race detection ensure production quality.

---

## License

MIT License - Copyright (c) 2025 Nicolas JUHEL

See [LICENSE](../../../../LICENSE) file for full details.

**Maintained By**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/client/udp`
