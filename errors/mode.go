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

var modeError = Default

func SetModeReturnError(mode ErrorMode) {
	modeError = mode
}

func GetModeReturnError() ErrorMode {
	return modeError
}

type ErrorMode uint8

const (
	Default ErrorMode = iota
	ErrorReturnCode
	ErrorReturnCodeFull
	ErrorReturnCodeError
	ErrorReturnCodeErrorFull
	ErrorReturnCodeErrorTrace
	ErrorReturnCodeErrorTraceFull
	ErrorReturnStringError
	ErrorReturnStringErrorFull
)

func (m ErrorMode) String() string {
	//nolint exhaustive
	switch m {
	case Default:
		return "default"
	case ErrorReturnCode:
		return "Code"
	case ErrorReturnCodeFull:
		return "CodeFull"
	case ErrorReturnCodeError:
		return "CodeError"
	case ErrorReturnCodeErrorFull:
		return "CodeErrorFull"
	case ErrorReturnCodeErrorTrace:
		return "CodeErrorTrace"
	case ErrorReturnCodeErrorTraceFull:
		return "CodeErrorTraceFull"
	case ErrorReturnStringError:
		return "StringError"
	case ErrorReturnStringErrorFull:
		return "StringErrorFull"
	}

	return Default.String()
}

func (m ErrorMode) error(e *ers) string {
	//nolint exhaustive
	switch m {
	case Default:
		return e.StringError()
	case ErrorReturnCode:
		return fmt.Sprintf("%v", e.Code())
	case ErrorReturnCodeFull:
		return fmt.Sprintf("%v", e.CodeSlice())
	case ErrorReturnCodeError:
		return e.CodeError("")
	case ErrorReturnCodeErrorFull:
		return strings.Join(e.CodeErrorSlice(""), ", ")
	case ErrorReturnCodeErrorTrace:
		return e.CodeErrorTrace("")
	case ErrorReturnCodeErrorTraceFull:
		return strings.Join(e.CodeErrorTraceSlice(""), ", ")
	case ErrorReturnStringError:
		return e.StringError()
	case ErrorReturnStringErrorFull:
		return strings.Join(e.StringErrorSlice(), ", ")
	}

	return Default.error(e)
}
