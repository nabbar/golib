# Logger Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)
[![Coverage](https://img.shields.io/badge/Coverage-90.9%25-brightgreen)](TESTING.md)

Production-ready structured logging system for Go applications with flexible output management, field injection, level-based filtering, and extensive integration capabilities.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Test Coverage](#test-coverage)
  - [Memory Profile](#memory-profile)
  - [Concurrency](#concurrency)
- [Subpackages](#subpackages)
  - [config](#config)
  - [entry](#entry)
  - [fields](#fields)
  - [level](#level)
  - [gorm](#gorm)
  - [hashicorp](#hashicorp)
  - [hookfile](#hookfile)
  - [hooksyslog](#hooksyslog)
  - [hookstdout / hookstderr](#hookstdout--hookstderr)
  - [hookwriter](#hookwriter)
  - [types](#types)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Logging](#basic-logging)
  - [Configured Logger](#configured-logger)
  - [With Persistent Fields](#with-persistent-fields)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Logger Interface](#logger-interface)
  - [Configuration](#configuration)
  - [Log Levels](#log-levels)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The logger package provides a comprehensive structured logging solution built on top of [logrus](https://github.com/sirupsen/logrus). It extends logrus with advanced features while maintaining compatibility with standard Go logging interfaces.

### Design Philosophy

1. **Structured Logging**: JSON-formatted logs with custom fields for better observability
2. **Flexible Output**: Multiple simultaneous outputs (file, syslog, stdout/stderr)
3. **Thread-Safe**: Concurrent logging without data races
4. **Performance**: Efficient buffering and minimal overhead
5. **Integration-Ready**: Compatible with popular frameworks (GORM, Hashicorp, spf13)
6. **Observable**: Level-based filtering and customizable formatting

### Key Features

- **8 Log Levels**: Panic, Fatal, Error, Warning, Info, Debug, Trace, + Null
- **Multiple Outputs**: File, syslog, stdout, stderr, custom writers
- **Field Injection**: Persistent and per-entry custom fields
- **Hooks System**: Extensible logging pipeline
- **File Rotation**: Size and age-based automatic rotation
- **JSON/Text Format**: Configurable output formatting
- **io.Writer Interface**: Standard Go writer compatibility
- **Thread-Safe**: Safe for concurrent use
- **Clone Support**: Duplicate loggers with custom settings
- **Integration**: GORM, Hashicorp, spf13/cobra adapters

---

## Architecture

### Component Diagram

```
logger/
├── logger               # Core logging implementation
│   ├── interface.go     # Logger interface
│   ├── model.go         # Logger state management
│   ├── log.go           # Log methods (Debug, Info, etc.)
│   ├── manage.go        # Configuration management
│   └── iowritecloser.go # io.Writer implementation
├── config/              # Configuration management
│   ├── model.go         # Config structure
│   ├── validation.go    # Config validation
│   └── options.go       # Logger options
├── entry/               # Log entry management
│   ├── interface.go     # Entry interface
│   ├── model.go         # Entry implementation
│   └── format.go        # Entry formatting
├── fields/              # Custom fields management
│   ├── interface.go     # Fields interface
│   ├── model.go         # Fields implementation
│   └── merge.go         # Fields merging
├── level/               # Log level enumeration
│   ├── interface.go     # Level interface
│   ├── model.go         # Level implementation
│   └── parse.go         # Level parsing
├── hooks/               # Output hooks
│   ├── hookfile/        # File output hook
│   ├── hooksyslog/      # Syslog output hook
│   ├── hookstdout/      # Stdout hook
│   ├── hookstderr/      # Stderr hook
│   └── hookwriter/      # Custom writer hook
├── integrations/        # Third-party integrations
│   ├── gorm/            # GORM logger adapter
│   ├── hashicorp/       # Hashicorp logger adapter
│   └── types/           # Common types
└── spf13.go             # spf13/jwalterweatherman integration
```

### Component Hierarchy

```
┌──────────────────────────────────────┐
│            Logger Package            │
│       Structured Logging System      │
└──────┬──────────┬────────┬───────────┘
       │          │        │
   ┌───▼──┐  ┌────▼───┐  ┌─▼──────┐
   │Entry │  │Fields  │  │ Level  │
   │      │  │        │  │        │
   │Format│  │Persist │  │Filter  │
   └───┬──┘  └────┬───┘  └─┬──────┘
       │          │        │
       └──────────┴────────┘
                  │
           ┌──────▼──────┐
           │    Hooks    │
           │             │
           ├─ File       │
           ├─ Syslog     │
           ├─ Stdout     │
           ├─ Stderr     │
           └─ Custom     │
           └─────────────┘
```

### Data Flow

```
Application Code
       │
       ▼
┌────────────────┐
│ Logger.Info()  │  Create log entry
└──────┬─────────┘
       │
       ▼
┌────────────────┐
│ Entry + Fields │  Merge persistent fields
└──────┬─────────┘
       │
       ▼
┌────────────────┐
│  Level Filter  │  Check minimum level
└──────┬─────────┘
       │
       ▼
┌────────────────┐
│ Formatter      │  JSON or Text
└──────┬─────────┘
       │
       ▼
┌────────────────┐
│    Hooks       │  Distribute to outputs
└──────┬─────────┘
       │
       ├─→ File
       ├─→ Syslog
       ├─→ Stdout
       └─→ Custom
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger
```

**Dependencies**:
- Go ≥ 1.24 (hookfile requires os.OpenRoot introduced in Go 1.24)
- github.com/sirupsen/logrus
- github.com/nabbar/golib/logger/config
- github.com/nabbar/golib/logger/level

### Basic Logging

```go
import (
    "github.com/nabbar/golib/logger"
    "github.com/nabbar/golib/logger/config"
    "github.com/nabbar/golib/logger/level"
)

// Create logger with default configuration
log, err := logger.New(context.Background)
if err != nil {
    panic(err)
}
defer log.Close()

// Set minimum level
log.SetLevel(level.InfoLevel)

// Log messages
log.Info("Application started", nil)
log.Debug("Debug message", map[string]interface{}{
    "key": "value",
})
log.Error("Error occurred", nil, fmt.Errorf("example error"))
```

### Configured Logger

```go
import (
    logcfg "github.com/nabbar/golib/logger/config"
    loglvl "github.com/nabbar/golib/logger/level"
)

// Create configuration
opts := &logcfg.Options{
    LogLevel:      loglvl.InfoLevel,
    LogFormatter:  logcfg.FormatJSON,
    EnableTrace:   true,
    EnableConsole: true,
    DisableStack:  false,
}

// Add file output
opts.LogFile = &logcfg.OptionsFile{
    LogFileName:      "/var/log/app.log",
    LogFileMaxSize:   100,  // MB
    LogFileMaxAge:    30,   // days
    LogFileMaxBackup: 10,   // files
    LogFileCompress:  true,
}

// Create logger
log, err := logger.New(context.Background)
if err != nil {
    panic(err)
}

if err := log.SetOptions(opts); err != nil {
    panic(err)
}
defer log.Close()
```

### With Persistent Fields

```go
import (
    "github.com/nabbar/golib/logger/fields"
)

// Create fields
flds := fields.New()
flds.Add("service", "api-gateway")
flds.Add("version", "1.2.3")
flds.Add("environment", "production")

// Set on logger
log.SetFields(flds)

// All log entries will include these fields
log.Info("Request processed", map[string]interface{}{
    "request_id": "abc-123",
    "duration_ms": 45,
})

// Output (JSON):
// {"level":"info","msg":"Request processed","service":"api-gateway",
//  "version":"1.2.3","environment":"production","request_id":"abc-123",
//  "duration_ms":45,"time":"2024-01-15T10:30:00Z"}
```

---

## Performance

### Benchmarks

Measured on: AMD Ryzen 9 5950X, 64GB RAM, Go 1.24+

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Logger Creation | 2.5 µs | 3.2 KB | 28 allocs |
| Info() call (no output) | 850 ns | 512 B | 6 allocs |
| Info() call (file) | 12 µs | 1.8 KB | 18 allocs |
| Info() call (JSON + fields) | 15 µs | 2.1 KB | 22 allocs |
| Field Add | 120 ns | 64 B | 1 alloc |
| Clone() | 3.2 µs | 3.5 KB | 30 allocs |

### Test Coverage

Latest test results (861 total specs):

| Package | Specs | Coverage | Status |
|---------|-------|----------|--------|
| **logger** | 81 | 74.3% | ✅ PASS |
| **config** | 125 | 85.3% | ✅ PASS |
| **entry** | 135 | 85.8% | ✅ PASS |
| **fields** | 114 | 95.7% | ✅ PASS |
| **gorm** | 34 | 100.0% | ✅ PASS |
| **hashicorp** | 89 | 96.6% | ✅ PASS |
| **hookfile** | 25 | 82.2% | ✅ PASS |
| **hookstderr** | 30 | 100.0% | ✅ PASS |
| **hookstdout** | 30 | 100.0% | ✅ PASS |
| **hooksyslog** | 41 | 83.2% | ✅ PASS |
| **hookwriter** | 31 | 90.2% | ✅ PASS |
| **level** | 94 | 98.0% | ✅ PASS |
| **types** | 32 | N/A | ✅ PASS |
| **TOTAL** | **861** | **90.9%** | ✅ **ALL PASS** |

### Memory Profile

- **Per Logger Instance**: ~3KB base overhead
- **Per Log Entry**: Amortized O(1), ~512B without outputs
- **With JSON Formatting**: +600B per entry
- **With 10 Fields**: +800B per entry

### Concurrency

- **Write Operations**: Thread-safe with mutex protection
- **Read Operations**: Lock-free reads where possible
- **File Hooks**: Buffered writes for performance
- **Race Detection**: Clean (no data races)

---

## Subpackages

### config

**Purpose**: Configuration management for logger options and validation.

**Key Features**:
- Complete logger configuration structure
- JSON/YAML/TOML serialization support
- File rotation settings (size, age, backup count)
- Syslog configuration (network, host, level)
- Format enumeration (JSON/Text)
- Validation logic

**Use Case**: Application configuration, dynamic reconfiguration, config file parsing

**Documentation**: [config/README.md](config/README.md)

---

### entry

**Purpose**: Log entry creation, formatting, and lifecycle management.

**Key Features**:
- Entry creation with context
- Field merging (persistent + per-entry)
- Level association
- Timestamp management
- Formatting for output
- Stack trace capture

**Use Case**: Structured log entry building, custom formatters, log aggregation

**Documentation**: [entry/README.md](entry/README.md)

---

### fields

**Purpose**: Custom field management and structured data injection.

**Key Features**:
- Thread-safe field operations
- Key-value storage with type preservation
- Field merging and cloning
- logrus.Fields conversion
- Add/Get/Delete/List operations

**Use Case**: Contextual logging, request tracking, application metadata

**Documentation**: [fields/README.md](fields/README.md)

---

### level

**Purpose**: Log level enumeration, parsing, and filtering.

**Key Features**:
- 8 log levels (Panic, Fatal, Error, Warning, Info, Debug, Trace, Null)
- String parsing and validation
- Comparison operations
- logrus level conversion
- Level-based filtering

**Use Case**: Log verbosity control, environment-based filtering, dynamic level changes

**Documentation**: [level/README.md](level/README.md)

---

### gorm

**Purpose**: GORM ORM integration adapter.

**Key Features**:
- Query logging with duration tracking
- Slow query detection (configurable threshold)
- Error logging
- Record count tracking
- Compatible with GORM v2 logger interface

**Performance**:
- Query logging overhead: <100µs
- No impact on query execution
- 100% test coverage

**Use Case**: Database query monitoring, slow query analysis, ORM debugging

**Documentation**: [gorm/README.md](gorm/README.md)

---

### hashicorp

**Purpose**: Hashicorp tools integration (Vault, Consul, Nomad, Terraform).

**Key Features**:
- hclog adapter implementation
- Level mapping (hclog ↔ logger levels)
- Structured logging support
- Context propagation
- Named logger support

**Performance**:
- Adapter overhead: <50µs
- 96.6% test coverage

**Use Case**: Vault client logging, Consul integration, Terraform provider logs

**Documentation**: [hashicorp/README.md](hashicorp/README.md)

---

### hookfile

**Purpose**: File output hook with rotation support.

**Key Features**:
- Size-based rotation (MB threshold)
- Age-based rotation (days threshold)
- Backup file management (count limit)
- Compression (gzip)
- Buffered writes

**Performance**:
- Write buffering reduces I/O calls
- Compression saves disk space
- Rotation overhead: <10ms

**Use Case**: Production log files, application logs, audit trails

**Documentation**: [hookfile/README.md](hookfile/README.md)

---

### hooksyslog

**Purpose**: Syslog protocol output (RFC 5424).

**Key Features**:
- TCP/UDP transport
- Local and network syslog
- Priority mapping
- Tag customization
- Facility and severity codes

**Performance**:
- Network latency dependent
- Async write support

**Use Case**: Centralized logging, syslog servers, system integration

**Documentation**: [hooksyslog/README.md](hooksyslog/README.md)

---

### hookstdout / hookstderr

**Purpose**: Standard output/error stream hooks.

**Key Features**:
- Console output (stdout/stderr)
- Color support (if TTY detected)
- Human-readable formatting
- Development-friendly output

**Performance**:
- Direct write (no buffering)
- Minimal overhead
- 100% test coverage (both packages)

**Use Case**: Development logging, console applications, CLI tools

**Documentation**: [hookstdout/README.md](hookstdout/README.md), [hookstderr/README.md](hookstderr/README.md)

---

### hookwriter

**Purpose**: Custom io.Writer integration hook.

**Key Features**:
- Adapt any io.Writer as log output
- Level filtering per writer
- Buffering support
- Error handling

**Performance**:
- Overhead depends on underlying writer
- 90.2% test coverage

**Use Case**: Custom outputs, network streams, database logging, message queues

**Documentation**: [hookwriter/README.md](hookwriter/README.md)

---

### types

**Purpose**: Common types and structures used across logger packages.

**Key Features**:
- Logger type definitions
- Configuration structures
- Compatibility types (GORM, Hashicorp)
- Interface definitions

**Use Case**: Type safety, interface compliance, API contracts

**Documentation**: [types/README.md](types/README.md)

---

## Use Cases

### 1. Web Application Logging

```go
// HTTP middleware logging
func LoggingMiddleware(log logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Clone logger for this request
            reqLog, _ := log.Clone()
            reqLog.SetFields(fields.NewFromMap(map[string]interface{}{
                "request_id": generateRequestID(),
                "method":     r.Method,
                "path":       r.URL.Path,
                "remote_ip":  r.RemoteAddr,
            }))
            
            // Wrap response writer
            wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
            
            next.ServeHTTP(wrapped, r)
            
            // Log request completion
            reqLog.Info("Request completed", map[string]interface{}{
                "status":      wrapped.statusCode,
                "duration_ms": time.Since(start).Milliseconds(),
            })
        })
    }
}
```

### 2. Database Query Logging

```go
import (
    "github.com/nabbar/golib/logger/gorm"
    "gorm.io/gorm"
)

// Integrate with GORM
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: gorm.New(log, gorm.Config{
        SlowThreshold: 200 * time.Millisecond,
        LogLevel:      gorm.Info,
    }),
})

// Queries are automatically logged
db.Find(&users) // Logs query, duration, rows affected
```

### 3. Background Job Logging

```go
func ProcessJob(log logger.Logger, job Job) error {
    // Clone logger with job context
    jobLog, _ := log.Clone()
    jobLog.SetFields(fields.NewFromMap(map[string]interface{}{
        "job_id":   job.ID,
        "job_type": job.Type,
    }))
    
    jobLog.Info("Job started", nil)
    
    if err := job.Execute(); err != nil {
        jobLog.Error("Job failed", nil, err)
        return err
    }
    
    jobLog.Info("Job completed", map[string]interface{}{
        "duration": job.Duration(),
    })
    return nil
}
```

### 4. Microservice Distributed Tracing

```go
// Add trace ID to logger
func WithTraceID(log logger.Logger, traceID string) logger.Logger {
    clone, _ := log.Clone()
    flds := clone.GetFields()
    flds.Add("trace_id", traceID)
    flds.Add("span_id", generateSpanID())
    clone.SetFields(flds)
    return clone
}

// Use in service calls
func HandleRequest(ctx context.Context, log logger.Logger) {
    traceID := ctx.Value("trace_id").(string)
    log = WithTraceID(log, traceID)
    
    log.Info("Service called", nil)
    // Trace ID included in all logs
}
```

### 5. Multi-Output Logging

```go
opts := &logcfg.Options{
    EnableConsole: true,  // Development visibility
}

// Production file logging
opts.LogFile = &logcfg.OptionsFile{
    LogFileName: "/var/log/app.log",
    LogFileMaxSize: 100,
}

// Critical errors to syslog
opts.LogSyslog = &logcfg.OptionsSyslog{
    LogSyslogNetwork: "tcp",
    LogSyslogHost:    "syslog.example.com:514",
    LogSyslogLevel:   loglvl.ErrorLevel,
}

log.SetOptions(opts)

// Logs go to console, file, AND syslog (if level >= Error)
log.Error("Critical error", nil, err)
```

---

## Best Practices

### 1. Use Structured Logging

```go
// DON'T: String formatting
log.Info(fmt.Sprintf("User %s logged in from %s", user, ip), nil)

// DO: Structured fields
log.Info("User logged in", map[string]interface{}{
    "user_id": user.ID,
    "username": user.Name,
    "ip_address": ip,
})
```

### 2. Clone for Context

```go
// DON'T: Modify shared logger
log.SetFields(requestFields)
processRequest()
log.SetFields(originalFields) // Error-prone

// DO: Clone for isolated context
reqLog, _ := log.Clone()
reqLog.SetFields(requestFields)
processRequest(reqLog)
```

### 3. Use Appropriate Levels

```go
// Trace: Very detailed (function entry/exit)
log.Trace("Entering function", map[string]interface{}{"args": args})

// Debug: Diagnostic information
log.Debug("Cache miss", map[string]interface{}{"key": key})

// Info: General information
log.Info("Request processed", map[string]interface{}{"duration_ms": 45})

// Warning: Unexpected but handled
log.Warning("Retry attempt", map[string]interface{}{"attempt": 2})

// Error: Error conditions
log.Error("Database query failed", nil, err)

// Fatal: Unrecoverable (exits program)
log.Fatal("Failed to start server", nil, err)

// Panic: Programming errors (panics)
log.Panic("Nil pointer", nil, err)
```

### 4. Cleanup Resources

```go
// DO: Always close logger
log, err := logger.New(ctx)
if err != nil {
    return err
}
defer log.Close()  // Flushes buffers, closes files
```

### 5. Configure Early

```go
// DO: Configure before logging
log, _ := logger.New(ctx)
log.SetOptions(opts)
log.SetLevel(level.InfoLevel)
log.SetFields(appFields)

// Now start logging
log.Info("Application started", nil)
```

---

## API Reference

### Logger Interface

```go
type Logger interface {
    io.WriteCloser
    
    // Level management
    SetLevel(lvl level.Level)
    GetLevel() level.Level
    SetIOWriterLevel(lvl level.Level)
    GetIOWriterLevel() level.Level
    
    // Configuration
    SetOptions(opt *config.Options) error
    GetOptions() *config.Options
    
    // Fields
    SetFields(field fields.Fields)
    GetFields() fields.Fields
    
    // Cloning
    Clone() (Logger, error)
    
    // Integrations
    GetStdLogger(lvl level.Level, flags int) *log.Logger
    SetStdLogger(lvl level.Level, flags int)
    SetSPF13Level(lvl level.Level, log *jww.Notepad)
    
    // Logging methods
    Debug(message string, data interface{}, args ...interface{})
    Info(message string, data interface{}, args ...interface{})
    Warning(message string, data interface{}, args ...interface{})
    Error(message string, data interface{}, args ...interface{})
    Fatal(message string, data interface{}, args ...interface{})
    Panic(message string, data interface{}, args ...interface{})
    Trace(message string, data interface{}, args ...interface{})
    Log(level level.Level, message string, data interface{}, args ...interface{})
    
    // Entry-based logging
    Entry(lvl level.Level, data interface{}, args ...interface{}) entry.Entry
    CheckIn(ent entry.Entry)
    CheckOut(ent entry.Entry)
}
```

### Configuration

**Options Structure**:
```go
type Options struct {
    LogLevel      level.Level    // Minimum log level
    LogFormatter  Format         // JSON or Text
    EnableConsole bool           // Console output
    EnableTrace   bool           // Source location tracking
    DisableStack  bool           // Disable stack traces
    LogFile       *OptionsFile   // File configuration
    LogSyslog     *OptionsSyslog // Syslog configuration
}
```

**Sub-packages**:
- `config`: Configuration management
- `entry`: Log entry handling
- `fields`: Custom field injection
- `level`: Log level enumeration
- `gorm`: GORM integration
- `hashicorp`: Hashicorp tools adapter

### Log Levels

**Level Hierarchy** (highest to lowest severity):
- `PanicLevel` (0): Calls panic() after logging
- `FatalLevel` (1): Calls os.Exit(1) after logging
- `ErrorLevel` (2): Error conditions
- `WarnLevel` (3): Warning conditions
- `InfoLevel` (4): Informational messages (default)
- `DebugLevel` (5): Debug information
- `TraceLevel` (6): Very verbose tracing
- `NullLevel` (7): Disables logging

### Error Handling

```go
// CheckError: Conditional logging based on error presence
hasError := log.CheckError(
    level.ErrorLevel,  // Log level if error
    level.InfoLevel,   // Log level if no error
    "Operation result",
    err,
)

// LogDetails: Advanced logging with multiple errors
log.LogDetails(
    level.ErrorLevel,
    "Complex operation failed",
    data,
    []error{err1, err2},
    fields,
    args...,
)
```

---


## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: >85%)
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
   - Maintain coverage above 85%

4. **Documentation**
   - Update GoDoc comments for public APIs
   - Add examples for new features
   - Update README.md if adding subpackages
   - Update TESTING.md if changing test structure

5. **Pull Request Process**
   - Fork the repository
   - Create a feature branch
   - Write clear commit messages
   - Ensure all tests pass
   - Update documentation
   - Submit PR with description of changes

---

## Improvements & Security

**Planned Improvements**:
- Structured query language for programmatic log querying
- Log sampling for high-volume scenarios
- Enhanced context integration with distributed trace IDs
- Prometheus metrics for log rate monitoring
- Direct integration with log aggregators (Elasticsearch, Loki, Grafana)

**Security Considerations**:
- All log outputs are protected by file permissions
- Sensitive data should be filtered before logging
- File rotation prevents disk exhaustion
- Syslog connections support TLS for secure transmission

**Reporting Security Issues**:
Please report security vulnerabilities privately via GitHub Security Advisories or by contacting the maintainer directly.

---

## Resources

### Internal Documentation
- [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger) - Complete API documentation
- [TESTING.md](TESTING.md) - Comprehensive testing guide
- Individual subpackage READMEs (linked in [Subpackages](#subpackages))

### Related Packages
- [github.com/nabbar/golib/context](../context) - Context management
- [github.com/nabbar/golib/ioutils](../ioutils) - I/O utilities
- [github.com/nabbar/golib/errors](../errors) - Error handling

### External References
- [Logrus](https://github.com/sirupsen/logrus) - Underlying logging library
- [GORM](https://gorm.io) - GORM integration
- [Hashicorp hclog](https://github.com/hashicorp/go-hclog) - Hashicorp adapter
- [GitHub Issues](https://github.com/nabbar/golib/issues) - Bug reports and feature requests

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
