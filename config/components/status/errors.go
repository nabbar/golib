/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

	libcfg "github.com/nabbar/golib/config"
	liberr "github.com/nabbar/golib/errors"
)

const (
	// ErrorParamEmpty indicates that a required parameter is empty.
	ErrorParamEmpty liberr.CodeError = iota + libcfg.MinErrorComponentStatus
	// ErrorParamInvalid indicates that a parameter is invalid.
	ErrorParamInvalid
	// ErrorComponentNotInitialized indicates that the component has not been initialized.
	ErrorComponentNotInitialized
	// ErrorConfigInvalid indicates that the configuration is invalid.
	ErrorConfigInvalid
	// ErrorValidatorError indicates an error during configuration validation.
	ErrorValidatorError
)

// init registers the error codes and messages with the global error manager.
// It panics if there is a collision with existing error codes.
func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/config/components/head"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the error message corresponding to the given error code.
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "at least one given parameters is empty"
	case ErrorParamInvalid:
		return "at least one given parameters is invalid"
	case ErrorComponentNotInitialized:
		return "this component seems to not be correctly initialized"
	case ErrorConfigInvalid:
		return "invalid config"
	case ErrorValidatorError:
		return "cannot validate config"
	}

	return liberr.NullMessage
}
