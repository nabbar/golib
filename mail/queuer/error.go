/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package queuer

import (
	"fmt"

	liberr "github.com/nabbar/golib/errors"
)

// Error codes for the queuer package.
//
// These error codes are registered with the github.com/nabbar/golib/errors
// package for centralized error handling and consistent error messages.
//
// All error codes start from liberr.MinPkgMailPooler to avoid collisions
// with other packages in the golib ecosystem.
const (
	// ErrorParamEmpty indicates that a required parameter is missing or nil.
	//
	// This error is returned when:
	//   - SMTP client is nil when calling Send(), Check(), Client(), or Monitor()
	//   - Required configuration parameters are missing
	//
	// Example scenarios:
	//   - Calling pooler.Send() when created with New(cfg, nil)
	//   - Calling pooler.Check() before UpdConfig() sets a valid client
	ErrorParamEmpty liberr.CodeError = iota + liberr.MinPkgMailPooler

	// ErrorMailPooler is a generic error for pooler operations.
	//
	// This error code is reserved for general pooler failures that don't
	// fit into more specific categories. Currently used as a base error
	// for wrapping other errors.
	ErrorMailPooler

	// ErrorMailPoolerContext indicates that the operation was cancelled
	// due to context cancellation or timeout.
	//
	// This error is returned when:
	//   - Context is cancelled before a Pool() operation begins
	//   - Context is cancelled while waiting during throttling
	//   - Context timeout expires during an operation
	//
	// This allows callers to distinguish between throttling failures
	// and context-related cancellations.
	ErrorMailPoolerContext
)

// init registers the error message function with the errors package.
// This ensures that error codes have consistent, human-readable messages.
//
// Panics if there is an error code collision with another package,
// which indicates a programming error in package initialization.
func init() {
	if liberr.ExistInMapMessage(ErrorParamEmpty) {
		panic(fmt.Errorf("error code collision with package golib/mailPooler"))
	}
	liberr.RegisterIdFctMessage(ErrorParamEmpty, getMessage)
}

// getMessage returns the human-readable error message for a given error code.
//
// This function is called by the errors package to generate error messages.
// It provides consistent error descriptions for all queuer error codes.
//
// Parameters:
//   - code: The error code to get a message for
//
// Returns:
//   - A human-readable error message string
//   - liberr.NullMessage if the code is not recognized
//
// This function is registered during init() and should not be called directly.
// Use ErrorCode.Error() method instead to create errors with these messages.
func getMessage(code liberr.CodeError) (message string) {
	switch code {
	case ErrorParamEmpty:
		return "given parameters is empty"
	case ErrorMailPooler:
		return "generic mail pooler error"
	case ErrorMailPoolerContext:
		return "context has trigger error"
	}

	return liberr.NullMessage
}
