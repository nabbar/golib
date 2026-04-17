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
	"strings"
)

// ers is the internal implementation of the Error interface.
// It stores the error message, numeric code, parent errors, and stack trace information.
type ers struct {
	e string   // e is the error message string.
	c uint16   // c is the numeric error code.
	p []Error  // p is the slice of parent errors for hierarchical chaining.
	t tracer   // t is the stack trace information captured at error creation.
}

// is compares two *ers internal structures for equality.
// It checks the stack trace, the error code, and performs a case-insensitive message comparison.
func (e *ers) is(err *ers) bool {
	if e == nil || err == nil {
		return false
	}

	// Trace comparison ensures the error was created at the same location.
	if !e.t.IsSame(err.t) {
		return false
	}

	// Compare numeric error codes.
	if e.c != err.c {
		return false
	}

	// Compare error messages (fast path for exact match).
	if e.e == err.e {
		return true
	}
	
	// Case-insensitive comparison if lengths match.
	if len(e.e) != len(err.e) {
		return false
	}
	return strings.EqualFold(e.e, err.e)
}

// Is implements the standard library's error comparison (errors.Is).
// It first attempts a fast type assertion to *ers, then falls back to message comparison.
func (e *ers) Is(err error) bool {
	if err == nil {
		return false
	}

	// Direct type assertion is faster and zero-alloc compared to errors.As.
	if er, ok := err.(*ers); ok {
		return e.is(er)
	}

	// Fallback to error message comparison for standard errors.
	msg := err.Error()
	if e.e == msg {
		return true
	}
	if len(e.e) != len(msg) {
		return false
	}
	return strings.EqualFold(e.e, msg)
}

// Add appends one or more errors as parents to the current error.
// This creates a hierarchical structure where one error can have multiple causes or contexts.
func (e *ers) Add(parent ...error) {
	if len(parent) == 0 {
		return
	}

	// Initialize parent slice if it's currently nil.
	if e.p == nil {
		// Optimization: Pre-allocate slice capacity to reduce reallocations.
		e.p = make([]Error, 0, len(parent)+1)
	}

	// Manual capacity management to ensure efficient append operations.
	if cap(e.p)-len(e.p) < len(parent) {
		if cap(e.p) < len(parent) {
			// Extend capacity to fit exactly the new parents + 1 extra space.
			newP := make([]Error, len(e.p), len(e.p)+len(parent)+1)
			copy(newP, e.p)
			e.p = newP
		} else {
			// Double the capacity if it's already large enough to handle some growth.
			newP := make([]Error, len(e.p), cap(e.p)*2)
			copy(newP, e.p)
			e.p = newP
		}
	}

	for i := range parent {
		if parent[i] == nil {
			continue
		}

		if er, ok := parent[i].(*ers); ok {
			// Prevent circular addition: if the error to add is already in the chain,
			// we only append its parents to avoid infinite recursion during traversal.
			if e.IsError(er) {
				if len(er.p) > 0 {
					e.p = append(e.p, er.p...)
				}
			} else {
				e.p = append(e.p, er)
			}
		} else if err, ok := parent[i].(Error); ok {
			e.p = append(e.p, err)
		} else {
			// Wrap standard Go errors into our internal structure without trace.
			e.p = append(e.p, &ers{
				c: 0,
				e: parent[i].Error(),
				t: trcNil{},
			})
		}
	}
}

// IsCode checks if the error's direct code matches the provided CodeError.
func (e *ers) IsCode(code CodeError) bool {
	return e.c == code.Uint16()
}

// IsError checks if the given error is equivalent to the current error.
// It compares the message (case-insensitive) and, if the error is our internal type, the stack trace.
func (e *ers) IsError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	if e.e != msg {
		if len(e.e) != len(msg) {
			return false
		}
		if !strings.EqualFold(e.e, msg) {
			return false
		}
	}

	// Identical trace check for precise identification of the error instance.
	if er, ok := err.(*ers); ok {
		return e.t.IsSame(er.t)
	}

	// If the error is not an *ers, but the message matched, we consider it the same error.
	return true
}

// HasCode recursively searches for a specific error code in the current error and all its parents.
func (e *ers) HasCode(code CodeError) bool {
	c := code.Uint16()
	if e.c == c {
		return true
	}

	// Fast path: if no parents, stop here.
	if len(e.p) == 0 {
		return false
	}

	for i := range e.p {
		if e.p[i].HasCode(code) {
			return true
		}
	}

	return false
}

// GetCode returns the direct numeric code of the current error.
func (e *ers) GetCode() CodeError {
	return CodeError(e.c)
}

// GetParentCode returns a unique list of all error codes present in the hierarchy, starting with the current error.
func (e *ers) GetParentCode() []CodeError {
	if len(e.p) == 0 {
		return []CodeError{e.GetCode()}
	}

	var (
		res  = make([]CodeError, 0, len(e.p)+1)
		seen = make(map[CodeError]struct{}, len(e.p)+1)
	)

	// Recursive internal function to traverse the hierarchy and collect unique codes.
	var fill func(err Error)
	fill = func(err Error) {
		c := err.GetCode()
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			res = append(res, c)
		}
		if er, ok := err.(*ers); ok {
			for i := range er.p {
				fill(er.p[i])
			}
		}
	}

	fill(e)
	return res
}

// HasError recursively checks if the given error (or an equivalent one) exists within the error hierarchy.
func (e *ers) HasError(err error) bool {
	if e.IsError(err) {
		return true
	}

	for i := range e.p {
		if e.p[i].IsError(err) || e.p[i].HasError(err) {
			return true
		}
	}

	return false
}

