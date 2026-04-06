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
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FuncMap defines a callback function used when traversing error hierarchies with the Map method.
// It receives an error and returns a boolean indicating whether to continue (true) or stop (false) the traversal.
type FuncMap func(e error) bool

// ReturnError defines a callback function signature for extracting detailed error information.
// It provides the numeric code, message, file/function name, and line number.
type ReturnError func(code int, msg string, file string, line int)

// Error is an enhanced error interface that supports numeric codes, stack tracing, and hierarchical chaining.
// It is fully compatible with Go's standard error interface and the errors.Is/As functions.
type Error interface {
	error

	// IsCode returns true if the current error (excluding parents) has the specified CodeError.
	IsCode(code CodeError) bool

	// HasCode recursively searches the current error and all its parents for the specified CodeError.
	HasCode(code CodeError) bool

	// GetCode returns the CodeError assigned to the current error instance.
	GetCode() CodeError

	// GetParentCode returns a unique list of all CodeError values found in the error hierarchy.
	GetParentCode() []CodeError

	// Is implements the standard errors.Is interface, allowing comparison with other error instances.
	Is(e error) bool

	// IsError checks if the given error is equivalent to the current error, considering message and stack trace.
	IsError(e error) bool

	// HasError recursively searches the error hierarchy to see if the given error (or an equivalent one) exists.
	HasError(err error) bool

	// HasParent returns true if the current error has at least one parent error attached.
	HasParent() bool

	// GetParent flattens the error hierarchy into a slice. If withMainError is true, the current error is included first.
	GetParent(withMainError bool) []error

	// Map traverses the error hierarchy and executes the provided FuncMap for each error.
	Map(fct FuncMap) bool

	// ContainsString recursively checks if the error message or any parent message contains the specified substring.
	ContainsString(s string) bool

	// Add appends one or more errors as parents to the current error hierarchy.
	Add(parent ...error)

	// SetParent clears all current parents and replaces them with the provided list of errors.
	SetParent(parent ...error)

	// Code returns the current error's numeric code as a uint16.
	Code() uint16

	// CodeSlice returns a flattened slice of all numeric codes in the error hierarchy.
	CodeSlice() []uint16

	// CodeError returns a formatted string of the current error (code + message) using the provided pattern.
	CodeError(pattern string) string

	// CodeErrorSlice returns a slice of formatted strings (code + message) for all errors in the hierarchy.
	CodeErrorSlice(pattern string) []string

	// CodeErrorTrace returns a formatted string (code + message + trace) for the current error.
	CodeErrorTrace(pattern string) string

	// CodeErrorTraceSlice returns a slice of formatted strings (code + message + trace) for all errors in the hierarchy.
	CodeErrorTraceSlice(pattern string) []string

	// Error returns a string representation of the error hierarchy based on the globally configured ErrorMode.
	Error() string

	// StringError returns the raw error message of the current error instance.
	StringError() string

	// StringErrorSlice returns a slice of raw error messages for all errors in the hierarchy.
	StringErrorSlice() []string

	// GetError converts the current error instance into a standard Go error.
	GetError() error

	// GetErrorSlice converts the entire error hierarchy into a slice of standard Go errors.
	GetErrorSlice() []error

	// Unwrap returns the list of parent errors, supporting Go's multi-error unwrapping (since Go 1.20).
	Unwrap() []error

	// GetTrace returns the formatted stack trace string for the current error.
	GetTrace() string

	// GetTraceSlice returns a slice of formatted stack trace strings for all errors in the hierarchy.
	GetTraceSlice() []string

	// Return populates a Return interface with the current error hierarchy's information.
	Return(r Return)

	// ReturnError executes the callback for the current error's details.
	ReturnError(f ReturnError)

	// ReturnParent recursively executes the callback for all parent errors in the hierarchy.
	ReturnParent(f ReturnError)
}

// Errors is an interface for objects that store a collection of errors.
type Errors interface {
	// ErrorsLast returns the most recently registered error.
	ErrorsLast() error

	// ErrorsList returns a slice of all registered errors.
	ErrorsList() []error
}

// Is is a helper function that checks if an error implements the Error interface.
func Is(e error) bool {
	var err Error
	return errors.As(e, &err)
}

// Get attempts to cast a standard error to the Error interface. Returns nil if the cast fails.
func Get(e error) Error {
	var err Error
	if errors.As(e, &err) {
		return err
	}

	return nil
}

