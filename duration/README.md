# Duration Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-blue)](https://golang.org/)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue)](https://pkg.go.dev/github.com/nabbar/golib/duration)

**Extended duration handling with days support, multiple encoding formats, and big integer durations for very large time intervals.**

> **AI Disclaimer (EU AI Act Article 50.4):** AI assistance was used solely for testing, documentation, and bug resolution under human supervision.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Duration Package](#duration-package)
- [Big Duration Package](#big-duration-package)
- [Encoding Support](#encoding-support)
- [Viper Integration](#viper-integration)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [Related Documentation](#related-documentation)
- [License](#license)

---

## Overview

The **duration** package provides two complementary duration types:

1. **Standard Duration** - Extended `time.Duration` with days support
2. **Big Duration** - Large durations beyond `time.Duration` limits (>290 years)

### Design Philosophy

- **Extended Syntax**: Support for days notation (`d`) in duration strings
- **Type Safety**: Compile-time safety with type wrappers
- **Interoperability**: JSON, YAML, TOML, CBOR, and text encoding
- **Flexibility**: Convert between duration types seamlessly
- **Range Generation**: PID controller-based smooth transitions

---

## Key Features

### Standard Duration

| Feature | Description |
|---------|-------------|
| **Extended Parsing** | Parse durations with days (`5d23h15m13s`) |
| **Multiple Formats** | JSON, YAML, TOML, CBOR, text encoding |
| **Viper Integration** | Automatic configuration decoding |
| **Arithmetic** | Add, subtract, multiply, divide durations |
| **Truncation** | Truncate to days, hours, minutes, etc. |
| **Range Generation** | PID controller-based ranges |
| **Thread-Safe** | Safe for concurrent use |

### Big Duration

| Feature | Description |
|---------|-------------|
| **Large Durations** | Support durations >290 years (time.Duration limit) |
| **int64 Seconds** | Based on seconds (not nanoseconds) |
| **Max Duration** | ~106 trillion days (~292 billion years) |
| **Same API** | Compatible API with standard duration |
| **Type Conversion** | Convert to/from time.Duration, int64, float64 |

---

## Architecture

### Package Structure

```
duration/
├── interface.go         # Standard duration (time.Duration wrapper)
├── parse.go            # Parsing logic with days support
├── format.go           # String formatting
├── encode.go           # JSON/YAML/TOML/CBOR encoding
├── operation.go        # Arithmetic operations
├── truncate.go         # Truncation helpers
├── model.go            # Core implementation
│
└── big/                # Big duration sub-package
    ├── interface.go    # Large duration (int64 seconds)
    ├── parse.go        # Parsing for big durations
    ├── format.go       # Formatting for big durations
    ├── encode.go       # Encoding for big durations
    ├── operation.go    # Arithmetic for big durations
    ├── truncate.go     # Truncation for big durations
    └── model.go        # Big duration implementation
```

### Type System

```
┌─────────────────────────────────────────────────────┐
│                 Duration Types                       │
│                                                      │
│  ┌────────────────────┐      ┌────────────────────┐│
│  │  duration.Duration │      │   big.Duration     ││
│  │                    │      │                    ││
│  │  Based on:         │      │  Based on:         ││
│  │  time.Duration     │      │  int64 (seconds)   ││
│  │  (int64 nanos)     │      │                    ││
│  │                    │      │                    ││
│  │  Range:            │      │  Range:            ││
│  │  ±290 years        │      │  ±292B years       ││
│  │                    │      │                    ││
│  │  Precision:        │      │  Precision:        ││
│  │  1 nanosecond      │      │  1 second          ││
│  └────────────────────┘      └────────────────────┘│
│           │                            │            │
│           └────────────┬───────────────┘            │
│                        ▼                             │
│  ┌──────────────────────────────────────────────┐  │
│  │         Common Features                      │  │
│  │  - Days notation (5d23h15m)                 │  │
│  │  - Multiple encodings (JSON, YAML, etc.)    │  │
│  │  - Arithmetic operations                    │  │
│  │  - Truncation & rounding                    │  │
│  │  - PID-based range generation               │  │
│  │  - Viper integration                        │  │
│  └──────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘

Conversion Flow:
  string ←→ Duration ←→ time.Duration ←→ big.Duration
         ←→ JSON/YAML/TOML/CBOR ←→
```

---

## Installation

```bash
# Standard duration
go get github.com/nabbar/golib/duration

# Big duration
go get github.com/nabbar/golib/duration/big
```

---

## Quick Start

### Standard Duration

```go
package main

import (
    "fmt"
    "github.com/nabbar/golib/duration"
)

func main() {
    // Parse duration with days
    d, _ := duration.Parse("5d23h15m13s")
    fmt.Println(d.String())  // Output: 5d23h15m13s
    
    // Create durations
    timeout := duration.Days(2) + duration.Hours(3)
    fmt.Println(timeout.String())  // Output: 2d3h
    
    // Convert to time.Duration
    std := timeout.Time()
    fmt.Println(std)  // Output: 51h0m0s
    
    // Truncate
    days := timeout.TruncateDays()
    fmt.Println(days.String())  // Output: 2d
}
```

### Big Duration

```go
package main

import (
    "fmt"
    durbig "github.com/nabbar/golib/duration/big"
)

func main() {
    // Very large duration (beyond time.Duration limits)
    d := durbig.Days(1000000)  // 1 million days
    fmt.Println(d.String())  // Output: 1000000d
    
    // Parse large duration
    large, _ := durbig.Parse("365000d")  // ~1000 years
    fmt.Println(large.String())
    
    // Arithmetic
    result := large + durbig.Days(100)
    fmt.Println(result.String())
}
```

---

## Duration Package

### Supported Units

```go
const (
    Nanosecond  Duration = 1
    Microsecond          = 1000 * Nanosecond
    Millisecond          = 1000 * Microsecond
    Second               = 1000 * Millisecond
    Minute               = 60 * Second
    Hour                 = 60 * Minute
    Day                  = 24 * Hour  // Extension!
)
```

### Parsing

```go
// Parse various formats
d1, _ := duration.Parse("5h30m")         // 5h30m0s
d2, _ := duration.Parse("2d12h")         // 2d12h0m0s
d3, _ := duration.Parse("1.5h")          // 1h30m0s
d4, _ := duration.Parse("-3h")           // -3h0m0s

// Parse from bytes
d5, _ := duration.ParseByte([]byte("7d"))

// Convert from time.Duration
td := 5 * time.Hour
d6 := duration.ParseDuration(td)
```

### Constructor Functions

```go
d1 := duration.Nanoseconds(1000)    // 1µs
d2 := duration.Microseconds(1000)   // 1ms
d3 := duration.Milliseconds(1000)   // 1s
d4 := duration.Seconds(60)          // 1m
d5 := duration.Minutes(60)          // 1h
d6 := duration.Hours(24)            // 1d
d7 := duration.Days(7)              // 7d (1 week)

// Combine
timeout := duration.Days(1) + duration.Hours(2) + duration.Minutes(30)
```

### Formatting

```go
d := duration.Days(5) + duration.Hours(23) + duration.Minutes(15)

// String representation
fmt.Println(d.String())           // Output: 5d23h15m0s

// Convert to time.Duration
std := d.Time()

// Convert to nanoseconds (int64)
nanos := d.Nanoseconds()
```

### Truncation & Rounding

```go
d := duration.Parse("5d23h45m30s")

// Truncate
days := d.TruncateDays()        // 5d
hours := d.TruncateHours()      // 5d23h
minutes := d.TruncateMinutes()  // 5d23h45m

// Round
rounded := d.Round(duration.Hour)  // 6d (rounds to nearest hour)
```

### Arithmetic Operations

```go
d1 := duration.Hours(5)
d2 := duration.Hours(3)

sum := d1 + d2              // 8h
diff := d1 - d2             // 2h
product := d1 * 2           // 10h
quotient := d1 / 2          // 2h30m

// Comparison
if d1 > d2 {
    fmt.Println("d1 is longer")
}
```

### Range Generation

Generate smooth transitions between durations using PID controller:

```go
start := duration.Seconds(10)
end := duration.Minutes(5)

// With custom PID rates
rateP := 0.1
rateI := 0.01
rateD := 0.001
ranges := start.RangeTo(end, rateP, rateI, rateD)

// With default rates
rangesDef := start.RangeDefTo(end)

for _, d := range rangesDef {
    fmt.Println(d.String())
}
```

---

## Big Duration Package

### Constants

```go
const (
    Second Duration = 1              // 1 second
    Minute          = 60 * Second    // 60 seconds
    Hour            = 60 * Minute    // 3600 seconds
    Day             = 24 * Hour      // 86400 seconds
)

// Maximum: ~106,751,991,167,300 days (~292 billion years)
```

### Why Use Big Duration?

**time.Duration Limitations:**
- Based on int64 nanoseconds
- Maximum: ~290 years
- Overflows for long-term scheduling

**big.Duration Solutions:**
- Based on int64 seconds
- Maximum: ~292 billion years
- Perfect for astronomical calculations
- No nanosecond precision needed

### Creating Big Durations

```go
import durbig "github.com/nabbar/golib/duration/big"

// Constructors
d1 := durbig.Seconds(86400)         // 1 day
d2 := durbig.Minutes(1440)          // 1 day
d3 := durbig.Hours(24)              // 1 day
d4 := durbig.Days(365)              // 1 year

// From float64 (seconds)
d5 := durbig.ParseFloat64(3600.5)   // 1h0m0.5s

// From time.Duration
td := 24 * time.Hour
d6 := durbig.ParseDuration(td)      // 1 day

// Very large
millennium := durbig.Days(365250)   // ~1000 years
```

### Type Conversions

```go
d := durbig.Days(7)

// To int64 (seconds)
seconds := d.Int64()  // 604800

// To uint64 (seconds, 0 if negative)
useconds := d.Uint64()  // 604800

// To float64 (seconds)
fseconds := d.Float64()  // 604800.0

// To time.Duration (if within range)
td, err := d.Duration()
if err != nil {
    // Duration too large for time.Duration
}

// To string
str := d.String()  // "7d"
```

### Parsing

```go
// Parse string
d1, _ := durbig.Parse("365d")        // 1 year
d2, _ := durbig.Parse("1000000d")    // 1 million days
d3, _ := durbig.Parse("5h30m")       // 5h30m
d4, _ := durbig.Parse("2.5h")        // 2h30m (fractional)

// Parse bytes
d5, _ := durbig.ParseByte([]byte("10d"))
```

### Operations

```go
d1 := durbig.Days(100)
d2 := durbig.Days(50)

// Arithmetic
sum := d1 + d2          // 150d
diff := d1 - d2         // 50d
product := d1 * 3       // 300d
quotient := d1 / 2      // 50d

// Comparison
if d1 > d2 {
    fmt.Println("d1 is longer")
}

// Absolute value
abs := durbig.Days(-5).Abs()  // 5d

// Truncation
t := durbig.Parse("7d12h30m")
days := t.TruncateDays()      // 7d
hours := t.TruncateHours()    // 7d12h
```

---

## Encoding Support

Both duration types support multiple encoding formats:

### JSON

```go
type Config struct {
    Timeout duration.Duration `json:"timeout"`
    MaxAge  durbig.Duration   `json:"max_age"`
}

cfg := Config{
    Timeout: duration.Days(1),
    MaxAge:  durbig.Days(365),
}

// Marshal
data, _ := json.Marshal(cfg)
// {"timeout":"1d","max_age":"365d"}

// Unmarshal
var cfg2 Config
json.Unmarshal(data, &cfg2)
```

### YAML

```go
import "gopkg.in/yaml.v3"

type Config struct {
    Timeout duration.Duration `yaml:"timeout"`
}

// Marshal
data, _ := yaml.Marshal(Config{Timeout: duration.Hours(24)})
// timeout: 1d

// Unmarshal
var cfg Config
yaml.Unmarshal(data, &cfg)
```

### TOML

```go
import "github.com/pelletier/go-toml/v2"

type Config struct {
    Timeout duration.Duration `toml:"timeout"`
}

// Works with toml.Marshal/Unmarshal
```

### CBOR

```go
import "github.com/fxamacker/cbor/v2"

type Config struct {
    Timeout duration.Duration `cbor:"timeout"`
}

// Works with cbor.Marshal/Unmarshal
```

### Text Encoding

```go
d := duration.Days(5)

// Marshal
text, _ := d.MarshalText()
// []byte("5d")

// Unmarshal
var d2 duration.Duration
d2.UnmarshalText([]byte("5d"))
```

---

## Viper Integration

Both packages provide decoder hooks for Viper configuration:

```go
import (
    "github.com/spf13/viper"
    "github.com/nabbar/golib/duration"
    durbig "github.com/nabbar/golib/duration/big"
)

type Config struct {
    Timeout  duration.Duration
    MaxAge   durbig.Duration
}

func loadConfig() Config {
    v := viper.New()
    v.SetConfigFile("config.yaml")
    v.ReadInConfig()
    
    var cfg Config
    v.Unmarshal(&cfg, viper.DecodeHook(
        duration.ViperDecoderHook(),
        // Add other hooks as needed
    ))
    
    return cfg
}
```

**config.yaml:**
```yaml
timeout: 2d3h
max_age: 365d
```

---

## Performance

### Memory Characteristics

**Standard Duration:**
- **Size**: 8 bytes (int64)
- **Range**: ±290 years
- **Precision**: 1 nanosecond

**Big Duration:**
- **Size**: 8 bytes (int64)
- **Range**: ±292 billion years
- **Precision**: 1 second

### Benchmarks

| Operation | Standard | Big | Notes |
|-----------|----------|-----|-------|
| Parse | ~500ns | ~500ns | Similar performance |
| String | ~200ns | ~200ns | Same formatting logic |
| JSON Marshal | ~300ns | ~300ns | Encoding overhead |
| JSON Unmarshal | ~800ns | ~800ns | Parsing + validation |
| Arithmetic | ~5ns | ~5ns | Native int64 operations |
| Conversion | ~10ns | ~10ns | Type casting |

*Benchmarks on AMD64, Go 1.21*

### When to Use Which

**Use Standard Duration when:**
- Durations < 290 years
- Need nanosecond precision
- Working with standard library (time.Duration)
- Short-term timers, timeouts, delays

**Use Big Duration when:**
- Durations > 290 years
- Astronomical calculations
- Long-term scheduling (centuries, millennia)
- Second precision is sufficient

---

## Use Cases

### Standard Duration

**HTTP Timeouts**
```go
type ServerConfig struct {
    ReadTimeout  duration.Duration `json:"read_timeout"`
    WriteTimeout duration.Duration `json:"write_timeout"`
    IdleTimeout  duration.Duration `json:"idle_timeout"`
}

cfg := ServerConfig{
    ReadTimeout:  duration.Seconds(10),
    WriteTimeout: duration.Seconds(10),
    IdleTimeout:  duration.Minutes(2),
}

server := &http.Server{
    ReadTimeout:  cfg.ReadTimeout.Time(),
    WriteTimeout: cfg.WriteTimeout.Time(),
    IdleTimeout:  cfg.IdleTimeout.Time(),
}
```

**Cache Expiration**
```go
type CacheConfig struct {
    TTL duration.Duration `yaml:"ttl"`
}

cfg := CacheConfig{
    TTL: duration.Hours(24),
}

cache.Set(key, value, cfg.TTL.Time())
```

**Retry Backoff**
```go
backoff := duration.Seconds(1)
maxBackoff := duration.Minutes(5)

for i := 0; i < maxRetries; i++ {
    if err := doSomething(); err == nil {
        break
    }
    
    time.Sleep(backoff.Time())
    backoff = backoff * 2
    if backoff > maxBackoff {
        backoff = maxBackoff
    }
}
```

### Big Duration

**Astronomy**
```go
// Earth's orbital period
earthYear := durbig.Days(365.25)

// Solar system age (4.6 billion years)
solarSystemAge := durbig.Days(365.25 * 4_600_000_000)

fmt.Println(solarSystemAge.String())
```

**Geological Time**
```go
// Age of dinosaur extinction (66 million years ago)
extinction := durbig.Days(365.25 * 66_000_000)

// Jurassic period duration (56 million years)
jurassic := durbig.Days(365.25 * 56_000_000)
```

**Long-Term Scheduling**
```go
// Archive retention (1000 years)
retention := durbig.Days(365 * 1000)

// Next solar eclipse in 500 years
nextEclipse := durbig.Days(365 * 500)
```

---

## Best Practices

### 1. Choose the Right Type

```go
// ✅ Good: Use standard duration for typical cases
timeout := duration.Seconds(30)

// ✅ Good: Use big duration for very large durations
archiveAge := durbig.Days(10000)

// ❌ Bad: Using big duration unnecessarily
delay := durbig.Seconds(5)  // Overkill, use standard
```

### 2. Use Constructor Functions

```go
// ✅ Good: Type-safe constructors
timeout := duration.Minutes(5)

// ❌ Bad: Manual conversion
timeout := duration.Duration(5 * 60 * time.Second)
```

### 3. Handle Parse Errors

```go
// ✅ Good: Check errors
d, err := duration.Parse(userInput)
if err != nil {
    log.Printf("Invalid duration: %v", err)
    d = duration.Minutes(5)  // Fallback
}

// ❌ Bad: Ignoring errors
d, _ := duration.Parse(userInput)
```

### 4. Use Days Notation

```go
// ✅ Good: Readable
retention := duration.Days(90)

// ❌ Less readable
retention := duration.Hours(24 * 90)
```

### 5. Truncate When Needed

```go
// ✅ Good: Truncate for cleaner output
d := duration.Parse("2d3h15m30s")
fmt.Println(d.TruncateDays().String())  // "2d"

// Context-dependent: Keep precision if needed
precise := d.String()  // "2d3h15m30s"
```

---

## Testing

Comprehensive testing documentation is available in [TESTING.md](TESTING.md).

**Quick Test:**
```bash
cd duration
go test -v -cover

cd big
go test -v -cover
```

**Test Metrics:**
- Standard duration: 93.5% coverage, 150+ specs
- Big duration: 91% coverage, 250+ specs
- Ginkgo v2 + Gomega framework

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass existing tests
- Maintain or improve test coverage
- Follow existing code style

**Documentation**
- Update README.md for new features
- Add examples for common use cases
- Keep TESTING.md synchronized
- Document all public APIs with GoDoc

**Testing**
- Write tests for all new features
- Test edge cases (overflow, underflow, invalid input)
- Verify encoding/decoding works correctly
- Include benchmarks for performance-critical code

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Standard Duration**
- More time units (weeks, months, years)
- Duration arithmetic helpers (min, max, clamp)
- Duration formatting templates
- Localization support for parsing/formatting

**Big Duration**
- Decimal precision (sub-second support)
- Overflow detection and handling
- More conversion helpers (to/from various types)
- Performance optimizations for very large values

**Both Packages**
- Interactive duration calculator
- Visual duration comparison tools
- More encoding formats (MessagePack, etc.)
- Duration validation helpers

Suggestions and contributions are welcome via GitHub issues.

---

## Related Documentation

### Go Standard Library
- **[time](https://pkg.go.dev/time)** - Standard time package
- **[time.Duration](https://pkg.go.dev/time#Duration)** - Standard duration type

### Encoding Libraries
- **[encoding/json](https://pkg.go.dev/encoding/json)** - JSON encoding
- **[gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)** - YAML encoding
- **[github.com/pelletier/go-toml](https://github.com/pelletier/go-toml)** - TOML encoding
- **[github.com/fxamacker/cbor](https://github.com/fxamacker/cbor)** - CBOR encoding

### Configuration
- **[Viper](https://github.com/spf13/viper)** - Configuration management

### Related Golib Packages
- **[config](../config/README.md)** - Configuration management
- **[logger](../logger/README.md)** - Logging (used in examples)

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/duration)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Source Code**: [GitHub Repository](https://github.com/nabbar/golib)

---

*This package is part of the [golib](https://github.com/nabbar/golib) project.*
