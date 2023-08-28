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
)

var (
	defaultPattern      = "[Error #%s] %s"
	defaultPatternTrace = "[Error #%s] %s (%s)"
)

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

type ers struct {
	c uint16
	e string
	p []Error
	t runtime.Frame
}

type FuncMap func(e error) bool

type Error interface {
	//IsCode check if the given error code is matching with the current Error
	IsCode(code CodeError) bool
	//HasCode check if current error or parent has the given error code
	HasCode(code CodeError) bool
	//GetCode return the CodeError value of the current error
	GetCode() CodeError
	//GetParentCode return a slice of CodeError value of all parent Error and the code of the current Error
	GetParentCode() []CodeError

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

func (e *ers) Add(parent ...error) {
	for _, v := range parent {
		if v == nil {
			continue
		}

		if err, ok := v.(Error); !ok {
			e.p = append(e.p, &ers{
				c: 0,
				e: v.Error(),
				p: nil,
			})
		} else {
			e.p = append(e.p, err)
		}
	}
}

func (e *ers) IsCode(code CodeError) bool {
	return e.c == code.GetUint16()
}

func (e *ers) IsError(err error) bool {
	return e.e == err.Error()
}

func (e *ers) HasCode(code CodeError) bool {
	if e.IsCode(code) {
		return true
	}

	for _, p := range e.p {
		if p.IsCode(code) {
			return true
		}
	}

	return false
}

func (e *ers) GetCode() CodeError {
	return CodeError(e.c)
}

func (e *ers) GetParentCode() []CodeError {
	var res = make([]CodeError, 0)

	res = append(res, e.GetCode())
	for _, p := range e.p {
		res = append(res, p.GetParentCode()...)
	}

	return unicCodeSlice(res)
}

func (e *ers) HasError(err error) bool {
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

func (e *ers) HasParent() bool {
	return len(e.p) > 0
}

func (e *ers) GetParent(withMainError bool) []error {
	var res = make([]error, 0)

	if withMainError {
		res = append(res, &ers{
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

func (e *ers) SetParent(parent ...error) {
	e.p = make([]Error, 0)
	e.Add(parent...)
}

func (e *ers) Map(fct FuncMap) bool {
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

func (e *ers) ContainsString(s string) bool {
	if strings.Contains(e.e, s) {
		return true
	} else {
		for _, i := range e.p {
			if i.ContainsString(s) {
				return true
			}
		}
	}

	return false
}

func (e *ers) Code() uint16 {
	return e.c
}

func (e *ers) CodeSlice() []uint16 {
	var r = []uint16{e.Code()}

	for _, v := range e.p {
		if v.Code() > 0 {
			r = append(r, v.Code())
		}
	}

	return r
}

func (e *ers) Error() string {
	return modeError.error(e)
}

func (e *ers) StringError() string {
	return e.e
}

func (e *ers) StringErrorSlice() []string {
	var r = []string{e.StringError()}

	for _, v := range e.p {
		r = append(r, v.Error())
	}

	return r
}

func (e *ers) GetError() error {
	//nolint goerr113
	return fmt.Errorf(e.e)
}

func (e *ers) GetErrorSlice() []error {
	var r = []error{e.GetError()}

	for _, v := range e.p {
		for _, s := range v.GetErrorSlice() {
			r = append(r, s)
		}
	}

	return r
}

func (e *ers) GetTrace() string {
	if e.t.File != "" {
		return fmt.Sprintf("%s#%d", filterPath(e.t.File), e.t.Line)
	} else if e.t.Function != "" {
		return fmt.Sprintf("%s#%d", e.t.Function, e.t.Line)
	}

	return ""
}

func (e *ers) GetTraceSlice() []string {
	var r = []string{e.GetTrace()}

	for _, v := range e.p {
		if t := v.GetTrace(); t != "" {
			r = append(r, v.GetTrace())
		}
	}

	return r
}

func (e *ers) CodeError(pattern string) string {
	if pattern == "" {
		pattern = defaultPattern
	}
	return fmt.Sprintf(pattern, e.Code(), e.StringError())
}

func (e *ers) CodeErrorSlice(pattern string) []string {
	var r = []string{e.CodeError(pattern)}

	for _, v := range e.p {
		r = append(r, v.CodeError(pattern))
	}

	return r
}

func (e *ers) CodeErrorTrace(pattern string) string {
	if pattern == "" {
		pattern = defaultPatternTrace
	}

	return fmt.Sprintf(pattern, e.Code(), e.StringError(), e.GetTrace())
}

func (e *ers) CodeErrorTraceSlice(pattern string) []string {
	var r = []string{e.CodeErrorTrace(pattern)}

	for _, v := range e.p {
		r = append(r, v.CodeErrorTrace(pattern))
	}

	return r
}

func (e *ers) Return(r Return) {
	e.ReturnError(r.SetError)
	e.ReturnParent(r.AddParent)
}

func (e *ers) ReturnError(f ReturnError) {
	if e.t.File != "" {
		f(int(e.c), e.e, e.t.File, e.t.Line)
	} else {
		f(int(e.c), e.e, e.t.Function, e.t.Line)
	}
}

func (e *ers) ReturnParent(f ReturnError) {
	for _, p := range e.p {
		p.ReturnError(f)
		p.ReturnParent(f)
	}
}
