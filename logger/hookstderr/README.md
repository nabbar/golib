# Logger HookStdErr Package

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen)](TESTING.md)

Logrus hook for writing log entries to standard error (stderr) with configurable field filtering, formatting options, and Unix-standard error output conventions.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The `hookstderr` package provides a specialized logrus hook for writing log entries to standard error (stderr), following Unix conventions where errors and diagnostic information belong on stderr while normal program output goes to stdout.

### Design Philosophy

1. **Unix Conventions**: Dedicated stderr output for errors and diagnostics
2. **Wrapper Simplicity**: Thin wrapper over hookwriter with stderr-specific defaults
3. **Configuration Consistency**: Uses OptionsStd for uniform configuration across logger hooks
4. **Cross-Platform Color**: Automatic color handling via mattn/go-colorable
5. **Testing Friendly**: Custom writer support for testing without polluting stderr

### Why Use This Package?

- **Standard Error Separation**: Proper Unix-style separation of errors (stderr) from output (stdout)
- **Field Filtering**: Remove sensitive or verbose fields from error logs
- **Flexible Formatting**: Support for JSON, Text, or custom formatters
- **Level-Based Routing**: Send only specific log levels to stderr
- **Color Control**: Automatic color stripping for non-terminal outputs
- **Zero Cost When Disabled**: Returns nil with no allocation overhead

### Key Features

- **Dedicated stderr output** for error and diagnostic messages
- **Automatic color support** detection and stripping via go-colorable
- **Selective field filtering** (stack traces, timestamps, caller info)
- **Access log mode** for message-only output
- **Multiple formatter support** (JSON, Text, custom)
- **Level-based filtering** (handle only specific log levels)
- **Custom writer support** for testing and advanced scenarios
- **100% test coverage** with 40 test specifications
- **Thread-safe** when used correctly (safe for concurrent logging)
- **Zero dependencies** beyond logrus, go-colorable, and internal golib packages

---

## Architecture

### Package Structure

```
logger/hookstderr/
├── interface.go         # HookStdErr interface and constructor functions
├── doc.go              # Package documentation with usage examples
├── example_test.go     # Runnable examples (10 examples)
├── hookstderr_test.go  # Creation and configuration tests
├── fire_test.go        # Fire method and integration tests
└── hookstderr_suite_test.go  # Ginkgo test suite entry point
```

### Component Overview

```
┌──────────────────────────────────────────────┐
│             logrus.Logger                    │
│                                              │
│  ┌────────────────────────────────────┐      │
│  │  logger.Error("error message")     │      │
│  └────────────────┬───────────────────┘      │
│                   │                          │
│                   ▼                          │
│         ┌──────────────────┐                 │
│         │  logrus.Entry    │                 │
│         └──────────┬───────┘                 │
│                    │                         │
└────────────────────┼─────────────────────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │   HookStdErr.Fire()        │
        │   (delegates to hookwriter)│
        └────────────┬───────────────┘
                     │
                     ▼
              ┌──────────────┐
              │  os.Stderr   │
              │  (or custom) │
              └──────────────┘
```

| Component | Role | Dependencies |
|-----------|------|--------------|
| **HookStdErr** | Interface extending logtps.Hook | hookwriter, OptionsStd |
| **New()** | Constructor using os.Stderr | NewWithWriter |
| **NewWithWriter()** | Constructor with custom writer | hookwriter.New |

### How It Works

1. **Initialization**: `New()` or `NewWithWriter()` creates hook with configuration
2. **Registration**: Hook registered with logrus via `logger.AddHook(hook)`
3. **Interception**: logrus calls `Fire()` for each log entry at matching levels
4. **Filtering**: hookwriter filters fields based on configuration
5. **Formatting**: Entry formatted using configured formatter or default
6. **Writing**: Formatted output written to stderr (or custom writer)

---

## Performance

### Memory Efficiency

**Low Overhead** - The package has minimal memory impact:

```
Hook creation:     ~200 bytes (interface wrapper)
Per log entry:     Delegated to hookwriter (entry duplication)
Disabled hook:     0 bytes (returns nil)
```

### Throughput

Performance depends on underlying hookwriter and stderr characteristics:

| Scenario | Performance | Notes |
|----------|-------------|-------|
| Terminal stderr | High | OS-buffered writes |
| Redirected stderr | Very High | File-backed, buffered |
| Network stderr | Variable | Depends on network latency |
| Custom buffer | Very High | In-memory writes |

