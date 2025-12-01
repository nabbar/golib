# Socket Config

[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-89.4%25-brightgreen)](TESTING.md)

Configuration structures for declarative socket client and server setup with validation.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Limitations](#limitations)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Recommendations](#recommendations)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic TCP Client](#basic-tcp-client)
  - [Basic TCP Server](#basic-tcp-server)
  - [Unix Socket with Permissions](#unix-socket-with-permissions)
- [Examples](#examples)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The `socket/config` package provides a configuration-first approach to socket programming in Go. Instead of creating sockets directly, you define configuration structures that can be validated before instantiation. This pattern is particularly useful when:

- Loading socket parameters from external sources (files, environment variables, databases)
- Validating configuration at application startup rather than during runtime
- Supporting multiple socket types through a unified interface
- Requiring platform-specific validation (e.g., Unix sockets on non-Windows systems)

### Design Philosophy

**Configuration Separation**: Socket parameters are defined independently from socket instances, enabling validation before connection attempts.

**Declarative API**: Configuration uses simple struct field assignments rather than complex builder patterns or method chaining.

**Fail-Fast Validation**: The `Validate()` methods catch configuration errors early, before network operations are attempted.

**Platform Awareness**: Built-in checks for platform-specific limitations (e.g., Unix sockets on Windows).

**Security by Default**: TLS/SSL configuration is validated to ensure proper certificate handling and protocol restrictions.

### Key Features

- **Protocol Flexibility**: Single configuration API supports TCP, UDP, Unix, and Unixgram sockets
- **TLS Support**: First-class TLS/SSL configuration for TCP connections with certificate validation
- **Unix Socket Permissions**: Fine-grained control over Unix socket file permissions and group ownership
- **Connection Management**: Configurable idle timeouts for connection-oriented protocols
- **Validation Guarantees**: Comprehensive validation catches address format errors, protocol mismatches, platform incompatibilities, and TLS configuration issues

---

## Architecture

### Component Diagram

The package consists of two main configuration structures:

```
┌────────────────────────────────────────────────────────────┐
│                     socket/config                          │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────┐              ┌───────────────┐           │
│  │   Client     │              │   Server      │           │
│  ├──────────────┤              ├───────────────┤           │
│  │ Network      │              │ Network       │           │
│  │ Address      │              │ Address       │           │
│  │ TLS          │              │ PermFile      │           │
│  │              │              │ GroupPerm     │           │
│  │ + Validate() │              │ ConIdleTimeout│           │
│  └──────────────┘              │ TLS           │           │
│                                │               │           │
│                                │ + Validate()  │           │
│                                │ + DefaultTLS()│           │
│                                │ + GetTLS()    │           │
│                                └───────────────┘           │
│                                                            │
│  ┌────────────────────────────────────────────┐            │
│  │              Error Types                   │            │
│  ├────────────────────────────────────────────┤            │
│  │ ErrInvalidProtocol                         │            │
│  │ ErrInvalidTLSConfig                        │            │
│  │ ErrInvalidGroup                            │            │
│  └────────────────────────────────────────────┘            │
└────────────────────────────────────────────────────────────┘
           │                              │
           │                              │
           ▼                              ▼
┌────────────────────┐         ┌────────────────────┐
│ socket/client      │         │ socket/server      │
│ implementations    │         │ implementations    │
└────────────────────┘         └────────────────────┘
```

Configuration flows from external sources → config structs → validation → socket implementations.

### Limitations

1. **Platform-Specific Features**: Unix domain sockets are not available on Windows
2. **TLS Protocol Restrictions**: TLS/SSL is only supported for TCP-based protocols
3. **No Dynamic Reconfiguration**: Changing configuration does not affect existing sockets
4. **Group Permission Limits**: Unix socket group IDs are limited to MaxGID (32767)
5. **No IPv6 Scope IDs**: Zone/scope IDs in link-local addresses may have platform-specific behavior

## Performance

The configuration structures are designed for infrequent creation (e.g., application startup) rather than high-frequency operations. 

### Benchmarks

From the test suite performance measurements:

- **Client Creation**: < 100µs per instance
- **Server Creation**: < 100µs per instance
- **TCP Validation**: < 1ms average
- **UDP Validation**: < 1ms average
- **Structure Copy**: < 10µs (small memory footprint)

### Recommendations

For hot-path operations:
- Cache validated configurations rather than re-validating
- Create socket instances once and reuse them
- Avoid calling `Validate()` in request handling loops

The structs are small (< 100 bytes) and safe to copy by value.

---

## Use Cases

### Configuration File Loading

Load socket parameters from YAML, JSON, or TOML files and validate them before starting services:

```yaml
# config.yaml
client:
  network: tcp
  address: localhost:8080
```

This catches configuration errors at startup rather than during operation.

### Environment-Based Configuration

Read socket settings from environment variables (12-factor app pattern):

```go
network := os.Getenv("SOCKET_NETWORK")  // "tcp"
address := os.Getenv("SOCKET_ADDRESS")  // ":8080"
```

Create properly validated socket instances for different deployment environments.

### Dynamic Service Discovery

Receive socket addresses from service discovery systems (Consul, etcd) and validate them before establishing connections.

### Multi-Protocol Services

Configure services to listen on multiple socket types simultaneously:

```go
// Network access via TCP
tcpCfg := config.Server{
    Network: protocol.NetworkTCP,
    Address: ":8080",
}

// Local IPC via Unix socket
unixCfg := config.Server{
    Network: protocol.NetworkUnix,
    Address: "/tmp/app.sock",
    PermFile: 0660,
}
```

### Secure Service Communication

Define TLS/SSL parameters for encrypted client-server communication with proper certificate validation.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/socket/config
```

### Basic TCP Client

```go
package main

import (
    "log"
    
    "github.com/nabbar/golib/socket/config"
    libptc "github.com/nabbar/golib/network/protocol"
)

func main() {
    // Create configuration
    cfg := config.Client{
        Network: libptc.NetworkTCP,
        Address: "localhost:8080",
    }
    
    // Validate before use
    if err := cfg.Validate(); err != nil {
        log.Fatal("Invalid configuration:", err)
    }
    
    // Configuration is valid, proceed with client creation
    log.Println("Client configuration validated successfully")
}
```

### Basic TCP Server

```go
package main

import (
    "log"
    
    "github.com/nabbar/golib/socket/config"
    libptc "github.com/nabbar/golib/network/protocol"
)

func main() {
    // Create configuration
    cfg := config.Server{
        Network: libptc.NetworkTCP,
        Address: ":8080",
    }
    
    // Validate before use
    if err := cfg.Validate(); err != nil {
        log.Fatal("Invalid configuration:", err)
    }
    
    // Configuration is valid, proceed with server creation
    log.Println("Server configuration validated successfully")
}
```

### Unix Socket with Permissions

```go
package main

import (
    "log"
    "runtime"
    
    "github.com/nabbar/golib/socket/config"
    libptc "github.com/nabbar/golib/network/protocol"
)

func main() {
    // Check platform compatibility
    if runtime.GOOS == "windows" {
        log.Fatal("Unix sockets not supported on Windows")
    }
    
    // Create configuration with permissions
    cfg := config.Server{
        Network:   libptc.NetworkUnix,
        Address:   "/tmp/app.sock",
        PermFile:  0660,  // Owner and group can read/write
        GroupPerm: 1000,  // Set to group 1000
    }
    
    // Validate
    if err := cfg.Validate(); err != nil {
        log.Fatal("Invalid configuration:", err)
    }
    
    log.Println("Unix socket configuration validated")
}
```

## Examples

### Multiple Server Types

```go
// Network server for remote access
tcpCfg := config.Server{
    Network:        libptc.NetworkTCP,
    Address:        ":8080",
    ConIdleTimeout: 10 * time.Minute,
}

// Unix socket for local IPC (if not on Windows)
var unixCfg *config.Server
if runtime.GOOS != "windows" {
    unixCfg = &config.Server{
        Network:  libptc.NetworkUnix,
        Address:  "/tmp/app.sock",
        PermFile: 0660,
    }
}

// Validate all configurations before starting
if err := tcpCfg.Validate(); err != nil {
    log.Fatal("TCP config error:", err)
}

if unixCfg != nil {
    if err := unixCfg.Validate(); err != nil {
        log.Fatal("Unix socket config error:", err)
    }
}
```

### Configuration from Environment

```go
// Read from environment variables
network := os.Getenv("SOCKET_NETWORK")
if network == "" {
    network = "tcp"  // Default
}

address := os.Getenv("SOCKET_ADDRESS")
if address == "" {
    address = ":8080"  // Default
}

// Parse network protocol
var proto libptc.NetworkProtocol
switch network {
case "tcp":
    proto = libptc.NetworkTCP
case "udp":
    proto = libptc.NetworkUDP
case "unix":
    proto = libptc.NetworkUnix
default:
    log.Fatalf("Unknown network type: %s", network)
}

// Create and validate configuration
cfg := config.Server{
    Network: proto,
    Address: address,
}

if err := cfg.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}
```

### Error Handling

```go
cfg := config.Client{
    Network: libptc.NetworkUnix,
    Address: "/tmp/test.sock",
}

if err := cfg.Validate(); err != nil {
    switch {
    case errors.Is(err, config.ErrInvalidProtocol):
        log.Println("Unsupported protocol (possibly Windows)")
    case errors.Is(err, config.ErrInvalidTLSConfig):
        log.Println("TLS configuration error")
    case errors.Is(err, config.ErrInvalidGroup):
        log.Println("Invalid group permission")
    default:
        log.Printf("Validation error: %v", err)
    }
    return
}
```

### Batch Validation

```go
configs := []config.Server{
    {Network: libptc.NetworkTCP, Address: ":8080"},
    {Network: libptc.NetworkTCP, Address: ":8081"},
    {Network: libptc.NetworkUDP, Address: ":9000"},
}

// Validate all configurations before starting any servers
for i, cfg := range configs {
    if err := cfg.Validate(); err != nil {
        log.Fatalf("Configuration %d invalid: %v", i, err)
    }
}

log.Printf("All %d configurations validated successfully", len(configs))
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
   - Use `gmeasure` for benchmarks
   - Ensure zero race conditions

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

- ✅ **89.4% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** configuration reading
- ✅ **Platform-aware** validation logic
- ✅ **Memory-safe** with no panics

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **YAML/JSON Serialization**: Built-in marshaling/unmarshaling support for configuration files
2. **Configuration Hot Reload**: Support for reloading socket configurations without restart
3. **Extended Protocol Support**: SCTP, raw sockets, or other specialized protocols
4. **Configuration Templates**: Predefined templates for common use cases

These are **optional improvements** and not required for production use. The current implementation is stable and well-tested.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/socket/config)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, performance considerations, and limitations. Provides detailed explanations of configuration structures and validation logic.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (89.4%), performance benchmarks, and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/network/protocol](https://pkg.go.dev/github.com/nabbar/golib/network/protocol)** - Network protocol definitions (TCP, UDP, Unix, etc.) used by the configuration structures. Provides the NetworkProtocol interface implemented by all supported protocols.

- **[github.com/nabbar/golib/certificates](https://pkg.go.dev/github.com/nabbar/golib/certificates)** - TLS/SSL certificate configuration used for secure connections. Provides the Config and TLSConfig interfaces for certificate management and validation.

- **[github.com/nabbar/golib/socket/client](https://pkg.go.dev/github.com/nabbar/golib/socket/client)** - Socket client implementations that use these configuration structures. Shows real-world usage of Client configuration for establishing connections.

- **[github.com/nabbar/golib/socket/server](https://pkg.go.dev/github.com/nabbar/golib/socket/server)** - Socket server implementations that use these configuration structures. Shows real-world usage of Server configuration for accepting connections.

### External References

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for struct design, error handling, and interface usage. The config package follows these conventions for idiomatic Go code.

- **[Go net Package](https://pkg.go.dev/net)** - Standard library networking package. The config package wraps net's address resolution and validation functions with additional type safety and platform checks.

- **[Unix Domain Sockets](https://en.wikipedia.org/wiki/Unix_domain_socket)** - Background on Unix domain sockets, their use cases, and how they differ from network sockets. Relevant for understanding Unix socket configuration options.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/socket/config`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
