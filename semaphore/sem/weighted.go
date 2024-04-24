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

package sem

import (
	"context"
	"time"

	semtps "github.com/nabbar/golib/semaphore/types"

	goxsem "golang.org/x/sync/semaphore"
)

type sem struct {
	c context.CancelFunc
	x context.Context

	s *goxsem.Weighted
	n int64 // max concurrent process
}

func (o *sem) Deadline() (deadline time.Time, ok bool) {
	return o.x.Deadline()
}

func (o *sem) Done() <-chan struct{} {
	return o.x.Done()
}

func (o *sem) Err() error {
	return o.x.Err()
}

func (o *sem) Value(key any) any {
	return o.x.Value(key)
}

func (o *sem) NewWorker() error {
	return o.s.Acquire(o.x, 1)
}

func (o *sem) NewWorkerTry() bool {
	return o.s.TryAcquire(1)
}

func (o *sem) DeferWorker() {
	o.s.Release(1)
}

func (o *sem) DeferMain() {
	if o.c != nil {
		o.c()
	}
}

func (o *sem) WaitAll() error {
	return o.s.Acquire(o.x, o.n)
}

func (o *sem) Weighted() int64 {
	return o.n
}

func (o *sem) New() semtps.Sem {
	var (
		x context.Context
		n context.CancelFunc
	)

	x, n = context.WithCancel(o)

	return &sem{
		c: n,
		x: x,
		s: goxsem.NewWeighted(o.n),
		n: o.n,
	}
}
