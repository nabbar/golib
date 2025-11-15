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

type ers struct {
	c uint16
	e string
	p []Error
	t runtime.Frame
}

func (e *ers) is(err *ers) bool {
	if e == nil || err == nil {
		return false
	}

	var (
		ss = e.GetTrace()
		sd = err.GetTrace()
		ts = len(ss) > 0
		td = len(sd) > 0
	)

	// XOR Trace Source & Destination != 0
	if (ts || td) && !(ts && td) { // nolint
		return false
	} else if ts && td {
		return strings.EqualFold(ss, sd)
	}

	ss = e.Error()
	sd = err.Error()
	ts = len(ss) > 0
	td = len(sd) > 0

	// XOR Message Source & Destination != 0
	if (ts || td) && !(ts && td) { // nolint
		return false
	} else if ts && td {
		return strings.EqualFold(ss, sd)
	}

	var (
		cs = e.Code()
		cd = err.Code()
	)

	ts = cs > 0
	td = cd > 0

	// XOR Message Source & Destination != 0
	if (ts || td) && !(ts && td) { // nolint
		return false
	} else if ts && td {
		return cs == cd
	}

	return false
}

func (e *ers) Is(err error) bool {
	if err == nil {
		return false
	}

	if er, ok := err.(*ers); ok {
		return e.is(er)
	} else {
		return e.IsError(err)
	}
}

func (e *ers) Add(parent ...error) {
	for _, v := range parent {
		if v == nil {
			continue
		}

		var (
			ok  bool
			er  *ers
			err Error
		)

		if er, ok = v.(*ers); ok {
			// prevent circular addition
			if e.IsError(er) {
				for _, erp := range er.p {
					e.Add(erp)
				}
			} else {
				e.p = append(e.p, er)
			}
		} else if err, ok = v.(Error); !ok {
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
	return e.c == code.Uint16()
}

func (e *ers) IsError(err error) bool {
	return strings.EqualFold(e.e, err.Error())
}

func (e *ers) HasCode(code CodeError) bool {
	if e.IsCode(code) {
		return true
	}

	for _, p := range e.p {
		if p.HasCode(code) {
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
		} else if p.HasError(err) {
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
	return errors.New(e.e)
}

func (e *ers) GetErrorSlice() []error {
	var r = []error{e.GetError()}

	if len(e.p) < 1 {
		return r
	}

	for _, v := range e.p {
		if v == nil {
			continue
		}

		r = append(r, v.GetErrorSlice()...)
	}

	return r
}

func (e *ers) Unwrap() []error {
	if len(e.p) < 1 {
		return nil
	}

	var r = make([]error, 0)

	for _, v := range e.p {
		if v == nil {
			continue
		}

		r = append(r, v)
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
