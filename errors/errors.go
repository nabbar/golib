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

// SetTracePathFilter define the glue string to be used to join main error with parents'errors.
func SetDefaultGlue(glue string) {
	defaultGlue = glue
}

// GetDefaultGlue return the current glue used to joins errors with parents.
func GetDefaultGlue() string {
	return defaultGlue
}

// GetDefaultPatternTrace define the pattern to be used for string of error with code.
// The pattern is fmt pattern with 2 inputs in order : code, message.
func SetDefaultPattern(pattern string) {
	defaultPattern = pattern
}

// GetDefaultPattern return the current pattern used for string of error with code.
// The pattern is fmt pattern with 2 inputs in order : code, message.
func GetDefaultPattern() string {
	return defaultPattern
}

// SetDefaultPatternTrace define the pattern to be used for string of error with code and trace.
// The pattern is fmt pattern with 3 inputs in order : code, message, trace.
func SetDefaultPatternTrace(patternTrace string) {
	defaultPatternTrace = patternTrace
}

// GetDefaultPatternTrace return the current pattern used for string of error with code and trace.
// The pattern is fmt pattern with 3 inputs in order : code, message, trace.
func GetDefaultPatternTrace() string {
	return defaultPatternTrace
}

// SetTracePathFilter customize the filter apply to filepath on trace.
func SetTracePathFilter(path string) {
	filterPkg = path
}

type errors struct {
	c uint16
	e string
	p []Error
	t runtime.Frame
}

type FuncMap func(e Error) bool

