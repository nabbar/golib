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
	errs "errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

var (
	defaultGlue         = ", "
	defaultPattern      = "[Error #%s] %s"
	defaultPatternTrace = "[Error #%s] %s (%s)"
)

func SetDefaultGlue(glue string) {
	defaultGlue = glue
}

func GetDefaultGlue() string {
	return defaultGlue
}

func SetDefaultPattern(pattern string) {
	defaultPattern = pattern
}

func GetDefaultPattern() string {
	return defaultPattern
}

func SetDefaultPatternTrace(patternTrace string) {
	defaultPatternTrace = patternTrace
}

func GetDefaultPatternTrace() string {
	return defaultPatternTrace
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

	CodeError(pattern string) string
	CodeErrorFull(pattern, glue string) string
	CodeErrorSlice(pattern string) []string

	CodeErrorTrace(pattern string) string
	CodeErrorTraceFull(pattern, glue string) string
	CodeErrorTraceSlice(pattern string) []string

	Error() string

	StringError() string
	StringErrorFull(glue string) string
	StringErrorSlice() []string

	GetError() error
	GetErrorFull(glue string) error
	GetErrorSlice() []error

	GetIError() Error
	GetIErrorSlice() []Error

	GetTrace() string
	GetTraceSlice() []string
}

func MakeErrorIfError(err ...Error) Error {
	var e Error = nil

	for _, p := range err {
		if p == nil {
			continue
		}
		if e == nil {
			e = p
		} else {
			e.AddParentError(p)
		}
	}

	return e
}

func NewError(code uint16, message string, parent Error) Error {
	var p = make([]Error, 0)

	if parent != nil {
		p = parent.GetIErrorSlice()
	}

	return &errors{
		c: code,
		e: message,
		p: p,
		t: getFrame(),
	}
}

func NewErrorIferror(code uint16, message string, parent error) Error {
	if parent == nil {
		return nil
	}

	p := make([]Error, 0)
	p = append(p, &errors{
		c: 0,
		e: parent.Error(),
		p: nil,
	})

	return &errors{
		c: code,
		e: message,
		p: p,
		t: getFrame(),
	}
}

func NewErrorIfError(code uint16, message string, parent Error) Error {
	if parent == nil {
		return nil
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

		if v == nil {
			continue
		}

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
		glue = defaultGlue
	}

	return strings.Join(e.CodeSlice(), glue)
}

func (e *errors) CodeSlice() []string {
	var r = []string{e.Code()}

	for _, v := range e.p {
		r = append(r, v.Code())
	}

	return r
}

func (e *errors) Error() string {
	return modeError.error(e)
}

func (e *errors) StringError() string {
	return e.e
}

func (e *errors) StringErrorFull(glue string) string {
	if glue == "" {
		glue = defaultGlue
	}

	return strings.Join(e.StringErrorSlice(), glue)
}

func (e *errors) StringErrorSlice() []string {
	var r = []string{e.StringError()}

	for _, v := range e.p {
		r = append(r, v.Error())
	}

	return r
}

func (e *errors) GetError() error {
	return errs.New(e.e)
}

func (e *errors) GetErrorFull(glue string) error {
	return errs.New(e.StringErrorFull(glue))
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

func (e *errors) CodeError(pattern string) string {
	if pattern == "" {
		pattern = defaultPattern
	}
	return fmt.Sprintf(pattern, e.Code(), e.StringError())
}

func (e *errors) CodeErrorFull(pattern, glue string) string {
	if glue == "" {
		glue = defaultGlue
	}

	return strings.Join(e.CodeErrorSlice(pattern), glue)
}

func (e *errors) CodeErrorSlice(pattern string) []string {
	var r = []string{e.CodeError(pattern)}

	for _, v := range e.p {
		r = append(r, v.CodeError(pattern))
	}

	return r
}

func (e *errors) CodeErrorTrace(pattern string) string {
	if pattern == "" {
		pattern = defaultPatternTrace
	}

	return fmt.Sprintf(pattern, e.Code(), e.StringError(), e.GetTrace())
}

func (e *errors) CodeErrorTraceFull(pattern, glue string) string {
	if glue == "" {
		glue = defaultGlue
	}

	return strings.Join(e.CodeErrorTraceSlice(pattern), glue)
}

func (e *errors) CodeErrorTraceSlice(pattern string) []string {
	var r = []string{e.CodeErrorTrace(pattern)}

	for _, v := range e.p {
		r = append(r, v.CodeErrorTrace(pattern))
	}

	return r
}
