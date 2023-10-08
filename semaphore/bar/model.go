/*
 * MIT License
 *
 * Copyright (c) 2019 Nicolas JUHEL
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

package bar

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/vbauerster/mpb/v8"
	"golang.org/x/sync/semaphore"
)

type bar struct {
	c context.CancelFunc
	x context.Context

	s *semaphore.Weighted
	n int64
	d bool

	b *mpb.Bar
	m *atomic.Int64
	t *atomic.Value
}

func (o *bar) isMPB() bool {
	return o.b != nil
}

func (o *bar) GetMPB() *mpb.Bar {
	return o.b
}

func (o *bar) getDur() time.Duration {
	i := o.t.Load()
	o.t.Store(time.Now())

	if i == nil {
		return time.Millisecond
	} else if t, k := i.(time.Time); !k {
		return time.Millisecond
	} else if t.IsZero() {
		return time.Millisecond
	} else {
		return time.Since(t)
	}
}
