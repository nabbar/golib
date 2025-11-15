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

// idMsgFct stores the mapping between error codes and their message functions.
// This allows customizing error messages for specific codes.
var idMsgFct = make(map[CodeError]Message)

// Message is a function type that generates error messages based on error codes.
// It allows dynamic message generation or customization per error code.
type Message func(code CodeError) (message string)

// CodeError represents a numeric error code similar to HTTP status codes.
// It is a uint16 allowing codes from 0 to 65535.
// Predefined codes follow HTTP conventions (200, 404, 500, etc.).
type CodeError uint16

const (
	// UnknownError represents an error with no specific code (0).
	// Used as a fallback when error code cannot be determined.
	UnknownError CodeError = 0

	// UnknownMessage is the default message for UnknownError.
	UnknownMessage = "unknown error"

	// NullMessage represents an empty error message.
	NullMessage = ""
)

// ParseCodeError returns a CodeError value based on the input int64 value.
// If the input value is negative, it returns UnknownError.
// If the input value is greater than or equal to math.MaxUint16, it returns math.MaxUint16.
// Otherwise, it returns the input value as a CodeError.
func ParseCodeError(i int64) CodeError {
	if i < 0 {
		return UnknownError
	} else if i >= int64(math.MaxUint16) {
		return math.MaxUint16
	} else {
		return CodeError(i)
	}
}

// NewCodeError returns a CodeError value based on the input uint16 value.
//
// It is a simple wrapper around the type conversion from uint16 to CodeError.
//
// The returned value is the same as the input uint16 value, but with a different type.
// This can be useful when interacting with functions or methods that expect or return a CodeError value.
func NewCodeError(code uint16) CodeError {
	return CodeError(code)
}

// Uint16 returns the CodeError value as a uint16.
//
// It is a simple wrapper around the type conversion from CodeError to uint16.
//
// The returned value is the same as the CodeError value, but with a different type.
// This can be useful when interacting with functions or methods that expect or return a uint16 value.
func (c CodeError) Uint16() uint16 {
	return uint16(c)
}

// Int returns the CodeError value as an int.
//
// It is a simple wrapper around the type conversion from CodeError to int.
//
// The returned value is the same as the CodeError value, but with a different type.
// This can be useful when interacting with functions or methods that expect or return an int value.
func (c CodeError) Int() int {
	return int(c)
}

// String returns a string representation of the CodeError value.
//
// The returned string is the same as the CodeError value, but with a different type.
// This can be useful when interacting with functions or methods that expect or return a string value.
func (c CodeError) String() string {
	return strconv.Itoa(c.Int())
}

// GetMessage returns the string representation of the CodeError value.
// Deprecated: see Message
func (c CodeError) GetMessage() string {
	return c.String()
}

// Message returns a string representation of the CodeError value.
// If the CodeError value is registered and not the UnknownError code,
// it returns the associated message string.
// Otherwise, it returns UnknownMessage.
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

// Error returns an error value based on the CodeError value.
// It takes a variable number of parent errors as arguments.
// The returned error value is a new instance of the `Error` type.
// The CodeError value is used to set the code and message of the returned error value.
// The parent errors are stored in the `Error.Parent` field.
// If the CodeError value is UnknownError, it sets the code to UnknownError and the message to UnknownMessage.
// If the CodeError value is registered, it sets the code to the CodeError value
// and the message to the associated message string.
// Otherwise, it sets the code to UnknownError and the message to UnknownMessage.
func (c CodeError) Error(p ...error) Error {
	return New(c.Uint16(), c.Message(), p...)
}

// Errorf returns an error value based on the CodeError value and arguments.
// It takes a variable number of arguments as arguments.
// The returned error value is a new instance of the `Error` type.
//
// The CodeError value is used to set the code and message of the returned error value.
// The arguments are used to format the message string.
//
// If the message string does not contain any "%" characters, it returns a new error value with the message string as is.
// If the message string contains "%" characters, it formats the message string using the arguments
// and returns a new error value with the formatted message string.
//
// The number of arguments must be less than or equal to the number of "%" characters in the message string.
// If the number of arguments is greater than the number of "%" characters in the message string,
// it ignores the extra arguments.
//
// If the CodeError value is UnknownError, it sets the code to UnknownError and the message to UnknownMessage.
// If the CodeError value is registered, it sets the code to the CodeError value
// and the message to the associated message string.
//
// Otherwise, it sets the code to UnknownError and the message to UnknownMessage.
func (c CodeError) Errorf(args ...interface{}) Error {
	m := c.Message()

	if !strings.Contains(m, "%") {
		return New(c.Uint16(), m)
	}

	if n := strings.Count(m, "%"); n < len(args) {
		return Newf(c.Uint16(), m, args[:n]...)
	} else {
		return Newf(c.Uint16(), m, args...)
	}
}

