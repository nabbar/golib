# HTTPServer Types

[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)](TESTING.md)

Core type definitions and constants for HTTP server implementations providing foundational types for handler registration, server field identification, and timeout management.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Limitations](#limitations)
- [Performance](#performance)
  - [Type Overhead](#type-overhead)
  - [Constant Access](#constant-access)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Field Type Usage](#basic-field-type-usage)
  - [Handler Registration](#handler-registration)
  - [Using Timeout Constants](#using-timeout-constants)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Field Types](#field-types)
  - [Handler Types](#handler-types)
  - [Constants](#constants)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **httpserver/types** package provides foundational type definitions and constants for HTTP server implementations. It serves as the shared type system for server configuration, handler registration, field identification, and timeout management across the httpserver ecosystem.

### Why a Separate Types Package?

Separating types into their own package provides several architectural benefits:

**Benefits of Dedicated Types:**
- ✅ **No Circular Dependencies**: Higher-level packages (httpserver, pool) can import types without dependency cycles
- ✅ **Shared Vocabulary**: Consistent type definitions across all HTTP server components
- ✅ **Minimal Import Footprint**: Packages importing types don't pull in heavy dependencies
- ✅ **Interface Stability**: Type definitions remain stable even as implementations evolve
- ✅ **Clear Contracts**: FuncHandler and FieldType define explicit contracts for server operations
- ✅ **Type Safety**: Compile-time guarantees prevent invalid field specifications and handler types

**Internally**, the types package uses only standard library primitives (`net/http`, `time`), ensuring zero external dependencies. This makes it suitable as a foundation layer that other packages can safely depend on without transitive dependency concerns.

### Design Philosophy

1. **Minimal Dependencies**: Depend only on standard library to serve as stable foundation for higher-level packages.
2. **Type Safety First**: Use custom types (`FieldType`) to provide compile-time safety and prevent invalid operations.
3. **Fail-Safe Defaults**: Provide safe fallbacks (`BadHandler`) that fail visibly rather than silently.
4. **Constants Over Magic Values**: Use named constants for all configuration values to improve code readability.
5. **Zero Allocation Types**: Design types for minimal runtime overhead (empty structs, uint8 enums).
6. **Interface Compliance**: Ensure types properly implement standard interfaces (`http.Handler`).

### Key Features

- ✅ **Field Type System**: `FieldType` enumeration enables type-safe server filtering by name, bind address, or expose URL.
- ✅ **Handler Registration**: `FuncHandler` defines the contract for dynamic handler registration.
- ✅ **Fallback Handling**: `BadHandler` provides a safe default when no valid handler is configured.
- ✅ **Timeout Management**: Pre-defined timeout constants standardize server lifecycle operations.
- ✅ **Zero Dependencies**: Only depends on standard library `net/http` and `time`.
- ✅ **Minimal Overhead**: Types designed for zero or minimal runtime allocation.

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────────┐
│                     httpserver/types                       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────────┐           ┌─────────────────────┐    │
│  │   Field Types    │           │   Handler Types     │    │
│  │  (Enumeration)   │           │   (Interfaces)      │    │
│  └──────┬───────────┘           └──────────┬──────────┘    │
│         │                                  │               │
│         ▼                                  ▼               │
│  ┌──────────────────┐           ┌─────────────────────┐    │
│  │ FieldType (uint8)│           │ FuncHandler (func)  │    │
│  ├──────────────────┤           ├─────────────────────┤    │
│  │ FieldName = 0    │           │ Returns map[string] │    │
│  │ FieldBind = 1    │           │   http.Handler      │    │
│  │ FieldExpose = 2  │           │                     │    │
│  └──────────────────┘           └─────────────────────┘    │
│                                                            │
│  ┌──────────────────┐           ┌─────────────────────┐    │
│  │    Constants     │           │   Default Handler   │    │
│  ├──────────────────┤           ├─────────────────────┤    │
│  │ HandlerDefault   │           │ BadHandler struct{} │    │
│  │ BadHandlerName   │           │ NewBadHandler()     │    │
│  │ TimeoutWaiting*  │           │ ServeHTTP() → 500   │    │
│  └──────────────────┘           └─────────────────────┘    │
│                                                            │
└────────────────────────────────────────────────────────────┘
                          │
                          ▼
         ┌────────────────────────────────────┐
         │   Consuming Packages               │
         ├────────────────────────────────────┤
         │ httpserver       (server impl)     │
         │ httpserver/pool  (server registry) │
         │ Application code (configuration)   │
         └────────────────────────────────────┘
```

### Data Flow

1. **Type Definition**:
   * Constants are compiled at build time (zero runtime cost).
   * `FieldType` values are `uint8` enums (1 byte each).
   * `BadHandler` is an empty struct (zero bytes).

2. **Handler Registration**:
   * `FuncHandler` is invoked to retrieve handler map.
   * Map keys use `HandlerDefault` constant or custom strings.
   * Handlers are registered in server configuration.

3. **Field Filtering**:
   * Server pool queries use `FieldType` for type-safe filtering.
   * Switch statements or map keys use enum values.
   * Compile-time type checking prevents invalid fields.

### Limitations

This package is intentionally minimal with the following design constraints:

1. **No Dynamic Field Types**: `FieldType` is a closed enumeration (0, 1, 2). Adding new field types requires modifying this package and recompiling consuming packages.

2. **No Handler Lifecycle Management**: `BadHandler` does not implement graceful shutdown, resource cleanup, or lifecycle hooks. It's a stateless fallback.

3. **Fixed Timeout Values**: Timeout constants (`TimeoutWaitingStop`, `TimeoutWaitingPortFreeing`) are compile-time constants and cannot be configured at runtime.

4. **No Validation**: The package does not validate handler maps returned by `FuncHandler`. Validation is the responsibility of consuming packages.

5. **Single Error Status**: `BadHandler` always returns HTTP 500 Internal Server Error. Custom error codes require implementing a custom `http.Handler`.

6. **No State Management**: Types are stateless primitives. State management (server instances, handler routing) is handled by consuming packages.

---

## Performance

### Benchmarks

Based on type characteristics (Go 1.25, standard library primitives):

| Operation | Runtime Cost | Memory Cost |
|-----------|--------------|-------------|
| **FieldType Comparison** | 0 ns | 0 bytes |
| **Constant Access** | 0 ns | 0 bytes |
| **BadHandler Creation** | <10 ns | 0 bytes (empty struct) |
| **BadHandler.ServeHTTP** | <100 ns | 0 allocations |
| **FuncHandler Invocation** | Depends on implementation | Depends on map size |

*Note: FieldType and constants are compile-time values with zero runtime overhead.*

### Memory Usage

- **Base Overhead**: Zero (all types are primitives or empty structs).
- **FieldType**: 1 byte per variable (`uint8`).
- **BadHandler**: 0 bytes (empty struct, allocated on stack).
- **Constants**: Compiled into binary, no runtime memory allocation.
- **Optimization**: Types designed for zero-allocation scenarios where possible.

### Scalability

- **Type Safety**: Compile-time checks scale to any codebase size.
- **Constant Propagation**: Compiler optimizes constant usage at compile time.
- **Zero Contention**: No shared mutable state, safe for unlimited concurrent use.
- **Handler Maps**: Scalability depends on `FuncHandler` implementation (not managed by this package).

---

## Use Cases

### 1. Server Pool Management

Filter servers in a pool by specific attributes:

```go
// Find all servers listening on a specific address
servers := pool.FilterByField(types.FieldBind, ":8080")
```

### 2. Multi-Handler Server Configuration

Register multiple handlers for different routes or purposes:

```go
cfg.HandlerFunc = func() map[string]http.Handler {
    return map[string]http.Handler{
        types.HandlerDefault: webHandler,
        "api":                apiHandler,
        "metrics":            metricsHandler,
    }
}
```

### 3. Graceful Shutdown

Use standard timeout for server shutdown:

```go
shutdownCtx, cancel := context.WithTimeout(
    context.Background(),
    types.TimeoutWaitingStop,
)
defer cancel()
server.Shutdown(shutdownCtx)
```

### 4. Safe Default Handler

Provide fallback when handler registration fails:

```go
handler := getConfiguredHandler()
if handler == nil {
    handler = types.NewBadHandler()
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/httpserver/types
```

### Basic Field Type Usage

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/httpserver/types"
)

func main() {
    // Use FieldType for server filtering
    filterByField := func(field types.FieldType, value string) {
        switch field {
        case types.FieldName:
            fmt.Println("Filtering by server name:", value)
        case types.FieldBind:
            fmt.Println("Filtering by bind address:", value)
        case types.FieldExpose:
            fmt.Println("Filtering by expose URL:", value)
        }
    }
    
    filterByField(types.FieldBind, ":8080")
}
```

### Handler Registration

```go
package main

import (
    "net/http"
    "github.com/nabbar/golib/httpserver/types"
)

func main() {
    // Define handler registration function
    var handlerFunc types.FuncHandler
    
    handlerFunc = func() map[string]http.Handler {
        return map[string]http.Handler{
            types.HandlerDefault: http.NotFoundHandler(),
            "api":                myAPIHandler,
            "admin":              myAdminHandler,
        }
    }
    
    // Use in server configuration
    handlers := handlerFunc()
    server.RegisterHandlers(handlers)
}
```

### Using Timeout Constants

```go
package main

import (
    "context"
    "time"
    "github.com/nabbar/golib/httpserver/types"
)

func main() {
    // Use predefined timeout for server shutdown
    ctx, cancel := context.WithTimeout(
        context.Background(),
        types.TimeoutWaitingStop,
    )
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **100.0% code coverage** and **32 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All type definitions and constants
- ✅ Field type enumeration and usage
- ✅ Handler creation and ServeHTTP behavior
- ✅ Constant values verification
- ✅ Interface compliance

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Use FieldType constants:**
```go
// ✅ GOOD: Type-safe field filtering
switch filterField {
case types.FieldName:
    // Filter by name
case types.FieldBind:
    // Filter by bind address
}
```

**Use HandlerDefault:**
```go
// ✅ GOOD: Standard handler registration
handlers := map[string]http.Handler{
    types.HandlerDefault: myHandler,
}
```

**Use timeout constants:**
```go
// ✅ GOOD: Consistent timeout management
ctx, cancel := context.WithTimeout(ctx, types.TimeoutWaitingStop)
defer cancel()
```

### ❌ DON'T

**Don't cast arbitrary integers to FieldType:**
```go
// ❌ BAD: Type-unsafe field creation
field := types.FieldType(99)  // No validation

// ✅ GOOD: Use defined constants
field := types.FieldName
```

**Don't rely on BadHandler for production:**
```go
// ❌ BAD: Using error handler for normal traffic
handler := types.NewBadHandler()
server.SetHandler(handler)  // Always returns 500!

// ✅ GOOD: Use as fallback only
handler := configuredHandler
if handler == nil {
    handler = types.NewBadHandler()
}
```

**Don't modify timeout constants:**
```go
// ❌ BAD: Can't modify package constants
types.TimeoutWaitingStop = 10 * time.Second  // Won't compile

// ✅ GOOD: Define your own if needed
const myTimeout = 10 * time.Second
```

---

## API Reference

### Field Types

```go
// FieldType identifies server fields for filtering and listing operations
type FieldType uint8

const (
    FieldName   FieldType = iota  // Server name field
    FieldBind                      // Bind address field (Listen)
    FieldExpose                    // Expose URL field
)
```

### Handler Types

```go
// FuncHandler is the function signature for handler registration
type FuncHandler func() map[string]http.Handler

// NewBadHandler creates a default error handler
func NewBadHandler() http.Handler

// BadHandler returns HTTP 500 for all requests
type BadHandler struct{}
func (o BadHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request)
```

### Constants

```go
const (
    // HandlerDefault is the default handler registration key
    HandlerDefault = "default"
    
    // TimeoutWaitingPortFreeing is the port availability check timeout
    TimeoutWaitingPortFreeing = 250 * time.Microsecond
    
    // TimeoutWaitingStop is the graceful server shutdown timeout
    TimeoutWaitingStop = 5 * time.Second
    
    // BadHandlerName is the identifier for BadHandler
    BadHandlerName = "no handler"
)
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
   - Ensure zero race conditions with `go test -race`

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

- ✅ **100.0% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** - all types are immutable or stateless
- ✅ **Minimal dependencies** - only standard library
- ✅ **Type-safe** - compile-time safety for field types

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Extended Field Types**: Additional server properties for filtering (status, protocol, etc.)
2. **Handler Validation**: Optional validation interface for FuncHandler implementations
3. **Custom Error Codes**: BadHandler with configurable HTTP status codes
4. **Handler Middleware**: Built-in middleware support for handler chains

These are **optional improvements** and not required for production use. The current implementation is stable and minimal by design.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/httpserver/types)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, type definitions, usage patterns, and limitations. Provides detailed explanations of each type and constant.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 100% coverage analysis, and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/httpserver](https://pkg.go.dev/github.com/nabbar/golib/httpserver)** - HTTP server implementation that uses these types. Shows real-world usage of field types and handler registration.

- **[github.com/nabbar/golib/httpserver/pool](https://pkg.go.dev/github.com/nabbar/golib/httpserver/pool)** - Server pool management that uses FieldType for filtering. Demonstrates server filtering by field attributes.

### External References

- **[net/http Package](https://pkg.go.dev/net/http)** - Standard library HTTP package. The types package extends net/http with additional type definitions for server management.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for type definitions, constants, and interface usage.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/httpserver/types`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
