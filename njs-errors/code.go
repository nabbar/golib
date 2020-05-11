/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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
 */

package njs_errors

import "fmt"

type ErrorCode interface {
	Error() error
	ErrorWithCode() error

	String() string
	StringWithCode() string

	Code() string
}

type errorCode struct {
	code string
	err  ErrorType
	ori  error
}

// Error return an error type of the current error code, with no code in reference
func (e errorCode) Error() error {
	return e.err.Error(e.ori)
}

// String return a string of the current error code, with no code in reference
func (e errorCode) String() string {
	return e.err.String()
}

// Code return a string of the current code, with no error in reference
func (e errorCode) Code() string {
	return e.code
}

// ErrorWithCode return a error type of the current code, with error and origin in reference
func (e errorCode) StringWithCode() string {
	return e.ErrorWithCode().Error()
}

// ErrorWithCode return a error type of the current code, with error and origin in reference
func (e errorCode) ErrorWithCode() error {
	return fmt.Errorf("(%s) %v", e.code, e.Error())
}
