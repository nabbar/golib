//go:build amd64 || arm64

/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package muxtun

import (
	"math"
	"time"
)

func (o *mdl) Counter() (uint64, uint64, time.Duration, time.Duration) {
	t := time.Now()
	r := o.cr.Swap(0)
	w := o.cw.Swap(0)
	n := o.cn.Swap(t.UnixNano())
	s := o.cs.Swap(t.Unix())
	d := o.cl.Load()

	return r, w, t.Sub(time.Unix(s, n)), time.Duration(d)
}

func (o *mdl) cntWrtAdd(i int) {
	if i <= 0 {
		return
	}

	if n := o.cw.Load(); n > (math.MaxUint64 - uint64(i)) {
		_, _, _, _ = o.Counter()
	}

	o.cw.Add(uint64(i))
}

func (o *mdl) cntRdrAdd(i int) {
	if i <= 0 {
		return
	}

	if n := o.cr.Load(); n > (math.MaxUint64 - uint64(i)) {
		_, _, _, _ = o.Counter()
	}

	o.cr.Add(uint64(i))
}

func (o *mdl) cntDelay(d time.Duration) {
	for o.Err() == nil {
		select {
		case <-o.Done():
			return
		case <-time.After(d):
			o.sendPING()
		}
	}
}
