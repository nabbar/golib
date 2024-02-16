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
	"sync"

	libtls "github.com/nabbar/golib/certificates"
	libctx "github.com/nabbar/golib/context"
)

type componentTls struct {
	m sync.RWMutex
	x libctx.Config[uint8]
	t libtls.TLSConfig
	c *libtls.Config
	f libtls.FctRootCA
}

func (o *componentTls) Config() *libtls.Config {
	o.m.Lock()
	defer o.m.Unlock()

	return o.c
}

func (o *componentTls) GetTLS() libtls.TLSConfig {
	o.m.Lock()
	defer o.m.Unlock()

	return o.t
}

func (o *componentTls) SetTLS(tls libtls.TLSConfig) {
	o.m.Lock()
	defer o.m.Unlock()

	o.t = tls
}
