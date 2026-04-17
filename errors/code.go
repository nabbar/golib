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
	"math"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// idMsgFct stores the mapping between base error codes and their message generator functions.
// This mapping allows dynamic message generation for specific ranges or individual error codes.
var idMsgFct = make(map[CodeError]Message)

// Message is a callback function type that generates an error message string for a given CodeError.
// This allows customization or dynamic message generation based on the code value.
type Message func(code CodeError) (message string)

// CodeError represents a numeric identifier for an error, similar to an HTTP status code.
// It is a uint16, allowing for a wide range of custom and predefined codes (0-65535).
type CodeError uint16

const (
	// UnknownError is the default code (0) used when no specific error code is defined.
	UnknownError CodeError = 0

	// UnknownMessage is the default message used for UnknownError.
	UnknownMessage = "unknown error"

	// NullMessage represents an empty error message.
	NullMessage = ""
)

// ParseCodeError safely converts an int64 into a CodeError, ensuring bounds checking.
// If the value is negative, it returns UnknownError. If it exceeds uint16, it returns math.MaxUint16.
func ParseCodeError(i int64) CodeError {
	if i < 0 {
		return UnknownError
	} else if i >= int64(math.MaxUint16) {
		return math.MaxUint16
	} else {
		return CodeError(i)
	}
}

// NewCodeError is a convenience constructor that converts a uint16 into a CodeError type.
func NewCodeError(code uint16) CodeError {
	return CodeError(code)
}

// Uint16 returns the underlying uint16 value of the CodeError.
func (c CodeError) Uint16() uint16 {
	return uint16(c)
}

// Int returns the CodeError value as an integer.
func (c CodeError) Int() int {
	return int(c)
}

// String returns the string representation of the numeric error code.
func (c CodeError) String() string {
	return strconv.Itoa(c.Int())
}

// GetMessage returns the string representation of the numeric error code.
// Deprecated: This method only returns the numeric string. Use Message() instead for registered messages.
func (c CodeError) GetMessage() string {
	return c.String()
}

// Message looks up the registered message for this error code.
// If the code is UnknownError or no message function is registered for the code's range,
// it returns UnknownMessage.
func (c CodeError) Message() string {
	if c == UnknownError {
		return UnknownMessage
	}

	if f, ok := idMsgFct[findCodeErrorInMapMessage(c)]; ok {
		if m := f(c); m != "" {
			return m
		}
	}

	return UnknownMessage
}

// Error creates a new Error instance with this CodeError as the code and its registered message.
// It also allows adding optional parent errors to the hierarchy.
func (c CodeError) Error(p ...error) Error {
	return New(c.Uint16(), c.Message(), p...)
}

// Errorf creates a new Error instance using this CodeError's message as a formatting pattern.
// If the registered message contains format specifiers (e.g., %s, %d), they are replaced by the provided args.
func (c CodeError) Errorf(args ...interface{}) Error {
	m := c.Message()

	// If the message is not a pattern, create a simple Error.
	if !strings.Contains(m, "%") {
		return New(c.Uint16(), m)
	}

	// Ensure we don't pass more arguments than there are format specifiers in the message.
	if n := strings.Count(m, "%"); n < len(args) {
		return Newf(c.Uint16(), m, args[:n]...)
	} else {
		return Newf(c.Uint16(), m, args...)
	}
}

// IfError returns a new Error only if at least one of the provided parent errors is not nil.
// If all provided errors are nil, it returns nil.
func (c CodeError) IfError(e ...error) Error {
	return IfError(c.Uint16(), c.Message(), e...)
}

// GetCodePackages returns a map where keys are registered CodeErrors and values are the file paths 
// where the message function was registered. This is useful for debugging code-message collisions.
// The rootPackage parameter allows filtering/cleaning the returned file paths.
func GetCodePackages(rootPackage string) map[CodeError]string {
	var res = make(map[CodeError]string)

	for i, f := range idMsgFct {
		p := reflect.ValueOf(f).Pointer()
		n, _ := runtime.FuncForPC(p).FileLine(p)

		// Clean up file paths by removing vendor and local package prefixes.
		if strings.Contains(n, "/vendor/") {
			a := strings.SplitN(n, "/vendor/", 2)
			n = a[1]
		}

		if strings.Contains(n, rootPackage) {
			a := strings.SplitN(n, rootPackage, 2)
			n = a[1]
		}

		if !strings.HasPrefix(n, "/") {
			n = "/" + n
		}

		res[i] = n
	}

	return res
}

// RegisterIdFctMessage associates a message generator function with a minimum CodeError value.
// Any error code equal to or greater than minCode (up to the next registered range) will use this function.
// This allows registering entire blocks of error codes at once.
func RegisterIdFctMessage(minCode CodeError, fct Message) {
	if idMsgFct == nil {
		idMsgFct = make(map[CodeError]Message)
	}

	idMsgFct[minCode] = fct
	orderMapMessage()
}

// ExistInMapMessage checks if a specific CodeError has a registered message generator.
func ExistInMapMessage(code CodeError) bool {
	if f, ok := idMsgFct[findCodeErrorInMapMessage(code)]; ok {
		if m := f(code); m != NullMessage {
			return true
		}
	}

	return false
}

// getMapMessageKey returns a sorted slice of all registered base CodeErrors.
func getMapMessageKey() []CodeError {
	var (
		keys = make([]int, 0, len(idMsgFct))
		res  = make([]CodeError, 0, len(idMsgFct))
	)

	for k := range idMsgFct {
		keys = append(keys, int(k.Uint16()))
	}

	sort.Ints(keys)

	for _, k := range keys {
		// Prevent overflow during conversion back to CodeError.
		var i CodeError
		if k < 0 {
			i = 0
		} else if k > math.MaxUint16 {
			i = math.MaxUint16
		} else {
			i = CodeError(k)
		}

		res = append(res, i)
	}

	return res
}

// orderMapMessage ensures the internal map is consistent with its sorted keys.
func orderMapMessage() {
	var res = make(map[CodeError]Message)

	for _, k := range getMapMessageKey() {
		res[k] = idMsgFct[k]
	}

	idMsgFct = res
}

// findCodeErrorInMapMessage finds the highest registered base CodeError that is less than or equal 
// to the provided code. This implements the range-based message lookup.
func findCodeErrorInMapMessage(code CodeError) CodeError {
	var res CodeError = 0

	for _, k := range getMapMessageKey() {
		if k <= code && k > res {
			res = k
		}
	}

	return res
}
