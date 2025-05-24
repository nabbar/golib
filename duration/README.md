# `duration` Package

The `duration` package provides an extended and user-friendly wrapper around Go's `time.Duration`, supporting parsing, formatting, encoding/decoding, and arithmetic operations with additional units (including days). It is designed for easy integration with JSON, YAML, TOML, CBOR, and Viper, and offers helper functions for common duration manipulations.

A `big` subpackage is also available for handling very large durations that exceed the limits of `time.Duration`.<br />See the [`big` subpackage documentation](#big-subpackage) for more details.

## Features

- Extended duration type supporting days (`d`), hours, minutes, seconds, milliseconds, microseconds, nanoseconds
- Parse and format durations as strings (e.g., `5d23h15m13s`)
- Marshal/unmarshal support for JSON, YAML, TOML, CBOR, and text
- Viper decoder hook for configuration loading
- Helper functions for creating durations from days, hours, minutes, etc.
- Truncation helpers (to days, hours, minutes, etc.)
- Range generation using PID controller logic (for smooth transitions)
- Thread-safe and compatible with Go generics

---

## Main Types & Functions

### `Duration` Type

A custom type based on `time.Duration`:

```go
type Duration time.Duration
```

### Creating Durations

```go
import "github.com/nabbar/golib/duration"

d1 := duration.Days(2)
d2 := duration.Hours(5)
d3 := duration.ParseDuration(time.Hour * 3)
d4, err := duration.Parse("1d2h30m")
```

### Parsing and Formatting

- `Parse(s string) (Duration, error)`: Parse a duration string (supports `d`, `h`, `m`, `s`, etc.)
- `ParseByte([]byte) (Duration, error)`: Parse from byte slice
- `String() string`: Format as string (e.g., `2d5h`)
- `Time() time.Duration`: Convert to standard Go duration

### Encoding/Decoding

Supports JSON, YAML, TOML, CBOR, and text:

```go
type Example struct {
    Value duration.Duration `json:"value" yaml:"value" toml:"value"`
}

// Marshal to JSON
b, _ := json.Marshal(Example{Value: duration.Days(1) + duration.Hours(2)})
// Unmarshal from YAML, TOML, etc.
```

### Viper Integration

Use the provided decoder hook for Viper:

```go
import "github.com/nabbar/golib/duration"

cfg := viper.New()
cfg.Set("timeout", "2d3h")
var d duration.Duration
cfg.UnmarshalKey("timeout", &d, viper.DecodeHook(duration.ViperDecoderHook()))
```

### Truncation Helpers

- `TruncateDays()`, `TruncateHours()`, `TruncateMinutes()`, etc.

### Range Generation

Generate a range of durations between two values using PID controller logic:

```go
r := d1.RangeTo(d2, rateP, rateI, rateD)
rDef := d1.RangeDefTo(d2) // uses default PID rates
```

---

## Error Handling

All parsing and conversion functions return standard Go `error` values. Always check errors when parsing or decoding durations.

---

## Notes

- Duration strings support days (`d`), which is not available in Go's standard library.
- The package is compatible with Go 1.18+ and supports generics.
- Integrates with `github.com/go-viper/mapstructure/v2` for configuration decoding.

---

## `big` Subpackage

The `big` subpackage provides an extended duration type supporting very large time intervals, including days, and offers advanced parsing, formatting, encoding/decoding, arithmetic, and range generation. It is designed for scenarios where Go's standard `time.Duration` is insufficient due to its limited range.

### Features

- Extended duration type (`Duration`) supporting days (`d`), hours, minutes, seconds
- Parse and format durations as strings (e.g., `5d23h15m13s`)
- Marshal/unmarshal support for JSON, YAML, TOML, CBOR, and text
- Viper decoder hook for configuration loading
- Helper functions for creating durations from days, hours, minutes, seconds
- Truncation and rounding helpers (to days, hours, minutes)
- Range generation using PID controller logic (for smooth transitions)
- Thread-safe and compatible with Go generics

---

### Main Types & Functions

#### `Duration` Type

A custom type based on `int64`, supporting very large values:

```go
type Duration int64
```

#### Creating Durations

```go
import "github.com/nabbar/golib/duration/big"

d1 := big.Days(2)
d2 := big.Hours(5)
d3 := big.ParseDuration(time.Hour * 3)
d4, err := big.Parse("1d2h30m")
```

#### Parsing and Formatting

- `Parse(s string) (Duration, error)`: Parse a duration string (supports `d`, `h`, `m`, `s`)
- `ParseByte([]byte) (Duration, error)`: Parse from byte slice
- `String() string`: Format as string (e.g., `2d5h`)
- `Time() (time.Duration, error)`: Convert to standard Go duration (with overflow check)

#### Encoding/Decoding

Supports JSON, YAML, TOML, CBOR, and text:

```go
type Example struct {
    Value big.Duration `json:"value" yaml:"value" toml:"value"`
}

// Marshal to JSON
b, _ := json.Marshal(Example{Value: big.Days(1) + big.Hours(2)})
// Unmarshal from YAML, TOML, etc.
```

#### Viper Integration

Use the provided decoder hook for Viper:

```go
import "github.com/nabbar/golib/duration/big"

cfg := viper.New()
cfg.Set("timeout", "2d3h")
var d big.Duration
cfg.UnmarshalKey("timeout", &d, viper.DecodeHook(big.ViperDecoderHook()))
```

#### Truncation & Rounding Helpers

- `TruncateDays()`, `TruncateHours()`, `TruncateMinutes()`
- `Round(unit Duration) Duration`

#### Range Generation

Generate a range of durations between two values using PID controller logic:

```go
r := d1.RangeTo(d2, rateP, rateI, rateD)
rDef := d1.RangeDefTo(d2) // uses default PID rates
```

---

### Error Handling

All parsing and conversion functions return standard Go `error` values. Always check errors when parsing or decoding durations.

---

### Notes

- Duration strings support days (`d`), which is not available in Go's standard library.
- The package is compatible with Go 1.18+ and supports generics.
- Integrates with `github.com/go-viper/mapstructure/v2` for configuration decoding.
- Maximum supported value: `106,751,991,167,300d15h30m7s`.

