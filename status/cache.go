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

package status

import (
	"sync/atomic"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
)

const timeCache = 3 * time.Second

type ch struct {
	m *atomic.Int32
	t *atomic.Value
	c *atomic.Int64
	f func() monsts.Status
}

func (o *ch) Max() time.Duration {
	if m := o.m.Load(); m > 0 {
		return time.Duration(m) * time.Second
	} else {
		return timeCache
	}
}

func (o *ch) Time() time.Time {
	if t := o.t.Load(); t != nil {
		return t.(time.Time)
	} else {
		return time.Time{}
	}
}

func (o *ch) IsCache() monsts.Status {
	if t := o.Time(); !t.IsZero() && time.Since(t) < o.Max() {
		r := o.c.Load()
		return monsts.NewFromInt(r)
	}

	if o.f != nil {
		c := o.f()
		o.c.Store(c.Int())

		t := time.Now()
		o.t.Store(t)

		return c
	}

	return monsts.KO
}
