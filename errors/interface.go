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

// Package errors provides advanced error handling with error codes, stack tracing, and hierarchy management.
//
// This package extends Go's standard error handling with features for enterprise applications:
//   - Error codes (numeric classification similar to HTTP status codes)
//   - Automatic stack trace capture (file, line, function)
//   - Error hierarchy (parent-child error chains)
//   - Error collection management (via pool sub-package)
//   - Compatibility with standard errors.Is and errors.As
//
// Key Features:
//   - Predefined HTTP-like error codes (404, 500, etc.)
//   - Custom error code support
//   - Automatic stack trace capture
//   - Error chaining and hierarchy
//   - Thread-safe error pool for collection
//   - Gin framework integration
//
// Example usage:
//
//	import liberr "github.com/nabbar/golib/errors"
//
//	// Create error with code
//	err := liberr.NotFoundError.Error(nil)
//	fmt.Println(err.Code()) // 404
//
//	// With parent error
//	baseErr := errors.New("database connection failed")
//	err := liberr.InternalError.Error(baseErr)
//
//	// Check error code
//	if e, ok := err.(liberr.Error); ok {
//	    if e.IsCode(liberr.InternalError) {
//	        log.Printf("Internal error at %s:%d", e.GetFile(), e.GetLine())
//	    }
//	}
//
//	// Error hierarchy
//	mainErr := liberr.UnknownError.Error(nil)
//	mainErr.Add(err1, err2, err3)
//
// Sub-packages:
//   - pool: Thread-safe error collection with automatic indexing
//
// Thread Safety:
//
//	The Error interface is not thread-safe for modification (Add, SetParent)
//	but is safe for concurrent reads. Use the pool sub-package for
//	thread-safe error collection across goroutines.
package errors

import (
	"errors"
	"fmt"
	"math"
	"runtime"
	"strings"
)

// FuncMap is a callback function type used for iterating over error hierarchies.
// It receives each error in the chain and returns true to continue iteration
// or false to stop.
//
// Example:
//
//	err.Map(func(e error) bool {
//	    log.Println(e.Error())
//	    return true // continue to next error
//	})
type FuncMap func(e error) bool

// ReturnError is a callback function type for custom error return handling.
// It receives error details (code, message, file, line) and can be used
// to implement custom error reporting or logging mechanisms.
type ReturnError func(code int, msg string, file string, line int)

// Error is the main interface extending Go's standard error with additional capabilities.
//
// This interface provides:
//   - Error code management (numeric classification)
//   - Error hierarchy (parent-child relationships)
//   - Stack trace information (file, line, function)
//   - Pattern matching and searching
//   - Compatibility with errors.Is and errors.As
//
// All methods are safe for concurrent reads but modification methods (Add, SetParent)
// are not thread-safe. Use the pool sub-package for thread-safe error collection.
type Error interface {
	error

	// IsCode checks if the error's code matches the given code.
	// Returns true only if the error's direct code equals the provided code.
	// Does not check parent errors.
	IsCode(code CodeError) bool
	//HasCode check if current error or parent has the given error code
	HasCode(code CodeError) bool
	//GetCode return the CodeError value of the current error
	GetCode() CodeError
	//GetParentCode return a slice of CodeError value of all parent Error and the code of the current Error
	GetParentCode() []CodeError

	//Is implement compatibility with root package errors Is function
	Is(e error) bool

	//IsError check if the given error params is a valid error and not a nil pointer
	IsError(e error) bool
	//HasError check if the given error in params is still in parent error
	HasError(err error) bool
	//HasParent check if the current Error has any valid parent
	HasParent() bool
	//GetParent return a slice of Error interface for each parent error with or without the first error.
	GetParent(withMainError bool) []error
	//Map run a function on each func and parent. If the function return false, the loop stop.
	Map(fct FuncMap) bool
	//ContainsString return true if any message into the main error or the parent message error contains the given part string
	ContainsString(s string) bool

	//Add will add all no empty given error into parent of the current Error pointer
	Add(parent ...error)
	//SetParent will replace all parent with the given error list
	SetParent(parent ...error)

	//Code is used to return the code of current Error, as string
	Code() uint16
	//CodeSlice is used to return a slice string of all code of current Error (main and parent)
	CodeSlice() []uint16

	//CodeError is used to return a composed string of current Error code with message, for current Error and no parent
	CodeError(pattern string) string
	//CodeErrorSlice is used to return a composed string slice of couple error code with message, for current Error and all parent
	CodeErrorSlice(pattern string) []string

	//CodeErrorTrace is used to return a composed string of current Error code with message and trace information, for current Error and no parent
	CodeErrorTrace(pattern string) string
	//CodeErrorTraceSlice is used to return a composed string slice of couple error code with message and trace information, for current Error and all parent
	CodeErrorTraceSlice(pattern string) []string

	//Error is used to match with error interface
	//this function will return a mixed result depends of the configuration defined by calling SetModeReturnError
	Error() string

	//StringError is used to return the error message, for current Error and no parent
	StringError() string
	//StringErrorSlice is used to return the error message, for current Error and all parent, as a slice of string
	StringErrorSlice() []string

	//GetError is used to return a new error interface based of the current error (and no parent)
	GetError() error
	//GetErrorSlice is used to return a slice of new error interface, based of the current error and all parent
	GetErrorSlice() []error
	//Unwrap will set compliance with errors As/Is functions
	Unwrap() []error

	//GetTrace will return a comped string for the trace of the current Error
	GetTrace() string
	//GetTrace will return a slice of comped string fpr the trace of the current Error and all parent
	GetTraceSlice() []string

	//Return will transform the current Error into a given pointer that implement the Return interface
	Return(r Return)
	//ReturnError will send the current Error value to the given function ReturnError
	ReturnError(f ReturnError)
	//ReturnParent will send all parent information of the current Error value to the given function ReturnError
	ReturnParent(f ReturnError)
}

