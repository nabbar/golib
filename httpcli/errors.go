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

package httpcli

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for HTTP client operations.
// These errors are registered with the golib/errors package for consistent error handling.
const (
	ErrorParamEmpty           liberr.CodeError = iota + liberr.MinPkgHttpCli // At least one given parameter is empty
	ErrorParamInvalid                                                        // At least one given parameter is invalid
	ErrorValidatorError                                                      // Configuration validation failed
	ErrorClientTransportHttp2                                                // HTTP/2 transport configuration error
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/httpcli"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "at least one given parameters is empty"
	case ErrorParamInvalid:
		return "at least one given parameters is invalid"
	case ErrorValidatorError:
		return "config seems to be invalid"
	case ErrorClientTransportHttp2:
		return "error while configure http2 transport for client"
	}

	return liberr.NullMessage
}