### Scalability

- **Concurrent loggers**: Safe for multiple goroutines logging simultaneously
- **Hook count**: Linear overhead per registered hook
- **Entry size**: Constant overhead regardless of message size
- **Level filtering**: Zero cost for filtered-out levels

---

## Use Cases

This package excels in scenarios requiring Unix-standard error handling:

**Separate Error and Info Logs**
- Send errors to stderr, info to stdout
- Follow Unix convention for scriptable applications
- Enable proper shell redirection (2>&1, 2>/dev/null)

**Structured Error Logging**
- JSON-formatted errors to stderr for parsing
- Text-formatted errors for human readability
- Filter fields for clean error messages

**Testing Without Pollution**
- Use custom writers to capture stderr in tests
- Validate error output without console clutter
- Mock stderr for integration testing

**Level-Based Routing**
- Error/Fatal/Panic to stderr
- Info/Debug/Trace to stdout or files
- Warn to either based on context

**Clean Error Messages**
- Filter stack traces for production
- Remove timestamps for cleaner output
- Strip caller info for less verbose errors

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger/hookstderr
```

### Basic Error Logging

```go
package main

import (
    "github.com/sirupsen/logrus"
    "github.com/nabbar/golib/logger/config"
    "github.com/nabbar/golib/logger/hookstderr"
)

func main() {
    // Configure hook for stderr
    opt := &config.OptionsStd{
        DisableStandard: false,
        DisableColor:    false,  // Enable color on terminal stderr
    }

    // Create and register hook
    hook, err := hookstderr.New(opt, nil, &logrus.TextFormatter{})
    if err != nil {
        panic(err)
    }

    logger := logrus.New()
    logger.AddHook(hook)

    // Errors go to stderr, message must be on field and not in message function
    logger.WithField("msg", "This error appears on stderr").Error("ignored")
}
```

### Separate Stdout and Stderr

```go
package main

import (
    "github.com/sirupsen/logrus"
    "github.com/nabbar/golib/logger/config"
    "github.com/nabbar/golib/logger/hookstderr"
    "github.com/nabbar/golib/logger/hookstdout"
)

func main() {
    logger := logrus.New()

    // Errors to stderr
    stderrOpt := &config.OptionsStd{DisableStandard: false}
    errHook, _ := hookstderr.New(stderrOpt, []logrus.Level{
        logrus.ErrorLevel,
        logrus.FatalLevel,
        logrus.PanicLevel,
    }, nil)
    logger.AddHook(errHook)

    // Info to stdout
    stdoutOpt := &config.OptionsStd{DisableStandard: false}
    infoHook, _ := hookstdout.New(stdoutOpt, []logrus.Level{
        logrus.InfoLevel,
        logrus.DebugLevel,
    }, nil)
    logger.AddHook(infoHook)

    logger.WithField("msg", "Info to stdout").Info("ignored message")
    logger.WithField("msg", "Error to stderr").Error("ignored message")
}
```

### JSON Error Logs

```go
opt := &config.OptionsStd{
    DisableStandard: false,
    DisableColor:    true,  // No color in JSON
}

hook, _ := hookstderr.New(opt, nil, &logrus.JSONFormatter{})
logger.AddHook(hook)

logger.WithFields(logrus.Fields{
    "msg":        "Database connection failed",
    "error_code": "E500",
    "request_id": "abc123",
}).Error("ignored message")
// {"error_code":"E500","level":"error","fields.msg":"Database connection failed","request_id":"abc123"}
```

### Testing with Custom Writer

```go
func TestErrorLogging(t *testing.T) {
    var buf bytes.Buffer

    opt := &config.OptionsStd{DisableStandard: false}
    hook, _ := hookstderr.NewWithWriter(&buf, opt, nil, nil)

    logger := logrus.New()
    logger.SetOutput(io.Discard)  // Disable default output
    logger.AddHook(hook)

    logger.WithField("msg", "test error").Error("ignored message")

    assert.Contains(t, buf.String(), "test error")
}
```

### Field Filtering

```go
opt := &config.OptionsStd{
    DisableStandard:  false,
    DisableStack:     true,   // Remove stack traces
    DisableTimestamp: true,   // Remove timestamps
    EnableTrace:      false,  // Remove caller info
}

hook, _ := hookstderr.New(opt, nil, &logrus.TextFormatter{})
logger.AddHook(hook)

