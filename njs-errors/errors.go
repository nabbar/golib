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

package njs_errors

import (
	errs "errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var defaultGlue = ", "

func SetDefaultGlue(glue string) {
	defaultGlue = glue
}

func GetDefaultGlue() string {
	return defaultGlue
}

type errors struct {
	c uint16
	e string
	p []Error
	t runtime.Frame
}

type Error interface {
	IsCodeError(code CodeError) bool
	HasCodeError(code CodeError) bool

	IsError(e error) bool
	HasError(err error) bool

	AddParent(parent ...error)
	SetParent(parent ...error)
	AddParentError(parent ...Error)
	SetParentError(parent ...Error)

	Code() string
	CodeFull(glue string) string
	CodeSlice() []string

	Error() string
	ErrorFull(glue string) string
	ErrorSlice() []string

	GetError() error
	GetErrorFull(glue string) error
	GetErrorSlice() []error

	GetIError() Error
	GetIErrorSlice() []Error

	GetTrace() string
	GetTraceSlice() []string
}

func NewError(code uint16, message string, parent Error) Error {
	if parent == nil {
		parent = &errors{
			c: 0,
			e: "",
			p: make([]Error, 0),
		}
	}

	return &errors{
		c: code,
		e: message,
		p: parent.GetIErrorSlice(),
		t: getFrame(),
	}
}

func (e *errors) AddParent(parent ...error) {
	for _, v := range parent {
		e.p = append(e.p, &errors{
			c: 0,
			e: v.Error(),
			p: nil,
		})
	}
}

func (e *errors) IsCodeError(code CodeError) bool {
	return e.c == code.GetUint16()
}

func (e *errors) IsError(err error) bool {
	return e.e == err.Error()
}

func (e *errors) HasCodeError(code CodeError) bool {
	if e.IsCodeError(code) {
		return true
	}

	for _, p := range e.p {
		if p.IsCodeError(code) {
			return true
		}
	}

	return false
}

func (e *errors) HasError(err error) bool {
	if e.IsError(err) {
		return true
	}

	for _, p := range e.p {
		if p.IsError(err) {
			return true
		}
	}

	return false
}

func (e *errors) SetParent(parent ...error) {
	e.p = make([]Error, 0)
	e.AddParent(parent...)
}

func (e *errors) AddParentError(parent ...Error) {
	e.p = append(e.p, parent...)
}

func (e *errors) SetParentError(parent ...Error) {
	e.p = parent
}

func (e *errors) Code() string {
	return strconv.Itoa(int(e.c))
}

func (e *errors) CodeFull(glue string) string {
	if glue == "" {
		return strings.Join(e.CodeSlice(), glue)
	}

	return strings.Join(e.CodeSlice(), defaultGlue)
}

func (e *errors) CodeSlice() []string {
	var r = []string{e.Code()}

	for _, v := range e.p {
		r = append(r, v.Code())
	}

	return r
}

func (e errors) Error() string {
	return e.e
}

func (e errors) ErrorFull(glue string) string {
	if glue == "" {
		return strings.Join(e.ErrorSlice(), glue)
	}

	return strings.Join(e.ErrorSlice(), defaultGlue)
}

func (e errors) ErrorSlice() []string {
	var r = []string{e.Error()}

	for _, v := range e.p {
		r = append(r, v.Error())
	}

	return r
}

func (e *errors) GetError() error {
	return errs.New(e.e)
}

func (e *errors) GetErrorFull(glue string) error {
	return errs.New(e.ErrorFull(glue))
}

func (e *errors) GetErrorSlice() []error {
	var r = []error{e.GetError()}

	for _, v := range e.p {
		r = append(r, v.GetError())
	}

	return r
}

func (e *errors) GetIError() Error {
	return e
}

func (e *errors) GetIErrorSlice() []Error {
	var r = []Error{e}

	for _, v := range e.p {
		r = append(r, v.GetIError())
	}

	return r
}

func (e *errors) GetTrace() string {
	if e.t.File != "" {
		return fmt.Sprintf("%s#%d", e.t.File, e.t.Line)
	} else if e.t.Function != "" {
		return fmt.Sprintf("%s#%d", e.t.Function, e.t.Line)
	}

	return ""
}

func (e *errors) GetTraceSlice() []string {
	var r = []string{e.GetTrace()}

	for _, v := range e.p {
		if t := v.GetTrace(); t != "" {
			r = append(r, v.GetTrace())
		}
	}

	return r
}
