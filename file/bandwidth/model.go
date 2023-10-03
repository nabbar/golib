/*
 * MIT License
 *
 * Copyright (c) 2023 Nicolas JUHEL
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

package bandwidth

import (
	"sync/atomic"
	"time"

	libfpg "github.com/nabbar/golib/file/progress"
	libsiz "github.com/nabbar/golib/size"
)

type bw struct {
	t *atomic.Value
	l libsiz.Size
}

func (o *bw) RegisterIncrement(fpg libfpg.Progress, fi libfpg.FctIncrement) {
	fpg.RegisterFctIncrement(func(size int64) {
		o.Increment(size)
		if fi != nil {
			fi(size)
		}
	})
}

func (o *bw) RegisterReset(fpg libfpg.Progress, fr libfpg.FctReset) {
	fpg.RegisterFctReset(func(size, current int64) {
		o.Reset(size, current)
		if fr != nil {
			fr(size, current)
		}
	})
}

func (o *bw) Increment(size int64) {
	if o == nil {
		return
	}

	var (
		i any
		t time.Time
		k bool
	)

	i = o.t.Load()
	if i == nil {
		t = time.Time{}
	} else if t, k = i.(time.Time); !k {
		t = time.Time{}
	}

	if !t.IsZero() && o.l > 0 {
		ts := time.Since(t)
		rt := float64(size) / ts.Seconds()
		if lm := o.l.Float64(); rt > lm {
			wt := time.Duration((rt / lm) * float64(time.Second))
			if wt.Seconds() > float64(time.Second) {
				time.Sleep(time.Second)
			} else {
				time.Sleep(wt)
			}
		}
	}

	o.t.Store(time.Now())
}

func (o *bw) Reset(size, current int64) {
	o.t.Store(time.Time{})
}
