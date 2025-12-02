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

package level_test

import (
	"fmt"

	"github.com/nabbar/golib/logger/level"
	"github.com/sirupsen/logrus"
)

// Example_basic demonstrates basic level usage and conversions.
func Example_basic() {
	// Create a level
	lvl := level.InfoLevel

	// Get string representation
	fmt.Println("String:", lvl.String())

	// Get code representation
	fmt.Println("Code:", lvl.Code())

	// Get integer representation
	fmt.Println("Int:", lvl.Int())

	// Get uint8 representation
	fmt.Println("Uint8:", lvl.Uint8())

	// Output:
	// String: Info
	// Code: Info
	// Int: 4
	// Uint8: 4
}

// Example_parse demonstrates parsing levels from strings.
func Example_parse() {
	// Parse from full name (case-insensitive)
	lvl1 := level.Parse("info")
	fmt.Println(lvl1.String())

	lvl2 := level.Parse("ERROR")
	fmt.Println(lvl2.String())

	lvl3 := level.Parse("Critical")
	fmt.Println(lvl3.String())

	// Parse from code
	lvl4 := level.Parse("Warn")
	fmt.Println(lvl4.String())

	// Invalid input returns InfoLevel
	lvl5 := level.Parse("unknown")
	fmt.Println(lvl5.String())

	// Output:
	// Info
	// Error
	// Critical
	// Warning
	// Info
}

// Example_parseFromInt demonstrates parsing levels from integers.
func Example_parseFromInt() {
	// Parse from integer value
	lvl1 := level.ParseFromInt(0)
	fmt.Println(lvl1.String())

	lvl2 := level.ParseFromInt(4)
	fmt.Println(lvl2.String())

	lvl3 := level.ParseFromInt(6)
	fmt.Println(lvl3.String())

	// Invalid value returns InfoLevel
	lvl4 := level.ParseFromInt(99)
	fmt.Println(lvl4.String())

	// Output:
	// Critical
	// Info
	//
	// Info
}

// Example_parseFromUint32 demonstrates parsing levels from uint32.
func Example_parseFromUint32() {
	// Parse from uint32 value
	lvl1 := level.ParseFromUint32(2)
	fmt.Println(lvl1.String())

	lvl2 := level.ParseFromUint32(5)
	fmt.Println(lvl2.String())

	// Large values are clamped
	lvl3 := level.ParseFromUint32(99)
	fmt.Println(lvl3.String())

	// Output:
	// Error
	// Debug
	// Info
}

// Example_listLevels demonstrates listing all available levels.
func Example_listLevels() {
	levels := level.ListLevels()

	fmt.Println("Available levels:")
	for _, lvl := range levels {
		fmt.Printf("  - %s\n", lvl)
	}

	// Output:
	// Available levels:
	//   - critical
	//   - fatal
	//   - error
	//   - warning
	//   - info
	//   - debug
}

// Example_comparison demonstrates comparing log levels.
func Example_comparison() {
	// More severe levels have lower values
	fmt.Println("PanicLevel < InfoLevel:", level.PanicLevel < level.InfoLevel)
	fmt.Println("ErrorLevel < DebugLevel:", level.ErrorLevel < level.DebugLevel)

	// Check if a level should be logged
	currentLevel := level.WarnLevel
	testLevel := level.ErrorLevel

	if testLevel <= currentLevel {
		fmt.Println("ErrorLevel would be logged")
	}

	// Output:
	// PanicLevel < InfoLevel: true
	// ErrorLevel < DebugLevel: true
	// ErrorLevel would be logged
}

// Example_logrus demonstrates integration with logrus.
func Example_logrus() {
	// Convert to logrus level
	goLibLevel := level.InfoLevel
	logrusLevel := goLibLevel.Logrus()

	fmt.Printf("GoLib level: %s\n", goLibLevel.String())
	fmt.Printf("Logrus level: %v\n", logrusLevel)
	fmt.Printf("Logrus level matches: %v\n", logrusLevel == logrus.InfoLevel)

	// Output:
	// GoLib level: Info
	// Logrus level: info
	// Logrus level matches: true
}