type Error interface {
	//IsCodeError check if the given error code is matching with the current Error
	IsCodeError(code CodeError) bool
	//HasCodeError check if current error or parent has the given error code
	HasCodeError(code CodeError) bool
	//GetCodeError return the CodeError value of the current error
	GetCodeError() CodeError
	//GetCodeErrorParent return a slice of CodeError value of all parent Error and the code of the current Error
	GetCodeErrorParent() []CodeError

	//IsError check if the given error params is a valid error and not a nil pointer
	IsError(e error) bool
	//HasError check if the given error in params is still in parent error
	HasError(err error) bool
	//HasParent check if the current Error has any valid parent
	HasParent() bool
	//GetParent return a slice of Error interface for each parent error with or without the first error.
	GetParent(withMainError bool) []Error
	//Map run a function on each func and parent. If the function return false, the loop stop.
	Map(fct FuncMap) bool

	//AddParent will add all no empty given error into parent of the current Error pointer
	AddParent(parent ...error)
	//SetParent will replace all parent with the given error list
	SetParent(parent ...error)
	//AddParentError will add all no empty given Error into parent of the current Error pointer
	AddParentError(parent ...Error)
	//SetParentError will replace all parent with the given Error list
	SetParentError(parent ...Error)

	//Code is used to return the code of current Error, as string
	Code() string
	//CodeFull is used to return a joint string of code of current Error and code of all parent Error
	CodeFull(glue string) string
	//CodeSlice is used to return a slice string of all code of current Error (main and parent)
	CodeSlice() []string

	//CodeError is used to return a composed string of current Error code with message, for current Error and no parent
	CodeError(pattern string) string
	//CodeErrorFull is used to return a composed string of couple error code with message, for current Error and all parent
	CodeErrorFull(pattern, glue string) string
	//CodeErrorSlice is used to return a composed string slice of couple error code with message, for current Error and all parent
	CodeErrorSlice(pattern string) []string

	//CodeErrorTrace is used to return a composed string of current Error code with message and trace information, for current Error and no parent
	CodeErrorTrace(pattern string) string
	//CodeErrorTraceFull is used to return a composed string of couple error code with message and trace information, for current Error and all parent
	CodeErrorTraceFull(pattern, glue string) string
	//CodeErrorTraceSlice is used to return a composed string slice of couple error code with message and trace information, for current Error and all parent
	CodeErrorTraceSlice(pattern string) []string

	//Error is used to match with error interface
	//this function will return a mixed result depends of the configuration defined by calling SetModeReturnError
	Error() string

	//StringError is used to return the error message, for current Error and no parent
	StringError() string
	//StringErrorFull is used to return the error message, for current Error and all parent
	StringErrorFull(glue string) string
	//StringErrorSlice is used to return the error message, for current Error and all parent, as a slice of string
	StringErrorSlice() []string

	//GetError is used to return a new error interface based of the current error (and no parent)
	GetError() error
	//GetErrorFull is used to return a new error interface based of the current error with all parent
	GetErrorFull(glue string) error
	//GetErrorSlice is used to return a slice of new error interface, based of the current error and all parent
	GetErrorSlice() []error

	//GetIError is used to return a Error interface pointer based of current Error
	GetIError() Error
	//GetIErrorSlice is used to return a slice of Error interface pointer, based of current Error and all parents
	GetIErrorSlice() []Error

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

func NewErrorTrace(code int, msg string, file string, line int, parent Error) Error {
	var p = make([]Error, 0)

	if parent != nil {
		p = parent.GetIErrorSlice()
	}

	return &errors{
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
		p = append(p, &errors{
			c: 0,
			e: recovered,
			p: nil,
		})
	}

	if len(parent) > 0 {
		for _, err := range parent {
			if err == nil {
				continue
			}

			p = append(p, &errors{
				c: 0,
				e: err.Error(),
				p: nil,
			})
		}
	}

	for _, t := range getFrameVendor() {
		if t == getNilFrame() {
			continue
		}
		msg += "\n " + fmt.Sprintf("Fct: %s - File: %s - Line: %d", t.Function, t.File, t.Line)
	}

	return &errors{
		c: 0,
		e: msg,
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
		if v != nil {
			e.p = append(e.p, &errors{
				c: 0,
				e: v.Error(),
				p: nil,
			})
		}
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

func (e *errors) GetCodeError() CodeError {
	return CodeError(e.c)
}

func (e *errors) GetCodeErrorParent() []CodeError {
	var res = make([]CodeError, 0)

	res = append(res, e.GetCodeError())
	for _, p := range e.p {
		res = append(res, p.GetCodeErrorParent()...)
	}

	return unicCodeSlice(res)
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

func (e *errors) HasParent() bool {
	return len(e.p) > 0
}

func (e *errors) GetParent(withMainError bool) []Error {
	var res = make([]Error, 0)

	if withMainError {
		res = append(res, &errors{
			c: e.c,
			e: e.e,
			p: nil,
			t: e.t,
		})
	}

	if len(e.p) > 0 {
		for _, er := range e.p {
			res = append(res, er.GetParent(true)...)
		}
	}

	return res
}

func (e *errors) SetParent(parent ...error) {
	e.p = make([]Error, 0)
	e.AddParent(parent...)
}

func (e *errors) Map(fct FuncMap) bool {
	if !fct(e) {
		return false
	} else if len(e.p) > 0 {
		for _, er := range e.p {
			if !er.Map(fct) {
				return false
			}
		}
	}

	return true
}

func (e *errors) AddParentError(parent ...Error) {
	for _, p := range parent {
		if p != nil {
			e.p = append(e.p, p)
		}
	}
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
	//nolint goerr113
	return errs.New(e.e)
}

func (e *errors) GetErrorFull(glue string) error {
	//nolint goerr113
	return errs.New(e.StringErrorFull(glue))
}

func (e *errors) GetErrorSlice() []error {
	var r = []error{e.GetError()}

	for _, v := range e.p {
		for _, s := range v.GetErrorSlice() {
			r = append(r, s)
		}
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
		return fmt.Sprintf("%s#%d", filterPath(e.t.File), e.t.Line)
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

func (e *errors) Return(r Return) {
	e.ReturnError(r.SetError)
	e.ReturnParent(r.AddParent)
}

func (e *errors) ReturnError(f ReturnError) {
	if e.t.File != "" {
		f(int(e.c), e.e, e.t.File, e.t.Line)
	} else {
		f(int(e.c), e.e, e.t.Function, e.t.Line)
	}
}

func (e *errors) ReturnParent(f ReturnError) {
	for _, p := range e.p {
		p.ReturnError(f)
		p.ReturnParent(f)
	}
}
