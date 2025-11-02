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

var modeError = ModeDefault

func SetModeReturnError(mode ErrorMode) {
	modeError = mode
}

func GetModeReturnError() ErrorMode {
	return modeError
}

type ErrorMode uint8

const (
	ModeDefault ErrorMode = iota
	ModeReturnCode
	ModeReturnCodeFull
	ModeReturnCodeError
	ModeReturnCodeErrorFull
	ModeReturnCodeErrorTrace
	ModeReturnCodeErrorTraceFull
	ModeReturnStringError
	ModeReturnStringErrorFull
)

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
