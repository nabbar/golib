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

package level

import (
	"math"
	"strings"
)

// Level represents a logging severity level as an uint8 value.
// It provides methods for conversion to various formats (string, int, logrus.Level)
// and parsing from multiple input types.
// Levels are ordered from most severe (PanicLevel=0) to least severe (DebugLevel=5).
// NilLevel (6) is a special value that disables logging.
type Level uint8

const (
	// PanicLevel is the highest severity level (value: 0).
	// Used for critical errors that will trigger a panic with stack trace.
	// String representation: "Critical", Code: "Crit"
	PanicLevel Level = iota

	// FatalLevel represents fatal errors (value: 1).
	// Used for errors that will cause the application to exit.
	// String representation: "Fatal", Code: "Fatal"
	FatalLevel

	// ErrorLevel represents errors (value: 2).
	// Used when the operation fails and execution stops, returning control to the caller.
	// String representation: "Error", Code: "Err"
	ErrorLevel

	// WarnLevel represents warnings (value: 3).
	// Used when an issue occurs but execution continues with degraded functionality.
	// String representation: "Warning", Code: "Warn"
	WarnLevel

	// InfoLevel represents informational messages (value: 4).
	// Used for general information about application state, events, or successful operations.
	// This is the default level returned by Parse() for invalid inputs.
	// String representation: "Info", Code: "Info"
	InfoLevel

	// DebugLevel is the lowest severity level for normal logging (value: 5).
	// Used for detailed diagnostic information useful during development and troubleshooting.
	// String representation: "Debug", Code: "Debug"
	DebugLevel

	// NilLevel is a special level that disables all logging (value: 6).
	// It cannot be parsed from string and returns empty strings for String() and Code().
	// Converts to math.MaxInt32 when used with Logrus().
	NilLevel
)

// ListLevels returns a slice containing lowercase string representations of all standard log levels.
// The returned slice contains: ["critical", "fatal", "error", "warning", "info", "debug"]
// NilLevel is not included as it's not meant to be parsed or used in configuration.
// All returned strings can be parsed back using Parse().
func ListLevels() []string {
	return []string{
		strings.ToLower(PanicLevel.String()),
		strings.ToLower(FatalLevel.String()),
		strings.ToLower(ErrorLevel.String()),
		strings.ToLower(WarnLevel.String()),
		strings.ToLower(InfoLevel.String()),
		strings.ToLower(DebugLevel.String()),
	}
}

// Parse converts a string to its corresponding Level value.
// Parsing is case-insensitive and supports both full names and short codes:
//   - "Critical", "CRITICAL", "critical", "Crit" -> PanicLevel
//   - "Fatal", "FATAL", "fatal" -> FatalLevel
//   - "Error", "ERROR", "error", "Err" -> ErrorLevel
//   - "Warning", "WARNING", "warning", "Warn" -> WarnLevel
//   - "Info", "INFO", "info" -> InfoLevel
//   - "Debug", "DEBUG", "debug" -> DebugLevel
//
// Returns InfoLevel for any unrecognized input (empty string, invalid values, etc.).
// Note: Parse does not trim leading/trailing whitespace.
// Note: NilLevel cannot be parsed from string and will return InfoLevel.
func Parse(l string) Level {
	switch {
	case strings.EqualFold(PanicLevel.String(), l), strings.EqualFold(PanicLevel.Code(), l):
		return PanicLevel

	case strings.EqualFold(FatalLevel.String(), l), strings.EqualFold(FatalLevel.Code(), l):
		return FatalLevel

	case strings.EqualFold(ErrorLevel.String(), l), strings.EqualFold(ErrorLevel.Code(), l):
		return ErrorLevel

	case strings.EqualFold(WarnLevel.String(), l), strings.EqualFold(WarnLevel.Code(), l):
		return WarnLevel

	case strings.EqualFold(InfoLevel.String(), l), strings.EqualFold(InfoLevel.Code(), l):
		return InfoLevel

	case strings.EqualFold(DebugLevel.String(), l), strings.EqualFold(DebugLevel.Code(), l):
		return DebugLevel
	}

	return InfoLevel
}

// ParseFromInt converts an integer to its corresponding Level value.
// Valid inputs: 0=PanicLevel, 1=FatalLevel, 2=ErrorLevel, 3=WarnLevel,
// 4=InfoLevel, 5=DebugLevel, 6=NilLevel.
// Returns InfoLevel for any value outside the valid range (negative or > 6).
// This function is useful for deserializing levels from numeric storage or APIs.
func ParseFromInt(i int) Level {
	switch i {
	case PanicLevel.Int():
		return PanicLevel
	case FatalLevel.Int():
		return FatalLevel
	case ErrorLevel.Int():
		return ErrorLevel
	case WarnLevel.Int():
		return WarnLevel
	case InfoLevel.Int():
		return InfoLevel
	case DebugLevel.Int():
		return DebugLevel
	case NilLevel.Int():
		return NilLevel
	default:
		return InfoLevel
	}
}

// ParseFromUint32 converts a uint32 to its corresponding Level value.
// Valid inputs: 0=PanicLevel, 1=FatalLevel, 2=ErrorLevel, 3=WarnLevel,
// 4=InfoLevel, 5=DebugLevel, 6=NilLevel.
// Values >= math.MaxInt are clamped to math.MaxInt before conversion (platform-dependent).
// Returns InfoLevel for any value outside the valid range (> 6).
// This function is useful for deserializing levels from 32-bit numeric storage.
func ParseFromUint32(i uint32) Level {
	if uint64(i) < uint64(math.MaxInt) {
		return ParseFromInt(int(i))
	} else {
		return ParseFromInt(math.MaxInt)
	}
}
