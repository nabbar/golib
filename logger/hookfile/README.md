# Logger HookFile

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-84.0%25-brightgreen)](TESTING.md)

Logrus hook for writing log entries to files with automatic rotation detection, efficient multi-writer aggregation, and configurable field filtering.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Logrus Hook Behavior](#logrus-hook-behavior)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Example](#basic-example)
  - [Production Setup](#production-setup)
  - [Access Log Mode](#access-log-mode)
  - [Level Filtering](#level-filtering)
  - [Field Filtering](#field-filtering)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [HookFile Interface](#hookfile-interface)
  - [Configuration](#configuration)
  - [Rotation Detection](#rotation-detection)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **hookfile** package provides a production-ready logrus.Hook that writes log entries to files with sophisticated features not found in standard file logging. It automatically detects log rotation, efficiently aggregates writes from multiple loggers, and provides fine-grained control over log formatting.

### Design Philosophy

1. **Rotation-Aware**: Automatically detect and handle external log rotation using inode tracking
2. **Resource Efficient**: Share file handles and aggregators across multiple hooks
3. **Production-Ready**: Handle edge cases like file deletion, permission errors, disk full
4. **Zero-Copy Writes**: Use aggregator pattern to minimize memory allocations
5. **Fail-Safe Operation**: Continue logging even when rotation fails

### Key Features

- ✅ **Automatic Rotation Detection**: Detects when log files are moved/renamed (inode tracking)
- ✅ **File Handle Sharing**: Multiple hooks to same file share single aggregator and file handle
- ✅ **Buffered Aggregation**: Uses ioutils/aggregator for efficient async writes
- ✅ **Reference Counting**: Automatically closes files when last hook is removed
- ✅ **Permission Management**: Configurable file and directory permissions
- ✅ **Field Filtering**: Remove stack traces, timestamps, caller info as needed
- ✅ **Access Log Mode**: Message-only output for HTTP access logs
- ✅ **Error Recovery**: Automatic file reopening on errors
- ✅ **84.0% Test Coverage**: 28 specs + 10 examples, zero race conditions

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────┐
│           Multiple logrus.Logger            │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐      │
│  │Logger 1 │  │Logger 2 │  │Logger 3 │      │
│  └────┬────┘  └────┬────┘  └────┬────┘      │
│       │            │            │           │
└───────┼────────────┼────────────┼───────────┘
        │            │            │
        ▼            ▼            ▼
    ┌────────────────────────────────┐
    │     HookFile Instances         │
    │   (3 hooks, same filepath)     │
    └────────────┬───────────────────┘
                 │
                 ▼
        ┌───────────────────┐
        │   File Aggregator │
        │   (RefCount: 3)   │
        │                   │
        │  • Shared File    │
        │  • Sync Timer     │
        │  • Rotation Check │
        └────────┬──────────┘
                 │
                 ▼
          ┌──────────────┐
          │  Aggregator  │
          │  (buffered)  │
          └──────┬───────┘
                 │
                 ▼
           ┌──────────┐
           │ app.log  │
           └──────────┘
```

### Data Flow

**File Aggregation:**
1. Multiple hooks created for same file path
2. First hook creates file aggregator with refcount=1
3. Subsequent hooks increment refcount (reuse aggregator)
4. Each hook.Close() decrements refcount
5. When refcount reaches 0, file and aggregator are closed

**Rotation Detection:**
```
Time T0: app.log (inode: 12345)
         Hook writes → file descriptor points to inode 12345

Time T1: logrotate renames app.log to app.log.1
         Creates new app.log (inode: 67890)
         Hook still writes → FD points to OLD inode 12345

Time T2: Sync timer runs (every 1 second)
         Compare: FD inode (12345) ≠ Disk inode (67890)
         Rotation detected!
         Close old FD → Open new file → Resume logging

Time T3: Hook writes → file descriptor points to NEW inode 67890
```

### Logrus Hook Behavior

**⚠️ CRITICAL**: Understanding how logrus hooks process log data:

**Standard Mode (Default)**:
- ✅ **Fields (logrus.Fields) ARE written** to output
- ❌ **Message parameter in Info/Error/etc. is IGNORED** by formatter
- To log a message: use `logger.WithField("msg", "text").Info("")`

**Access Log Mode (EnableAccessLog=true)**:
- ❌ **Fields (logrus.Fields) are IGNORED**
- ✅ **Message parameter IS written** to output
- To log a message: use `logger.Info("GET /api - 200 OK")`

**Example of Standard Mode:**
```go
// ❌ WRONG: Message will NOT appear in logs
logger.Info("User logged in")  // Output: (empty)

// ✅ CORRECT: Use fields
logger.WithField("msg", "User logged in").Info("ignored")
// Output: level=info fields.msg="User logged in"
```

**Example of Access Log Mode:**
```go
// ✅ CORRECT in AccessLog mode
logger.Info("GET /api/users - 200 OK - 45ms")
// Output: GET /api/users - 200 OK - 45ms

// ❌ WRONG in AccessLog mode: Fields are ignored
logger.WithField("status", 200).Info("")  // Output: (empty)
```

---

## Performance

### Benchmarks

Based on test results with gmeasure:

| Metric | Value | Notes |
|--------|-------|-------|
| **Write Latency (Median)** | 106ms | Includes formatting + buffer |
| **Write Latency (Mean)** | 119ms | Average under normal load |
| **Write Latency (P99)** | 169ms | 99th percentile |
| **Memory Usage** | ~280KB | Per file aggregator |
| **Throughput** | ~5000-10000 entries/sec | Depends on formatter |
| **Rotation Detection** | 1s | Sync timer interval |
| **File Reopen** | 1-5ms | During rotation |

### Memory Usage

```
Hook struct:        ~120 bytes (minimal footprint)
File aggregator:    ~280 KB (includes buffers)
Per operation:      0 allocations (zero-copy delegation)
Reference counting: ~16 bytes per hook
Total per file:     ~280 KB (shared across hooks)
```

**Memory characteristics:**
- File handles shared across multiple hooks (reference counted)
- Aggregator reuses buffers to minimize GC pressure
- No heap allocations during normal operation
- Suitable for high-volume applications (thousands of concurrent hooks)

### Scalability

- ✅ **Concurrent Writers**: Multiple goroutines can log safely
- ✅ **File Sharing**: Multiple hooks efficiently share single file
- ✅ **Reference Counting**: Automatic resource cleanup
- ✅ **Thread-Safe**: All operations safe for concurrent use
- ✅ **Zero Race Conditions**: Tested with `-race` detector
- ✅ **Rotation Resilience**: Continues logging during rotation

**Tested Scalability:**
- 100+ concurrent goroutines writing to same file
- 10+ loggers sharing single file
- Millions of log entries without memory leaks
- Sub-second rotation detection and recovery

---

## Use Cases

### 1. Production Application with Log Rotation

**Problem**: Application needs persistent file logging with external log rotation (logrotate).

**Solution**: Use HookFile with CreatePath=true for automatic rotation detection.

**Advantages**:
- Automatic rotation detection via inode comparison
- No application restart required after rotation
- Continues logging to new file after rotation
- Compatible with all rotation tools (logrotate, etc.)

**Suited for**: Production servers, long-running daemons, system services, containerized apps with volume mounts.

### 2. Multi-Logger Applications

**Problem**: Multiple loggers in same application need to write to shared log file.

**Solution**: Create multiple HookFile instances for same file path.

**Advantages**:
- Single file descriptor shared across all hooks
- Reference counting prevents premature file closure
- Thread-safe concurrent writes via aggregator
- No file handle exhaustion

**Suited for**: Microservices, multi-tenant applications, plugin architectures, distributed logging.

### 3. Separate Access and Application Logs

**Problem**: HTTP access logs need different format and file from application logs.

**Solution**: Use two hooks - one in AccessLog mode, one in standard mode.

**Advantages**:
- Clean access log format (message-only)
- Structured application logs (JSON/fields)
- Independent rotation policies
- Easy parsing with standard tools

**Suited for**: Web servers, API gateways, reverse proxies, HTTP middleware.

### 4. Level-Specific Log Files

**Problem**: Different log levels need separate files (errors separate from info).

**Solution**: Create multiple hooks with different LogLevel configurations.

**Advantages**:
- Errors written to separate file for easy monitoring
- Debug logs isolated from production logs
- Independent retention policies per level
- Efficient filtering at hook level

**Suited for**: Debugging environments, error monitoring, compliance logging, audit trails.

### 5. Structured Logging for Log Aggregation

**Problem**: Logs need structured format for ELK, Splunk, CloudWatch, etc.

**Solution**: Use HookFile with JSON formatter for machine-readable logs.

**Advantages**:
- Structured JSON for easy parsing
- Compatible with log aggregation tools
- Field filtering for sensitive data
- Automatic rotation for log shippers

**Suited for**: Cloud-native apps, microservices, observability platforms, SIEM integration.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger/hookfile
```

**Requirements:**
- Go 1.24 or higher (requires os.OpenRoot)
- Compatible with Linux, macOS, Windows

### Basic Example

Write logs to a file with automatic rotation detection:

```go
package main

import (
    "github.com/sirupsen/logrus"
    "github.com/nabbar/golib/logger/config"
    "github.com/nabbar/golib/logger/hookfile"
)

func main() {
    // Configure file hook
    opts := config.OptionsFile{
        Filepath:   "/var/log/myapp/app.log",
        FileMode:   0644,
        PathMode:   0755,
        CreatePath: true,  // Enable rotation detection
        Create:     true,  // Enable file creation after rotation
    }

    // Create hook
    hook, err := hookfile.New(opts, &logrus.TextFormatter{})
    if err != nil {
        panic(err)
    }
    defer hook.Close()

    // Configure logger
    logger := logrus.New()
    logger.AddHook(hook)

    // IMPORTANT: Message parameter "ignored" is NOT used.
    // Only fields are written to the file.
    logger.WithField("msg", "Application started").Info("ignored")
    // Output to file: level=info fields.msg="Application started"
}
```

### Production Setup

Production-ready configuration with JSON formatter:

```go
opts := config.OptionsFile{
    Filepath:         "/var/log/myapp/app.log",
    FileMode:         0644,  // Readable by others
    PathMode:         0755,  // Standard directory permissions
    CreatePath:       true,  // Create dirs + rotation detection
    LogLevel:         []string{"info", "warning", "error"},
    DisableStack:     true,  // Don't log stack traces
    DisableTimestamp: false, // Include timestamps
}

hook, _ := hookfile.New(opts, &logrus.JSONFormatter{})
defer hook.Close()

logger := logrus.New()
logger.AddHook(hook)

// IMPORTANT: Use fields, not message parameter
logger.WithFields(logrus.Fields{
    "msg":    "Request processed",
    "method": "GET",
    "status": 200,
}).Info("ignored")
// Output: {"fields.msg":"Request processed","level":"info","method":"GET","status":200}
```

### Access Log Mode

Use message-only mode for HTTP access logs:

```go
accessOpts := config.OptionsFile{
    Filepath:        "/var/log/myapp/access.log",
    CreatePath:      true,
    EnableAccessLog: true,  // Message-only mode
}

accessHook, _ := hookfile.New(accessOpts, nil)
defer accessHook.Close()

accessLogger := logrus.New()
accessLogger.AddHook(accessHook)

// IMPORTANT: In AccessLog mode, MESSAGE is output, fields ignored
accessLogger.WithFields(logrus.Fields{
    "method": "GET",  // This field is IGNORED
    "path":   "/api", // This field is IGNORED
}).Info("GET /api/users - 200 OK - 45ms")
// Output: GET /api/users - 200 OK - 45ms
```

### Level Filtering

Route different log levels to different files:

```go
// Error log file
errorOpts := config.OptionsFile{
    Filepath: "/var/log/myapp/error.log",
    CreatePath: true,
    LogLevel: []string{"error", "fatal"},
}
errorHook, _ := hookfile.New(errorOpts, &logrus.JSONFormatter{})
defer errorHook.Close()

// Info log file
infoOpts := config.OptionsFile{
    Filepath: "/var/log/myapp/info.log",
    CreatePath: true,
    LogLevel: []string{"info", "debug"},
}
infoHook, _ := hookfile.New(infoOpts, &logrus.TextFormatter{})
defer infoHook.Close()

logger := logrus.New()
logger.AddHook(errorHook)
logger.AddHook(infoHook)

// IMPORTANT: Use fields, message parameter is ignored
logger.WithField("msg", "Normal operation").Info("ignored")    // → info.log
logger.WithField("msg", "Error occurred").Error("ignored")     // → error.log
```

### Field Filtering

Filter out verbose fields for cleaner output:

```go
opts := config.OptionsFile{
    Filepath:         "/var/log/myapp/app.log",
    CreatePath:       true,
    DisableStack:     true,  // Remove stack traces
    DisableTimestamp: true,  // Remove timestamps
    EnableTrace:      false, // Remove caller info
}

hook, _ := hookfile.New(opts, &logrus.TextFormatter{
    DisableTimestamp: true,
})
defer hook.Close()

logger := logrus.New()
logger.AddHook(hook)

// IMPORTANT: Fields are used, message parameter is ignored
logger.WithFields(logrus.Fields{
    "msg":    "Clean log",
    "stack":  "will be filtered",  // Removed by DisableStack
    "caller": "will be filtered",  // Removed by EnableTrace=false
    "user":   "john",               // Kept
}).Info("ignored")
// Output: level=info fields.msg="Clean log" user=john
```

---

## Best Practices

### Testing

The package includes comprehensive tests with **84.0% code coverage** and **28 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ Hook creation and configuration
- ✅ File writing and rotation detection
- ✅ Multiple loggers sharing same file
- ✅ Access log mode and field filtering
- ✅ Concurrency and race conditions
- ✅ Integration with logrus formatters

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Configure log rotation externally:**
```bash
# /etc/logrotate.d/myapp
/var/log/myapp/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 myapp myapp
}
```

**Use fields for structured logging:**
```go
// ✅ GOOD: Fields are output
logger.WithFields(logrus.Fields{
    "user_id": 123,
    "action":  "login",
    "msg":     "User logged in",
}).Info("ignored")
```

**Enable rotation detection:**
```go
opts := config.OptionsFile{
    Filepath:   "/var/log/app.log",
    CreatePath: true,  // Required for rotation detection
}
```

**Separate stdout and file logs:**
```go
logger := logrus.New()
logger.SetOutput(os.Stdout)  // Console output
logger.AddHook(fileHook)     // File output
```

**Close hooks on shutdown:**
```go
hook, _ := hookfile.New(opts, formatter)
defer hook.Close()  // Ensures proper cleanup
```

### ❌ DON'T

**Don't rely on message parameter in standard mode:**
```go
// ❌ BAD: Message "important" is NOT output
logger.Info("important")

// ✅ GOOD: Put text in field
logger.WithField("msg", "important").Info("ignored")
```

**Don't manually rotate files from application:**
```go
// ❌ BAD: Manual rotation
os.Rename("/var/log/app.log", "/var/log/app.log.1")
// Hook will detect and handle rotation automatically

// ✅ GOOD: Use external tools
// Configure logrotate or similar
```

**Don't create hundreds of hooks to same file:**
```go
// ❌ BAD: Excessive hooks
for i := 0; i < 1000; i++ {
    hook, _ := hookfile.New(sameOpts, formatter)
    logger.AddHook(hook)
}

// ✅ GOOD: One hook per logger, multiple loggers OK
hook, _ := hookfile.New(opts, formatter)
logger1.AddHook(hook)  // Reuse same hook
```

**Don't mix AccessLog and standard logging:**
```go
// ❌ BAD: Single logger with AccessLog mode
hook, _ := hookfile.New(&config.OptionsFile{
    EnableAccessLog: true,
}, nil)
logger.AddHook(hook)
logger.Info("app message")  // Confusing behavior

// ✅ GOOD: Separate loggers
appLogger.AddHook(appHook)
accessLogger.AddHook(accessHook)
```

**Don't ignore errors:**
```go
// ❌ BAD: Ignore errors
hook, _ := hookfile.New(opts, formatter)

// ✅ GOOD: Handle errors
hook, err := hookfile.New(opts, formatter)
if err != nil {
    log.Fatalf("Failed to create hook: %v", err)
}
```

---

## API Reference

### HookFile Interface

**`HookFile`**

Extends `logtps.Hook` interface with file-specific functionality:

```go
type HookFile interface {
    logtps.Hook
    // Inherits: Fire, Levels, RegisterHook, Run, IsRunning, Close, Write
}
```

### Configuration

**`New(opt config.OptionsFile, format logrus.Formatter) (HookFile, error)`**

Creates a new file hook with specified options and formatter.

- **Parameters:**
  - `opt`: File configuration including path, permissions, log levels
  - `format`: Logrus formatter (JSON, Text, or custom). If nil, uses entry.Bytes()

- **Returns:**
  - `HookFile`: Configured hook instance
  - `error`: Error if file cannot be created or accessed

**`config.OptionsFile`** struct:

```go
type OptionsFile struct {
    Filepath         string      // Required: Path to log file
    FileMode         FileMode    // File permissions (default: 0644)
    PathMode         FileMode    // Directory permissions (default: 0755)
    CreatePath       bool        // Create parent directories (required for rotation)
    Create           bool        // Create file if missing (required for rotation)
    LogLevel         []string    // Log levels to handle (default: all)
    DisableStack     bool        // Filter "stack" field
    DisableTimestamp bool        // Filter "time" field
    EnableTrace      bool        // Include "caller", "file", "line" fields
    EnableAccessLog  bool        // Message-only mode (ignores fields)
}
```

### Rotation Detection

The hook automatically detects log rotation when `CreatePath=true`:

**How it works:**
1. Sync timer runs every 1 second
2. Compares file descriptor inode with disk file inode
3. If different, closes old FD and opens new file
4. Logging continues seamlessly

**Supported rotation tools:**
- logrotate (Linux)
- newsyslog (BSD)
- Any tool that renames/moves log files

**Detection latency:** Up to 1 second (sync timer interval)

### Error Handling

**Construction Errors:**
```go
hook, err := hookfile.New(config.OptionsFile{
    Filepath: "",  // Missing
}, nil)
// err: "missing file path"
```

**Runtime Errors:**
- Formatter errors during Fire() are returned
- Write errors during Fire() are returned
- File rotation errors logged to stderr, continue with old FD

**Silent Behaviors:**
- Empty log data: Fire() returns nil without writing
- Empty access log message: Fire() returns nil
- Entry level not in LogLevel: Fire() returns nil (normal filtering)

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
- Update examples if needed

### Documentation Requirements

- Update GoDoc comments for public APIs
- Add runnable examples for new features
- Update README.md and TESTING.md if needed

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

- ✅ **84.0% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation with file aggregation
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Rotation-resilient** with automatic detection

### Security Considerations

**No Security Vulnerabilities Identified:**
- No external dependencies beyond Go stdlib + internal golib
- File permissions configurable (FileMode, PathMode)
- No privilege escalation paths
- Safe inode comparison for rotation detection
- Proper error handling prevents crashes

**Best Practices Applied:**
- Defensive nil checks in constructors
- Proper error propagation
- No panic paths in normal operation
- Resource cleanup with defer and Close()
- Reference counting prevents leaks

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Custom Sync Timer**: Configurable rotation detection interval
2. **Size-Based Rotation**: Built-in rotation based on file size
3. **Compression**: Automatic compression of rotated files
4. **Metrics Export**: Integration with Prometheus for hook metrics
5. **Custom Rotation Callbacks**: User-defined rotation handlers

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger/hookfile)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, rotation detection algorithm, typical use cases, and comprehensive usage examples. Provides detailed explanations of file-specific behavior.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (82.2%), and guidelines for writing new tests. Includes troubleshooting and CI integration examples.

### Related golib Packages

- **[github.com/nabbar/golib/ioutils/aggregator](https://pkg.go.dev/github.com/nabbar/golib/ioutils/aggregator)** - Core aggregator used for buffered writes. Provides the underlying write aggregation and sync timer functionality.

- **[github.com/nabbar/golib/logger/hookwriter](https://pkg.go.dev/github.com/nabbar/golib/logger/hookwriter)** - Generic hook for any io.Writer. Provides field filtering and formatting logic used by hookfile.

- **[github.com/nabbar/golib/logger/hookstdout](https://pkg.go.dev/github.com/nabbar/golib/logger/hookstdout)** - Companion package for stdout output. Use together with hookfile for console + file logging.

- **[github.com/nabbar/golib/logger/hookstderr](https://pkg.go.dev/github.com/nabbar/golib/logger/hookstderr)** - Companion package for stderr output. Use for error logs separate from main logs.

- **[github.com/nabbar/golib/logger/config](https://pkg.go.dev/github.com/nabbar/golib/logger/config)** - Configuration types used by hooks. Provides OptionsFile structure and FileMode types.

- **[github.com/nabbar/golib/logger/types](https://pkg.go.dev/github.com/nabbar/golib/logger/types)** - Hook interface definition and field constants. Defines the Hook interface implemented by hookfile.

- **[github.com/nabbar/golib/logger](https://pkg.go.dev/github.com/nabbar/golib/logger)** - Main logger package that uses hooks internally. Provides higher-level logging abstractions built on top of logrus and hooks.

### External References

- **[Logrus](https://github.com/sirupsen/logrus)** - Underlying structured logging framework. Essential for understanding hook behavior, formatters, and field handling.

- **[Logrotate](https://linux.die.net/man/8/logrotate)** - Standard Linux log rotation utility. Compatible with this package's rotation detection.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and logging patterns. The hookfile package follows these conventions.

- **[Go Standard Library](https://pkg.go.dev/std)** - Standard library documentation for `os`, `io`, and logging-related packages. Understanding os.File and inode handling is essential for rotation detection.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger/hookfile`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
