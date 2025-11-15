/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package level

import (
	"math"

	"github.com/sirupsen/logrus"
)

// Uint8 Convert the current Level type to a uint8 value. E.g. FatalLevel becomes 1.
func (l Level) Uint8() uint8 {
	return uint8(l)
}

// Uint32 Convert the current Level type to a uint32 value. E.g. FatalLevel becomes 1.
func (l Level) Uint32() uint32 {
	return uint32(l)
}

// Int Convert the current Level type to a int value. E.g. FatalLevel becomes 1.
func (l Level) Int() int {
	return int(l)
}

// String Convert the current Level type to a string. E.g. PanicLevel becomes "Critical Error".
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