// Has checks if the given error or any of its parents matches the specified CodeError.
func Has(e error, code CodeError) bool {
	if err := Get(e); err == nil {
		return false
	} else {
		return err.HasCode(code)
	}
}

// ContainsString checks if the error's message or any parent message contains the specified substring.
func ContainsString(e error, s string) bool {
	if e == nil {
		return false
	} else if err := Get(e); err == nil {
		return strings.Contains(e.Error(), s)
	} else {
		return err.ContainsString(s)
	}
}

// IsCode checks if the direct error code matches the specified CodeError.
func IsCode(e error, code CodeError) bool {
	if err := Get(e); err == nil {
		return false
	} else {
		return err.IsCode(code)
	}
}

// Make converts a standard error into an Error interface.
// If the error already implements Error, it is returned as is.
// Otherwise, it's wrapped in a new Error instance with code 0 and no stack trace.
func Make(e error) Error {
	if e == nil {
		return nil
	}

	if err, ok := e.(*ers); ok {
		return err
	}

	return &ers{
		c: 0,
		e: e.Error(),
		p: nil,
		t: trcNil{},
	}
}

// MakeIfError combines multiple errors into a single Error hierarchy.
// If all input errors are nil, it returns nil.
func MakeIfError(err ...error) Error {
	var e Error = nil

	for i := range err {
		if err[i] == nil {
			continue
		} else if e == nil {
			e = Make(err[i])
		} else {
			e.Add(err[i])
		}
	}

	return e
}

// AddOrNew adds a sub-error and optional parents to a main error.
// If errMain is nil, it creates a new Error from errSub.
func AddOrNew(errMain, errSub error, parent ...error) Error {
	var e Error

	if errMain != nil {
		if e = Get(errMain); e == nil {
			e = New(0, errMain.Error())
		}
		e.Add(errSub)
		e.Add(parent...)
		return e
	} else if errSub != nil {
		return New(0, errSub.Error(), parent...)
	}

	return nil
}

// New creates a new Error instance with the given code, message, and optional parents.
// It automatically captures the stack trace of the caller.
func New(code uint16, message string, parent ...error) Error {
	return newErs(code, message, getFrame(), parent...)
}

// Newf creates a new Error instance with a formatted message and optional parents.
// It automatically captures the stack trace of the caller.
func Newf(code uint16, pattern string, args ...any) Error {
	return newErs(code, fmt.Sprintf(pattern, args...), getFrame(), nil)
}

// NewErrorTrace creates a new Error instance with a manually specified stack trace (file and line).
func NewErrorTrace(code int, msg string, file string, line int, parent ...error) Error {
	trc := trcNil{
		Func: "",
		File: file,
		Line: line,
	}
	return newErs(intToUint16(code), msg, trc, parent...)
}

// NewErrorRecovered creates a new Error instance specially formatted for panic recovery scenarios.
// It captures a multi-level stack trace excluding internal package frames.
func NewErrorRecovered(msg string, recovered string, parent ...error) Error {
	var b = strings.Builder{}

	if recovered != "" {
		parent = append([]error{errors.New(recovered)}, parent...)
	}

	b.WriteString(msg)
	for _, pc := range getFrameVendor() {
		fc, fl, ln := pc.Compose()
		b.WriteRune('\n')
		b.WriteString("Fct: " + fc)
		b.WriteString(" - ")
		b.WriteString("File: " + fl)
		b.WriteString(" - ")
		b.WriteString("Line: " + strconv.Itoa(ln))
	}

	return newErs(0, b.String(), getFrame(), parent...)
}

// IfError returns a new Error only if at least one of the provided parent errors is not nil.
// If all parents are nil, it returns nil.
func IfError(code uint16, message string, parent ...error) Error {
	for i := range parent {
		if parent[i] != nil {
			return newErs(code, message, getFrame(), parent...)
		}
	}

	return nil
}

// NewDefaultReturn creates a new instance of DefaultReturn, useful for API response standardization.
func NewDefaultReturn() *DefaultReturn {
	return &DefaultReturn{
		Code:    "",
		Message: "",
	}
}

// newErs is the internal constructor for the ers struct.
func newErs(c uint16, m string, f tracer, p ...error) Error {
	var r = &ers{
		c: c,
		e: m,
		p: nil,
		t: f,
	}
	r.Add(p...)
	return r
}

// intToUint16 safely converts an int to a uint16 with bounds checking.
func intToUint16(i int) uint16 {
	if i < 0 {
		return 0
	}

	if i > math.MaxUint16 {
		return math.MaxUint16
	}

	return uint16(i)
}
