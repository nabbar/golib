# Size Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Type-safe, human-readable size representation library for Go with parsing, formatting, arithmetic operations, and comprehensive marshaling support.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This library provides production-ready size value handling for Go applications. It emphasizes type safety, overflow protection, and flexibility while supporting multiple serialization formats and integration with popular configuration systems.

### Design Philosophy

1. **Type-Safe**: Size wraps `uint64` to prevent accidental misuse as plain integers
2. **Overflow-Protected**: All arithmetic operations detect and handle overflow/underflow
3. **Human-Readable**: Automatic formatting with appropriate units (KB, MB, GB, etc.)
4. **Format-Agnostic**: Marshaling/unmarshaling for JSON, YAML, TOML, CBOR, and plain text
5. **Configuration-Ready**: Native integration with Viper through decode hooks

---

## Key Features

- **Flexible Parsing**: Parse size strings like "10MB", "1.5GB", "1GB500MB" with automatic unit detection
- **Safe Arithmetic**: Multiply, divide, add, subtract with overflow/underflow protection
- **Multiple Units**: Support for B, KB, MB, GB, TB, PB, EB (binary powers of 1024)
- **Rich Formatting**: Convert to various numeric types with overflow detection
- **Universal Marshaling**: JSON, YAML, TOML, CBOR, binary, and text encoding
- **Viper Integration**: Seamless configuration file parsing with `ViperDecoderHook()`
- **Zero Dependencies**: Only relies on standard library (except Viper hook)

---

## Installation

```bash
go get github.com/nabbar/golib/size
```

---

## Architecture

### Type Structure

```
Size (uint64)
│
├── Parsing      → Parse(), ParseByte(), ParseInt64(), ParseFloat64()
├── Formatting   → String(), Format(), Unit()
├── Arithmetic   → Mul(), Div(), Add(), Sub() (with overflow protection)
├── Conversion   → Int64(), Uint64(), Float64(), KiloBytes(), etc.
└── Marshaling   → JSON, YAML, TOML, CBOR, Text, Binary
```

### Component Overview

```
┌─────────────────────────────────────────────────────┐
│                   Size Package                      │
│  Parse, Format, Arithmetic, Marshal/Unmarshal       │
└───────┬──────────┬───────────┬──────────┬───────────┘
        │          │           │          │
   ┌────▼───┐ ┌────▼───┐ ┌─────▼────┐ ┌───▼─────────┐
   │ Parser │ │ Format │ │Arithmetic│ │  Encoding   │
   │        │ │        │ │          │ │             │
   │ String │ │ Human  │ │ Overflow │ │ JSON, YAML  │
   │ Numeric│ │ Machine│ │  Safety  │ │ TOML, CBOR  │
   └────────┘ └────────┘ └──────────┘ └─────────────┘
```

| Component | Purpose | Safety | Example |
|-----------|---------|--------|---------|
| **Parser** | Convert strings/numbers to Size | Overflow checks | `Parse("10MB")` |
| **Formatter** | Convert Size to strings/numbers | Overflow detection | `size.String()` |
| **Arithmetic** | Math operations on Size values | Error on overflow | `size.MulErr(2.0)` |
| **Encoding** | Serialize/deserialize | Type-safe | `json.Marshal(size)` |

### Unit System

The package uses **binary prefixes** (powers of 1024):

```
B   = 1
KB  = 1024¹  = 1,024
MB  = 1024²  = 1,048,576
GB  = 1024³  = 1,073,741,824
TB  = 1024⁴  = 1,099,511,627,776
PB  = 1024⁵  = 1,125,899,906,842,624
EB  = 1024⁶  = 1,152,921,504,606,846,976
```

**Note**: This follows the traditional computing convention (1 KB = 1024 bytes), not the SI decimal system (1 kB = 1000 bytes).

---

## Performance

### Memory Efficiency

- **Size Type**: 8 bytes (underlying `uint64`)
- **Zero Allocation**: Parsing and formatting operations minimize heap allocations
- **Value Semantics**: Pass by value is efficient; methods use pointer receivers only when modifying

### Parsing Performance

```
BenchmarkParse          5000000    250 ns/op     32 B/op    2 allocs/op
BenchmarkParseComplex   2000000    650 ns/op     96 B/op    4 allocs/op
BenchmarkFormat         10000000   120 ns/op     48 B/op    2 allocs/op
```

