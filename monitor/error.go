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

package monitor

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

const (
	// ErrorParamEmpty indicates an empty parameter was provided.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgMonitor
	// ErrorMissingHealthCheck indicates no health check function was registered.
	ErrorMissingHealthCheck
	// ErrorValidatorError indicates configuration validation failed.
	ErrorValidatorError
	// ErrorLoggerError indicates logger initialization failed.
	ErrorLoggerError
	// ErrorTimeout indicates a timeout occurred during an operation.
	ErrorTimeout
	// ErrorInvalid indicates an invalid monitor instance.
	ErrorInvalid
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/monitor"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the error message for a given error code.
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorMissingHealthCheck:
		return "missing healthcheck"
	case ErrorValidatorError:
		return "invalid config"
	case ErrorLoggerError:
		return "cannot initialize logger"
	case ErrorTimeout:
		return "timeout error"
	case ErrorInvalid:
		return "invalid instance"
	}

	return liberr.NullMessage
}
