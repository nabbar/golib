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

package version

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the version package.
// These codes are used with the github.com/nabbar/golib/errors package.
const (
	// ErrorParamEmpty indicates that a required parameter is empty or missing.
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgVersion

	// ErrorGoVersionInit indicates a failure to initialize the Go version constraint parser.
	// This typically occurs when the constraint string is malformed.
	ErrorGoVersionInit

	// ErrorGoVersionRuntime indicates a failure to extract the current Go runtime version.
	// This is rare and usually indicates a problem with the Go installation.
	ErrorGoVersionRuntime

	// ErrorGoVersionConstraint indicates that the current Go version does not satisfy
	// the required version constraint.
	ErrorGoVersionConstraint
)

func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/version"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorGoVersionInit:
		return "init GoVersion contraint error"
	case ErrorGoVersionRuntime:
		return "extract GoVersion runtime error"
	case ErrorGoVersionConstraint:
		return "current binary is build with a non-compatible version of Go"
	}

	return liberr.NullMessage
}