*Benchmarks on AMD64, Go 1.21*

### Thread Safety

- **Value Type**: Size is a simple value type and safe to copy
- **Concurrent Reads**: Safe across goroutines
- **Concurrent Writes**: Use explicit synchronization for pointer receivers (`Mul`, `Add`, etc.)

---

## Use Cases

This library is designed for scenarios requiring type-safe size handling:

**Configuration Files**
- Parse size limits from YAML/TOML/JSON configurations
- Human-readable format (e.g., `max_upload: "100MB"`)
- Type-safe validation with Viper integration

**Resource Management**
- Memory allocation limits
- File size restrictions
- Quota enforcement with overflow protection

**CLI Applications**
- Parse user input for size flags (e.g., `--max-size 1.5GB`)
- Display progress with formatted output
- Validate size constraints

**Storage Systems**
- Calculate disk usage with safe arithmetic
- Compare file sizes with type safety
- Aggregate storage metrics

**Logging & Monitoring**
- Format byte counts in human-readable form
- Track memory usage over time
- Alert on size thresholds

---

## Quick Start

### Basic Parsing

Parse size strings with automatic unit detection:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/size"
)

func main() {
    // Parse from string
    s, err := size.Parse("10MB")
    if err != nil {
        panic(err)
    }
    fmt.Println(s.String())        // Output: "10.00 MB"
    fmt.Println(s.Uint64())        // Output: 10485760
    fmt.Println(s.MegaBytes())     // Output: 10
}
```

### Arithmetic Operations

Perform safe arithmetic with overflow detection:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/size"
)

func main() {
    s := size.ParseUint64(1024 * 1024) // 1 MB
    
    // Multiply (with error checking)
    err := s.MulErr(2.5)
    if err != nil {
        fmt.Println("Overflow detected!")
    }
    fmt.Println(s.String()) // Output: "2.50 MB"
    
    // Add
    s.Add(512 * 1024) // Add 512 KB
    fmt.Println(s.String()) // Output: "3.00 MB"
    
    // Subtract
    s.Sub(1024 * 1024) // Subtract 1 MB
    fmt.Println(s.String()) // Output: "2.00 MB"
}
```

### JSON Marshaling

Automatic serialization to human-readable format:

```go
package main

import (
    "encoding/json"
    "fmt"
    "github.com/nabbar/golib/size"
)

type Config struct {
    MaxUpload size.Size `json:"max_upload"`
    MaxFile   size.Size `json:"max_file"`
}

func main() {
    cfg := Config{
        MaxUpload: size.ParseUint64(100 * 1024 * 1024), // 100 MB
        MaxFile:   size.ParseUint64(10 * 1024 * 1024),  // 10 MB
    }
    
    // Marshal to JSON
    data, _ := json.MarshalIndent(cfg, "", "  ")
    fmt.Println(string(data))
    // Output:
    // {
    //   "max_upload": "100.00 MB",
    //   "max_file": "10.00 MB"
    // }
    
    // Unmarshal from JSON
    input := `{"max_upload": "200MB", "max_file": "20MB"}`
    var newCfg Config
    json.Unmarshal([]byte(input), &newCfg)
    fmt.Println(newCfg.MaxUpload.MegaBytes()) // Output: 200
}
```

### Viper Integration

Seamless configuration parsing with Viper:

```go
package main

import (
    "github.com/nabbar/golib/size"
    "github.com/spf13/viper"
    libmap "github.com/go-viper/mapstructure/v2"
)

type ServerConfig struct {
    MaxRequestSize  size.Size `mapstructure:"max_request_size"`
    MaxUploadSize   size.Size `mapstructure:"max_upload_size"`
    CacheSize       size.Size `mapstructure:"cache_size"`
}

func main() {
    v := viper.New()
    v.SetConfigFile("config.yaml")
    v.ReadInConfig()
    
    var cfg ServerConfig
    err := v.Unmarshal(&cfg, viper.DecodeHook(
        libmap.ComposeDecodeHookFunc(
            size.ViperDecoderHook(),
            libmap.StringToTimeDurationHookFunc(),
        ),
    ))
    
    if err != nil {
        panic(err)
    }
    
    // Now cfg.MaxRequestSize is properly parsed
    fmt.Println(cfg.MaxRequestSize.MegaBytes())
}
```

