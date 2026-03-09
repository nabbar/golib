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

package status

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	// ErrorParamEmpty occurs when required parameters are missing. For example,
	// this can happen when attempting to marshal a status response without having
	// first set the application info (name, release, hash) via `SetInfo` or `SetVersion`.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgStatus

	// ErrorValidatorError occurs when the configuration validation fails. This
	// error is returned by `Config.Validate()` when the provided configuration
	// does not meet the required constraints defined by the struct tags (e.g., a
	// required field is missing).
	ErrorValidatorError
)

// init registers the error codes and their corresponding messages with the `golib/errors`
// package. This allows for centralized error management and consistent error messages.
// It will panic on startup if there is an error code collision to prevent ambiguity.
func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/status"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the human-readable message for a given error code. This function
// is used by the `golib/errors` package to resolve an error code to its string
// representation.
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters are empty"
	case ErrorValidatorError:
		return "invalid config"
	}

	return liberr.NullMessage
}
