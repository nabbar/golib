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

package njs_errors

var modeError = ERROR_RETURN_Default

func SetModeReturnError(mode ErrorMode) {
	modeError = mode
}

func GetModeReturnError() ErrorMode {
	return modeError
}

type ErrorMode uint8

const (
	ERROR_RETURN_Default ErrorMode = iota
	ERROR_RETURN_Code
	ERROR_RETURN_CodeFull
	ERROR_RETURN_CodeError
	ERROR_RETURN_CodeErrorFull
	ERROR_RETURN_CodeErrorTrace
	ERROR_RETURN_CodeErrorTraceFull
	ERROR_RETURN_StringError
	ERROR_RETURN_StringErrorFull
)

func (m ErrorMode) String() string {
	switch m {
	case ERROR_RETURN_Code:
		return "Code"
	case ERROR_RETURN_CodeFull:
		return "CodeFull"
	case ERROR_RETURN_CodeError:
		return "CodeError"
	case ERROR_RETURN_CodeErrorFull:
		return "CodeErrorFull"
	case ERROR_RETURN_CodeErrorTrace:
		return "CodeErrorTrace"
	case ERROR_RETURN_CodeErrorTraceFull:
		return "CodeErrorTraceFull"
	case ERROR_RETURN_StringError:
		return "StringError"
	case ERROR_RETURN_StringErrorFull:
		return "StringErrorFull"

	default:
		return "default"
	}
}

func (m ErrorMode) error(e *errors) string {
	switch m {
	case ERROR_RETURN_Code:
		return e.Code()
	case ERROR_RETURN_CodeFull:
		return e.CodeFull("")
	case ERROR_RETURN_CodeError:
		return e.CodeError("")
	case ERROR_RETURN_CodeErrorFull:
		return e.CodeErrorFull("", "")
	case ERROR_RETURN_CodeErrorTrace:
		return e.CodeErrorTrace("")
	case ERROR_RETURN_CodeErrorTraceFull:
		return e.CodeErrorTraceFull("", "")
	case ERROR_RETURN_StringError:
		return e.StringError()
	case ERROR_RETURN_StringErrorFull:
		return e.StringErrorFull("")

	default:
		return e.StringError()
	}
}
