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

package console

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the console package.
// These codes are registered with the golib/errors package for structured error handling.
const (
	// ErrorParamEmpty indicates that required parameters were not provided or are empty.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgConsole

	// ErrorColorIOFprintf indicates a failure to write formatted output to an io.Writer.
	ErrorColorIOFprintf

	// ErrorColorBufWrite indicates a failure to write data to a buffer.
	ErrorColorBufWrite

	// ErrorColorBufUndefined indicates that a required buffer was nil or not defined.
	// This occurs when BuffPrintf is called with a nil io.Writer.
	ErrorColorBufUndefined
)

func init() {
	// Register error messages with the golib/errors package.
	// Panic if there's a code collision with another package.
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/console"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the human-readable error message for a given error code.
// This function is registered with the golib/errors package and called automatically
// when error messages are needed.
//
// Parameters:
//   - code: The error code to get the message for
//
// Returns:
//   - string: The error message, or liberr.NullMessage if code is unknown
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorColorIOFprintf:
		return "cannot write on IO"
	case ErrorColorBufWrite:
		return "cannot write on buffer"
	case ErrorColorBufUndefined:
		return "buffer is not defined"
	}

	return liberr.NullMessage
}