type Errors interface {
	// ErrorsLast return the last registered error
	ErrorsLast() error

	// ErrorsList return a slice of all registered errors
	ErrorsList() []error
}

// Is checks if the given error is of type Error.
// It uses the errors.As function to check if the given error can be asserted to the Error interface.
//
// Parameters:
// - e (error): the error to be checked
//
// Returns:
// - bool: true if the given error is of type Error, false otherwise
func Is(e error) bool {
	var err Error
	return errors.As(e, &err)
}

// Get checks if the given error is of type Error and returns the Error interface if it is.
//
// Parameters:
// - e (error): the error to be checked
//
// Returns:
// - Error: the given error as an Error interface if it is of type Error, nil otherwise
func Get(e error) Error {
	var err Error
	if errors.As(e, &err) {
		return err
	}

	return nil
}

// Has checks if the given error or its parent has the given error code.
//
// Parameters:
// - e (error): the error to be checked
// - code (CodeError): the error code to be checked
//
// Returns:
// - bool: true if the given error or its parent has the given error code, false otherwise
func Has(e error, code CodeError) bool {
	if err := Get(e); err == nil {
		return false
	} else {
		return err.HasCode(code)
	}
}

// ContainsString checks if the given error message contains the given string.
//
// Parameters:
// - e (error): the error to be checked
// - s (string): the string to be searched for in the error message
//
// Returns:
// - bool: true if the given error message contains the given string, false otherwise
func ContainsString(e error, s string) bool {
	if e == nil {
		return false
	} else if err := Get(e); err == nil {
		return strings.Contains(e.Error(), s)
	} else {
		return err.ContainsString(s)
	}
}

// IsCode checks if the given error has the given error code.
//
// Parameters:
// - e (error): the error to be checked
// - code (CodeError): the error code to be checked
//
// Returns:
// - bool: true if the given error has the given error code, false otherwise
func IsCode(e error, code CodeError) bool {
	if err := Get(e); err == nil {
		return false
	} else {
		return err.IsCode(code)
	}
}

// Make takes an error and returns an Error interface if the given error is of type Error.
// If the given error is not of type Error, it will be wrapped in an Error interface with code 0.
//
// Parameters:
// - e (error): the error to be converted to an Error interface
//
// Returns:
// - Error: the given error as an Error interface if it is of type Error,
// otherwise a new Error interface wrapping the given error with code 0.
func Make(e error) Error {
	var err Error

	if e == nil {
		return nil
	} else if errors.As(e, &err) {
		return err
	} else {
		return &ers{
			c: 0,
			e: e.Error(),
			p: nil,
			t: getNilFrame(),
		}
	}
}

