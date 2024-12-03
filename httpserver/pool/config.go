/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package pool

import (
	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
	liberr "github.com/nabbar/golib/errors"
	libhtp "github.com/nabbar/golib/httpserver"
	srvtps "github.com/nabbar/golib/httpserver/types"
	liblog "github.com/nabbar/golib/logger"
)

type Config []libhtp.Config
type FuncWalkConfig func(cfg libhtp.Config) bool

func (p Config) SetHandlerFunc(hdl srvtps.FuncHandler) {
	for i, c := range p {
		c.RegisterHandlerFunc(hdl)
		p[i] = c
	}
}

func (p Config) SetDefaultTLS(f libtls.FctTLSDefault) {
	for i, c := range p {
		c.SetDefaultTLS(f)
		p[i] = c
	}
}

func (p Config) SetContext(f libctx.FuncContext) {
	for i, c := range p {
		c.SetContext(f)
		p[i] = c
	}
}

func (p Config) Pool(ctx libctx.FuncContext, hdl srvtps.FuncHandler, defLog liblog.FuncLog) (Pool, liberr.Error) {
	var (
		r = New(ctx, hdl)
		e = ErrorPoolAdd.Error(nil)
	)

	p.Walk(func(cfg libhtp.Config) bool {
		if err := r.StoreNew(cfg, defLog); err != nil {
			e.Add(err)
		}
		return true
	})

	if !e.HasParent() {
		e = nil
	}

	return r, e
}

func (p Config) Walk(fct FuncWalkConfig) {
	if fct == nil {
		return
	}

	for _, c := range p {
		if !fct(c) {
			return
		}
	}
}

func (p Config) Validate() error {
	var e = ErrorPoolValidate.Error(nil)

	p.Walk(func(cfg libhtp.Config) bool {
		var err error

		if err = cfg.Validate(); err != nil {
			e.Add(err)
		}

		return true
	})

	if !e.HasParent() {
		e = nil
	}

	return e
}
