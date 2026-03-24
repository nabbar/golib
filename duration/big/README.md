# Big Duration Package

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-86.4%25-brightgreen)](TESTING.md)

The `big` package provides a custom duration type that extends the standard `time.Duration` to support much larger time scales, up to billions of years. It is designed for applications that need to handle very long durations, such as in astronomical calculations, geological time scales, or long-term planning simulations, where the standard nanosecond precision of `time.Duration` limits the maximum representable duration to approximately 290 years.

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
- [Resources](#resources)

---

## Overview

This package introduces `big.Duration`, a type based on `int64` representing seconds (instead of nanoseconds). This trade-off allows for representing durations of up to ~292 billion years. It provides a compatible API with standard Go time utilities while offering extended features like JSON/YAML/TOML serialization, advanced parsing (including days), and arithmetic operations.

### Design Philosophy

The design focuses on:

1. **Extended Range**: Prioritizing duration magnitude over nanosecond precision.
2. **Compatibility**: Mimicking `time.Duration` methods and behavior where possible.
3. **Usability**: Providing robust parsing strings (e.g., "100d", "5y" contextually) and serialization support.

### Key Features

- ✅ **Huge Range**: Supports durations up to ~292 billion years.
- ✅ **Standard Units**: Supports parsing and formatting of days (`d`), hours (`h`), minutes (`m`), and seconds (`s`).
- ✅ **Serialization**: Built-in support for `JSON`, `YAML`, `TOML`, and `CBOR`.
- ✅ **Arithmetic**: Includes helper functions for rounding, truncation, and absolute values.
- ✅ **Integration**: Easy conversion to/from `time.Duration` (with overflow checks).

### Key Benefits

- ✅ **Long-term Scheduling**: Handle dates and durations far beyond the 290-year limit of standard Go.
- ✅ **Simplified Config**: Use "days" directly in configuration files (e.g., "30d" for retention policies).
- ❌ **Precision Trade-off**: Minimum resolution is 1 second (no milliseconds/nanoseconds).

---

## Architecture

### Package Structure

```
big/
├── doc.go                   # Package documentation
├── model.go                 # Core type definition and basic helpers
├── interface.go             # Public interface and factory functions
├── parse.go                 # String parsing logic
├── format.go                # String formatting and conversion
├── encode.go                # Serialization (JSON, YAML, TOML, CBOR)
├── operation.go             # Math operations (Abs, Range)
├── truncate.go              # Rounding and truncation logic
├── examples_test.go         # Runnable examples
├── big_test.go              # Integration tests
└── ..._test.go              # Unit tests for each file
```

### Package Architecture

The core of the package is the `Duration` type, which is an alias for `int64`. All operations revolve around manipulating this integer value, which represents the number of seconds.

Parsing and formatting logic interprets string representations (e.g., "1d2h") and converts them to/from this seconds-based integer. Serialization hooks interface with standard libraries to allow seamless encoding/decoding.

### Dataflow

1. **Input**: String ("1h"), Integer (3600), or Byte Slice.
2. **Parsing**: Converted to `big.Duration` (int64 seconds).
3. **Storage/Manipulation**: Stored as int64; supports math operations.
4. **Output**: Formatted String ("1h0m0s"), Serialized Data (JSON/YAML), or converted standard types (`time.Duration`, `float64`).

### Data Explain

The `big.Duration` type is a simple scalar value in memory.

```go
type Duration int64
```

- **Storage**: 8 bytes (64-bit integer).
- **Unit**: Seconds.
- **Range**: [-9,223,372,036,854,775,808, 9,223,372,036,854,775,807] seconds.
- **Conversion**:
    - To `time.Duration`: Multiplied by `time.Second` (1,000,000,000). Check for overflow.
    - To `float64`: Direct conversion, precision loss possible for very large values (> 2^53).

---

## Performance

The package is designed to be lightweight. Operations are primarily simple integer arithmetic. Parsing involves string manipulation which is optimized but naturally heavier than raw integer math.

| Operation   | Time/Op  | Alloc/Op |
|-------------|----------|----------|
| `String()`  | ~782 ns  | 96 B     |
| `Parse()`   | ~226 ns  | 0 B      |
| `IsDays()`  | ~0.35 ns | 0 B      |
| `Round()`   | ~11.6 ns | 0 B      |

*Note: `String()` formatting involves allocations due to string concatenation. Benchmarks run on Intel(R) Core(TM) i7-4700HQ CPU @ 2.40GHz.*

---

## Use Cases

### 1. Data Retention Policies

Defining long-term data retention policies in configuration files.

```go
// Config struct
type Config struct {
    Retention big.Duration `json:"retention"`
}

// JSON: {"retention": "3650d"} // 10 years
```

### 2. Astronomical Simulations

Modeling events that occur over millennia.

```go
orbitPeriod := big.Days(90000) // ~246 years (Pluto)
```

### 3. Long-running Token Expiration

Setting expiration times for tokens that should last for decades (e.g., root CA certificates).

```go
caValidity := big.Days(365 * 20) // 20 years
```

---

## Quick Start

### Installation

```bash
go get github.com/nabbar/golib/duration/big
```

### Basic Implementation

```go
package main

import (
    "fmt"
	
    "github.com/nabbar/golib/duration/big"
)

func main() {
    // Parse from string
    d, _ := big.Parse("1d2h")
    fmt.Printf("Duration: %s\n", d) // Output: 1d2h

    // Create from units
    d2 := big.Days(10) + big.Hours(5)
    fmt.Printf("Seconds: %d\n", d2.Int64())
}
```

### Viper Configuration Implementation

Integration with `spf13/viper` for configuration loading.

```go
import (
    "github.com/spf13/viper"
    durbig "github.com/nabbar/golib/duration/big"
)

type Config struct {
    Timeout durbig.Duration `mapstructure:"timeout"`
}

func loadConfig() {
    // Register the decoder hook
    viper.Unmarshal(&config, viper.DecodeHook(durbig.ViperDecoderHook()))
}
```

---

## Best Practices

### ✅ DO
- Use `big.Duration` for durations > 290 years or when "days" are a primary unit.
- Check errors when parsing user input.
- Use `Truncate` or `Round` to clean up calculated values.
- Use `ViperDecoderHook` when working with `spf13/viper`.

### ❌ DON'T
- Use `big.Duration` if you need sub-second precision (use `time.Duration`).
- Ignore overflow errors when converting *to* `time.Duration`.
- Rely on `float64` conversion for precise equality checks on very large values.

---

## API Reference

### Core Type

| Type       | Description                              |
|------------|------------------------------------------|
| `Duration` | Underlying `int64` representing seconds. |

### Factory Functions

| Function  | Parameters | Result            | Description                      |
|-----------|------------|-------------------|----------------------------------|
| `Parse`   | `string`   | `Duration, error` | Parses a string like "1d2h".     |
| `Days`    | `int64`    | `Duration`        | Creates a duration of N days.    |
| `Hours`   | `int64`    | `Duration`        | Creates a duration of N hours.   |
| `Minutes` | `int64`    | `Duration`        | Creates a duration of N minutes. |
| `Seconds` | `int64`    | `Duration`        | Creates a duration of N seconds. |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
    - Follow Go best practices and idioms
    - Maintain or improve code coverage (target: >80%)
    - Pass all tests including race detector
    - Use `gofmt`, `golangci-lint` and `gosec`

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
    - Maintain coverage above 80%

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

---

## Resources

- **[GoDoc](https://pkg.go.dev/github.com/nabbar/golib/duration/big)** - Full API documentation.
- **[TESTING.md](TESTING.md)** - Comprehensive testing guide.

### Related golib Packages

- **[github.com/nabbar/golib/pidcontroller](https://pkg.go.dev/github.com/nabbar/golib/pidcontroller)** - Used for range generation logic.

### External References

- **[Go time Package](https://pkg.go.dev/time)** - Standard library reference.
- **[Ginkgo Framework](https://onsi.github.io/ginkgo/)** - BDD Testing Framework.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for documentation, testing, and bug resolution under human supervision. All core functionality is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022 Nicolas JUHEL