**config.yaml**:
```yaml
max_request_size: "10MB"
max_upload_size: "100MB"
cache_size: 1073741824  # Can use numeric values too
```

### Complex Parsing

Parse complex size expressions:

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/size"
)

func main() {
    // Multiple units in one string
    s, _ := size.Parse("1GB500MB")
    fmt.Println(s.String()) // Output: "1.49 GB"
    
    // Fractional values
    s, _ = size.Parse("2.5TB")
    fmt.Println(s.TeraBytes()) // Output: 2
    
    // Different unit variations
    examples := []string{
        "10MB", "10Mb", "10mb", "10M",  // All parse to 10 megabytes
        "1.5GB", "1GB500MB",             // Same value
    }
    
    for _, ex := range examples {
        s, _ := size.Parse(ex)
        fmt.Printf("%-12s = %d bytes\n", ex, s.Uint64())
    }
}
```

---

## API Reference

### Parsing Functions

```go
// Parse from string (most common)
size, err := size.Parse("10MB")

// Parse from byte slice
size, err := size.ParseByte([]byte("10MB"))

// Convert from numeric types
size := size.ParseUint64(1048576)
size := size.ParseInt64(-1048576)   // Converts to absolute value
size := size.ParseFloat64(1048576.5)
```

### Arithmetic Operations

All arithmetic operations have two variants:
- **Non-error**: Ignores overflow (e.g., `Mul()`)
- **Error-returning**: Returns error on overflow (e.g., `MulErr()`)

```go
size := size.ParseUint64(1024)

// Multiplication
size.Mul(2.0)              // size *= 2
err := size.MulErr(2.0)    // Returns error on overflow

// Division
size.Div(2.0)              // size /= 2
err := size.DivErr(2.0)    // Returns error if divisor ≤ 0

// Addition
size.Add(512)              // size += 512
err := size.AddErr(512)    // Returns error on overflow

// Subtraction
size.Sub(256)              // size -= 256
err := size.SubErr(256)    // Returns error on underflow
```

### Formatting

```go
size := size.ParseUint64(1572864) // 1.5 MB

// Human-readable string
size.String()                   // "1.50 MB"
size.Format(size.FormatRound0)  // "2"
size.Format(size.FormatRound1)  // "1.5"
size.Format(size.FormatRound2)  // "1.50"
size.Format(size.FormatRound3)  // "1.500"

// Get unit only
size.Unit('B')  // "MB"
size.Unit('o')  // "Mo" (useful for French/other locales)
```

### Type Conversion

```go
size := size.ParseUint64(1048576)

// Integer types (with overflow protection)
size.Int64()   // 1048576
size.Int32()   // 1048576 (or MaxInt32 if too large)
size.Int()     // 1048576
size.Uint64()  // 1048576
size.Uint32()  // 1048576 (or MaxUint32 if too large)
size.Uint()    // 1048576

// Float types (with overflow protection)
size.Float64() // 1048576.0
size.Float32() // 1048576.0

// Unit-specific conversions
size.KiloBytes() // 1024
size.MegaBytes() // 1
size.GigaBytes() // 0
```

### Constants

```go
size.SizeNul   // 0
size.SizeUnit  // 1 B
size.SizeKilo  // 1024 B
size.SizeMega  // 1048576 B
size.SizeGiga  // 1073741824 B
size.SizeTera  // 1099511627776 B
size.SizePeta  // 1125899906842624 B
size.SizeExa   // 1152921504606846976 B
```

### Format Constants

```go
size.FormatRound0 // "%.0f" - no decimal places
size.FormatRound1 // "%.1f" - 1 decimal place
size.FormatRound2 // "%.2f" - 2 decimal places (default)
size.FormatRound3 // "%.3f" - 3 decimal places
```

### Marshaling

The Size type implements multiple encoding interfaces:

```go
// JSON
json.Marshal(size)     // Outputs: "1.50 MB"
json.Unmarshal(data, &size)

// YAML
yaml.Marshal(size)     // Outputs: 1.50 MB
yaml.Unmarshal(data, &size)

// TOML
toml.Marshal(size)
toml.Unmarshal(data, &size)

// Text encoding
size.MarshalText()
size.UnmarshalText([]byte("10MB"))

