# Logger Types

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-100%25-brightgreen)](TESTING.md)

Core types, interfaces, and constants for structured logging with standardized field names and extensible hook mechanisms.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Field Constants](#field-constants)
  - [Hook Interface](#hook-interface)
- [Performance](#performance)
  - [Field Constants](#field-constants-1)
  - [Hook Implementation](#hook-implementation)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Field Usage](#basic-field-usage)
  - [Error Logging](#error-logging)
  - [Basic Hook Implementation](#basic-hook-implementation)
  - [Hook Lifecycle](#hook-lifecycle)
  - [Multiple Hooks](#multiple-hooks)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Field Constants](#field-constants-2)
  - [Hook Interface](#hook-interface-1)
  - [Interface Composition](#interface-composition)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **types** package provides the foundational types for the logger subsystem in github.com/nabbar/golib. It defines standardized field names for structured logging and an extensible Hook interface for advanced log processing, filtering, and multi-destination output.

### Design Philosophy

1. **Standardization**: Consistent field names across all logger implementations
2. **Extensibility**: Hook interface for custom log processors without core modifications
3. **Type Safety**: Strong typing prevents typos and ensures field name consistency
4. **Minimal Dependencies**: Only standard library and logrus required
5. **Logrus Integration**: Full compatibility with logrus.Hook and io.WriteCloser interfaces

### Key Features

- ✅ **9 Standard Field Constants**: Type-safe field names for structured logging
- ✅ **Extended Hook Interface**: Adds lifecycle management to logrus.Hook
- ✅ **Background Processing**: Run() method for async log processing
- ✅ **Context Integration**: Context-based cancellation and lifecycle control
- ✅ **Thread-Safe**: Constants are immutable, Hook implementations guide concurrency
- ✅ **Zero Runtime Overhead**: Constants inlined at compile time
- ✅ **io.WriteCloser Support**: Direct write capabilities for hooks
- ✅ **Multiple Loggers**: Single hook can register with multiple logger instances

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    logger/types                         │
├──────────────────────┬──────────────────────────────────┤
│                      │                                  │
│   Field Constants    │         Hook Interface           │
│   (fields.go)        │         (hook.go)                │
│                      │                                  │
│  - FieldTime         │  Extends:                        │
│  - FieldLevel        │    • logrus.Hook                 │
│  - FieldStack        │    • io.WriteCloser              │
│  - FieldCaller       │                                  │
│  - FieldFile         │  Methods:                        │
│  - FieldLine         │    • RegisterHook(log)           │
│  - FieldMessage      │    • Run(ctx)                    │
│  - FieldError        │    • IsRunning()                 │
│  - FieldData         │    • Fire(entry)                 │
│                      │    • Levels()                    │
│                      │    • Write(p)                    │
│                      │    • Close()                     │
└──────────────────────┴──────────────────────────────────┘
```

### Field Constants

The package provides 9 standard field name constants organized into three categories:

**Metadata fields**: Information about the log entry
- `FieldTime` - Timestamp (RFC3339 format)
- `FieldLevel` - Severity level (debug, info, warn, error, fatal)

**Trace fields**: Execution context and debugging information
- `FieldStack` - Full stack trace (multi-line)
- `FieldCaller` - Function/method identifier (package.function)
- `FieldFile` - Source code file name
- `FieldLine` - Line number in source file

**Content fields**: Log message and associated data
- `FieldMessage` - Primary log message text
- `FieldError` - Error description or message
- `FieldData` - Additional structured data (maps, objects)

Example JSON log output:
```json
{
  "time": "2025-01-01T12:00:00Z",
  "level": "error",
  "message": "operation failed",
  "error": "connection timeout",
  "file": "main.go",
  "line": 42,
  "caller": "main.processRequest",
  "data": {"user_id": 123}
}
```

### Hook Interface

The Hook interface extends `logrus.Hook` with lifecycle management:

**Interface composition**:
- `logrus.Hook` - Fire(entry) and Levels() methods
- `io.WriteCloser` - Write(p) and Close() methods

**Additional methods**:
- `RegisterHook(log)` - Self-registration with logger
- `Run(ctx)` - Background processing with context cancellation
- `IsRunning()` - State checking for monitoring

**Key characteristics**:
- Fire() called synchronously for each log entry
- Run() executes in background goroutine
- Context-based graceful shutdown
- Thread-safe when properly implemented

---

## Performance

### Field Constants

- **Zero runtime overhead**: Constants inlined at compile time
- **No allocations**: String constants don't allocate memory
- **Optimized comparisons**: Compiler optimizes constant string comparisons
- **Thread-safe**: Constants are immutable by definition

### Hook Implementation

**Fast patterns**:
```go
// Fire() offloads work to background goroutine
func (h *Hook) Fire(entry *logrus.Entry) error {
    select {
    case h.queue <- entry:
        return nil
    default:
        return errors.New("queue full")
    }
}
```

**Slow patterns to avoid**:
```go
// DON'T: Synchronous I/O in Fire() blocks all logging
func (h *Hook) Fire(entry *logrus.Entry) error {
    return h.sendToRemoteAPI(entry) // BLOCKS LOGGING!
}
```

**Performance guidelines**:
- Keep Fire() < 1ms execution time
- Use buffered channels between Fire() and Run()
- Perform heavy operations (I/O, formatting) in Run()
- Use atomic operations for state management

---

## Use Cases

### 1. Consistent Structured Logging

**Problem**: Inconsistent field names across application components.

**Solution**: Use standard field constants.

```go
import "github.com/nabbar/golib/logger/types"

log.WithFields(logrus.Fields{
    types.FieldFile:  "handler.go",
    types.FieldLine:  123,
    types.FieldError: err.Error(),
}).Error("request failed")
```

### 2. Multi-Destination Logging

**Problem**: Log to multiple destinations (file, syslog, metrics) simultaneously.

**Solution**: Implement Hook interface for each destination.

```go
fileHook := &FileHook{path: "/var/log/app.log"}
syslogHook := &SyslogHook{facility: "daemon"}
metricsHook := &MetricsHook{registry: prometheus.DefaultRegisterer}

fileHook.RegisterHook(logger)
syslogHook.RegisterHook(logger)
metricsHook.RegisterHook(logger)
```

### 3. Log Filtering and Transformation

**Problem**: Filter sensitive data or transform log format before output.

**Solution**: Hook Fire() method processes entries.

```go
type SensitiveDataFilter struct{}

func (f *SensitiveDataFilter) Fire(entry *logrus.Entry) error {
    if pwd, ok := entry.Data["password"]; ok {
        entry.Data["password"] = "***REDACTED***"
    }
    return nil
}
```

### 4. Async Log Aggregation

**Problem**: Aggregate logs from high-frequency sources without blocking.

**Solution**: Buffer entries in Fire(), batch process in Run().

```go
type BatchHook struct {
    queue chan *logrus.Entry
}

func (h *BatchHook) Fire(entry *logrus.Entry) error {
    h.queue <- entry
    return nil
}

func (h *BatchHook) Run(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    batch := make([]*logrus.Entry, 0, 100)
    
    for {
        select {
        case entry := <-h.queue:
            batch = append(batch, entry)
            if len(batch) >= 100 {
                h.writeBatch(batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                h.writeBatch(batch)
                batch = batch[:0]
            }
        case <-ctx.Done():
            h.writeBatch(batch)
            return
        }
    }
}
```

### 5. Log Metrics Collection

**Problem**: Track log volume and error rates.

**Solution**: Hook increments metrics in Fire().

```go
type MetricsHook struct {
    totalLogs    atomic.Int64
    errorLogs    atomic.Int64
}

func (h *MetricsHook) Fire(entry *logrus.Entry) error {
    h.totalLogs.Add(1)
    if entry.Level <= logrus.ErrorLevel {
        h.errorLogs.Add(1)
    }
    return nil
}
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger/types
```

### Basic Field Usage

```go
package main

import (
    "github.com/nabbar/golib/logger/types"
    "github.com/sirupsen/logrus"
)

func main() {
    log := logrus.New()
    
    log.WithFields(logrus.Fields{
        types.FieldFile: "main.go",
        types.FieldLine: 42,
    }).Info("application started")
}
```

### Error Logging

```go
func processRequest(req *Request) error {
    if err := validate(req); err != nil {
        log.WithFields(logrus.Fields{
            types.FieldError:  err.Error(),
            types.FieldFile:   "handler.go",
            types.FieldLine:   123,
            types.FieldCaller: "processRequest",
        }).Error("validation failed")
        return err
    }
    return nil
}
```

### Basic Hook Implementation

```go
package main

import (
    "context"
    "io"
    "sync/atomic"
    
    "github.com/nabbar/golib/logger/types"
    "github.com/sirupsen/logrus"
)

type SimpleHook struct {
    running atomic.Bool
    output  io.Writer
}

func (h *SimpleHook) Fire(entry *logrus.Entry) error {
    line, _ := entry.String()
    _, err := h.output.Write([]byte(line))
    return err
}

func (h *SimpleHook) Levels() []logrus.Level {
    return logrus.AllLevels
}

func (h *SimpleHook) RegisterHook(log *logrus.Logger) {
    log.AddHook(h)
}

func (h *SimpleHook) Run(ctx context.Context) {
    h.running.Store(true)
    defer h.running.Store(false)
    <-ctx.Done()
}

func (h *SimpleHook) IsRunning() bool {
    return h.running.Load()
}

func (h *SimpleHook) Write(p []byte) (n int, err error) {
    return h.output.Write(p)
}

func (h *SimpleHook) Close() error {
    return nil
}
```

### Hook Lifecycle

```go
func main() {
    log := logrus.New()
    hook := &SimpleHook{output: os.Stdout}
    
    // Register hook
    hook.RegisterHook(log)
    
    // Start background processing
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go hook.Run(ctx)
    
    // Use logger
    log.Info("processing started")
    
    // Cleanup
    cancel()
    hook.Close()
}
```

### Multiple Hooks

```go
func main() {
    log := logrus.New()
    
    // Multiple hooks for different purposes
    fileHook := &FileHook{path: "app.log"}
    consoleHook := &ConsoleHook{colored: true}
    metricsHook := &MetricsHook{}
    
    fileHook.RegisterHook(log)
    consoleHook.RegisterHook(log)
    metricsHook.RegisterHook(log)
    
    // All hooks receive log entries
    log.Info("this goes to file, console, and metrics")
}
```

---

## Best Practices

### Testing

The package includes comprehensive tests with **100% code coverage** for constant definitions and interface compliance. Hook implementations are tested via mock implementations in **32 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Quick test commands:**
```bash
go test ./...                          # Run all tests
go test -cover ./...                   # With coverage
CGO_ENABLED=1 go test -race ./...      # With race detection
```

See **[TESTING.md](TESTING.md)** for comprehensive testing documentation.

### ✅ DO

**Use Field Constants**:
```go
// ✅ GOOD: Type-safe field names
log.WithField(types.FieldError, err.Error())
```

**Fast Fire() Implementation**:
```go
// ✅ GOOD: Quick return, offload to Run()
func (h *Hook) Fire(entry *logrus.Entry) error {
    select {
    case h.queue <- entry:
        return nil
    default:
        return errors.New("queue full")
    }
}
```

**Context Management**:
```go
// ✅ GOOD: Proper lifecycle
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go hook.Run(ctx)
```

**Thread Safety**:
```go
// ✅ GOOD: Atomic state
type Hook struct {
    running atomic.Bool
    mu      sync.Mutex
    data    map[string]interface{}
}
```

### ❌ DON'T

**Hardcode Field Names**:
```go
// ❌ BAD: Prone to typos
log.WithField("eror", err.Error()) // Typo!
```

**Slow Fire() Method**:
```go
// ❌ BAD: Blocks all logging
func (h *Hook) Fire(entry *logrus.Entry) error {
    time.Sleep(100 * time.Millisecond)
    return h.sendToAPI(entry)
}
```

**Forget Context Cancellation**:
```go
// ❌ BAD: Goroutine leak
go hook.Run(context.Background())
```

**Race Conditions**:
```go
// ❌ BAD: Unprotected shared state
type Hook struct {
    count int // Race condition!
}

func (h *Hook) Fire(entry *logrus.Entry) error {
    h.count++ // UNSAFE!
    return nil
}
```

---

## API Reference

### Field Constants

```go
const (
    FieldTime    = "time"     // Timestamp (RFC3339)
    FieldLevel   = "level"    // Severity level
    FieldStack   = "stack"    // Stack trace
    FieldCaller  = "caller"   // Function identifier
    FieldFile    = "file"     // Source file name
    FieldLine    = "line"     // Line number
    FieldMessage = "message"  // Log message
    FieldError   = "error"    // Error description
    FieldData    = "data"     // Structured data
)
```

**Usage**: Keys in `logrus.Fields` maps for structured logging.

**Thread safety**: Immutable constants, safe for concurrent use.

### Hook Interface

```go
type Hook interface {
    logrus.Hook        // Fire(entry), Levels()
    io.WriteCloser     // Write(p), Close()
    
    RegisterHook(log *logrus.Logger)
    Run(ctx context.Context)
    IsRunning() bool
}
```

**Methods**:

- **`Fire(entry *logrus.Entry) error`** - Process log entry (from logrus.Hook)
  - Called synchronously for every matching log entry
  - MUST return quickly (< 1ms recommended)
  - Offload heavy work to Run() via channels

- **`Levels() []logrus.Level`** - Return handled log levels (from logrus.Hook)
  - Return `logrus.AllLevels` to process all levels
  - Filter to reduce Fire() call frequency

- **`Write(p []byte) (n int, err error)`** - Direct write (from io.Writer)
  - Bypass logrus for external log sources
  - Must handle concurrent calls safely

- **`Close() error`** - Cleanup resources (from io.Closer)
  - Idempotent (safe to call multiple times)
  - Wait for in-flight operations

- **`RegisterHook(log *logrus.Logger)`** - Register with logger
  - Call `log.AddHook(h)`
  - Perform initialization
  - Can register with multiple loggers

- **`Run(ctx context.Context)`** - Background processing
  - Execute in goroutine: `go hook.Run(ctx)`
  - Perform heavy operations (I/O, formatting)
  - Respect `ctx.Done()` for graceful shutdown

- **`IsRunning() bool`** - Check operational state
  - Return true while Run() executes
  - Use atomic.Bool for thread safety

### Interface Composition

**logrus.Hook** (embedded):
- Standard logrus hook integration
- All logrus features available

**io.WriteCloser** (embedded):
- Direct I/O capabilities
- Standard cleanup interface

**Additional Methods**:
- Lifecycle management (RegisterHook, Run, IsRunning)
- Context-based control
- Monitoring support

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain 100% code coverage for constant definitions
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
   - Test Hook implementations with mocks

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

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Improvements & Security

### Current Status

The package is **production-ready** with no urgent improvements or security vulnerabilities identified.

### Code Quality Metrics

- ✅ **100% coverage** for constant definitions (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** constants and interface guidance
- ✅ **Type-safe** field names prevent typos
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**Field Constants**:
1. Field namespacing mechanism for custom extensions
2. Field validation helpers
3. Field grouping utilities
4. Standard field sets for common scenarios

**Hook Interface**:
1. Priority system for hook execution order
2. Hook chaining helpers
3. Built-in buffering/batching utilities
4. Standard hook implementations (file, syslog, etc.)

**Quality of Life**:
1. Field builder pattern for complex log entries
2. Hook factory functions
3. Testing utilities for Hook implementations
4. Performance profiling helpers

These are **optional improvements** and not required for production use. The current implementation is stable and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger/types)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, field constant categories, Hook interface details, performance considerations, and best practices for production use.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 100% coverage analysis, example tests, and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/logger/config](https://pkg.go.dev/github.com/nabbar/golib/logger/config)** - Logger configuration using types defined in this package. Provides configuration structures and validation for logger setup.

- **[github.com/nabbar/golib/logger/entry](https://pkg.go.dev/github.com/nabbar/golib/logger/entry)** - Log entry management using standard field names. Implements advanced entry handling and formatting.

- **[github.com/nabbar/golib/logger/fields](https://pkg.go.dev/github.com/nabbar/golib/logger/fields)** - Field manipulation utilities. Provides helpers for working with structured log fields.

- **[github.com/nabbar/golib/logger/gorm](https://pkg.go.dev/github.com/nabbar/golib/logger/gorm)** - GORM logger integration using Hook interface. Bridges GORM ORM logging with logrus.

### External References

- **[Logrus Documentation](https://github.com/sirupsen/logrus)** - The structured logging library that this package extends. Understanding logrus.Hook interface is essential for implementing custom hooks.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, constants, and concurrent programming. This package follows these conventions.

- **[Context Package](https://pkg.go.dev/context)** - Standard library documentation for context.Context. The Hook interface uses context for lifecycle management and graceful shutdown.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger/types`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