// Example_configurationParsing demonstrates parsing log levels from configuration.
func Example_configurationParsing() {
	// Simulate configuration values (use slice to maintain order)
	configs := []struct {
		key   string
		value string
	}{
		{"app.log.level", "debug"},
		{"api.log.level", "INFO"},
		{"default.log.level", "warning"},
	}

	for _, cfg := range configs {
		lvl := level.Parse(cfg.value)
		fmt.Printf("%s: %s (code: %s)\n", cfg.key, lvl.String(), lvl.Code())
	}

	// Output:
	// app.log.level: Debug (code: Debug)
	// api.log.level: Info (code: Info)
	// default.log.level: Warning (code: Warn)
}

// Example_dynamicLevelChange demonstrates changing log levels at runtime.
func Example_dynamicLevelChange() {
	// Initial level
	currentLevel := level.InfoLevel
	fmt.Printf("Initial level: %s\n", currentLevel.String())

	// Simulate level change request
	newLevelStr := "debug"
	newLevel := level.Parse(newLevelStr)

	if newLevel.String() != "unknown" {
		currentLevel = newLevel
		fmt.Printf("Changed to: %s\n", currentLevel.String())
	}

	// Another change
	newLevelStr = "error"
	newLevel = level.Parse(newLevelStr)
	if newLevel.String() != "unknown" {
		currentLevel = newLevel
		fmt.Printf("Changed to: %s\n", currentLevel.String())
	}

	// Output:
	// Initial level: Info
	// Changed to: Debug
	// Changed to: Error
}

// Example_allLevels demonstrates all log level constants.
func Example_allLevels() {
	levels := []level.Level{
		level.PanicLevel,
		level.FatalLevel,
		level.ErrorLevel,
		level.WarnLevel,
		level.InfoLevel,
		level.DebugLevel,
		level.NilLevel,
	}

	fmt.Println("Level constants:")
	for _, lvl := range levels {
		fmt.Printf("  %d: %s (%s)\n", lvl.Uint8(), lvl.String(), lvl.Code())
	}

	// Output:
	// Level constants:
	//   0: Critical (Crit)
	//   1: Fatal (Fatal)
	//   2: Error (Err)
	//   3: Warning (Warn)
	//   4: Info (Info)
	//   5: Debug (Debug)
	//   6:  ()
}

// Example_roundtrip demonstrates roundtrip conversions.
func Example_roundtrip() {
	// Start with a level
	original := level.WarnLevel

	// Convert to string and back
	str := original.String()
	parsed := level.Parse(str)
	fmt.Printf("String roundtrip: %s -> %s -> %s\n",
		original.String(), str, parsed.String())

	// Convert to int and back
	i := original.Int()
	fromInt := level.ParseFromInt(i)
	fmt.Printf("Int roundtrip: %s -> %d -> %s\n",
		original.String(), i, fromInt.String())

	// Convert to uint32 and back
	u32 := original.Uint32()
	fromUint32 := level.ParseFromUint32(u32)
	fmt.Printf("Uint32 roundtrip: %s -> %d -> %s\n",
		original.String(), u32, fromUint32.String())

	// Output:
	// String roundtrip: Warning -> Warning -> Warning
	// Int roundtrip: Warning -> 3 -> Warning
	// Uint32 roundtrip: Warning -> 3 -> Warning
}

// Example_validation demonstrates input validation.
func Example_validation() {
	testInputs := []string{"info", "DEBUG", "invalid", "", "trace"}

	fmt.Println("Validation results:")
	for _, input := range testInputs {
		lvl := level.Parse(input)
		// Parse returns InfoLevel for invalid inputs
		if lvl == level.InfoLevel {
			// Check if this was actually "info" or a fallback
			if input != "info" && input != "INFO" {
				fmt.Printf("  %q -> %s (fallback for invalid)\n", input, lvl.String())
				continue
			}
		}
		fmt.Printf("  %q -> %s\n", input, lvl.String())
	}

	// Output:
	// Validation results:
	//   "info" -> Info
	//   "DEBUG" -> Debug
	//   "invalid" -> Info (fallback for invalid)
	//   "" -> Info (fallback for invalid)
	//   "trace" -> Info (fallback for invalid)
}
