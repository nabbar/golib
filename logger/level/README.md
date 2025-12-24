# Logger Level

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://go.dev/doc/install)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](../../../../LICENSE)
[![Coverage](https://img.shields.io/badge/Coverage-98.0%25-brightgreen)](TESTING.md)

Type-safe logging level definitions with parsing, validation, and logrus integration, providing standardized severity levels for structured logging applications.

---

## Table of Contents

- [Overview](#overview)
  - [Design Philosophy](#design-philosophy)
  - [Key Features](#key-features)
- [Architecture](#architecture)
  - [Component Diagram](#component-diagram)
  - [Level Hierarchy](#level-hierarchy)
- [Performance](#performance)
  - [Benchmarks](#benchmarks)
  - [Memory Usage](#memory-usage)
  - [Scalability](#scalability)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
  - [Installation](#installation)
  - [Basic Level Usage](#basic-level-usage)
  - [Parsing from Strings](#parsing-from-strings)
  - [Logrus Integration](#logrus-integration)
  - [Level Comparison](#level-comparison)
  - [Configuration Parsing](#configuration-parsing)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
  - [Level Type](#level-type)
  - [Constants](#constants)
  - [Parsing Functions](#parsing-functions)
  - [Conversion Methods](#conversion-methods)
- [Contributing](#contributing)
- [Improvements & Security](#improvements--security)
- [Resources](#resources)
- [AI Transparency](#ai-transparency)
- [License](#license)

---

## Overview

The **level** package provides a type-safe representation of logging severity levels with multiple formats (string, code, integer) and seamless logrus integration. It offers flexible parsing from various input sources while maintaining strict type safety and performance.

### Design Philosophy

1. **Type Safety**: Strong typing with custom Level type prevents invalid values
2. **Framework Compatibility**: Direct conversion to logrus levels for zero-overhead integration
3. **Multiple Representations**: String, integer, and code formats for different use cases
4. **Parse Flexibility**: Case-insensitive parsing from strings and integers
5. **Simplicity**: Minimal API surface with clear semantics and zero dependencies

### Key Features

- ✅ **Type-Safe Levels**: Compile-time safety with custom uint8 type
- ✅ **Multiple Formats**: String ("Info"), Code ("Info"), Integer (4) representations
- ✅ **Logrus Integration**: Direct conversion to logrus.Level
- ✅ **Flexible Parsing**: Case-insensitive string and integer parsing
- ✅ **Level Comparison**: Ordered severity for threshold filtering
- ✅ **Zero Dependencies**: Only Go stdlib and logrus
- ✅ **High Performance**: <10ns for conversions, O(1) operations

---

## Architecture

### Component Diagram

```
┌────────────────────────────────────────────────────────┐
│                    Level Package                       │
├────────────────────────────────────────────────────────┤
│                                                        │
│  Level Type (uint8)                                    │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Constants (Severity Order)                      │  │
│  │  ┌────────────────────────────────────────────┐  │  │
│  │  │  PanicLevel   = 0  (Most severe)           │  │  │
│  │  │  FatalLevel   = 1                          │  │  │
│  │  │  ErrorLevel   = 2                          │  │  │
│  │  │  WarnLevel    = 3                          │  │  │
│  │  │  InfoLevel    = 4  (Default)               │  │  │
│  │  │  DebugLevel   = 5                          │  │  │
│  │  │  NilLevel     = 6  (Disable logging)       │  │  │
│  │  └────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────┘  │
│                                                        │
│  Parsing Functions                                     │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Parse(s string) → Level                         │  │
│  │  ParseFromInt(i int) → Level                     │  │
│  │  ParseFromUint32(u uint32) → Level               │  │
│  │  ListLevels() → []string                         │  │
│  └──────────────────────────────────────────────────┘  │
│                                                        │
│  Conversion Methods                                    │
│  ┌──────────────────────────────────────────────────┐  │
│  │  String() → string    ("Critical", "Info", ...)  │  │
│  │  Code() → string      ("Crit", "Info", ...)      │  │
│  │  Int() → int          (0, 4, ...)                │  │
│  │  Uint8() → uint8      (0, 4, ...)                │  │
│  │  Uint32() → uint32    (0, 4, ...)                │  │
│  │  Logrus() → logrus.Level                         │  │
│  └──────────────────────────────────────────────────┘  │
│                                                        │
└────────────────────────────────────────────────────────┘
```

### Level Hierarchy

Levels are ordered from most severe (0) to least severe (6):

```
┌───────────────────────────────────────────────────────┐
│  Level        │ Value │ String    │ Code   │ Logrus   │
├───────────────┼───────┼───────────┼────────┼──────────┤
│  PanicLevel   │   0   │ Critical  │ Crit   │ Panic    │
│  FatalLevel   │   1   │ Fatal     │ Fatal  │ Fatal    │
│  ErrorLevel   │   2   │ Error     │ Err    │ Error    │
│  WarnLevel    │   3   │ Warning   │ Warn   │ Warning  │
│  InfoLevel    │   4   │ Info      │ Info   │ Info     │
│  DebugLevel   │   5   │ Debug     │ Debug  │ Debug    │
│  NilLevel     │   6   │ (empty)   │ (empty)│ MaxInt32 │
└───────────────────────────────────────────────────────┘
```

**Design Choices:**
- Lower values = higher severity (enables simple `level <= threshold` checks)
- NilLevel (6) disables logging when converted to logrus
- String() returns full name, Code() returns compact form
- Case-insensitive parsing for user-friendly configuration

---

## Performance

### Benchmarks

Performance results from typical usage scenarios (AMD Ryzen 9 7900X3D):

| Operation | Time/op | Throughput | Notes |
|-----------|---------|------------|-------|
| **Parse("info")** | ~10 ns | 100M ops/s | String parsing |
| **ParseFromInt(4)** | ~5 ns | 200M ops/s | Integer parsing |
| **String()** | ~5 ns | 200M ops/s | String conversion |
| **Code()** | ~5 ns | 200M ops/s | Code conversion |
| **Logrus()** | ~5 ns | 200M ops/s | Logrus conversion |
| **Int()** | ~2 ns | 500M ops/s | Integer conversion |

**Key Insights:**
- **Ultra-Fast**: All operations complete in <10ns
- **Zero Allocations**: All operations are stack-based
- **Constant Time**: O(1) for all operations (switch statement)
- **Cache-Friendly**: uint8 type fits in CPU register

### Memory Usage

| Component | Size | Notes |
|-----------|------|-------|
| **Level value** | 1 byte | uint8 storage |
| **Parse result** | 1 byte | No allocation |
| **Conversion** | 0 bytes | Stack-based |
| **Total** | **1 byte** | Per level value |

**Memory Characteristics:**
- ✅ **Minimal Footprint**: 1 byte per level value
- ✅ **Zero Heap**: All operations use stack memory
- ✅ **Cache Efficient**: Fits in CPU registers
- ✅ **No Allocations**: Conversions are allocation-free

### Scalability

**Concurrency:**
- ✅ Thread-safe: All operations are read-only after initialization
- ✅ No synchronization needed: Immutable constants
- ✅ No contention: No shared mutable state
- ✅ Tested: All tests pass with `-race` detector (0 races)

**Performance Characteristics:**
- Parse operations scale linearly with CPU cores
- No lock contention or synchronization overhead
- Suitable for high-throughput logging (millions of ops/s)
- Constant memory usage regardless of scale

---

## Use Cases

### 1. Configuration File Parsing

**Problem**: Parse log levels from YAML/JSON configuration files with user-friendly strings.

**Solution**: Use Parse() for case-insensitive string parsing.

**Advantages**:
- Accepts multiple formats ("info", "INFO", "Info")
- Returns default (InfoLevel) for invalid inputs
- No error handling needed for configuration defaults
- Fast parsing suitable for startup configuration

**Suited for**: Web applications, microservices, CLI tools loading configuration files.

### 2. Dynamic Log Level Changes

**Problem**: Change log levels at runtime based on user input or signals.

**Solution**: Parse user input and update logger configuration dynamically.

**Advantages**:
- Validation through parsing (invalid → default)
- List available levels with ListLevels()
- Immediate effect on logging behavior
- No restart required

**Suited for**: Production systems requiring runtime debugging, admin panels, monitoring tools.

### 3. Log Level Filtering

**Problem**: Filter log messages based on severity threshold.

**Solution**: Compare level values using integer comparison.

**Advantages**:
- Simple comparison: `if currentLevel <= threshold`
- Ordered by severity (0 = most severe)
- Fast integer comparison (<5ns)
- Type-safe operations

**Suited for**: Log aggregators, filtering proxies, monitoring systems requiring threshold-based filtering.

### 4. Logrus Integration

**Problem**: Use standardized level definitions with logrus logger.

**Solution**: Convert Level to logrus.Level using Logrus() method.

**Advantages**:
- Direct mapping to logrus levels
- Zero-overhead conversion
- Type-safe integration
- NilLevel disables logging (returns MaxInt32)

**Suited for**: Applications using logrus for structured logging, libraries exposing logging configuration.

### 5. Multi-Format Level Display

**Problem**: Display levels in different formats (full name, code, integer) for different audiences.

**Solution**: Use String() for humans, Code() for compact output, Int() for storage/APIs.

**Advantages**:
- Multiple representations without duplication
- Consistent conversion from single source
- Flexible output formatting
- Suitable for dashboards, logs, APIs

**Suited for**: Monitoring dashboards, log viewers, REST APIs exposing logging configuration.

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/logger/level
```

**Requirements:**
- Go 1.24 or higher
- Compatible with Linux, macOS, Windows

### Basic Level Usage

Create and use levels:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/logger/level"
)

func main() {
    // Use predefined constants
    currentLevel := level.InfoLevel
    
    fmt.Printf("Level value: %d\n", currentLevel)          // 4
    fmt.Printf("Level string: %s\n", currentLevel.String()) // "Info"
    fmt.Printf("Level code: %s\n", currentLevel.Code())     // "Info"
}
```

### Parsing from Strings

Parse levels from configuration:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/logger/level"
)

func main() {
    // Case-insensitive parsing
    lvl1 := level.Parse("info")      // InfoLevel
    lvl2 := level.Parse("ERROR")     // ErrorLevel
    lvl3 := level.Parse("Critical")  // PanicLevel
    lvl4 := level.Parse("unknown")   // InfoLevel (fallback)
    
    fmt.Printf("Parsed: %s, %s, %s, %s\n",
        lvl1.String(), lvl2.String(), lvl3.String(), lvl4.String())
    
    // List available levels
    levels := level.ListLevels()
    fmt.Printf("Available levels: %v\n", levels)
    // ["critical", "fatal", "error", "warning", "info", "debug"]
}
```

### Logrus Integration

Use with logrus logger:

```go
package main

import (
    "github.com/nabbar/golib/logger/level"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Parse level from config
    configLevel := "debug"
    lvl := level.Parse(configLevel)
    
    // Set logger level
    logger.SetLevel(lvl.Logrus())
    
    // Use logger
    logger.Debug("Debug message")  // Will be logged
    logger.Info("Info message")    // Will be logged
}
```

### Level Comparison

Filter messages by severity:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/logger/level"
)

func shouldLog(messageLevel, threshold level.Level) bool {
    // Lower values = higher severity
    return messageLevel <= threshold
}

func main() {
    threshold := level.InfoLevel
    
    fmt.Printf("Error logged: %v\n", shouldLog(level.ErrorLevel, threshold))  // true
    fmt.Printf("Debug logged: %v\n", shouldLog(level.DebugLevel, threshold))  // false
}
```

### Configuration Parsing

Parse from different sources:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/logger/level"
)

func main() {
    // From string (config file)
    strLevel := level.Parse("warning")
    
    // From integer (database/API)
    intLevel := level.ParseFromInt(2)  // ErrorLevel
    
    // From uint32 (network protocol)
    uintLevel := level.ParseFromUint32(4)  // InfoLevel
    
    fmt.Printf("String: %s, Int: %s, Uint: %s\n",
        strLevel.String(), intLevel.String(), uintLevel.String())
}
```

---

## Best Practices

### Testing

The package includes comprehensive tests with **98.0% code coverage** and **94 test specifications** using BDD methodology (Ginkgo v2 + Gomega).

For detailed test documentation, see **[TESTING.md](TESTING.md)**.

### ✅ DO

**Use Parse() for configuration values**:
```go
// ✅ GOOD: Case-insensitive, defaults to InfoLevel
configLevel := os.Getenv("LOG_LEVEL")
level := level.Parse(configLevel)
logger.SetLevel(level.Logrus())
```

**Use constants for known levels**:
```go
// ✅ GOOD: Type-safe, compile-time checked
currentLevel := level.InfoLevel
if messageLevel <= currentLevel {
    log(message)
}
```

**Use String() for human output**:
```go
// ✅ GOOD: Human-readable
fmt.Printf("Current log level: %s\n", level.InfoLevel.String())  // "Info"
```

**Use Code() for compact display**:
```go
// ✅ GOOD: Compact log prefixes
logEntry := fmt.Sprintf("[%s] %s", level.ErrorLevel.Code(), message)  // "[Err] ..."
```

**Use Int() for storage/comparison**:
```go
// ✅ GOOD: Efficient storage and comparison
config.LogLevel = level.InfoLevel.Int()  // Store as integer
if level.ParseFromInt(config.LogLevel) <= threshold {
    // Compare using integers
}
```

**Check ListLevels() for validation**:
```go
// ✅ GOOD: List valid levels for user
levels := level.ListLevels()
fmt.Printf("Valid levels: %s\n", strings.Join(levels, ", "))
```

### ❌ DON'T

**Don't cast arbitrary integers to Level**:
```go
// ❌ BAD: Unchecked cast
var lvl level.Level = 99  // Invalid value!

// ✅ GOOD: Use parsing functions
lvl := level.ParseFromInt(99)  // Returns InfoLevel (safe default)
```

**Don't expect Parse() to handle whitespace**:
```go
// ❌ BAD: No whitespace trimming
lvl := level.Parse(" info ")  // Returns InfoLevel (fallback)

// ✅ GOOD: Trim before parsing
lvl := level.Parse(strings.TrimSpace(configValue))
```

**Don't try to parse NilLevel from strings**:
```go
// ❌ BAD: NilLevel not parseable from string
lvl := level.Parse("nil")  // Returns InfoLevel (fallback)

// ✅ GOOD: Use constant directly
lvl := level.NilLevel
```

**Don't rely on unknown level behavior in production**:
```go
// ❌ BAD: Assuming unknown returns specific value
lvl := level.Parse(userInput)
// No validation, might get InfoLevel fallback

// ✅ GOOD: Validate input
lvl := level.Parse(userInput)
if lvl.String() == "unknown" {
    return fmt.Errorf("invalid log level: %s", userInput)
}
```

**Don't use String() for code comparison**:
```go
// ❌ BAD: String comparison
if level.String() == "Info" { ... }

// ✅ GOOD: Direct comparison
if level == level.InfoLevel { ... }
```

---

## API Reference

### Level Type

```go
type Level uint8
```

Custom type representing logging severity levels. Ordered from most severe (0) to least severe (6).

**Properties:**
- Underlying type: uint8 (1 byte)
- Comparable: Supports `==`, `!=`, `<`, `<=`, `>`, `>=`
- Serializable: Can be stored as integer
- Thread-safe: Immutable after creation

### Constants

```go
const (
    PanicLevel  Level = 0  // Critical errors causing panic
    FatalLevel  Level = 1  // Fatal errors causing exit
    ErrorLevel  Level = 2  // Error conditions
    WarnLevel   Level = 3  // Warning conditions
    InfoLevel   Level = 4  // Informational messages (default)
    DebugLevel  Level = 5  // Debug-level messages
    NilLevel    Level = 6  // Disable logging
)
```

**Constants Usage:**
- Lower values indicate higher severity
- InfoLevel (4) is the recommended default
- NilLevel (6) disables logging when converted to logrus
- All constants are immutable and thread-safe

### Parsing Functions

**`ListLevels() []string`**
- Returns list of valid level strings
- Excludes NilLevel (not user-parseable)
- Returned slice: `["critical", "fatal", "error", "warning", "info", "debug"]`
- Use for validation, help messages, dropdowns

**`Parse(s string) Level`**
- Parses string to Level (case-insensitive)
- Returns InfoLevel for invalid input (safe default)
- Accepts: full names ("Critical", "Info") and codes ("Crit", "Err")
- No error return (always succeeds with fallback)

**`ParseFromInt(i int) Level`**
- Converts integer to Level
- Valid range: 0-6
- Returns InfoLevel for out-of-range values
- Use for database/API integer representations

**`ParseFromUint32(u uint32) Level`**
- Converts uint32 to Level
- Clamps large values to math.MaxInt
- Returns InfoLevel for out-of-range values
- Use for network protocols, binary formats

### Conversion Methods

**Level Methods:**

```go
func (l Level) Uint8() uint8        // Returns underlying uint8 value
func (l Level) Uint32() uint32      // Returns value as uint32
func (l Level) Int() int            // Returns value as int
func (l Level) String() string      // Returns full name ("Info", "Error", ...)
func (l Level) Code() string        // Returns code ("Info", "Err", ...)
func (l Level) Logrus() logrus.Level // Converts to logrus.Level
```

**Method Details:**

- **Uint8()**: Direct access to underlying value, fastest conversion
- **Uint32()**: For protocols requiring uint32, zero-allocation cast
- **Int()**: For storage in databases, configuration files
- **String()**: Human-readable full name, use for logs, UI
- **Code()**: Compact form, use for log prefixes, narrow displays
- **Logrus()**: Direct mapping to logrus levels, NilLevel → math.MaxInt32

**Return Values by Level:**

| Level | Uint8() | Int() | String() | Code() | Logrus() |
|-------|---------|-------|----------|--------|----------|
| PanicLevel | 0 | 0 | "Critical" | "Crit" | logrus.PanicLevel |
| FatalLevel | 1 | 1 | "Fatal" | "Fatal" | logrus.FatalLevel |
| ErrorLevel | 2 | 2 | "Error" | "Err" | logrus.ErrorLevel |
| WarnLevel | 3 | 3 | "Warning" | "Warn" | logrus.WarnLevel |
| InfoLevel | 4 | 4 | "Info" | "Info" | logrus.InfoLevel |
| DebugLevel | 5 | 5 | "Debug" | "Debug" | logrus.DebugLevel |
| NilLevel | 6 | 6 | "" | "" | math.MaxInt32 |
| Unknown | - | - | "unknown" | "unknown" | logrus.InfoLevel |

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

- ✅ **98.0% test coverage** (target: >80%)
- ✅ **Zero race conditions** detected with `-race` flag
- ✅ **Thread-safe** by design (immutable constants)
- ✅ **Memory-safe** with proper validation
- ✅ **Standard interfaces** for maximum compatibility

### Security Considerations

**No Security Vulnerabilities Identified:**
- No external dependencies (only Go stdlib and logrus)
- No network operations or file system access
- No cryptographic operations
- No user input parsing vulnerabilities (safe defaults)
- Integer parsing prevents overflow (ParseFromUint32 clamps values)

**Best Practices Applied:**
- Input validation with safe defaults (Parse returns InfoLevel for invalid)
- No panic on invalid input
- Immutable constants prevent modification
- Type safety prevents invalid values at compile time
- Defensive nil checks in Logrus() method

### Future Enhancements (Non-urgent)

The following enhancements could be considered for future versions:

**Level Features:**
1. **Trace Level**: Add TraceLevel below DebugLevel for ultra-verbose logging
2. **Custom Levels**: Support for user-defined levels beyond predefined set
3. **Level Ranges**: Parse level ranges ("error-fatal") for filtering
4. **Level Groups**: Predefined level groups (PRODUCTION, DEVELOPMENT)

**Parsing Improvements:**
1. **Whitespace Handling**: Automatic trimming in Parse()
2. **Alias Support**: Additional aliases ("err" → ErrorLevel, "warn" → WarnLevel)
3. **Localization**: Support for non-English level names
4. **Strict Parsing**: Optional ParseStrict() returning error for invalid input

**Integration:**
1. **Zap Integration**: Direct conversion to uber/zap levels
2. **Zerolog Integration**: Direct conversion to zerolog levels
3. **Standard log Integration**: Conversion to standard library log levels
4. **Custom Mappers**: Plugin system for custom logger frameworks

These are **optional improvements** and not required for production use. The current implementation is stable, performant, and feature-complete for its intended use cases.

Suggestions and contributions are welcome via [GitHub issues](https://github.com/nabbar/golib/issues).

---

## Resources

### Package Documentation

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/logger/level)** - Complete API reference with function signatures, method descriptions, and runnable examples. Essential for understanding the public interface, level constants, and conversion methods. Includes detailed examples for parsing, comparison, and logrus integration.

- **[doc.go](doc.go)** - In-depth package documentation including design philosophy (type safety, framework compatibility, multiple representations, parse flexibility, simplicity), architecture diagrams (level hierarchy, data flow), representations (type value, string, code, integer), parsing strategies, logrus integration, use cases (5 detailed scenarios), advantages and limitations, performance considerations, best practices, thread safety, compatibility, and comprehensive examples.

- **[TESTING.md](TESTING.md)** - Comprehensive test suite documentation covering test architecture (test matrix with 94 specs, detailed inventory), BDD methodology with Ginkgo v2, ISTQB alignment (test levels, types, design techniques), testing pyramid distribution, 98.0% coverage analysis with uncovered code justification (ParseFromUint32 edge case), thread safety assurance, and guidelines for writing new tests. Includes troubleshooting, CI integration examples, and bug reporting templates.

### Related golib Packages

- **[github.com/nabbar/golib/logger](https://pkg.go.dev/github.com/nabbar/golib/logger)** - Main logger package that uses level definitions. The level package provides standardized severity levels consumed by the logger for log filtering, threshold comparisons, and configuration. Understanding both packages together enables effective logging implementation with proper level management.

- **[github.com/nabbar/golib/logger/config](https://pkg.go.dev/github.com/nabbar/golib/logger/config)** - Logger configuration structures using level package. The config package references level strings in LogLevel arrays for output filtering. Integration enables per-output level configuration (stdout, files, syslog) with validation and inheritance.

### External References

- **[github.com/sirupsen/logrus](https://pkg.go.dev/github.com/sirupsen/logrus)** - Structured logger for Go. The level package provides seamless integration through direct logrus.Level conversion, enabling structured logging with standardized severity levels. All golib level constants map directly to logrus levels except NilLevel which disables logging.

- **[Effective Go](https://go.dev/doc/effective_go)** - Official Go programming guide covering best practices for type definitions, constants, error handling, and package design. The level package follows these conventions for idiomatic Go code, particularly in type safety, immutable constants, and zero-value handling.

### Community & Support

- **[GitHub Issues](https://github.com/nabbar/golib/issues)** - Report bugs, request features, or ask questions about the level package. Check existing issues before creating new ones to avoid duplicates. Use appropriate labels (bug, enhancement, documentation) for faster triage.

- **[Contributing Guide](../../CONTRIBUTING.md)** - Detailed guidelines for contributing code, tests, and documentation to the project. Includes code style requirements (gofmt, golint), testing procedures (Ginkgo/Gomega, coverage targets >80%), pull request process, and AI usage policy. Essential reading before submitting contributions.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for testing, documentation, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../../../LICENSE) file for details.

Copyright (c) 2021 Nicolas JUHEL

---

**Maintained by**: [Nicolas JUHEL](https://github.com/nabbar)  
**Package**: `github.com/nabbar/golib/logger/level`  
**Version**: See [releases](https://github.com/nabbar/golib/releases) for versioning