// MakeIfError takes a variable number of errors and returns an Error interface if any of the given errors is not nil.
// If all the given errors are nil, it will return nil.
// If any of the given errors are of type Error, it will be added to the Error interface.
// If any of the given errors are not of type Error, it will be wrapped in an Error interface with code 0
// and added to the Error interface.
//
// Parameters:
// - err (error): a variable number of errors to be converted to an Error interface
//
// Returns:
// - Error: the given errors as an Error interface if any of the given errors is not nil, nil otherwise.
func MakeIfError(err ...error) Error {
	var e Error = nil

	for _, p := range err {
		if p == nil {
			continue
		} else if e == nil {
			e = Make(p)
		} else {
			e.Add(p)
		}
	}

	return e
}

// AddOrNew takes an error and an error to be added to it. It also takes a variable number of parent errors.
//
// If the main error is not nil, it will be converted to an Error interface if it is not of type Error.
// The given error to be added and the parent errors will be added to the Error interface.
//
// If the main error is nil and the error to be added is not nil, a new Error interface will be created
// with the given error and parent errors.
//
// If both the main error and the error to be added are nil, nil will be returned.
//
// Parameters:
// - errMain (error): the main error to be converted to an Error interface
// - errSub (error): the error to be added to the main error
// - parent (error): a variable number of parent errors to be added to the main error
//
// Returns:
// - Error: the given errors as an Error interface if the main error or the error to be added is not nil, nil otherwise.
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

// New creates a new Error interface with the given code, message, and parent errors.
//
// Parameters:
// - code (uint16): the code of the error
// - message (string): the message of the error
// - parent (error): a variable number of parent errors to be added to the error
//
// Returns:
// - Error: the given code, message, and parent errors as an Error interface
func New(code uint16, message string, parent ...error) Error {
	var p = make([]Error, 0)

	if len(parent) > 0 {
		for _, e := range parent {
			if er := Make(e); er != nil {
				p = append(p, er)
			}
		}
	}

	return &ers{
		c: code,
		e: message,
		p: p,
		t: getFrame(),
	}
}

// Newf creates a new Error interface with the given code and a message generated by calling
// fmt.Sprintf with the given pattern and arguments.
//
// Parameters:
// - code (uint16): the code of the error
// - pattern (string): the pattern to be used with fmt.Sprintf
// - args (any): the arguments to be used with fmt.Sprintf
//
// Returns:
// - Error: the given code and message as an Error interface
func Newf(code uint16, pattern string, args ...any) Error {
	return &ers{
		c: code,
		e: fmt.Sprintf(pattern, args...),
		p: make([]Error, 0),
		t: getFrame(),
	}
}

func NewErrorTrace(code int, msg string, file string, line int, parent ...error) Error {
	var p = make([]Error, 0)

	if len(parent) > 0 {
		for _, e := range parent {
			if er := Make(e); er != nil {
				p = append(p, er)
			}
		}
	}

	// Prevent overflow
	var i uint16
	if code < 0 {
		i = 0
	} else if code > math.MaxUint16 {
		i = math.MaxUint16
	} else {
		i = uint16(code)
	}

	return &ers{
		c: i,
		e: msg,
		p: p,
		t: runtime.Frame{
			File: file,
			Line: line,
		},
	}
}

func NewErrorRecovered(msg string, recovered string, parent ...error) Error {
	var p = make([]Error, 0)

	if recovered != "" {
		p = append(p, &ers{
			c: 0,
			e: recovered,
			p: nil,
		})
	}

	if len(parent) > 0 {
		for _, e := range parent {
			if er := Make(e); er != nil {
				p = append(p, er)
			}
		}
	}

	for _, t := range getFrameVendor() {
		if t == getNilFrame() {
			continue
		}
		msg += "\n " + fmt.Sprintf("Fct: %s - File: %s - Line: %d", t.Function, t.File, t.Line)
	}

	return &ers{
		c: 0,
		e: msg,
		p: p,
		t: getFrame(),
	}
}

func IfError(code uint16, message string, parent ...error) Error {
	p := make([]Error, 0)

	if len(parent) > 0 {
		for _, e := range parent {
			if er := Make(e); er != nil {
				p = append(p, er)
			}
		}
	}

	if len(p) < 1 {
		return nil
	}

	return &ers{
		c: code,
		e: message,
		p: p,
		t: getFrame(),
	}
}

func NewDefaultReturn() *DefaultReturn {
	return &DefaultReturn{
		Code:    "",
		Message: "",
	}
}
