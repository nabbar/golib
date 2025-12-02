# UNIX Domain Socket Client

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-76.1%25-brightgreen)](TESTING.md)

Thread-safe UNIX domain socket client for high-performance local inter-process communication with atomic state management, connection lifecycle callbacks, and one-shot request/response operations.

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
  - [One-Shot Request](#one-shot-request)
  - [With Callbacks](#with-callbacks)
  - [Read and Write](#read-and-write)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [ClientUnix Interface](#clientunix-interface)
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

The **unix** package provides a high-performance, thread-safe UNIX domain socket client implementation for local inter-process communication. UNIX sockets offer lower latency and higher throughput than TCP for same-machine communication, with filesystem-based access control and no network overhead.

### Design Philosophy

1. **Thread Safety First**: All operations safe for concurrent use without external synchronization
2. **Performance Oriented**: Zero network stack overhead, kernel-space only communication
3. **Context Integration**: First-class support for cancellation, timeouts, and deadline propagation
4. **Connection-Oriented**: Reliable, ordered, bidirectional communication like TCP
5. **Local-Only Security**: Filesystem permissions for access control, no network exposure

### Key Features

- ✅ **Connection-Oriented**: Reliable stream protocol (SOCK_STREAM) like TCP
- ✅ **Thread-Safe State**: Atomic map for lock-free concurrent operations
- ✅ **Context-Aware**: Full `context.Context` integration for lifecycle management
- ✅ **Event Callbacks**: Asynchronous error and connection state notifications
- ✅ **One-Shot Operations**: Convenient `Once()` for request/response patterns
- ✅ **Panic Recovery**: Automatic callback panic recovery with detailed logging
- ✅ **Zero Network Overhead**: Kernel-space communication without TCP/IP stack
- ✅ **Platform-Specific**: Linux and Darwin (macOS) support with conditional compilation

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                  ClientUnix Interface                   │
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
│  • net.UnixConn for socket operations                   │
│  • Async callback execution in goroutines               │
│  • Panic recovery with runner.RecoveryCaller            │
└──────────────┬──────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────┐
│             Go Standard Library                         │
│     net.DialUnix │ net.UnixConn │ context.Context       │
└─────────────────────────────────────────────────────────┘
```

### State Management

The client uses `atomic.Map` for lock-free state storage:

| Key | Type | Purpose |
|-----|------|---------|
| `keyNetAddr` | `string` | UNIX socket file path |
| `keyNetConn` | `*net.UnixConn` | Active socket connection |
| `keyFctErr` | `FuncError` | Error callback function |
| `keyFctInfo` | `FuncInfo` | Info callback function |

All state transitions are atomic, ensuring thread-safe concurrent access without mutexes or locks.

### Connection Lifecycle

```
New() → [Disconnected] → Connect() → [Connected] → Read/Write → Close() → [Disconnected]
                              ↓                                      ↑
                         Once() (auto-connect and auto-close)
```

**Key lifecycle methods:**
- `New(path)`: Create client instance (disconnected state)
- `Connect(ctx)`: Establish connection to socket
- `IsConnected()`: Check connection state (atomic)
- `Read()/Write()`: I/O operations (requires connection)
- `Close()`: Terminate connection and cleanup
- `Once(ctx, req, rsp)`: One-shot request/response with automatic lifecycle

---

## Performance

### Benchmarks

Performance characteristics on typical development machine:

| Operation | Median | Mean | Notes |
|-----------|--------|------|-------|
| Client Creation | <100µs | ~50µs | Memory allocation only |
| Connect | <500µs | ~200µs | Socket association |
| Write (Small) | <100µs | ~50µs | 13-byte message |
| Write (Large) | <500µs | ~200µs | 1400-byte message |
| Read (Small) | <100µs | ~50µs | 13-byte message |
| State Check | <10µs | ~5µs | Atomic operation |
| Close | <200µs | ~100µs | Socket cleanup |

**vs TCP Loopback:**
- 2-3x lower latency
- 50% less CPU usage
- No network stack overhead
- Better for high-frequency local IPC

### Memory Usage

- **Base Client**: ~200 bytes per instance
- **Per Connection**: ~8KB kernel buffer
- **Callback Storage**: Negligible (function pointers)
- **Zero Allocations**: After initial setup for read/write operations

### Scalability

- **Concurrent Clients**: Limited by system file descriptors (typically 1024+)
- **Throughput**: 100,000+ messages/sec on modern hardware
- **Thread Safety**: Zero contention on state checks (atomic operations)
- **Path Length**: Maximum 108 bytes (UNIX_PATH_MAX on Linux)

---

## Use Cases

1. **Container Communication**
   - Docker container to host service communication
   - Kubernetes sidecar proxy patterns
   - Example: Docker daemon API (`/var/run/docker.sock`)

2. **Database Connections**
   - High-performance local database access
   - Example: PostgreSQL, MySQL, Redis local connections

3. **Microservices IPC**
   - Fast communication between services on same machine
   - Example: Service mesh data plane

4. **System Daemon Control**
   - Controlling system daemons and services
   - Example: systemd, containerd, kubelet sockets

5. **Development Tools**
   - IDE plugins, build tools, development servers
   - Example: Language servers, debug adapters

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/client/unix
```

### Basic Connection

```go
package main

import (
    "context"
    "log"

    "github.com/nabbar/golib/socket/client/unix"
)

func main() {
    // Create client
    client := unix.New("/tmp/app.sock")
    if client == nil {
        log.Fatal("Invalid socket path")
    }
    defer client.Close()

    // Connect
    ctx := context.Background()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Send data
    if _, err := client.Write([]byte("Hello, UNIX!")); err != nil {
        log.Fatal(err)
    }
}
```

### One-Shot Request

```go
client := unix.New("/tmp/app.sock")

// Automatic connect, write, read, and close
request := bytes.NewBufferString("QUERY")
err := client.Once(ctx, request, func(r io.Reader) {
    response, _ := io.ReadAll(r)
    fmt.Printf("Response: %s\n", response)
})
```

### With Callbacks

```go
client := unix.New("/tmp/app.sock")
defer client.Close()

// Register error callback
client.RegisterFuncError(func(errs ...error) {
    for _, err := range errs {
        log.Printf("Error: %v", err)
    }
})

// Register state callback
client.RegisterFuncInfo(func(local, remote net.Addr, state libsck.ConnState) {
    log.Printf("State: %s (%s -> %s)", state, local, remote)
})

client.Connect(ctx)
```

### Read and Write

```go
// Write request
request := []byte("GET /status")
n, err := client.Write(request)
if err != nil {
    log.Fatal(err)
}

// Read response
response := make([]byte, 4096)
n, err = client.Read(response)
if err != nil && err != io.EOF {
    log.Fatal(err)
}

fmt.Printf("Received: %s\n", response[:n])
```

---

## Best Practices

### General Recommendations

1. **Always Close**: Use `defer client.Close()` to prevent socket leaks
2. **Set Timeouts**: Use context deadlines for all blocking operations
3. **Handle Errors**: Check all `Write()` and `Read()` errors
4. **Use Callbacks**: Prefer async error handling over polling
5. **Socket Cleanup**: Server should remove socket file on shutdown
6. **File Permissions**: Use restrictive permissions (0600 or 0660)

### Thread Safety

- All client methods are safe for concurrent use
- Callbacks execute asynchronously in separate goroutines
- Use synchronization when accessing shared state in callbacks
- Only one `Read()` or `Write()` at a time per connection

### Security Considerations

- Set restrictive file permissions: `chmod 600 /path/to/socket`
- Use `chown` to limit access to specific users/groups
- Consider SELinux/AppArmor policies for additional security
- Implement authentication at application level
- Validate all input data

### Error Handling

```go
// Check for specific errors
if err := client.Connect(ctx); err != nil {
    if errors.Is(err, unix.ErrAddress) {
        log.Fatal("Invalid socket path")
    } else if errors.Is(err, unix.ErrConnection) {
        log.Fatal("Connection failed")
    }
}
```

---

## API Reference

### ClientUnix Interface

```go
type ClientUnix interface {
    libsck.Client
}
```

**Methods:**
- `Connect(ctx context.Context) error`: Establish connection
- `IsConnected() bool`: Check connection state
- `Read(p []byte) (n int, err error)`: Read data
- `Write(p []byte) (n int, err error)`: Write data
- `Close() error`: Close connection
- `Once(ctx, request, response) error`: One-shot operation
- `SetTLS(...) error`: No-op for UNIX (always returns nil)
- `RegisterFuncError(f FuncError)`: Register error callback
- `RegisterFuncInfo(f FuncInfo)`: Register state callback

### Configuration

**Socket Path Requirements:**
- Must not be empty
- Maximum 108 bytes (UNIX_PATH_MAX)
- Parent directory must exist and be accessible
- Examples:
  - `/tmp/app.sock` - Temporary socket
  - `/var/run/app.sock` - System daemon socket
  - `/run/user/$UID/app.sock` - User-specific socket
  - `./app.sock` - Relative path

### Callbacks

**Error Callback:**
```go
type FuncError func(errs ...error)

client.RegisterFuncError(func(errs ...error) {
    // Executed asynchronously
    // Handle connection and I/O errors
})
```

**Info Callback:**
```go
type FuncInfo func(local, remote net.Addr, state ConnState)

client.RegisterFuncInfo(func(local, remote net.Addr, state ConnState) {
    // Executed asynchronously
    // Monitor: Connect, Read, Write, Close events
})
```

### Error Codes

- `ErrInstance`: Client instance is nil or invalid
- `ErrConnection`: No active connection established
- `ErrAddress`: Invalid or inaccessible socket path

Standard errors:
- `syscall.ECONNREFUSED`: Server not running
- `syscall.EACCES`: Permission denied
- `syscall.ENOENT`: Socket file doesn't exist
- `io.EOF`: Connection closed by remote
- `syscall.EPIPE`: Write to closed connection

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Style**: Follow Go idioms and `gofmt` standards
2. **Testing**: Include tests for new features (target >75% coverage)
3. **Documentation**: Update GoDoc comments and examples
4. **Race Detection**: Run `go test -race` before submitting
5. **Commit Messages**: Use clear, descriptive messages

**Pull Request Process:**
1. Fork the repository
2. Create a feature branch
3. Add tests and documentation
4. Run full test suite with race detector
5. Submit PR with clear description

---

## Improvements & Security

### Reporting Issues

- **Bugs**: Create GitHub issue with reproduction steps
- **Security**: Report privately via GitHub Security Advisories
- **Features**: Discuss in GitHub Discussions before implementing

### Known Limitations

- **Platform Support**: Linux and Darwin (macOS) only
- **Path Length**: 108 bytes maximum
- **No TLS**: Use application-level encryption if needed
- **Socket File**: Persists after server shutdown

### Future Improvements

- Abstract namespace support (Linux-specific)
- Automatic socket file cleanup helpers
- Built-in retry mechanisms
- Connection pooling support

---

## Resources

### Documentation

- [Go net Package](https://golang.org/pkg/net/)
- [UNIX(7) Man Page](https://man7.org/linux/man-pages/man7/unix.7.html)
- [UNIX Sockets Tutorial](https://beej.us/guide/bgipc/html/multi/unixsock.html)

### Related Packages

- [socket/client/tcp](../tcp) - TCP client implementation
- [socket/client/udp](../udp) - UDP client implementation
- [socket/client/unixgram](../unixgram) - UNIX datagram client
- [socket/server/unix](../../server/unix) - UNIX server implementation

### Examples

See [example_test.go](example_test.go) for comprehensive usage examples.

---

## AI Transparency

**EU AI Act Compliance (Article 50.4)**: AI assistance was used for documentation generation, code review, and testing under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details

**Package**: `github.com/nabbar/golib/socket/client/unix`  
**Maintained By**: [Nicolas JUHEL](https://github.com/nabbar)

---
