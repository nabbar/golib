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
	"strings"
)

// Level a uint8 type customized with function to log message with the current log level.
type Level uint8

const (
	// PanicLevel Panic level for entry log, will result on a Panic() call (trace + fatal).
	PanicLevel Level = iota
	// FatalLevel Fatal level for entry log, will result on os.Exit with error.
	FatalLevel
	// ErrorLevel Error level for entry log who's meaning the caller stop his process and return to the pre caller.
	ErrorLevel
	// WarnLevel Warning level for entry log who's meaning the caller don't stop his process and try to continue it.
	WarnLevel
	// InfoLevel Info level for entry log who's meaning it is just an information who's have no impact on caller's process but can be useful to inform human of a state, event, success, ...
	InfoLevel
	// DebugLevel Debug level for entry log who's meaning the caller has no problem and the information is only useful to identify a potential problem who's can arrive later.
	DebugLevel
	// NilLevel Nil level will never log anything and is used to completely disable current log entry. It cannot be used in the SetLogLevel function.
	NilLevel
)

// ListLevels return a list ([]string) of all string loglevel available.
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

// Parse return a valid Level Type matching the given string parameter.
// If the given parameter don't represent a valid level, the InfoLevel will be return.
/*
	level the string representation of a Level type
*/
func Parse(l string) Level {
	switch {
	case strings.EqualFold(PanicLevel.String(), l):
		return PanicLevel

	case strings.EqualFold(FatalLevel.String(), l):
		return FatalLevel

	case strings.EqualFold(ErrorLevel.String(), l):
		return ErrorLevel

	case strings.EqualFold(WarnLevel.String(), l):
		return WarnLevel

	case strings.EqualFold(InfoLevel.String(), l):
		return InfoLevel

	case strings.EqualFold(DebugLevel.String(), l):
		return DebugLevel
	}

	return InfoLevel
}
