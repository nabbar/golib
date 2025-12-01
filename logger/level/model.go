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

	"github.com/sirupsen/logrus"
)

// Uint8 converts the Level to its underlying uint8 value.
// Examples: PanicLevel -> 0, FatalLevel -> 1, InfoLevel -> 4, NilLevel -> 6.
// This is useful for compact binary serialization or when interfacing with
// systems that expect 8-bit level values.
func (l Level) Uint8() uint8 {
	return uint8(l)
}

// Uint32 converts the Level to a uint32 value.
// Examples: PanicLevel -> 0, FatalLevel -> 1, InfoLevel -> 4, NilLevel -> 6.
// This is useful when interfacing with 32-bit APIs or for compatibility
// with systems that use uint32 for level representation.
// Use ParseFromUint32() to convert back to Level.
func (l Level) Uint32() uint32 {
	return uint32(l)
}

// Int converts the Level to an int value.
// Examples: PanicLevel -> 0, FatalLevel -> 1, InfoLevel -> 4, NilLevel -> 6.
// This is useful for level comparison, storage, or when interfacing with
// systems that expect integer level values.
// Use ParseFromInt() to convert back to Level.
func (l Level) Int() int {
	return int(l)
}

// String converts the Level to its full human-readable string representation.
// Returns:
//   - PanicLevel -> "Critical"
//   - FatalLevel -> "Fatal"
//   - ErrorLevel -> "Error"
//   - WarnLevel -> "Warning"
//   - InfoLevel -> "Info"
//   - DebugLevel -> "Debug"
//   - NilLevel -> "" (empty string)
//   - Unknown levels -> "unknown"
//
// The returned string can be parsed back using Parse().
// This method implements the fmt.Stringer interface.
func (l Level) String() string {
	//nolint exhaustive
	switch l {
	case DebugLevel:
		return "Debug"
	case InfoLevel:
		return "Info"
	case WarnLevel:
		return "Warning"
	case ErrorLevel:
		return "Error"
	case FatalLevel:
		return "Fatal"
	case PanicLevel:
		return "Critical"
	case NilLevel:
		return ""
	}

	return "unknown"
}

// Code converts the Level to its short code representation.
// Returns:
//   - PanicLevel -> "Crit"
//   - FatalLevel -> "Fatal"
//   - ErrorLevel -> "Err"
//   - WarnLevel -> "Warn"
//   - InfoLevel -> "Info"
//   - DebugLevel -> "Debug"
//   - NilLevel -> "" (empty string)
//   - Unknown levels -> "unknown"
//
// Short codes are useful for compact log output or when space is limited.
// The returned code can be parsed back using Parse().
func (l Level) Code() string {
	//nolint exhaustive
	switch l {
	case DebugLevel:
		return "Debug"
	case InfoLevel:
		return "Info"
	case WarnLevel:
		return "Warn"
	case ErrorLevel:
		return "Err"
	case FatalLevel:
		return "Fatal"
	case PanicLevel:
		return "Crit"
	case NilLevel:
		return ""
	}

	return "unknown"
}

// Logrus converts the Level to its equivalent logrus.Level value.
// Mappings:
//   - PanicLevel -> logrus.PanicLevel
//   - FatalLevel -> logrus.FatalLevel
//   - ErrorLevel -> logrus.ErrorLevel
//   - WarnLevel -> logrus.WarnLevel
//   - InfoLevel -> logrus.InfoLevel
//   - DebugLevel -> logrus.DebugLevel
//   - NilLevel -> math.MaxInt32 (effectively disables logging)
//   - Unknown levels -> math.MaxInt32
//
// This method enables seamless integration with the logrus logging library.
// Use this to set logrus logger levels: logger.SetLevel(level.InfoLevel.Logrus())
func (l Level) Logrus() logrus.Level {
	switch l {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	case FatalLevel:
		return logrus.FatalLevel
	case PanicLevel:
		return logrus.PanicLevel
	default:
		return math.MaxInt32
	}
}