// HasParent returns true if the current error has at least one parent error attached.
func (e *ers) HasParent() bool {
	return len(e.p) > 0
}

// GetParent flattens the error hierarchy into a slice of errors.
// If withMainError is true, the current error (isolated from its parents) is included first.
func (e *ers) GetParent(withMainError bool) []error {
	var res = make([]error, 0, len(e.p)+1)

	if withMainError {
		res = append(res, &ers{
			c: e.c,
			e: e.e,
			p: nil,
			t: e.t,
		})
	}

	for i := range e.p {
		res = append(res, e.p[i].GetParent(true)...)
	}

	return res
}

// SetParent replaces any existing parents with the provided list of errors.
func (e *ers) SetParent(parent ...error) {
	e.p = nil
	e.Add(parent...)
}

// Map executes a function for each error in the hierarchy.
// Iteration continues as long as the function returns true.
func (e *ers) Map(fct FuncMap) bool {
	if !fct(e) {
		return false
	}
	for i := range e.p {
		if !e.p[i].Map(fct) {
			return false
		}
	}

	return true
}

// ContainsString recursively checks if the error message (or any parent's message) contains the specified substring.
func (e *ers) ContainsString(s string) bool {
	if strings.Contains(e.e, s) {
		return true
	}
	for i := range e.p {
		if e.p[i].ContainsString(s) {
			return true
		}
	}

	return false
}

// Code returns the current error's numeric code as a uint16.
func (e *ers) Code() uint16 {
	return e.c
}

// CodeSlice returns a flattened slice of all numeric codes in the error hierarchy.
func (e *ers) CodeSlice() []uint16 {
	var r = make([]uint16, 0, len(e.p)+1)
	r = append(r, e.c)

	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i].CodeSlice()...)
		}
	}

	return r
}

// Error returns a string representation of the error hierarchy based on the globally configured ModeError.
func (e *ers) Error() string {
	return modeError.error(e)
}

// StringError returns the direct error message string of the current error.
func (e *ers) StringError() string {
	return e.e
}

// StringErrorSlice returns a flattened slice of all error message strings in the hierarchy.
func (e *ers) StringErrorSlice() []string {
	var r = make([]string, 0, len(e.p)+1)
	r = append(r, e.e)

	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i].StringErrorSlice()...)
		}
	}

	return r
}

// GetError creates a standard Go error from the current error message.
func (e *ers) GetError() error {
	return errors.New(e.e)
}

// GetErrorSlice returns a slice of standard Go errors, one for each error in the hierarchy.
func (e *ers) GetErrorSlice() []error {
	var r = make([]error, 0, len(e.p)+1)
	r = append(r, errors.New(e.e))

	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i].GetErrorSlice()...)
		}
	}

	return r
}

// Unwrap implements the Go 1.20 multi-error unwrapping interface.
// It returns the direct parent errors as a slice.
func (e *ers) Unwrap() []error {
	if len(e.p) == 0 {
		return nil
	}

	var r = make([]error, 0, len(e.p))
	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i])
		}
	}

	return r
}

// GetTrace returns a string representation of the error's stack trace.
func (e *ers) GetTrace() string {
	return e.t.String()
}

// GetTraceSlice returns a slice containing the stack traces of the current error and all its parents.
func (e *ers) GetTraceSlice() []string {
	var r = make([]string, 0, len(e.p)+1)
	if t := e.GetTrace(); t != "" {
		r = append(r, t)
	}

	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i].GetTraceSlice()...)
		}
	}

	return r
}

// CodeError returns a formatted string containing the code and message based on the given pattern.
// If the pattern is empty, the default pattern is used.
func (e *ers) CodeError(pattern string) string {
	if pattern == "" {
		pattern = defaultPattern
	}
	return fmt.Sprintf(pattern, e.c, e.e)
}

// CodeErrorSlice returns a slice of formatted code-message strings for the entire hierarchy.
func (e *ers) CodeErrorSlice(pattern string) []string {
	var r = make([]string, 0, len(e.p)+1)
	r = append(r, e.CodeError(pattern))

	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i].CodeErrorSlice(pattern)...)
		}
	}

	return r
}

// CodeErrorTrace returns a formatted string containing the code, message, and trace.
func (e *ers) CodeErrorTrace(pattern string) string {
	if pattern == "" {
		pattern = defaultPatternTrace
	}

	return fmt.Sprintf(pattern, e.c, e.e, e.GetTrace())
}

// CodeErrorTraceSlice returns a slice of formatted code-message-trace strings for the entire hierarchy.
func (e *ers) CodeErrorTraceSlice(pattern string) []string {
	var r = make([]string, 0, len(e.p)+1)
	r = append(r, e.CodeErrorTrace(pattern))

	for i := range e.p {
		if e.p[i] != nil {
			r = append(r, e.p[i].CodeErrorTraceSlice(pattern)...)
		}
	}

	return r
}

// Return invokes the SetError and AddParent methods on the provided Return interface for the entire hierarchy.
func (e *ers) Return(r Return) {
	e.ReturnError(r.SetError)
	e.ReturnParent(r.AddParent)
}

// ReturnError executes the provided ReturnError callback with the current error's details.
func (e *ers) ReturnError(f ReturnError) {
	fc, fl, ln := e.t.Compose()
	if fl != "" {
		f(int(e.c), e.e, fl, ln)
	} else {
		f(int(e.c), e.e, fc, ln)
	}
}

// ReturnParent recursively executes the ReturnError callback for all parent errors in the hierarchy.
func (e *ers) ReturnParent(f ReturnError) {
	for i := range e.p {
		e.p[i].ReturnError(f)
		e.p[i].ReturnParent(f)
	}
}
