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

package router

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the router package.
// These codes are used with github.com/nabbar/golib/errors for structured error handling.
const (
	// ErrorParamEmpty indicates that a required parameter was not provided or is empty.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgRouter

	// ErrorConfigValidator indicates that configuration validation failed.
	ErrorConfigValidator

	// ErrorHeaderAuth indicates an invalid response code from authorization check.
	ErrorHeaderAuth

	// ErrorHeaderAuthMissing indicates that the Authorization header is missing from the request.
	ErrorHeaderAuthMissing

	// ErrorHeaderAuthEmpty indicates that the Authorization header is present but empty.
	ErrorHeaderAuthEmpty

	// ErrorHeaderAuthRequire indicates that authorization check failed and authentication is required.
	// This typically results in HTTP 401 Unauthorized.
	ErrorHeaderAuthRequire

	// ErrorHeaderAuthForbidden indicates that authorization check succeeded but the client is not authorized.
	// This typically results in HTTP 403 Forbidden.
	ErrorHeaderAuthForbidden
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/router"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the error message for a given error code.
// This function is registered with the errors package during init.
//
// See also: github.com/nabbar/golib/errors
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorConfigValidator:
		return "invalid config, validation error"
	case ErrorHeaderAuthMissing:
		return "missing authorization header"
	case ErrorHeaderAuthEmpty:
		return "authorization header is empty"
	case ErrorHeaderAuthRequire:
		return "authorization check failed, authorization still require"
	case ErrorHeaderAuthForbidden:
		return "authorization check success but unauthorized client"
	case ErrorHeaderAuth:
		return "authorization check return an invalid response code"
	}

	return liberr.NullMessage
}
