/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

package errors

import (
	"fmt"
	"strings"
)

// modeError holds the current global error output mode.
// It defaults to ModeDefault.
var modeError = ModeDefault

// SetModeReturnError sets the global mode for the Error() method's string output.
// This affects all Error instances in the application.
func SetModeReturnError(mode ErrorMode) {
	modeError = mode
}

// GetModeReturnError returns the current global ErrorMode.
func GetModeReturnError() ErrorMode {
	return modeError
}

// ErrorMode defines how an Error instance is rendered as a string when Error() is called.
type ErrorMode uint8

const (
	// ModeDefault is the default mode, returning only the current error's message.
	ModeDefault ErrorMode = iota

	// ModeReturnCode returns only the current error's numeric code as a string.
	ModeReturnCode

	// ModeReturnCodeFull returns a slice-formatted string of all error codes in the hierarchy.
	ModeReturnCodeFull

	// ModeReturnCodeError returns the formatted code and message of the current error.
	ModeReturnCodeError

	// ModeReturnCodeErrorFull returns a newline-separated list of formatted code/message for all errors in the hierarchy.
	ModeReturnCodeErrorFull

	// ModeReturnCodeErrorTrace returns the formatted code, message, and trace for the current error.
	ModeReturnCodeErrorTrace

	// ModeReturnCodeErrorTraceFull returns a newline-separated list of formatted code/message/trace for all errors in the hierarchy.
	ModeReturnCodeErrorTraceFull

	// ModeReturnStringError returns only the current error's message.
	ModeReturnStringError

	// ModeReturnStringErrorFull returns a newline-separated list of all error messages in the hierarchy.
	ModeReturnStringErrorFull
)

// String returns a human-readable name for the ErrorMode.
func (m ErrorMode) String() string {
	//nolint exhaustive
	switch m {
	case ModeDefault:
		return "default"
	case ModeReturnCode:
		return "Code"
	case ModeReturnCodeFull:
		return "CodeFull"
	case ModeReturnCodeError:
		return "CodeError"
	case ModeReturnCodeErrorFull:
		return "CodeErrorFull"
	case ModeReturnCodeErrorTrace:
		return "CodeErrorTrace"
	case ModeReturnCodeErrorTraceFull:
		return "CodeErrorTraceFull"
	case ModeReturnStringError:
		return "StringError"
	case ModeReturnStringErrorFull:
		return "StringErrorFull"
	}

	return ModeDefault.String()
}

// error is an internal helper that renders the provided ers struct based on the current ErrorMode.
func (m ErrorMode) error(e *ers) string {
	//nolint exhaustive
	switch m {
	case ModeDefault:
		return e.StringError()
	case ModeReturnCode:
		return fmt.Sprintf("%v", e.Code())
	case ModeReturnCodeFull:
		return fmt.Sprintf("%v", e.CodeSlice())
	case ModeReturnCodeError:
		return e.CodeError("")
	case ModeReturnCodeErrorFull:
		return strings.Join(e.CodeErrorSlice(""), "\n")
	case ModeReturnCodeErrorTrace:
		return e.CodeErrorTrace("")
	case ModeReturnCodeErrorTraceFull:
		return strings.Join(e.CodeErrorTraceSlice(""), "\n")
	case ModeReturnStringError:
		return e.StringError()
	case ModeReturnStringErrorFull:
		return strings.Join(e.StringErrorSlice(), "\n")
	}

	return ModeDefault.error(e)
}
