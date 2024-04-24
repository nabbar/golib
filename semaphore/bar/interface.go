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
	"sync/atomic"
	"time"

	semtps "github.com/nabbar/golib/semaphore/types"
	sdkmpb "github.com/vbauerster/mpb/v8"
)

func New(sem semtps.SemPgb, tot int64, drop bool, opts ...sdkmpb.BarOption) semtps.SemBar {
	if drop {
		opts = append(opts, sdkmpb.BarRemoveOnComplete())
	}

	var b *sdkmpb.Bar
	if m := sem.GetMPB(); m != nil {
		b = m.AddBar(tot, opts...)
	}

	ts := new(atomic.Value)
	ts.Store(time.Now())

	mx := new(atomic.Int64)
	mx.Store(tot)

	return &bar{
		s: sem.New(),
		d: drop,
		b: b,
		m: mx,
		t: ts,
	}
}
