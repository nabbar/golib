# Logger Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Production-ready structured logging system for Go applications with flexible output management, field injection, level-based filtering, and extensive integration capabilities.

> **AI Disclaimer**: AI tools are used solely to assist with testing, documentation, and bug fixes under human supervision, in compliance with EU AI Act Article 50.4.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Subpackages](#subpackages)
- [Configuration](#configuration)
- [Log Levels](#log-levels)
- [Fields Management](#fields-management)
- [Hooks](#hooks)
- [Integrations](#integrations)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
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

### What Problems Does It Solve?

- **Centralized Logging**: Unified logging interface across your application
- **Log Aggregation**: Send logs to multiple destinations simultaneously
- **Structured Data**: Add contextual information to every log entry
- **Level Filtering**: Control verbosity per output destination
- **Rotation**: Automatic log file rotation based on size/age
- **Integration**: Adapt third-party library logging to your system

---

## Key Features

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

## Installation

```bash
go get github.com/nabbar/golib/logger
```

**Dependencies**:
- Go ≥ 1.18
- github.com/sirupsen/logrus
- github.com/nabbar/golib/context
- github.com/nabbar/golib/ioutils

---

## Architecture

### Package Structure

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
┌────────────────────────────────────────────┐
│            Logger Package                   │
│       Structured Logging System             │
└──────┬──────────┬────────┬─────────────────┘
       │          │        │
   ┌───▼──┐  ┌────▼───┐ ┌─▼──────┐
   │Entry │  │Fields  │ │ Level  │
   │      │  │        │ │        │
   │Format│  │Persist │ │Filter  │
   └───┬──┘  └────┬───┘ └─┬──────┘
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
┌──────────────┐
│ Logger.Info()│  Create log entry
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Entry + Fields│  Merge persistent fields
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Level Filter │  Check minimum level
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Formatter    │  JSON or Text
└──────┬───────┘
       │
       ▼
┌──────────────┐
│    Hooks     │  Distribute to outputs
└──────┬───────┘
       │
       ├─→ File
       ├─→ Syslog
       ├─→ Stdout
       └─→ Custom
```

---

## Quick Start

### Basic Logger

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

Measured on: AMD Ryzen 9 5950X, 64GB RAM, Go 1.21

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Logger Creation | 2.5 µs | 3.2 KB | 28 allocs |
| Info() call (no output) | 850 ns | 512 B | 6 allocs |
| Info() call (file) | 12 µs | 1.8 KB | 18 allocs |
| Info() call (JSON + fields) | 15 µs | 2.1 KB | 22 allocs |
| Field Add | 120 ns | 64 B | 1 alloc |
| Clone() | 3.2 µs | 3.5 KB | 30 allocs |

### Test Coverage

Latest test results (705 total specs):

| Package | Specs | Coverage | Status |
|---------|-------|----------|--------|
| **logger** | 81 | 75.0% | ✅ PASS |
| **config** | 127 | 85.3% | ✅ PASS |
| **entry** | 119 | 85.1% | ✅ PASS |
| **fields** | 49 | 78.4% | ✅ PASS |
| **gorm** | 34 | 100.0% | ✅ PASS |
| **hashicorp** | 89 | 96.6% | ✅ PASS |
| **hookfile** | 22 | 20.1% | ✅ PASS |
| **hookstderr** | 30 | 100.0% | ✅ PASS |
| **hookstdout** | 30 | 100.0% | ✅ PASS |
| **hooksyslog** | 20 | 53.5% | ✅ PASS |
| **hookwriter** | 31 | 90.2% | ✅ PASS |
| **level** | 42 | 65.9% | ✅ PASS |
| **types** | 32 | N/A | ✅ PASS |
| **TOTAL** | **705** | **~77%** | ✅ **ALL PASS** |

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

## Subpackages

### config

**Purpose**: Configuration management for logger options

**Key Types**:
- `Options`: Complete logger configuration
- `OptionsFile`: File output configuration
- `OptionsSyslog`: Syslog output configuration
- `Format`: Output format enumeration (JSON/Text)

**Features**:
- JSON/YAML/TOML serialization
- Validation
- Default configuration templates
- File rotation settings

### entry

**Purpose**: Log entry creation and management

**Key Types**:
- `Entry`: Individual log entry
- Interface for entry manipulation

**Features**:
- Field merging
- Level association
- Timestamp management
- Formatting

### fields

**Purpose**: Custom field management and injection

**Key Types**:
- `Fields`: Field container
- Thread-safe field operations

**Features**:
- Key-value storage
- Field merging
- Clone support
- logrus.Fields conversion

### level

**Purpose**: Log level enumeration and filtering

**Key Types**:
- `Level`: Log level enumeration
- 8 levels: Panic, Fatal, Error, Warning, Info, Debug, Trace, Null

**Features**:
- String parsing
- Comparison operations
- logrus level conversion
- Validation

### hookfile

**Purpose**: File output with rotation

**Features**:
- Size-based rotation
- Age-based rotation
- Compression
- Backup management
- Buffered writes

### hooksyslog

**Purpose**: Syslog protocol output

**Features**:
- TCP/UDP transport
- RFC 5424 format
- Priority mapping
- Network and local syslog

### hookstdout / hookstderr

**Purpose**: Standard output/error streams

**Features**:
- Console logging
- Color support (if TTY)
- Development mode

### hookwriter

**Purpose**: Custom io.Writer integration

**Features**:
- Adapt any io.Writer
- Level filtering
- Buffering

### gorm

**Purpose**: GORM ORM integration

**Features**:
- Query logging
- Slow query detection
- Error logging
- Record count tracking

### hashicorp

**Purpose**: Hashicorp tools integration (Vault, Consul, etc.)

**Features**:
- hclog adapter
- Level mapping
- Structured logging

---

## Configuration

### Options Structure

```go
type Options struct {
    // Log level (Panic, Fatal, Error, Warning, Info, Debug, Trace)
    LogLevel level.Level
    
    // Output format (JSON or Text)
    LogFormatter Format
    
    // Enable console output (stdout/stderr)
    EnableConsole bool
    
    // Enable source location tracking
    EnableTrace bool
    
    // Disable stack trace on errors
    DisableStack bool
    
    // File output configuration
    LogFile *OptionsFile
    
    // Syslog output configuration
    LogSyslog *OptionsSyslog
}
```

### File Configuration

```go
type OptionsFile struct {
    // Log file path
    LogFileName string
    
    // Maximum size in MB before rotation
    LogFileMaxSize int64
    
    // Maximum age in days
    LogFileMaxAge int64
    
    // Maximum number of backup files
    LogFileMaxBackup int64
    
    // Compress rotated files
    LogFileCompress bool
}
```

### Syslog Configuration

```go
type OptionsSyslog struct {
    // Network type: "tcp", "udp", "unix", or "" for local
    LogSyslogNetwork string
    
    // Syslog server address
    LogSyslogHost string
    
    // Minimum level for syslog
    LogSyslogLevel level.Level
    
    // Syslog tag/application name
    LogSyslogTag string
}
```

---

## Log Levels

### Level Hierarchy

```
Panic   (0) - Highest severity, calls panic() after logging
Fatal   (1) - Logs then exits with os.Exit(1)
Error   (2) - Error conditions
Warning (3) - Warning conditions
Info    (4) - Informational messages
Debug   (5) - Debug-level messages
Trace   (6) - Trace-level messages (very verbose)
Null    (7) - Disable logging
```

### Usage

```go
// Set minimum level
log.SetLevel(level.InfoLevel)

// Only Info, Warning, Error, Fatal, Panic will be logged
log.Trace("Not logged")
log.Debug("Not logged")
log.Info("Logged!")
log.Warning("Logged!")
log.Error("Logged!", nil, err)
```

### Per-Output Levels

```go
// Console: Debug and above
log.SetLevel(level.DebugLevel)

// File: Info and above (set in Options)
opts.LogFile = &OptionsFile{
    // Implicitly uses logger's level
}

// Syslog: Only errors
opts.LogSyslog = &OptionsSyslog{
    LogSyslogLevel: level.ErrorLevel,  // Override
}
```

---

## Fields Management

### Creating Fields

```go
import "github.com/nabbar/golib/logger/fields"

// Empty fields
flds := fields.New()

// From map
flds := fields.NewFromMap(map[string]interface{}{
    "service": "api",
    "version": "1.0.0",
})

// Add fields
flds.Add("key", "value")
flds.Add("count", 42)
flds.Add("enabled", true)
```

### Field Operations

```go
// Get value
val := flds.Get("key")

// Check existence
exists := flds.Exists("key")

// Delete field
flds.Del("key")

// List keys
keys := flds.List()

// Merge fields
other := fields.NewFromMap(map[string]interface{}{"new": "field"})
merged := flds.Merge(other)

// Clone
clone := flds.Clone(nil)
```

### Logger Fields

```go
// Set persistent fields
log.SetFields(flds)

// All logs include these fields
log.Info("Message", nil)  // Includes flds

// Get current fields
current := log.GetFields()

// Per-entry fields
log.Info("Message", map[string]interface{}{
    "request_id": "123",  // Merged with persistent fields
})
```

---

## Hooks

### File Hook

```go
opts.LogFile = &logcfg.OptionsFile{
    LogFileName:      "/var/log/app.log",
    LogFileMaxSize:   100,  // MB
    LogFileMaxAge:    30,   // days
    LogFileMaxBackup: 10,
    LogFileCompress:  true,
}
```

**Features**:
- Automatic rotation when size limit reached
- Age-based cleanup
- Gzip compression
- Backup file management

### Syslog Hook

```go
opts.LogSyslog = &logcfg.OptionsSyslog{
    LogSyslogNetwork: "tcp",
    LogSyslogHost:    "syslog.example.com:514",
    LogSyslogLevel:   level.WarnLevel,
    LogSyslogTag:     "myapp",
}
```

**Features**:
- RFC 5424 format
- TCP/UDP transport
- Priority mapping
- Tag customization

### Console Hook

```go
opts.EnableConsole = true
```

**Features**:
- Stdout/stderr output
- Color support (TTY)
- Human-readable format

### Custom Hook

```go
import "github.com/nabbar/golib/logger/hookwriter"

// Any io.Writer
customWriter := &MyWriter{}
hook := hookwriter.New(customWriter)

// Add to logger (via logrus integration)
```

---

## Integrations

### GORM

```go
import (
    loggorm "github.com/nabbar/golib/logger/gorm"
    "gorm.io/gorm"
)

// Create GORM logger
gormLogger := loggorm.New(log, loggorm.Config{
    SlowThreshold: 200 * time.Millisecond,
    LogLevel:      loggorm.Info,
    Colorful:      false,
})

// Use with GORM
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: gormLogger,
})

// Queries automatically logged
db.Where("age > ?", 18).Find(&users)
// Output: [info] Query executed in 45ms, 10 rows affected
```

### Hashicorp Tools

```go
import loghc "github.com/nabbar/golib/logger/hashicorp"

// Create Hashicorp logger
hcLogger := loghc.New(log, "myapp")

// Use with Vault, Consul, etc.
client, err := vault.NewClient(&vault.Config{
    Logger: hcLogger,
})
```

### spf13/cobra

```go
import "github.com/spf13/jwalterweatherman"

// Link spf13 logger to main logger
notepad := &jww.Notepad{}
log.SetSPF13Level(level.InfoLevel, notepad)

// spf13 logs now go through main logger
```

### Standard Library log

```go
// Get stdlib-compatible logger
stdLog := log.GetStdLogger(level.InfoLevel, log.LstdFlags)

// Use anywhere stdlib log is needed
http.Server{
    ErrorLog: stdLog,
}

// Or set as default
log.SetStdLogger(level.InfoLevel, log.LstdFlags)
log.Println("Uses main logger")
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

---

## Testing

See [TESTING.md](./TESTING.md) for comprehensive testing documentation.

**Test Statistics**:
- **Total Specs**: 705
- **Average Coverage**: ~77%
- **All Tests**: ✅ PASSING
- **Race Detection**: ✅ CLEAN

**Coverage Highlights**:
- Perfect (100%): gorm, hookstderr, hookstdout
- Excellent (>90%): hashicorp (96.6%), hookwriter (90.2%)
- Good (75-85%): config (85.3%), entry (85.1%), fields (78.4%), logger (75.0%)

**Quick Test**:
```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection
CGO_ENABLED=1 go test -race ./...

# Detailed results
go test -v -cover ./...
```

---

## Contributing

Contributions are welcome! Please follow these guidelines:

### Code Standards
- Write tests for new features
- Update documentation
- Add GoDoc comments for public APIs
- Run `go fmt` and `go vet`
- Test with race detector (`-race`)

### AI Usage Policy
- **DO NOT** use AI tools to generate package code or core logic
- **DO** use AI to assist with:
  - Writing and improving tests
  - Documentation and comments
  - Debugging and bug fixes

All AI-assisted work must be reviewed and validated by a human maintainer.

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Write tests
4. Update documentation
5. Run full test suite with race detection
6. Submit PR with clear description

---

## Future Enhancements

Potential improvements under consideration:

- **Structured Query Language**: Query logs programmatically
- **Log Sampling**: Sample high-volume logs
- **Context Integration**: Context-aware logging with trace IDs
- **Metrics Export**: Prometheus metrics for log rates
- **Remote Backends**: Direct integration with log aggregators (Elasticsearch, Loki)
- **Performance Profiling**: Built-in performance profiling hooks
- **Log Encryption**: Encrypted log output for sensitive data

Contributions and suggestions are welcome!

---

## License

MIT License - Copyright (c) 2021 Nicolas JUHEL

See [LICENSE](../LICENSE) for full details.

---

## Resources

- **GoDoc**: [pkg.go.dev/github.com/nabbar/golib/logger](https://pkg.go.dev/github.com/nabbar/golib/logger)
- **Logrus**: [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus)
- **Issues**: [github.com/nabbar/golib/issues](https://github.com/nabbar/golib/issues)

**Related Packages**:
- [github.com/nabbar/golib/context](https://github.com/nabbar/golib/context)
- [github.com/nabbar/golib/ioutils](https://github.com/nabbar/golib/ioutils)
