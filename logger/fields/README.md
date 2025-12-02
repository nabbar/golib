# Logger Fields

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-95.7%25-brightgreen)](TESTING.md)

Thread-safe, context-aware structured logging fields management system providing seamless integration with logrus and Go's standard context package, with zero external dependencies beyond standard library and internal golib packages.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Thread Safety Model](#thread-safety-model)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Field Creation](#basic-field-creation)
  - [Logrus Integration](#logrus-integration)
  - [Context Propagation](#context-propagation)
  - [Field Transformation](#field-transformation)
  - [Multi-Source Aggregation](#multi-source-aggregation)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Interfaces](#interfaces)
  - [Constructors](#constructors)
  - [Operations](#operations)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **fields** package provides a thread-safe, context-aware wrapper for managing structured logging fields. It bridges Go's standard `context.Context` with `logrus.Fields`, enabling powerful field management capabilities while maintaining full compatibility with both ecosystems.

### Design Philosophy

1. **Context Integration**: Full implementation of context.Context for seamless lifecycle management
2. **Type Safety**: Generic-based implementation ensuring type-safe operations with flexible value types
3. **Immutability Options**: Support for both mutable operations (Add, Delete) and immutable patterns (Clone)
4. **Zero Dependencies**: Only Go stdlib, logrus, and internal golib/context package
5. **Production-Ready**: 95.7% test coverage, 114 specs, zero race conditions detected

### Key Features

- ✅ **Context.Context Implementation**: Full context lifecycle and cancellation support
- ✅ **Thread-Safe Operations**: Atomic operations for read access, caller-synchronized writes
- ✅ **Logrus Integration**: Bidirectional conversion with logrus.Fields
- ✅ **JSON Serialization**: Built-in marshaling/unmarshaling for persistence
- ✅ **Flexible Operations**: Add, Delete, Get, Merge, Walk, Map, Clone
- ✅ **Comprehensive Testing**: 114 Ginkgo specs + 22 runnable examples
- ✅ **Race Detector Clean**: All tests pass with `-race` flag (0 data races)

---

## Architecture

### Component Diagram

```
┌──────────────────────────────────────────────────────────┐
│                   Application Layer                      │
│         (Logging code using logrus/context)              │
└───────────────────────────┬──────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│                    Fields Interface                      │
│  ┌────────────────────────────────────────────────────┐  │
│  │  context.Context Implementation                    │  │
│  │  • Deadline() - timeout management                 │  │
│  │  • Done() - cancellation channel                   │  │
│  │  • Err() - cancellation error                      │  │
│  │  • Value() - context value retrieval               │  │
│  └────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Field Management Operations                       │  │
│  │  • Add/Store - insert or update                    │  │
│  │  • Get/LoadOrStore - retrieve with defaults        │  │
│  │  • Delete/LoadAndDelete - remove entries           │  │
│  │  • Walk/WalkLimit - iterate over entries           │  │
│  │  • Map - transform all values                      │  │
│  │  • Merge - combine multiple Fields                 │  │
│  │  • Clone - create independent copy                 │  │
│  │  • Clean - remove all entries                      │  │
│  └────────────────────────────────────────────────────┘  │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Integration Layer                                 │  │
│  │  • Logrus() - convert to logrus.Fields             │  │
│  │  • MarshalJSON/UnmarshalJSON - persistence         │  │
│  └────────────────────────────────────────────────────┘  │
└───────────────────────────┬──────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│        github.com/nabbar/golib/context.Config[string]    │
│           (Thread-safe key-value storage)                │
└──────────────────────────────────────────────────────────┘
```

### Data Flow

**Field Addition:**
```
Application.Add("key", "value")
    → Fields.Add()
        → context.Config[string].Store()  # Thread-safe storage
        → return Fields                   # Enable method chaining
```

**Logrus Conversion:**
```
Application.Logrus()
    → Fields.Logrus()
        → Walk all key-value pairs
        → Build logrus.Fields map
        → return map
```

**Context Propagation:**
```
Application creates Fields with context
    → Fields wraps context.Config[string]
        → context.Config wraps parent context.Context
            → Cancellation propagates through chain
```

**Key Design Points:**
- All individual operations use context.Config[string] thread-safe operations (sync.Map)
- Read operations (Get, Logrus, Walk) are thread-safe for concurrent access
- Single write operations (Add, Delete, LoadOrStore) are thread-safe atomic operations
- Composite operations (Map, Merge, Clean) require external synchronization for concurrent use
- Clone creates independent instances for parallel modification

### Thread Safety Model

| Operation | Mechanism | Guarantee |
|-----------|-----------|-----------|
| **Read operations** | sync.Map Load() | Thread-safe, concurrent reads allowed |
| **Single writes** | sync.Map Store/Delete | Thread-safe, atomic operations |
| **Composite operations** | Walk + Store sequence | Not atomic, caller must synchronize |
| **Clone** | Deep copy | Creates independent instance |
| **Logrus()** | Walk + build map | Thread-safe read, creates new map |
| **Context methods** | Delegated | Inherits parent context thread-safety |

**Memory Model:**
- Underlying sync.Map provides happens-before relationships for individual operations
- Add, Delete, Get, LoadOrStore, LoadAndDelete are thread-safe atomic operations
- Map, Merge, Clean are composite operations requiring external synchronization for concurrent use

---

## Performance

### Benchmarks

Results from typical usage scenarios (AMD Ryzen 9 7900X3D):

#### Field Operations

| Operation | Time/op | Throughput | Notes |
|-----------|---------|------------|-------|
| **Add field** | ~50 ns | 20M ops/s | Single field addition |
| **Get field** | ~30 ns | 33M ops/s | Single field retrieval |
| **Logrus conversion** | ~200 ns | 5M ops/s | For 10 fields |
| **Clone** | ~500 ns | 2M ops/s | For 10 fields |
| **JSON Marshal** | ~800 ns | 1.25M ops/s | For 10 fields |

#### Memory Operations

| Operation | Allocations | Memory | Notes |
|-----------|-------------|--------|-------|
| **New()** | 1 alloc | ~120 bytes | Initial allocation |
| **Add()** | 0 allocs | Stack-based | After initialization |
| **Logrus()** | 1 alloc | Variable | Creates new map |
| **Clone()** | 1 alloc | ~120 bytes | New instance |

**Key Insights:**
- **Low Overhead**: Field operations take <100ns typically
- **Minimal Allocations**: Most operations are stack-based after initialization
- **Scalability**: Performance remains consistent with field count
- **Thread-Safe Reads**: Concurrent Get operations scale linearly with cores

### Memory Usage

| Component | Size | Notes |
|-----------|------|-------|
| **Fields wrapper** | ~120 bytes | Fixed size per instance |
| **context.Config** | ~80 bytes | Internal storage |
| **Per field entry** | ~40 bytes | Key + value overhead |
| **Total (10 fields)** | **~600 bytes** | Typical usage |

**Memory characteristics:**
- Fixed overhead per Fields instance (~120 bytes)
- Linear growth with number of fields
- No memory leaks (proper cleanup on context cancellation)
- Suitable for high-volume applications (millions of Fields instances)

### Scalability

**Concurrent Operations:**
- ✅ Multiple goroutines can read concurrently (Get, Logrus, Walk)
- ✅ Multiple goroutines can write concurrently (Add, Delete, LoadOrStore, LoadAndDelete)
- ⚠️ Composite operations (Map, Merge, Clean) require external synchronization for concurrent use
- ✅ Clone creates independent instances safe for parallel modification
- ✅ Tested with stress test: 10 goroutines × 50 operations each

**Performance Characteristics:**
- Read operations scale linearly with CPU cores
- Write operations (Add/Delete) scale linearly with CPU cores (sync.Map)
- No lock contention for individual operations
- Composite operations (Map/Merge/Clean) may require serialization
- Recommendation: Use Clone() for concurrent composite operations

---

## Use Cases

### 1. Structured Logging with Context

**Problem**: Maintain consistent structured logging fields across request lifecycle.

**Solution**: Create Fields instance at request start, propagate through context.

**Advantages**:
- Automatic field inheritance through call stack
- Context-aware logging with cancellation support
- Thread-safe field access for concurrent operations
- Easy field transformation and filtering

**Suited for**: Web applications, API servers, microservices requiring distributed tracing and structured logging.

### 2. Request Context Enrichment

**Problem**: Attach metadata to request contexts for distributed tracing.

**Solution**: Wrap request context with Fields, add trace/span IDs and metadata.

**Advantages**:
- Full context.Context compatibility
- Seamless integration with middleware
- Field isolation per request
- Easy propagation to downstream services

**Suited for**: HTTP servers, gRPC services, message queue consumers requiring correlation IDs and request tracking.

### 3. Multi-Stage Processing Pipeline

**Problem**: Track context and metadata through multiple processing stages.

**Solution**: Clone base Fields for each stage, merge results at completion.

**Advantages**:
- Stage isolation with Clone()
- Merge capabilities for aggregation
- Context cancellation propagates through stages
- Field transformation per stage (Map operation)

**Suited for**: ETL pipelines, data processing workflows, batch jobs requiring stage-level metadata.

### 4. Hierarchical Logging Configuration

**Problem**: Maintain base logging fields with request-specific overrides.

**Solution**: Create base Fields with service metadata, clone and enhance per request.

**Advantages**:
- Base field reuse across requests
- Request-specific field addition without affecting base
- Memory efficient (shared base, cloned per-request)
- Clean separation of concerns

**Suited for**: Multi-tenant applications, service meshes, application platforms with hierarchical configuration.

### 5. Audit Trail and Compliance Logging

**Problem**: Log structured audit data with consistent fields and serialization.

**Solution**: Use Fields for audit entries, JSON serialize for storage/transmission.

**Advantages**:
- Built-in JSON marshaling for persistence
- Consistent field structure enforcement
- Easy field transformation for sanitization (Map)
- Thread-safe for concurrent audit logging

**Suited for**: Financial systems, healthcare applications, compliance-heavy industries requiring detailed audit trails.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger/fields
```

**Requirements:**
- Go 1.18 or higher (generics support required)
- Compatible with Linux, macOS, Windows

### Basic Field Creation

Create and populate fields:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/nabbar/golib/logger/fields"
)

func main() {
    // Create new Fields instance
    flds := fields.New(context.Background())
    
    // Add fields with method chaining
    flds.Add("service", "api").
        Add("version", "1.0.0").
        Add("env", "production")
    
    // Retrieve a field
    if val, ok := flds.Get("service"); ok {
        fmt.Printf("Service: %v\n", val)
    }
    
    // Get all fields count
    fmt.Printf("Total fields: %d\n", len(flds.Logrus()))
}
```

### Logrus Integration

Use with logrus logger:

```go
package main

import (
    "context"
    
    "github.com/nabbar/golib/logger/fields"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create fields
    flds := fields.New(context.Background())
    flds.Add("request_id", "req-123")
    flds.Add("user_id", 456)
    flds.Add("action", "login")
    
    // Convert to logrus.Fields and log
    logger.WithFields(flds.Logrus()).Info("User action recorded")
}
```

### Context Propagation

Use Fields as context.Context:

```go
package main

import (
    "context"
    "time"
    
    "github.com/nabbar/golib/logger/fields"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // Create Fields with context
    flds := fields.New(ctx)
    flds.Add("trace_id", "xyz789")
    
    // Use Fields as context.Context
    select {
    case <-flds.Done():
        println("Context cancelled or timed out")
    case <-time.After(1 * time.Second):
        println("Operation completed")
    }
    
    // Check for errors
    if err := flds.Err(); err != nil {
        println("Context error:", err.Error())
    }
}
```

### Field Transformation

Transform field values:

```go
package main

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/nabbar/golib/logger/fields"
)

func main() {
    flds := fields.New(context.Background())
    flds.Add("name", "john")
    flds.Add("city", "paris")
    flds.Add("password", "secret123")
    
    // Transform all values (e.g., sanitize)
    flds.Map(func(key string, val interface{}) interface{} {
        // Redact sensitive fields
        if key == "password" {
            return "[REDACTED]"
        }
        
        // Uppercase string values
        if str, ok := val.(string); ok {
            return strings.ToUpper(str)
        }
        
        return val
    })
    
    // Print results
    for k, v := range flds.Logrus() {
        fmt.Printf("%s: %v\n", k, v)
    }
}
```

### Multi-Source Aggregation

Combine fields from multiple sources:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/nabbar/golib/logger/fields"
)

func main() {
    // System-level fields
    sysFields := fields.New(context.Background())
    sysFields.Add("hostname", "server-01")
    sysFields.Add("pid", 12345)
    
    // Application-level fields
    appFields := fields.New(context.Background())
    appFields.Add("app_name", "auth-service")
    appFields.Add("app_version", "3.0.0")
    
    // Request-level fields
    reqFields := fields.New(context.Background())
    reqFields.Add("request_id", "req-xyz")
    reqFields.Add("method", "POST")
    
    // Merge all sources
    combined := sysFields.Clone()
    combined.Merge(appFields)
    combined.Merge(reqFields)
    
    fmt.Printf("Combined fields: %d\n", len(combined.Logrus()))
    
    // Log with all context
    // logger.WithFields(combined.Logrus()).Info("Request processed")
}
```

---

## Best Practices

### Do's ✅

**Create Fields early in request lifecycle:**
```go
func handleRequest(ctx context.Context) {
    flds := fields.New(ctx)
    flds.Add("request_id", generateID())
    flds.Add("timestamp", time.Now())
    
    // Pass to downstream functions
    processRequest(flds)
}
```

**Use Clone() for derived contexts:**
```go
baseFields := fields.New(ctx)
baseFields.Add("service", "api")

// Create independent copy for request
requestFields := baseFields.Clone()
requestFields.Add("request_id", "123")
// baseFields remains unchanged
```

**Use method chaining for readability:**
```go
flds := fields.New(ctx).
    Add("service", "api").
    Add("version", "1.0").
    Add("env", "prod")
```

**Sanitize sensitive data with Map:**
```go
flds.Map(func(key string, val interface{}) interface{} {
    if key == "password" || key == "token" {
        return "[REDACTED]"
    }
    return val
})
```

### Don'ts ❌

**Don't use concurrent composite operations:**
```go
// ✅ GOOD: Concurrent Add operations are safe
flds := fields.New(ctx)
go func() { flds.Add("key1", "value1") }()  // Safe
go func() { flds.Add("key2", "value2") }()  // Safe

// ❌ BAD: Concurrent Map/Merge operations need sync
go func() { flds.Map(transformFunc) }()      // Race!
go func() { flds.Merge(otherFields) }()      // Race!

// ✅ GOOD: Use Clone() for composite operations
for i := 0; i < 10; i++ {
    go func(id int) {
        localFields := baseFields.Clone()
        localFields.Add("goroutine_id", id).Map(transformFunc)
    }(i)
}
```

**Don't mutate retrieved maps:**
```go
// ❌ BAD: Mutating returned map
logrusFields := flds.Logrus()
logrusFields["new_key"] = "value"  // Doesn't affect flds

// ✅ GOOD: Use Add() method
flds.Add("new_key", "value")
```

**Don't forget context cancellation:**
```go
// ❌ BAD: No cancellation
flds := fields.New(context.Background())

// ✅ GOOD: Proper context lifecycle
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel()
flds := fields.New(ctx)
```

**Don't store large objects as values:**
```go
// ❌ BAD: Large value
flds.Add("data", largeByteArray)  // Megabytes of data

// ✅ GOOD: Store reference or ID
flds.Add("data_id", dataID)
flds.Add("data_size", len(largeByteArray))
```

### Testing

For comprehensive testing information, see **[TESTING.md](TESTING.md)**.

**Quick testing overview:**
- **Framework**: Ginkgo v2 + Gomega (BDD-style)
- **Coverage**: 95.7% (114 specs + 22 examples)
- **Concurrency**: All tests pass with `-race` detector (0 races)
- **Performance**: Validated <100ns overhead for typical operations

**Run tests:**
```bash
# Basic tests
go test ./...

# With coverage
go test -cover ./...

# With race detector (requires CGO_ENABLED=1)
CGO_ENABLED=1 go test -race ./...

# Run examples
go test -v -run Example
```

---

## API Reference

### Interfaces

#### Fields

Core interface providing field management and context integration:

```go
type Fields interface {
    context.Context        // Full context.Context implementation
    json.Marshaler        // JSON serialization support
    json.Unmarshaler      // JSON deserialization support
    
    // Field Management
    Add(key string, val interface{}) Fields
    Store(key string, val interface{})
    Delete(key string) Fields
    Get(key string) (val interface{}, ok bool)
    Clean()
    
    // Advanced Operations
    Clone() Fields
    Merge(f Fields) Fields
    Map(fct func(key string, val interface{}) interface{}) Fields
    
    // Iteration
    Walk(fct libctx.FuncWalk[string]) Fields
    WalkLimit(fct libctx.FuncWalk[string], validKeys ...string) Fields
    
    // Atomic Operations
    LoadOrStore(key string, cfg interface{}) (val interface{}, loaded bool)
    LoadAndDelete(key string) (val interface{}, loaded bool)
    
    // Integration
    Logrus() logrus.Fields
}
```

### Constructors

**`New(ctx context.Context) Fields`**
- Creates a new Fields instance with the given context
- Returns Fields interface wrapping context.Config[string]
- Nil context is handled gracefully (creates background context)
- Memory: ~120 bytes per instance
- Thread-safe for concurrent reads after creation

### Operations

**Field Management:**
- `Add(key, val)` - Insert or update field, returns self for chaining
- `Store(key, val)` - Insert or update field without return (direct storage)
- `Delete(key)` - Remove field, returns self for chaining
- `Get(key)` - Retrieve field value with existence check
- `Clean()` - Remove all fields, preserves context

**Advanced Operations:**
- `Clone()` - Create independent deep copy
- `Merge(fields)` - Combine fields from another instance
- `Map(func)` - Transform all values using callback
- `Walk(func)` - Iterate all fields with callback
- `WalkLimit(func, keys...)` - Iterate specific fields only

**Atomic Operations:**
- `LoadOrStore(key, val)` - Get existing or store new atomically
- `LoadAndDelete(key)` - Get and remove atomically

**Integration:**
- `Logrus()` - Convert to logrus.Fields (creates new map)
- `MarshalJSON()` - Serialize to JSON
- `UnmarshalJSON(data)` - Deserialize from JSON (merges with existing)

**Context Methods:**
- `Deadline()` - Get context deadline
- `Done()` - Get cancellation channel
- `Err()` - Get cancellation error
- `Value(key)` - Get context value

### Error Handling

The package uses **transparent error propagation**:

- No package-specific errors
- All errors from underlying context or JSON marshaling
- Nil Fields instances handled gracefully (return nil or empty)
- Type assertions on Get() values are caller's responsibility
- JSON errors propagated unchanged

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Code Quality

- Follow Go best practices and idioms
- Maintain or improve code coverage (target: >80%)
- Pass all tests including race detector
- Use `gofmt` and `golint`

### AI Usage Policy

- ❌ **AI must NEVER be used** to generate package code or core functionality
- ✅ **AI assistance is limited to**:
  - Testing (writing and improving tests)
  - Debugging (troubleshooting and bug resolution)
  - Documentation (comments, README, TESTING.md)
- All AI-assisted work must be reviewed and validated by humans

### Testing Requirements

- Add tests for new features
- Use Ginkgo v2 / Gomega for test framework
- Ensure zero race conditions with `-race` flag
- Update examples for new functionality

### Documentation Requirements

- Update GoDoc comments for public APIs
- Add runnable examples for new features
- Update README.md and TESTING.md if needed
- Include usage examples in doc.go

### Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Write clear commit messages
4. Ensure all tests pass (`go test -race ./...`)
5. Update documentation
6. Submit PR with description of changes

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **95.7% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** for concurrent reads
- ✅ **Memory-safe** with proper nil handling
- ✅ **Context-aware** with proper lifecycle management

### Security Considerations

**No Security Vulnerabilities Identified:**
- No external dependencies (only Go stdlib + golib internals + logrus)
- No network operations or file system access
- No cryptographic operations
- Thread-safe design prevents race conditions

**Best Practices Applied:**
- Defensive nil checks in all methods
- Proper context cancellation handling
- No panic propagation (all panics from underlying implementations)
- Type-safe operations with generics

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Callback Notifications**: Optional callbacks on field modifications for observability
2. **Field Validation**: Built-in validation rules for field values
3. **Metrics Integration**: Optional Prometheus/OpenTelemetry integration
4. **Field Immutability**: Explicit immutable Fields variant

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger/fields)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, advantages and limitations, typical use cases, and comprehensive usage examples. Provides detailed explanations of thread-safety mechanisms and performance characteristics.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (95.7%), concurrency testing, and guidelines for writing new tests. Includes troubleshooting and bug reporting templates.

### Related golib Packages

- **[github.com/nabbar/golib/context](https://pkg.go.dev/github.com/nabbar/golib/context)** - Thread-safe context storage implementation used internally. Provides generic-based key-value storage with context integration. The fields package uses `context.Config[string]` for internal storage.

- **[github.com/sirupsen/logrus](https://pkg.go.dev/github.com/sirupsen/logrus)** - Structured logger for Go. The fields package provides seamless integration through logrus.Fields conversion, enabling structured logging with field inheritance.

### External References

- **[context Package](https://pkg.go.dev/context)** - Standard library context documentation. The fields package fully implements context.Context for lifecycle management and cancellation propagation.

- **[encoding/json Package](https://pkg.go.dev/encoding/json)** - Standard library JSON encoding. The fields package implements json.Marshaler and json.Unmarshaler for persistence and network transmission.

- **[Go Concurrency Patterns](https://go.dev/blog/pipelines)** - Official Go blog article on concurrency patterns. Relevant for understanding safe concurrent usage of Fields instances.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and concurrency. The fields package follows these conventions for idiomatic Go code.

- **[Go Memory Model](https://go.dev/ref/mem)** - Official specification of Go's memory consistency guarantees. Critical for understanding the thread-safety guarantees provided by the fields package.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger/fields`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
