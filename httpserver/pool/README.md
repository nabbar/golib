# HTTP Server Pool Manager

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-80.4%25-brightgreen)](TESTING.md)

Thread-safe pool manager for multiple HTTP servers with unified lifecycle control, advanced filtering, and monitoring integration.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
- [Performance](#performance)
  - [Characteristics](#characteristics)
  - [Complexity](#complexity)
  - [Memory Usage](#memory-usage)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Usage](#basic-usage)
  - [Multiple Servers](#multiple-servers)
  - [Filtering](#filtering)
  - [Lifecycle Management](#lifecycle-management)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Pool Interface](#pool-interface)
  - [Configuration](#configuration)
  - [Filtering](#filtering-1)
  - [Error Codes](#error-codes)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **pool** package provides unified management for multiple HTTP server instances through a thread-safe pool abstraction. It enables simultaneous operation of servers with different configurations while providing centralized lifecycle control, advanced filtering capabilities, and integrated monitoring.

### Design Philosophy

1. **Unified Management**: Control multiple heterogeneous servers as a single logical unit.
2. **Thread Safety First**: All operations protected by sync.RWMutex for concurrent safety.
3. **Flexibility**: Support for dynamic server addition, removal, and configuration updates.
4. **Observability**: Built-in monitoring and health check integration for all pooled servers.
5. **Error Aggregation**: Collect and report errors from all servers systematically.

### Key Features

- ✅ **Unified Lifecycle**: Start, stop, and restart all servers with single operations.
- ✅ **Thread-Safe Operations**: Concurrent-safe server management using sync.RWMutex.
- ✅ **Advanced Filtering**: Query servers by name, bind address, or expose address with regex support.
- ✅ **Dynamic Management**: Add, remove, and update servers during operation without downtime.
- ✅ **Monitoring Integration**: Collect health and metrics data from all servers.
- ✅ **Configuration Helpers**: Bulk operations on server configurations with validation.
- ✅ **Extensive Testing**: 80.4% coverage with race detection and comprehensive test scenarios.

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────┐
│                        Pool                        │
├────────────────────────────────────────────────────┤
│                                                    │
│  ┌──────────────┐        ┌──────────────────────┐  │
│  │   Context    │        │   Handler Function   │  │
│  │   Provider   │        │   (shared optional)  │  │
│  └──────┬───────┘        └──────────┬───────────┘  │
│         │                           │              │
│         ▼                           ▼              │
│  ┌──────────────────────────────────────────────┐  │
│  │     Server Map (libctx.Config[string])       │  │
│  │     Key: Bind Address (e.g., "0.0.0.0:8080") │  │
│  │     Value: libhtp.Server instance            │  │
│  └──────────────────────────────────────────────┘  │
│         │                                          │
│         ▼                                          │
│  ┌─────────────────────────────────────────────┐   │
│  │  Individual Server Instances                │   │
│  │                                             │   │
│  │  Server 1 ──┐  Server 2 ──┐  Server N ──┐   │   │
│  │  :8080      │  :8443      │  :9000      │   │   │
│  │  HTTP       │  HTTPS      │  Custom     │   │   │
│  └─────────────────────────────────────────────┘   │
│                                                    │
│  ┌─────────────────────────────────────────────┐   │
│  │          Pool Operations                    │   │
│  │  - Walk: Iterate all servers                │   │
│  │  - WalkLimit: Iterate specific servers      │   │
│  │  - Filter: Query by criteria                │   │
│  │  - Start/Stop/Restart: Lifecycle            │   │
│  │  - Monitor: Health and metrics              │   │
│  └─────────────────────────────────────────────┘   │
│                                                    │
└────────────────────────────────────────────────────┘
```

### Data Flow

**Server Lifecycle:**
1. **Configuration Phase**: Servers defined via libhtp.Config
2. **Pool Creation**: New() creates empty pool or with initial servers
3. **Server Addition**: StoreNew() validates config and adds server to map
4. **Lifecycle Control**: Start() initiates all servers concurrently
5. **Runtime Operations**: Filter, Walk, Monitor during operation
6. **Graceful Shutdown**: Stop() drains and closes all servers

**Error Handling:**
1. Validation errors collected during config validation
2. Startup errors aggregated during Start()
3. Shutdown errors collected during Stop()
4. All errors use liberr.Error with proper code hierarchy

---

## Performance

### Characteristics

**Operation Complexity:**
- Store/Load/Delete: O(1) average, O(n) worst case (map operations)
- Walk/WalkLimit: O(n) where n is number of servers
- Filter: O(n) with regex matching overhead
- List: O(n) + O(m) where m is filtered result size
- Start/Stop/Restart: O(n) parallel server operations

**Memory Usage:**
- Base pool overhead: ~200 bytes
- Per-server overhead: ~100 bytes (map entry)
- Total: Base + (n × Server size) + (n × Overhead)
- Typical pool with 10 servers: ~50KB

**Concurrency:**
- Read operations scale with goroutines (RLock)
- Write operations serialize (Lock)
- Server lifecycle operations run concurrently
- No goroutine leaks during normal operation
- Zero race conditions verified with -race detector

### Complexity

| Operation | Time | Space | Notes |
|-----------|------|-------|-------|
| New | O(1) | O(1) | Constant initialization |
| StoreNew | O(1) | O(1) | Map insertion |
| Load | O(1) | O(1) | Map lookup |
| Delete | O(1) | O(1) | Map deletion |
| Walk | O(n) | O(1) | Iterate all servers |
| Filter | O(n) | O(m) | m = result size |
| Start | O(n) | O(n) | Parallel goroutines |
| Stop | O(n) | O(n) | Parallel goroutines |

### Memory Usage

- **Sequential Write**: Zero allocations per operation
- **Parallel Write**: ~1 allocation per writer (goroutine stack)
- **Struct Overhead**: ~1KB base size (atomic values, maps)

---

## Use Cases

### Multi-Port HTTP Server

Run HTTP and HTTPS servers simultaneously:

```go
httpCfg := libhtp.Config{
    Name:   "http",
    Listen: ":80",
    Expose: "http://example.com",
}

httpsCfg := libhtp.Config{
    Name:   "https",
    Listen: ":443",
    Expose: "https://example.com",
}

p := pool.New(nil, sharedHandler)
p.StoreNew(httpCfg, nil)
p.StoreNew(httpsCfg, nil)
p.Start(context.Background())
```

### Microservices Gateway

Route different services on different ports:

```go
configs := pool.Config{
    makeConfig("users-api", ":8081"),
    makeConfig("orders-api", ":8082"),
    makeConfig("payments-api", ":8083"),
}

p, _ := configs.Pool(nil, nil, logger)
p.Start(context.Background())
```

### Development vs Production

Different configurations per environment:

```go
var configs pool.Config
if isProd {
    configs = makeTLSConfigs()
} else {
    configs = makeHTTPConfigs()
}

p, _ := configs.Pool(ctx, handler, logger)
```

### Admin and Public Separation

Isolate administrative interfaces:

```go
publicCfg := libhtp.Config{
    Name:   "public",
    Listen: "0.0.0.0:8080",
    Expose: "https://api.example.com",
}

adminCfg := libhtp.Config{
    Name:   "admin",
    Listen: "127.0.0.1:9000", // localhost only
    Expose: "http://localhost:9000",
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/httpserver/pool
```

### Basic Usage

```go
package main

import (
    "context"
    "net/http"
    
    libhtp "github.com/nabbar/golib/httpserver"
    "github.com/nabbar/golib/httpserver/pool"
)

func main() {
    // Create pool
    p := pool.New(nil, nil)
    
    // Configure server
    cfg := libhtp.Config{
        Name:   "api-server",
        Listen: "0.0.0.0:8080",
        Expose: "http://localhost:8080",
    }
    cfg.RegisterHandlerFunc(func() map[string]http.Handler {
        return map[string]http.Handler{
            "/": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte("Hello from pool!"))
            }),
        }
    })
    
    // Add to pool
    p.StoreNew(cfg, nil)
    
    // Start server
    p.Start(context.Background())
    
    // Graceful shutdown
    defer p.Stop(context.Background())
}
```

### Multiple Servers

```go
configs := pool.Config{
    {Name: "api", Listen: ":8080", Expose: "http://localhost:8080"},
    {Name: "admin", Listen: ":9000", Expose: "http://localhost:9000"},
}

configs.SetHandlerFunc(sharedHandler)

p, err := configs.Pool(nil, nil, nil)
if err != nil {
    log.Fatal(err)
}

p.Start(context.Background())
```

### Filtering

```go
// Filter by name pattern
apiServers := p.Filter(srvtps.FieldName, "", "^api-.*")

// Filter by bind address
localServers := p.Filter(srvtps.FieldBind, "", "^127\\.0\\.0\\.1:.*")

// List server names
names := p.List(srvtps.FieldName, srvtps.FieldName, "", ".*")
```

### Lifecycle Management

```go
// Check if servers are running
if p.IsRunning() {
    log.Println("Servers are active")
}

// Get uptime
uptime := p.Uptime()

// Restart all servers
p.Restart(context.Background())

// Get monitoring data
version := libver.Version{}
monitors, _ := p.Monitor(version)
```

---

## Best Practices

**DO:**
- ✅ Validate all configurations before pool creation
- ✅ Use unique bind addresses for each server
- ✅ Set appropriate context timeouts for Start/Stop operations
- ✅ Check error codes for specific failure types
- ✅ Use Filter operations to manage subsets of servers
- ✅ Clean up pools with defer Stop(ctx)
- ✅ Use monitoring integration for production observability

**DON'T:**
- ❌ Don't assume all operations succeed (check errors)
- ❌ Don't use the same bind address for multiple servers
- ❌ Don't ignore validation errors
- ❌ Don't block indefinitely on Start/Stop (use context timeouts)
- ❌ Don't modify server configurations directly (use pool methods)
- ❌ Don't forget to handle partial failures during batch operations

---

## API Reference

### Pool Interface

**Lifecycle Management:**
- `Start(ctx context.Context) error` - Start all servers
- `Stop(ctx context.Context) error` - Stop all servers gracefully
- `Restart(ctx context.Context) error` - Restart all servers
- `IsRunning() bool` - Check if at least one server is running
- `Uptime() time.Duration` - Get maximum uptime of all servers

**Server Management:**
- `Store(srv libhtp.Server)` - Add pre-configured server
- `StoreNew(cfg libhtp.Config, def liblog.FuncLog) error` - Create and add server
- `Load(bindAddress string) libhtp.Server` - Retrieve server by bind address
- `Delete(bindAddress string)` - Remove server from pool
- `LoadAndDelete(bindAddress string) (libhtp.Server, bool)` - Atomic load and delete
- `Clean()` - Remove all servers

**Iteration:**
- `Walk(fct FuncWalk)` - Iterate all servers
- `WalkLimit(fct FuncWalk, bindAddress ...string)` - Iterate specific servers

**Filtering:**
- `Has(bindAddress string) bool` - Check if server exists
- `Len() int` - Get server count
- `Filter(field srvtps.Field, pattern, regex string) Pool` - Create filtered pool
- `List(source, target srvtps.Field, pattern, regex string) []string` - Extract field values

**Pool Operations:**
- `Clone(ctx context.Context) Pool` - Create independent copy
- `Merge(p Pool, def liblog.FuncLog) error` - Merge another pool
- `Handler(fct srvtps.FuncHandler)` - Register shared handler

**Monitoring:**
- `MonitorNames() []string` - Get monitor identifiers
- `Monitor(vrs libver.Version) ([]montps.Monitor, liberr.Error)` - Collect monitoring data

### Configuration

**Config Type:**
```go
type Config []libhtp.Config
```

**Methods:**
- `SetHandlerFunc(fct srvtps.FuncHandler)` - Set handler for all configs
- `SetDefaultTLS(cfg *tls.Config)` - Set TLS configuration
- `SetContext(ctx context.Context)` - Set context provider
- `Pool(ctx context.Context, hdl srvtps.FuncHandler, log liblog.FuncLog) (Pool, liberr.Error)` - Create pool
- `Walk(fct FuncWalkConfig)` - Iterate configurations
- `Validate() liberr.Error` - Validate all configurations

### Filtering

**Field Types:**
- `FieldName` - Server name
- `FieldBind` - Bind address (e.g., "127.0.0.1:8080")
- `FieldExpose` - Expose address (e.g., "http://localhost:8080")

**Pattern Matching:**
- Exact match: Use pattern parameter
- Regex match: Use regex parameter
- Case-insensitive for exact matches
- Go regex syntax for regex matches

### Error Codes

- `ErrorParamEmpty` - Invalid or empty parameters
- `ErrorPoolAdd` - Failed to add server to pool
- `ErrorPoolValidate` - Configuration validation failure
- `ErrorPoolStart` - One or more servers failed to start
- `ErrorPoolStop` - One or more servers failed to stop
- `ErrorPoolRestart` - One or more servers failed to restart
- `ErrorPoolMonitor` - Monitoring operation failure

---

## Contributing

Contributions are welcome! Please ensure:
- ✅ All tests pass with race detector enabled
- ✅ Code coverage remains ≥80%
- ✅ GoDoc comments for all exported functions
- ✅ Follow existing code style and patterns
- ✅ Add tests for new features

See [TESTING.md](TESTING.md) for detailed test documentation.

---

## Improvements & Security

### Current Status

The pool package is **production-ready** with:
- ✅ **Thread-safe** concurrent operations
- ✅ **80.4% test coverage** with comprehensive scenarios
- ✅ **Zero race conditions** verified with -race detector
- ✅ **Error handling** with liberr.Error hierarchy

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Automatic Port Allocation**: Dynamic port assignment for servers
2. **Health Checks**: Automatic server health monitoring and restart
3. **Load Balancing**: Built-in traffic distribution between servers
4. **Configuration Reload**: Hot-reload of server configurations
5. **Metrics Export**: Optional integration with Prometheus

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/httpserver/pool)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, thread-safety guarantees, and implementation details. Provides detailed explanations of internal mechanisms and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 80.4% coverage analysis, and guidelines for writing new tests. Includes troubleshooting and test inventory.

### Related golib Packages

- **[github.com/nabbar/golib/httpserver](https://pkg.go.dev/github.com/nabbar/golib/httpserver)** - Individual HTTP server implementation with TLS support, graceful shutdown, and monitoring integration.

- **[github.com/nabbar/golib/httpserver/types](https://pkg.go.dev/github.com/nabbar/golib/httpserver/types)** - Server type definitions and interfaces used throughout the pool package.

- **[github.com/nabbar/golib/context](https://pkg.go.dev/github.com/nabbar/golib/context)** - Context management utilities for server operations.

### External References

- **[net/http](https://pkg.go.dev/net/http)** - Go standard library HTTP package. The pool manages multiple http.Server instances with unified control.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and concurrency patterns.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)
**Package**: `github.com/nabbar/golib/httpserver/pool`
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