logger.WithFields(logrus.Fields{
    "msg":   "Clean error message",
    "stack": "will be filtered",
    "user":  "will appear",
}).Error("ignored message")
// Output: level=error fields.msg="Clean error message" user="will appear"
```

### Access Log Mode

```go
opt := &config.OptionsStd{
    DisableStandard: false,
    EnableAccessLog: true,  // Message-only mode
}

hook, _ := hookstderr.New(opt, nil, nil)
logger.AddHook(hook)

logger.WithField("status", 500).Error("500 Internal Server Error")
// Output: 500 Internal Server Error
```

---

## Best Practices

### Testing

The package includes comprehensive tests with **100% code coverage** and **40 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

**Key test coverage:**
- ✅ All constructor combinations and configurations
- ✅ Field filtering (stack, timestamp, trace)
- ✅ Formatter integration (JSON, Text)
- ✅ Level filtering and multiple hooks
- ✅ Integration with logrus.Logger
- ✅ Custom writer scenarios

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Proper Hook Registration:**
```go
// ✅ GOOD: Register hook properly
hook, err := hookstderr.New(opt, nil, nil)
if err != nil {
    log.Fatal(err)
}
if hook != nil {  // Check for disabled hook
    logger.AddHook(hook)
}
```

**Level-Based Routing:**
```go
// ✅ GOOD: Route errors to stderr
errLevels := []logrus.Level{
    logrus.ErrorLevel,
    logrus.FatalLevel,
    logrus.PanicLevel,
}
hook, _ := hookstderr.New(opt, errLevels, nil)
```

**Testing Without Pollution:**
```go
// ✅ GOOD: Use custom writer for tests
var buf bytes.Buffer
hook, _ := hookstderr.NewWithWriter(&buf, opt, nil, nil)
logger.AddHook(hook)
// Test assertions on buf.String()
```

**Color Control:**
```go
// ✅ GOOD: Disable color for non-terminal
opt := &config.OptionsStd{
    DisableStandard: false,
    DisableColor:    !terminal.IsTerminal(int(os.Stderr.Fd())),
}
```

### ❌ DON'T

**Don't Ignore Disabled Hook:**
```go
// ❌ BAD: Not checking for nil
hook, _ := hookstderr.New(&config.OptionsStd{DisableStandard: true}, nil, nil)
logger.AddHook(hook)  // hook is nil!

// ✅ GOOD: Check before adding
if hook != nil {
    logger.AddHook(hook)
}
```

**Don't Mix Streams Carelessly:**
```go
// ❌ BAD: Both stdout and stderr for same levels
stderrHook, _ := hookstderr.New(opt, logrus.AllLevels, nil)
stdoutHook, _ := hookstdout.New(opt, logrus.AllLevels, nil)
// Logs appear on both streams!

// ✅ GOOD: Separate levels
stderrHook, _ := hookstderr.New(opt, []logrus.Level{logrus.ErrorLevel}, nil)
stdoutHook, _ := hookstdout.New(opt, []logrus.Level{logrus.InfoLevel}, nil)
```

**Don't Ignore Errors:**
```go
// ❌ BAD: Ignoring constructor errors
hook, _ := hookstderr.New(opt, nil, nil)

