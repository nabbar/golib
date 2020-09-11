/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package progress

import (
	"time"

	"github.com/nabbar/golib/errors"
	"github.com/nabbar/golib/semaphore"
	"github.com/vbauerster/mpb/v5"
)

type bar struct {
	i time.Time
	s semaphore.Sem
	t int64
	b *mpb.Bar
	u bool
	w bool
}

type Bar interface {
	Current() int64
	Completed() bool
	Reset(total, current int64)
	ResetDefined(current int64)
	Done()

	Increment(n int)
	Increment64(n int64)

	NewWorker() errors.Error
	NewWorkerTry() bool
	DeferWorker()
	DeferMain(dropBar bool)

	WaitAll() errors.Error

	GetBarMPB() *mpb.Bar
}

func newBar(b *mpb.Bar, s semaphore.Sem, total int64, isModeUnic bool) Bar {
	return &bar{
		u: total > 0,
		t: total,
		b: b,
		s: s,
		w: isModeUnic,
	}
}

func (b bar) GetBarMPB() *mpb.Bar {
	return b.b
}

func (b bar) Current() int64 {
	return b.b.Current()
}

func (b bar) Completed() bool {
	return b.b.Completed()
}

func (b *bar) Increment(n int) {
	if n > 0 {
		b.b.IncrBy(n)

		if b.i != semaphore.EmptyTime() {
			b.b.DecoratorEwmaUpdate(time.Since(b.i))
		}
	}

	b.i = time.Now()
}

func (b *bar) Increment64(n int64) {
	if n > 0 {
		b.b.IncrInt64(n)

		if b.i != semaphore.EmptyTime() {
			b.b.DecoratorEwmaUpdate(time.Since(b.i))
		}
	}

	b.i = time.Now()
}

func (b *bar) ResetDefined(current int64) {
	if current >= b.t {
		b.b.SetTotal(b.t, true)
		b.b.SetRefill(b.t)
	} else {
		b.b.SetTotal(b.t, false)
		b.b.SetRefill(current)
	}
}

func (b *bar) Reset(total, current int64) {
	b.u = total > 0
	b.t = total
	b.ResetDefined(current)
}

func (b *bar) Done() {
	b.b.SetRefill(b.t)
	b.b.SetTotal(b.t, true)
}

func (b *bar) NewWorker() errors.Error {
	if !b.u {
		b.t++
		b.b.SetTotal(b.t, false)
	}

	if !b.w {
		return b.s.NewWorker()
	}

	return nil
}

func (b *bar) NewWorkerTry() bool {

	if !b.w {
		return b.s.NewWorkerTry()
	}

	return false
}

func (b *bar) DeferWorker() {
	b.Increment(1)
	b.s.DeferWorker()
}

func (b *bar) DeferMain(dropBar bool) {
	b.b.Abort(dropBar)
	if !b.w {
		b.s.DeferMain()
	}
}

func (b *bar) WaitAll() errors.Error {
	if !b.w {
		return b.s.WaitAll()
	}

	return nil
}
