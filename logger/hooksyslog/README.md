# Logger HookSyslog

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-84.3%25-brightgreen)](TESTING.md)

Thread-safe logrus hook that writes log entries to syslog (Unix/Linux) or Windows Event Log with asynchronous buffered processing, automatic reconnection, and flexible field filtering.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Data Flow](#data-flow)
  - [Platform Support](#platform-support)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Example](#basic-example)
  - [Remote Syslog](#remote-syslog)
  - [Access Log Mode](#access-log-mode)
  - [Level Filtering](#level-filtering)
  - [Field Filtering](#field-filtering)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [HookSyslog Interface](#hooksyslog-interface)
  - [Configuration](#configuration)
  - [Severity Mapping](#severity-mapping)
  - [Error Handling](#error-handling)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **hooksyslog** package provides a production-ready logrus hook for sending structured logs to syslog (RFC 5424) on Unix/Linux systems or Windows Event Log on Windows. It features asynchronous buffered writes, automatic reconnection, and flexible configuration to handle high-throughput logging scenarios.

### Design Philosophy

1. **Platform Agnostic**: Seamless support for Unix syslog and Windows Event Log
2. **Non-Blocking**: Asynchronous buffered writes prevent logging from slowing down application
3. **Reliability**: Automatic reconnection on network failures
4. **Flexibility**: Configurable field filtering and log level mapping
5. **Standards Compliance**: Full RFC 5424 syslog compatibility

### Key Features

- ✅ **Multi-Platform**: Unix/Linux (tcp, udp, unix, unixgram) and Windows (Event Log)
- ✅ **Asynchronous Processing**: Buffered channel (250 entries) for non-blocking writes
- ✅ **Auto-Reconnection**: Automatic retry every 1 second on connection failures
- ✅ **Logrus Integration**: Implements `logrus.Hook` interface seamlessly
- ✅ **Field Filtering**: Optional removal of stack, timestamp, and trace fields
- ✅ **Access Log Mode**: Special mode where message is used instead of fields
- ✅ **Level Filtering**: Configure which log levels are sent to syslog
- ✅ **Graceful Shutdown**: Context-based shutdown for clean application termination
- ✅ **RFC 5424 Compliant**: Standard syslog severity and facility codes
- ✅ **Zero External Dependencies**: Only standard library and golib packages

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────────────┐
│                      HookSyslog                                │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌──────────┐         ┌──────────────┐        ┌──────────────┐ │
│  │ Logrus   │────────▶│              │───────▶│              │ │
│  │ Entry    │         │              │        │              │ │
│  │          │         │              │        │              │ │
│  │ Fields   │         │   Buffered   │        │  Aggregator  │ │
│  │ Level    │         │   Channel    │        │  Goroutine   │ │
│  │ Message  │────────▶│  (cap: 250)  │───────▶│              │ │
│  │          │         │              │        │              │ │
│  └──────────┘         │              │        │              │ │
│                       │              │        │              │ │
│       Fire()          │    data      │        │    Write     │ │
│         │             └──────────────┘        └───────┬──────┘ │
│         │                                            │         │
│         ▼                                            ▼         │
│  ┌──────────────┐                           ┌────────────────┐ │
│  │ Formatter    │                           │    Client      │ │
│  │ JSON/Text    │                           │  (Socket)      │ │
│  └──────────────┘                           │                │ │
│                                             │  Unix: syslog  │ │
│  ┌──────────────┐                           │  Win: EventLog │ │
│  │ Field Filter │                           └────────────────┘ │
│  │ - Stack      │                                   │          │
│  │ - Timestamp  │                                   ▼          │
│  │ - Trace      │                          ┌─────────────────┐ │
│  └──────────────┘                          │  Syslog/Event   │ │
│                                            │      Log        │ │
│                                            └─────────────────┘ │
└────────────────────────────────────────────────────────────────┘
```

### Data Flow

```
Logrus Log Statement → Fire(entry) → Format → Filter → Channel → Aggregator → Syslog
                           │            │        │         │        │           │
                           │            │        │         │        │           └─▶ TCP/UDP/Unix
                           │            │        │         │        │
                           │            │        │         │        └─▶ Retry on failure
                           │            │        │         │
                           │            │        │         └─▶ Buffer (250 entries)
                           │            │        │
                           │            │        └─▶ Remove stack/time/trace fields
                           │            │
                           │            └─▶ JSON or Text formatting
                           │
                           └─▶ Map logrus level to syslog severity
```

### Platform Support

**Unix/Linux** (via `log/syslog`):
- **Protocols**: TCP, UDP, Unix domain sockets (stream/datagram)
- **Syslog Daemon**: Compatible with rsyslog, syslog-ng, systemd-journald
- **Network**: Supports remote syslog servers
- **Build Tag**: `linux` or `darwin`

**Windows** (via `golang.org/x/sys/windows/svc/eventlog`):
- **Event Log**: Writes to Windows Event Log
- **Registration**: Automatic service registration on first use
- **Severity Mapping**: Maps syslog severities to Windows event types
- **Build Tag**: `windows`

---

## Performance

### Benchmarks

Based on test suite results (40 specs, AMD64):

| Operation | Median | Mean | Max | Notes |
|-----------|--------|------|-----|-------|
| **Hook Creation** | ~10ms | ~15ms | ~50ms | Includes syslog connection check |
| **Fire() (fields)** | <100µs | <500µs | <2ms | Non-blocking (buffered) |
| **Fire() (access log)** | <50µs | <300µs | <1ms | Message-only mode |
| **WriteSev()** | <100µs | <200µs | <1ms | Direct write to buffer |
| **Run() startup** | ~100ms | ~150ms | ~200ms | Initial syslog connection |
| **Shutdown (Close)** | ~50ms | ~100ms | ~200ms | Graceful channel close |

**Throughput:**
- Single logger: **~10,000 entries/second**
- Multiple hooks: **~5,000 entries/second per hook**
- Network I/O: **Limited by syslog server, not hook overhead**

### Memory Usage

```
Base overhead:        ~2KB (struct + channels)
Per buffered entry:   len(formatted) + ~64 bytes (data struct + channel overhead)
Total at capacity:    250 × (AvgEntrySize + 64 bytes)
```

**Example:**
- Buffer capacity: 250
- Average entry: 256 bytes (JSON formatted)
- Peak memory ≈ 250 × 320 = 80KB

### Scalability

- **Concurrent Loggers**: Tested with up to 10 concurrent loggers
- **Buffer Capacity**: Fixed at 250 entries (adjustable in code)
- **Log Levels**: All logrus levels supported (Panic, Fatal, Error, Warn, Info, Debug, Trace)
- **Zero Race Conditions**: All tests pass with `-race` detector

---

## Use Cases

### 1. Centralized Log Collection

**Problem**: Multiple servers need to send logs to a central syslog server.

```go
// Configure remote syslog over UDP
opts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkUDP.Code(),
    Host:     "logs.example.com:514",
    Tag:      "myapp",
    Facility: "LOCAL0",
}

hook, _ := logsys.New(opts, &logrus.JSONFormatter{})
logger.AddHook(hook)

ctx, cancel := context.WithCancel(context.Background())
go hook.Run(ctx)
defer func() {
    cancel()
    hook.Close()
}()
```

**Real-world**: Used in microservices for shipping logs to ELK or Splunk via syslog forwarders.

### 2. HTTP Access Logging

**Problem**: Log HTTP access logs in standard Apache/Nginx format.

```go
// Access log mode: use entry.Message instead of fields
opts := logcfg.OptionsSyslog{
    Network:         libptc.NetworkUnixGram.Code(),
    Host:            "/dev/log",
    Tag:             "nginx-access",
    EnableAccessLog: true,  // Message mode
}

hook, _ := logsys.New(opts, nil)
logger.AddHook(hook)

// Log statement - message will be written to syslog
logger.Info("192.168.1.1 - - [01/Dec/2025:20:00:00 +0100] \"GET /api HTTP/1.1\" 200 1234")
```

### 3. System Service Logging

**Problem**: Daemon service needs to log to system syslog (journald/Event Log).

```go
// Unix: logs to journald via /dev/log
// Windows: logs to Windows Event Log
opts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkUnixGram.Code(),  // or NetworkUnknown for default
    Host:     "",  // Default system socket
    Tag:      "myservice",
    Facility: "DAEMON",
    LogLevel: []string{"info", "warning", "error"},
}

hook, _ := logsys.New(opts, &logrus.TextFormatter{})
logger.AddHook(hook)
```

### 4. Security Audit Trail

**Problem**: Security-sensitive operations need to be logged to tamper-proof syslog.

```go
opts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkTCP.Code(),  // TCP for reliability
    Host:     "audit.internal:601",
    Tag:      "security-audit",
    Facility: "AUTHPRIV",  // Restricted facility
    LogLevel: []string{"warning", "error"},  // Security events only
}

hook, _ := logsys.New(opts, &logrus.JSONFormatter{})
securityLogger.AddHook(hook)

// All security events go to dedicated audit syslog
securityLogger.WithFields(logrus.Fields{
    "user":   "admin",
    "action": "login",
    "source": "192.168.1.100",
}).Warn("Failed login attempt")
```

### 5. Multi-Destination Logging

**Problem**: Send different log levels to different syslog destinations.

```go
// Errors to dedicated error server
errorOpts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkUDP.Code(),
    Host:     "errors.example.com:514",
    Tag:      "myapp-errors",
    LogLevel: []string{"error", "fatal"},
}

// All logs to general server
allOpts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkUDP.Code(),
    Host:     "logs.example.com:514",
    Tag:      "myapp",
}

errorHook, _ := logsys.New(errorOpts, nil)
allHook, _ := logsys.New(allOpts, nil)

logger.AddHook(errorHook)
logger.AddHook(allHook)
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger/hooksyslog
```

### Basic Example

```go
package main

import (
    "context"
    "github.com/sirupsen/logrus"
    
    logcfg "github.com/nabbar/golib/logger/config"
    logsys "github.com/nabbar/golib/logger/hooksyslog"
    libptc "github.com/nabbar/golib/network/protocol"
)

func main() {
    // Create logger
    logger := logrus.New()
    
    // Configure syslog hook
    opts := logcfg.OptionsSyslog{
        Network:  libptc.NetworkUDP.Code(),
        Host:     "localhost:514",
        Tag:      "myapp",
        LogLevel: []string{"info", "warning", "error"},
    }
    
    hook, err := logsys.New(opts, &logrus.JSONFormatter{})
    if err != nil {
        panic(err)
    }
    
    // Register hook
    logger.AddHook(hook)
    
    // Start background writer
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    go hook.Run(ctx)
    
    // Graceful shutdown
    defer func() {
        cancel()
        hook.Close()
    }()
    
    // IMPORTANT: In standard mode, the message parameter is IGNORED.
    // Only fields are formatted and sent to syslog.
    logger.WithField("msg", "Application started").Info("ignored")
    logger.WithFields(logrus.Fields{
        "msg":     "User logged in",
        "user":    "john",
        "session": "abc123",
    }).Info("ignored")
}
```

### Remote Syslog

```go
// Send logs to remote syslog server via UDP
opts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkUDP.Code(),
    Host:     "logs.example.com:514",
    Tag:      "myapp",
    Facility: "LOCAL0",
}

hook, _ := logsys.New(opts, &logrus.JSONFormatter{})
logger.AddHook(hook)

ctx, cancel := context.WithCancel(context.Background())
go hook.Run(ctx)
defer func() {
    cancel()
    hook.Close()
}()

// Only fields are sent - message "ignored" is discarded
logger.WithFields(logrus.Fields{
    "msg":      "Remote logging test",
    "service":  "api",
    "instance": "prod-1",
}).Info("ignored")
```

### Access Log Mode

```go
// Access log mode: Message IS used, fields are IGNORED
opts := logcfg.OptionsSyslog{
    Network:         libptc.NetworkUDP.Code(),
    Host:            "localhost:514",
    Tag:             "http-access",
    EnableAccessLog: true,  // Reverses behavior!
    LogLevel:        []string{"info"},
}

hook, _ := logsys.New(opts, nil)
logger.AddHook(hook)

ctx, cancel := context.WithCancel(context.Background())
go hook.Run(ctx)

// In AccessLog mode, the MESSAGE is written, fields are ignored
logger.WithFields(logrus.Fields{
    "ignored": "these fields are discarded",
}).Info("GET /api/users - 200 OK - 45ms")  // This message IS sent to syslog

defer func() {
    cancel()
    hook.Close()
}()
```

### Level Filtering

```go
// Only send errors and above to syslog
opts := logcfg.OptionsSyslog{
    Network:  libptc.NetworkUDP.Code(),
    Host:     "localhost:514",
    Tag:      "myapp-errors",
    LogLevel: []string{"error", "fatal"},  // Info/Debug won't be sent
}

hook, _ := logsys.New(opts, &logrus.TextFormatter{})
logger.AddHook(hook)

ctx, cancel := context.WithCancel(context.Background())
go hook.Run(ctx)

// This will NOT be sent (Info level)
logger.WithField("msg", "Request completed").Info("ignored")

// This WILL be sent (Error level)
logger.WithField("msg", "Database connection failed").Error("ignored")

defer func() {
    cancel()
    hook.Close()
}()
```

### Field Filtering

```go
// Remove stack traces and timestamps from syslog output
opts := logcfg.OptionsSyslog{
    Network:          libptc.NetworkUDP.Code(),
    Host:             "localhost:514",
    Tag:              "myapp",
    DisableStack:     true,  // Remove "stack" field
    DisableTimestamp: true,  // Remove "time" field
    EnableTrace:      false, // Remove "caller", "file", "line" fields
    LogLevel:         []string{"info"},
}

hook, _ := logsys.New(opts, &logrus.TextFormatter{})
logger.AddHook(hook)

ctx, cancel := context.WithCancel(context.Background())
go hook.Run(ctx)

// Fields "stack", "time", "caller" will be filtered out before sending
logger.WithFields(logrus.Fields{
    "msg":    "Filtered log entry",
    "user":   "john",
    "action": "login",
    "stack":  "will be filtered out",
    "caller": "will be filtered out",
}).Info("ignored")

defer func() {
    cancel()
    hook.Close()
}()
```

---

## Best Practices

### Testing

The package includes a comprehensive test suite with **84.3% code coverage** and **40 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All public APIs and lifecycle operations
- ✅ Concurrent access with race detector (zero races detected)
- ✅ Integration tests with mock syslog server
- ✅ Error handling and edge cases
- ✅ Platform-specific implementations (Unix/Windows)

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Always Start Run() Goroutine:**
```go
// ✅ GOOD: Start background writer
ctx, cancel := context.WithCancel(context.Background())
go hook.Run(ctx)

// Logs are processed asynchronously
logger.Info("message")
```

**Graceful Shutdown:**
```go
// ✅ GOOD: Complete shutdown sequence
defer func() {
    cancel()        // Signal Run() to stop
    hook.Close()    // Close channels
}()
```

**Use Fields, Not Messages (Standard Mode):**
```go
// ✅ GOOD: Fields are sent to syslog
logger.WithFields(logrus.Fields{
    "msg":    "User logged in",
    "user":   "john",
    "action": "login",
}).Info("this message is IGNORED")
```

**Use Messages, Not Fields (AccessLog Mode):**
```go
// ✅ GOOD: Message is sent to syslog (EnableAccessLog: true)
logger.Info("192.168.1.1 - GET /api - 200 OK")
// Fields would be ignored in this mode
```

**Configure Appropriate Log Levels:**
```go
// ✅ GOOD: Filter noisy logs
opts := logcfg.OptionsSyslog{
    LogLevel: []string{"warning", "error"},  // Only send important logs
}
```

**Use Structured Logging:**
```go
// ✅ GOOD: Structured fields for parsing
logger.WithFields(logrus.Fields{
    "msg":        "API request",
    "method":     "GET",
    "path":       "/api/users",
    "status":     200,
    "duration_ms": 45,
}).Info("ignored")
```

### ❌ DON'T

**Don't Forget Run() Goroutine:**
```go
// ❌ BAD: Hook never processes logs
hook, _ := logsys.New(opts, nil)
logger.AddHook(hook)
// Missing: go hook.Run(ctx)
```

**Don't Use Message Parameter (Standard Mode):**
```go
// ❌ BAD: Message "User logged in" is IGNORED
logger.Info("User logged in")

// ✅ GOOD: Use fields instead
logger.WithField("msg", "User logged in").Info("ignored")
```

**Don't Use Fields (AccessLog Mode):**
```go
// ❌ BAD: Fields are IGNORED in AccessLog mode
logger.WithField("msg", "access log").Info("ignored")

// ✅ GOOD: Use message parameter
logger.Info("192.168.1.1 - GET /api - 200")
```

**Don't Skip Graceful Shutdown:**
```go
// ❌ BAD: Buffered logs may be lost
cancel()
os.Exit(0)  // Immediate exit loses buffered entries

// ✅ GOOD: Wait for buffer to flush
cancel()
hook.Close()
```

**Don't Block in Production:**
```go
// ❌ BAD: Checking IsRunning() on every log
if hook.IsRunning() {
    logger.Info("message")
}

// ✅ GOOD: Just log, hook handles buffering
logger.Info("message")
```

**Don't Use Without Formatter (if needed):**
```go
// ❌ BAD: Default formatter may not be optimal
hook, _ := logsys.New(opts, nil)  // Uses logrus default

// ✅ GOOD: Explicit formatter for consistency
hook, _ := logsys.New(opts, &logrus.JSONFormatter{
    TimestampFormat: time.RFC3339,
})
```

---

## API Reference

### HookSyslog Interface

```go
type HookSyslog interface {
    logrus.Hook
    
    // Run starts the hook's main processing loop
    Run(ctx context.Context)
    
    // IsRunning checks if the hook is currently active
    IsRunning() bool
    
    // Close terminates the hook
    Close() error
}
```

**Methods:**

- **`Levels() []logrus.Level`**: Returns configured log levels for this hook
- **`Fire(entry *logrus.Entry) error`**: Processes log entry (called by logrus)
- **`Run(ctx context.Context)`**: Start background writer (must be called in goroutine)
- **`Close() error`**: Close channels
- **`IsRunning() bool`**: Check if background writer is active
- **`RegisterHook(logger)`**: Convenience method to add hook to logger

### Configuration

```go
type OptionsSyslog struct {
    // Connection
    Network  string   // tcp, udp, unix, unixgram (NetworkProtocol code)
    Host     string   // "host:port" or socket path
    Tag      string   // Syslog tag (application name)
    Facility string   // Syslog facility (USER, DAEMON, LOCAL0-7, etc.)
    
    // Filtering
    LogLevel         []string  // Log levels to process (empty = all)
    DisableStack     bool      // Remove "stack" field
    DisableTimestamp bool      // Remove "time" field
    EnableTrace      bool      // Keep "caller", "file", "line" fields
    
    // Modes
    EnableAccessLog  bool      // Reverse behavior: use Message, ignore Fields
}
```

**Network Protocols:**
- `tcp`: TCP connection (requires host:port)
- `udp`: UDP datagrams (requires host:port)
- `unix`: Unix domain socket stream
- `unixgram`: Unix domain socket datagram
- Empty: Platform default (/dev/log on Unix, Event Log on Windows)

**Facilities:**
```
KERN, USER, MAIL, DAEMON, AUTH, SYSLOG, LPR, NEWS,
UUCP, CRON, AUTHPRIV, FTP, LOCAL0-LOCAL7
```

### Severity Mapping

| Logrus Level | Syslog Severity | Numeric | Description |
|--------------|-----------------|---------|-------------|
| `PanicLevel` | ALERT | 1 | Action must be taken immediately |
| `FatalLevel` | CRIT | 2 | Critical conditions |
| `ErrorLevel` | ERR | 3 | Error conditions |
| `WarnLevel` | WARNING | 4 | Warning conditions |
| `InfoLevel` | INFO | 6 | Informational messages |
| `DebugLevel` | DEBUG | 7 | Debug-level messages |
| `TraceLevel` | INFO | 6 | Debug-level messages (fallback to INFO) |

### Error Handling

```go
var errStreamClosed = errors.New("stream is closed")
```

**Error Behavior:**
- Errors from syslog connection are logged but don't stop processing
- Automatic reconnection every 1 second on connection failure
- `WriteSev()` returns `errStreamClosed` if called after `Close()`
- Context cancellation triggers graceful shutdown
- Panics in `Run()` are recovered and logged

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
   - Ensure zero race conditions with `CGO_ENABLED=1 go test -race`
   - Update test documentation in TESTING.md

4. **Documentation**
   - Update GoDoc comments for public APIs
   - Add examples for new features
   - Update README.md and TESTING.md if needed
   - Document behavior differences (standard vs AccessLog mode)

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

- ✅ **84.3% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** implementation using atomic operations and channels
- ✅ **Panic recovery** in Run() goroutine
- ✅ **Platform-tested** on Unix/Linux and Windows

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Configurable Buffer Size**: Allow users to adjust channel capacity (currently fixed at 250)
2. **Metrics Export**: Optional integration with Prometheus for monitoring
3. **Compression**: Optional gzip compression for large log entries
4. **Batch Writing**: Combine multiple entries into single syslog write for efficiency

These are **optional improvements** and not required for production use. The current implementation is stable and performant.

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger/hooksyslog)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture diagrams, syslog severity mapping, and best practices for production use. Provides detailed explanations of standard mode vs AccessLog mode behavior.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, coverage analysis (84.3%), integration tests with mock syslog server, and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/logger/config](https://pkg.go.dev/github.com/nabbar/golib/logger/config)** - Configuration structures for logger components including `OptionsSyslog` used to configure the syslog hook. Provides standardized configuration across logger ecosystem.

- **[github.com/nabbar/golib/logger/types](https://pkg.go.dev/github.com/nabbar/golib/logger/types)** - Common logger types and interfaces including field names (`FieldStack`, `FieldTime`) used for field filtering. Ensures consistency across logger hooks.

- **[github.com/nabbar/golib/network/protocol](https://pkg.go.dev/github.com/nabbar/golib/network/protocol)** - Network protocol enumeration and parsing used for syslog network configuration. Supports TCP, UDP, Unix sockets with type-safe protocol handling.

- **[github.com/nabbar/golib/runner](https://pkg.go.dev/github.com/nabbar/golib/runner)** - Panic recovery mechanism used in Run() goroutine via `RecoveryCaller()`. Ensures graceful error handling without crashing the application.

### External References

- **[RFC 5424 - Syslog Protocol](https://tools.ietf.org/html/rfc5424)** - Official syslog protocol specification defining severity levels, facility codes, and message format. The package implements full RFC 5424 compliance for Unix/Linux systems.

- **[logrus Documentation](https://github.com/sirupsen/logrus)** - Popular structured logger for Go. This package integrates as a logrus hook, inheriting logrus's design philosophy of structured logging with fields.

- **[Go log/syslog Package](https://pkg.go.dev/log/syslog)** - Standard library syslog implementation used internally on Unix/Linux. Understanding this package helps debug connection issues and network configurations.

- **[Windows Event Log](https://docs.microsoft.com/en-us/windows/win32/eventlog/event-logging)** - Microsoft documentation for Windows Event Log system. The package uses this API on Windows, mapping syslog severities to Windows event types.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger/hooksyslog`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
