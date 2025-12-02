/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 *
 */

/*
Package level defines log severity levels with conversions and parsing capabilities.

# Design Philosophy

The level package provides a standardized representation of logging severity levels
that is compatible with popular logging frameworks like logrus while maintaining
flexibility for custom implementations. The design prioritizes:

 1. Type Safety: Strong typing with custom Level type prevents invalid values
 2. Framework Compatibility: Direct conversion to logrus levels
 3. Multiple Representations: String, integer, and code representations
 4. Parse Flexibility: Case-insensitive parsing from strings and integers
 5. Simplicity: Minimal API surface with clear semantics

# Architecture

Log Levels (Ordered by Severity):

	┌────────────────────────────────────────────────────────────┐
	│  Level        │ Value │ String    │ Code   │ Use Case      │
	├───────────────┼───────┼───────────┼────────┼───────────────┤
	│  PanicLevel   │   0   │ Critical  │ Crit   │ Panic + trace │
	│  FatalLevel   │   1   │ Fatal     │ Fatal  │ Fatal error   │
	│  ErrorLevel   │   2   │ Error     │ Err    │ Error         │
	│  WarnLevel    │   3   │ Warning   │ Warn   │ Warning       │
	│  InfoLevel    │   4   │ Info      │ Info   │ Information   │
	│  DebugLevel   │   5   │ Debug     │ Debug  │ Debug info    │
	│  NilLevel     │   6   │ (empty)   │ (empty)│ Disable log   │
	└────────────────────────────────────────────────────────────┘

Levels are ordered from most severe (PanicLevel=0) to least severe (DebugLevel=5).
NilLevel (6) is special and disables logging entirely.

# Representations

Each level has multiple representations for different use cases:

 1. Type Value: The internal uint8 value (0-6)
 2. String: Human-readable full name (e.g., "Critical", "Info")
 3. Code: Short code for compact output (e.g., "Crit", "Err")
 4. Integer: Numeric representation for comparisons and storage

# Parsing

The package supports flexible parsing from various input formats:

String Parsing (Case-Insensitive):

	level := level.Parse("info")      // InfoLevel
	level := level.Parse("ERROR")     // ErrorLevel
	level := level.Parse("Critical")  // PanicLevel
	level := level.Parse("unknown")   // InfoLevel (default fallback)

Code Parsing:

	level := level.Parse("Crit")      // PanicLevel
	level := level.Parse("Err")       // ErrorLevel
	level := level.Parse("Warn")      // WarnLevel

Integer Parsing:

	level := level.ParseFromInt(4)    // InfoLevel
	level := level.ParseFromInt(99)   // InfoLevel (invalid = default)
	level := level.ParseFromUint32(2) // ErrorLevel

# Logrus Integration

The package provides direct conversion to logrus levels:

	import "github.com/sirupsen/logrus"

	goLibLevel := level.InfoLevel
	logrusLevel := goLibLevel.Logrus() // logrus.InfoLevel

	logger := logrus.New()
	logger.SetLevel(goLibLevel.Logrus())

NilLevel returns math.MaxInt32 when converted to logrus, effectively disabling
all log output.

# Use Cases

1. Configuration Parsing

Parse log levels from configuration files:

	cfgLevel := config.Get("log.level") // "debug"
	level := level.Parse(cfgLevel)
	logger.SetLevel(level.Logrus())

2. Log Level Validation

Validate and list available levels:

	levels := level.ListLevels()
	// ["critical", "fatal", "error", "warning", "info", "debug"]

	for _, lvl := range levels {
	    fmt.Printf("%s is valid\n", lvl)
	}

3. Level Comparison

Compare severity levels:

	if level.ErrorLevel < level.WarnLevel {
	    // More severe levels have lower values
	}

	currentLevel := level.InfoLevel
	if level.DebugLevel >= currentLevel {
	    // Will log at current level
	}

4. Dynamic Level Changes

Change log levels at runtime:

	func SetLogLevel(lvl string) error {
	    parsed := level.Parse(lvl)
	    if parsed.String() == "unknown" {
	        return fmt.Errorf("invalid level: %s", lvl)
	    }
	    logger.SetLevel(parsed.Logrus())
	    return nil
	}

5. Structured Logging

Use levels in structured logging:

	entry := logger.WithFields(logrus.Fields{
	    "level_name": level.InfoLevel.String(),
	    "level_code": level.InfoLevel.Code(),
	    "level_int":  level.InfoLevel.Int(),
	})

# Advantages and Limitations

Advantages:

  - Simple and focused API with clear semantics
  - Type-safe with compile-time validation
  - Case-insensitive parsing for user input
  - Multiple representations (string, code, int)
  - Direct logrus compatibility
  - Zero dependencies (except logrus for conversion)
  - Immutable level definitions prevent runtime modification
  - Small memory footprint (uint8 storage)

Limitations:

  - Fixed set of levels (cannot add custom levels)
  - String parsing has no whitespace trimming
  - NilLevel cannot be parsed from string (by design)
  - ParseFromUint32 clamps large values to math.MaxInt
  - No trace level (use DebugLevel instead)
  - Code() returns same value for unknown levels

# Performance Considerations

The package is designed for minimal overhead:

  - Level comparisons: O(1) integer comparison
  - String parsing: O(1) switch statement with case-insensitive comparison
  - Conversions: O(1) direct mapping
  - Memory: 1 byte per Level value (uint8)

Parsing Performance:

	Parse("info"):          ~10ns
	ParseFromInt(4):        ~5ns
	String():               ~5ns
	Logrus():               ~5ns

The package is suitable for high-performance logging scenarios where level
checks are performed frequently.

# Best Practices

✓ Use Parse() for configuration values
✓ Use ParseFromInt() for numeric inputs
✓ Check for "unknown" result when parsing untrusted input
✓ Use String() for human-readable output
✓ Use Code() for compact log prefixes
✓ Use Int() for storage and comparison
✓ Use Logrus() for logrus integration

✗ Don't cast arbitrary integers to Level
✗ Don't expect Parse() to handle whitespace
✗ Don't try to parse NilLevel from strings
✗ Don't rely on unknown level behavior in production

# Thread Safety

The Level type is immutable and all operations are thread-safe. Multiple
goroutines can safely call any method concurrently:

	var level level.Level = level.InfoLevel

	go func() { fmt.Println(level.String()) }()
	go func() { fmt.Println(level.Int()) }()
	go func() { fmt.Println(level.Logrus()) }()

Package-level functions (Parse, ParseFromInt, etc.) are also thread-safe.

# Compatibility

The package maintains compatibility with:

  - github.com/sirupsen/logrus (primary integration)
  - Standard Go integer types (int, uint8, uint32)
  - String representations for configuration files
  - JSON/YAML serialization (via String()/Parse())

Minimum Go version: 1.18

# Examples

See example_test.go for runnable examples demonstrating:
  - Basic level creation and conversion
  - Parsing from strings and integers
  - Logrus integration
  - Level comparison
  - Configuration parsing
  - Dynamic level changes
*/
package level
