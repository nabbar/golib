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

package njs_progress

import (
	"context"

	njs_semaphore "github.com/nabbar/golib/njs-semaphore"
	"github.com/vbauerster/mpb/v5"
)

type bar struct {
	u bool
	t int64
	b *mpb.Bar
	s njs_semaphore.Sem
}

type Bar interface {
	Current() int64
	Completed() bool
	Increment(n int)
	Refill(amount int64)

	NewWorker() error
	NewWorkerTry() bool
	DeferWorker()
	DeferMain(dropBar bool)

	WaitAll() error
	Context() context.Context
	Cancel()

	GetBarMPB() *mpb.Bar
}

func newBar(b *mpb.Bar, s njs_semaphore.Sem, total int64) Bar {
	return &bar{
		u: total > 0,
		t: total,
		b: b,
		s: s,
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
	if n == 0 {
		n = 1
	}
	b.b.IncrBy(n)
}

func (b *bar) Refill(amount int64) {
	b.b.SetRefill(amount)
}

func (b *bar) NewWorker() error {
	if b.c == 0 {
		b.t++
		b.b.SetTotal(b.t, false)
	}
	return b.s.NewWorker()
}

func (b *bar) NewWorkerTry() bool {
	return b.s.NewWorkerTry()
}

func (b *bar) DeferWorker() {
	b.b.Increment()
	b.s.DeferWorker()
}

func (b *bar) DeferMain(dropBar bool) {
	b.b.Abort(dropBar)
	b.s.DeferMain()
}

func (b *bar) WaitAll() error {
	return b.s.WaitAll()
}

func (b bar) Context() context.Context {
	return b.s.Context()
}

func (b bar) Cancel() {
	b.s.Cancel()
}