// ✅ GOOD: Handle errors
hook, err := hookstderr.New(opt, nil, nil)
if err != nil {
    log.Fatalf("Failed to create stderr hook: %v", err)
}
```

---

## API Reference

### Types

#### HookStdErr Interface

```go
type HookStdErr interface {
    logtps.Hook
}
```

Extends `logtps.Hook` which provides:
- `Fire(*logrus.Entry) error` - Process log entry
- `Levels() []logrus.Level` - Return handled levels
- `RegisterHook(*logrus.Logger)` - Register with logger
- `Run(context.Context)` - No-op for this hook
- `Write([]byte) (int, error)` - io.Writer interface

### Functions

#### New

```go
func New(opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error)
```

Creates a new HookStdErr writing to os.Stderr.

**Parameters:**
- `opt` - Configuration options. If nil or DisableStandard=true, returns (nil, nil)
- `lvls` - Log levels to handle. If empty, defaults to logrus.AllLevels
- `f` - Optional formatter. If nil, uses entry.Bytes()

**Returns:**
- `HookStdErr` - Configured hook or nil if disabled
- `error` - Error from underlying hookwriter (rarely occurs)

#### NewWithWriter

```go
func NewWithWriter(w io.Writer, opt *logcfg.OptionsStd, lvls []logrus.Level, f logrus.Formatter) (HookStdErr, error)
```

Creates a new HookStdErr with custom writer.

**Parameters:**
- `w` - Target writer. If nil, defaults to os.Stderr
- `opt` - Configuration options. If nil or DisableStandard=true, returns (nil, nil)
- `lvls` - Log levels to handle. If empty, defaults to logrus.AllLevels
- `f` - Optional formatter. If nil, uses entry.Bytes()

**Returns:**
- `HookStdErr` - Configured hook or nil if disabled
- `error` - Error from underlying hookwriter

### Configuration

#### OptionsStd Structure

```go
type OptionsStd struct {
    DisableStandard  bool  // If true, returns nil hook (disabled)
    DisableColor     bool  // If true, strips ANSI color codes
    DisableStack     bool  // If true, filters "stack" field
    DisableTimestamp bool  // If true, filters "time" field
    EnableTrace      bool  // If false, filters "caller", "file", "line" fields
    EnableAccessLog  bool  // If true, message-only mode (ignores fields)
}
```

**Field Filtering:**

| Option | Filtered Fields | Use Case |
|--------|----------------|----------|
| `DisableStack: true` | `stack` | Remove stack traces from errors |
| `DisableTimestamp: true` | `time` | Clean output without timestamps |
| `EnableTrace: false` | `caller`, `file`, `line` | Remove caller information |
| `EnableAccessLog: true` | All fields | Message-only access logs |

### Error Handling

The hook can return errors in these situations:

**Construction Errors:**
```go
// None specific to hookstderr
// Errors propagate from hookwriter.New if writer is nil (shouldn't happen)
```

**Runtime Errors:**
```go
// Formatter errors during Fire()
err := hook.Fire(entry)  // Returns formatter.Format() error

// Writer errors during Fire()
err := hook.Fire(entry)  // Returns writer.Write() error
```

**Silent Behaviors:**
- Empty log data: `Fire()` returns nil without writing
- Empty access log message: `Fire()` returns nil without writing
- Disabled hook: `New()` returns (nil, nil) - not an error

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
   - Follow Go best practices and idioms
   - Maintain or improve code coverage (target: 100%)
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
   - Update examples if API changes

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

- ✅ **100% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** when used per logger instance
- ✅ **Memory-safe** with proper resource cleanup
- ✅ **Standard interfaces** for maximum compatibility

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

1. **Async Writing**: Optional async write buffer for high-throughput scenarios
2. **Metrics Integration**: Built-in metrics for error rates and volumes
3. **Dynamic Filtering**: Runtime-adjustable field filters without recreating hook
4. **Batch Writing**: Optional batching of multiple errors for efficiency

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger/hookstderr)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface and usage patterns.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy, architecture, configuration options, use cases, performance considerations, thread safety, error handling, limitations, and best practices.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture, BDD methodology with Ginkgo v2, 100% coverage analysis, and guidelines for writing new tests.

### Related golib Packages

- **[github.com/nabbar/golib/logger/config](https://pkg.go.dev/github.com/nabbar/golib/logger/config)** - Configuration types including OptionsStd used for hook configuration. Provides standardized configuration structure across all logger hooks.

- **[github.com/nabbar/golib/logger/hookwriter](https://pkg.go.dev/github.com/nabbar/golib/logger/hookwriter)** - Core hook implementation that hookstderr wraps. Handles actual field filtering, formatting, and writing operations.

- **[github.com/nabbar/golib/logger/hookstdout](https://pkg.go.dev/github.com/nabbar/golib/logger/hookstdout)** - Equivalent package for stdout output. Use together with hookstderr to properly separate error and info streams.

- **[github.com/nabbar/golib/logger/types](https://pkg.go.dev/github.com/nabbar/golib/logger/types)** - Hook interface definition and common types. Defines the Hook interface that HookStdErr implements.

### External References

- **[logrus](https://github.com/sirupsen/logrus)** - Structured logger for Go. The hookstderr package integrates with logrus via its Hook interface.

- **[go-colorable](https://github.com/mattn/go-colorable)** - Cross-platform colored terminal support. Used for automatic color detection and stripping on Windows and Unix systems.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for interfaces, error handling, and logging patterns.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the hookstderr package. Check existing issues before creating new ones.

- **[Contributing Guide](../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements, testing procedures, and pull request process.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2025 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger/hookstderr`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
