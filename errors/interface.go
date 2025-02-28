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
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

type FuncMap func(e error) bool

type ReturnError func(code int, msg string, file string, line int)

type Error interface {
	//IsCode check if the given error code is matching with the current Error
	IsCode(code CodeError) bool
	//HasCode check if current error or parent has the given error code
	HasCode(code CodeError) bool
	//GetCode return the CodeError value of the current error
	GetCode() CodeError
	//GetParentCode return a slice of CodeError value of all parent Error and the code of the current Error
	GetParentCode() []CodeError

	//Is implement compatiblity with GOROOT/src/errors/wrap Is fucntion
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

type Return interface {
	SetError(code int, msg string, file string, line int)
	AddParent(code int, msg string, file string, line int)
	JSON() []byte
}

type ReturnGin interface {
	Return

	GinTonicAbort(ctx *gin.Context, httpCode int)
	GinTonicErrorAbort(ctx *gin.Context, httpCode int)
}

func Is(e error) bool {
	var err Error
	return errors.As(e, &err)
}

func Get(e error) Error {
	var err Error
	if errors.As(e, &err) {
		return err
	}

	return nil
}

func Has(e error, code CodeError) bool {
	if err := Get(e); err == nil {
		return false
	} else {
		return err.HasCode(code)
	}
}

func ContainsString(e error, s string) bool {
	if e == nil {
		return false
	} else if err := Get(e); err == nil {
		return strings.Contains(e.Error(), s)
	} else {
		return err.ContainsString(s)
	}
}

func IsCode(e error, code CodeError) bool {
	if err := Get(e); err == nil {
		return false
	} else {
		return err.IsCode(code)
	}
}

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

func NewErrorTrace(code int, msg string, file string, line int, parent ...error) Error {
	var p = make([]Error, 0)

	if len(parent) > 0 {
		for _, e := range parent {
			if er := Make(e); er != nil {
				p = append(p, er)
			}
		}
	}

	return &ers{
		c: uint16(code),
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
