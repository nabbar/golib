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

package tls

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
)

type mod struct {
	x libctx.Config[uint8]
	t libatm.Value[libtls.TLSConfig]
	c libatm.Value[func() *libtls.Config]
	f libtls.FctRootCACert
	r *atomic.Bool
}

func (o *mod) Config() *libtls.Config {
	if f := o.c.Load(); f == nil {
		return nil
	} else if c := f(); c == nil {
		return nil
	} else {
		v := *c
		return &v
	}
}

func (o *mod) GetTLS() libtls.TLSConfig {
	if t := o.t.Load(); t != nil {
		return t.Clone()
	}

	return nil
}

func (o *mod) SetTLS(t libtls.TLSConfig) {
	if t == nil {
		o.r.Store(false)
		t = libtls.New()
	} else {
		o.r.Store(true)
	}

	o.t.Store(t)
}