// Binary (CBOR)
size.MarshalBinary()
size.UnmarshalBinary(data)
```

---

## Best Practices

**Use Parse() for String Input**
```go
// ✅ Good: Parse and handle errors
func setLimit(input string) error {
    limit, err := size.Parse(input)
    if err != nil {
        return fmt.Errorf("invalid size: %w", err)
    }
    server.SetLimit(limit)
    return nil
}

// ❌ Bad: No validation
func setLimitBad(input string) {
    limit, _ := size.Parse(input)
    server.SetLimit(limit)
}
```

**Check Arithmetic Errors**
```go
// ✅ Good: Check for overflow
func calculateTotal(sizes []size.Size) (size.Size, error) {
    total := size.SizeNul
    for _, s := range sizes {
        if err := total.AddErr(s.Uint64()); err != nil {
            return 0, fmt.Errorf("size overflow: %w", err)
        }
    }
    return total, nil
}

// ❌ Bad: Silent overflow
func calculateTotalBad(sizes []size.Size) size.Size {
    total := size.SizeNul
    for _, s := range sizes {
        total.Add(s.Uint64())
    }
    return total
}
```

**Use Type Conversion Safely**
```go
// ✅ Good: Check for overflow potential
func toInt32(s size.Size) (int32, error) {
    if s.Uint64() > math.MaxInt32 {
        return 0, errors.New("size too large for int32")
    }
    return s.Int32(), nil
}

// ❌ Bad: Assume no overflow
func toInt32Bad(s size.Size) int32 {
    return s.Int32() // May return MaxInt32 silently
}
```

**Consistent Units in Configuration**
```go
// ✅ Good: Human-readable config
type Config struct {
    MaxFileSize   size.Size `json:"max_file_size"`   // User can write "100MB"
    MaxUploadSize size.Size `json:"max_upload_size"` // User can write "1GB"
}

// ❌ Bad: Raw byte counts
type ConfigBad struct {
    MaxFileSize   uint64 `json:"max_file_size"`   // User must calculate 104857600
    MaxUploadSize uint64 `json:"max_upload_size"` // User must calculate 1073741824
}
```

**Marshal to Human-Readable Format**
```go
// ✅ Good: Easy to read
type APIResponse struct {
    TotalSize     size.Size `json:"total_size"`
    AvailableSize size.Size `json:"available_size"`
}
// JSON output: {"total_size": "10.50 GB", "available_size": "5.25 GB"}

// ❌ Bad: Requires client-side formatting
type APIResponseBad struct {
    TotalSize     uint64 `json:"total_size"`
    AvailableSize uint64 `json:"available_size"`
}
// JSON output: {"total_size": 11274289152, "available_size": 5637144576}
```

---

## Testing

**Test Suite**: 352 specs using Ginkgo v2 and Gomega (95.4% coverage)

```bash
# Run tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Coverage Areas**
- Parsing (strings, numbers, complex expressions)
- Arithmetic operations (overflow/underflow handling)
- Type conversions (all numeric types)
- Formatting (various precision levels)
- Marshaling (JSON, YAML, TOML, CBOR, Text, Binary)
- Viper integration (decode hook with multiple types)
- Edge cases (maximum values, zero, negative inputs)

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ 95.4% code coverage
- ✅ All overflow/underflow scenarios tested
- ✅ Cross-format marshaling verified

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥95%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify overflow/underflow protection
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Unit Systems**
- SI decimal prefixes (kB, MB, GB = powers of 1000)
- IEC binary prefixes (KiB, MiB, GiB = powers of 1024) with explicit notation
- User-selectable unit system

**Parsing Features**
- Negative size values with explicit semantics
- Unit aliases (e.g., "megabytes", "megs")
- Localized unit names
- Scientific notation support (e.g., "1e6")

**Formatting Options**
- Custom format strings
- Locale-aware number formatting
- Compact notation (e.g., "10M" instead of "10.00 MB")
- Adaptive precision based on magnitude

**Arithmetic**
- Checked arithmetic mode (panic on overflow)
- Saturation arithmetic mode (clamp to max/min)
- Fixed-point arithmetic for precise fractional operations

**Performance**
- Zero-allocation parsing for common patterns
- String interning for frequently used values
- SIMD-accelerated parsing

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/size)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../CONTRIBUTING.md)
- **Related Packages**:
  - [duration](../duration) - Time duration handling with similar API
  - [duration/big](../duration/big) - Arbitrary-precision durations
  - [viper](../viper) - Configuration management helpers
  - [config](../config) - Complete configuration system