// IfError returns an error value based on the CodeError value and parent error.
// The Error is returned only if the filtered parent list contain a valid error (not nil and with a string result).
// Otherwise, the return value is nil.
//
// It takes the CodeError value, its associated message string, and a parent error as arguments.
// The returned error value is a new instance of the `Error` type.
// The CodeError value is used to set the code and message of the returned error value.
// The parent error is stored in the `Error.Parent` field.
//
// If the CodeError value is UnknownError, it sets the code to UnknownError and the message to UnknownMessage.
// If the CodeError value is registered, it sets the code to the CodeError value
// and the message to the associated message string.
// Otherwise, it sets the code to UnknownError and the message to UnknownMessage.
func (c CodeError) IfError(e ...error) Error {
	return IfError(c.Uint16(), c.Message(), e...)
}

// GetCodePackages returns a map of CodeError to string that contains the path
// of the files associated with the CodeError.
//
// It takes the root package name as argument and return a map where the key is the CodeError
// and the value is the file path associated with the CodeError.
// The returned map contains only the CodeError that are registered and has a file path associated with them.
//
// The file path is a relative path from the root package.
// If the file path contains "/vendor/", it is removed from the path.
// If the file path starts with the root package, it is removed from the path.
// If the file path does not start with "/", it is prefixed with "/".
//
// The returned map is empty if no CodeError is registered or if no file path is associated with any CodeError.
func GetCodePackages(rootPackage string) map[CodeError]string {
	var res = make(map[CodeError]string)

	for i, f := range idMsgFct {
		p := reflect.ValueOf(f).Pointer()
		n, _ := runtime.FuncForPC(p).FileLine(p)

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

// RegisterIdFctMessage registers a message function associated with a CodeError value.
//
// The message function takes a CodeError value as argument and returns a string.
// The returned string is the message associated with the CodeError value.
// The message function must return a non-empty string.
//
// The registered message function is stored in a map where the key is the CodeError value
// and the value is the message function.
// The map is ordered by the CodeError value.
//
// The message function is used to get the message associated with a CodeError value.
// If the CodeError value is not registered, it returns an empty string.
//
// The minimum CodeError value is required to ensure that the message function is registered
// with the correct CodeError value.
//
// The returned message string is empty if the CodeError value is not registered.
//
// Example:
//
//	func testMessage(code CodeError) string {
//	 switch code {
//	 case MyErrorCode:
//	     return "Test error message"
//	 default:
//	     return UnknownMessage
//	 }
//	}
//
// RegisterIdFctMessage(MyErrorCode, testMessage)
//
// err := MyErrorCode.Error()
// fmt.Println(err.Message()) // "Test error message"
func RegisterIdFctMessage(minCode CodeError, fct Message) {
	if idMsgFct == nil {
		idMsgFct = make(map[CodeError]Message)
	}

	idMsgFct[minCode] = fct
	orderMapMessage()
}

// ExistInMapMessage checks if a message is registered for a CodeError value.
//
// It takes a CodeError value as argument and returns a boolean value.
// The returned boolean value is true if the CodeError value is registered and has a message associated with it.
// Otherwise, it returns false.
//
// The function is used to check if a CodeError value is registered and has a message associated with it.
// If the CodeError value is not registered, it returns false.
//
// The function is used in conjunction with the RegisterIdFctMessage function to register and check for CodeError values.
//
// Example:
//
//	func testMessage(code CodeError) string {
//	 switch code {
//	 case MyErrorCode:
//	     return "Test error message"
//	 default:
//	     return UnknownMessage
//	 }
//	}
//
//	if ExistInMapMessage(MyErrorCode) {
//	   panic("CodeError collision detected")
//	}
//
// RegisterIdFctMessage(MyErrorCode, testMessage)
func ExistInMapMessage(code CodeError) bool {
	if f, ok := idMsgFct[findCodeErrorInMapMessage(code)]; ok {
		if m := f(code); m != NullMessage {
			return true
		}
	}

	return false
}

func getMapMessageKey() []CodeError {
	var (
		keys = make([]int, 0)
		res  = make([]CodeError, 0)
	)

	for k := range idMsgFct {
		keys = append(keys, int(k.Uint16()))
	}

	sort.Ints(keys)

	for _, k := range keys {
		// prevent overflow
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

func orderMapMessage() {
	var res = make(map[CodeError]Message)

	for _, k := range getMapMessageKey() {
		res[k] = idMsgFct[k]
	}

	idMsgFct = res
}

func findCodeErrorInMapMessage(code CodeError) CodeError {
	var res CodeError = 0

	for _, k := range getMapMessageKey() {
		if k <= code && k > res {
			res = k
		}
	}

	return res
}

func isCodeInSlice(code CodeError, slice []CodeError) bool {
	for _, c := range slice {
		if c == code {
			return true
		}
	}

	return false
}

func unicCodeSlice(slice []CodeError) []CodeError {
	var res = make([]CodeError, 0)

	for _, c := range slice {
		if !isCodeInSlice(c, res) {
			res = append(res, c)
		}
	}

	return res
}
