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
	"sync"
	"time"

	semtps "github.com/nabbar/golib/semaphore/types"
)

type wg struct {
	c context.CancelFunc
	x context.Context

	w sync.WaitGroup
}

func (o *wg) Deadline() (deadline time.Time, ok bool) {
	return o.x.Deadline()
}

func (o *wg) Done() <-chan struct{} {
	return o.x.Done()
}

func (o *wg) Err() error {
	return o.x.Err()
}

func (o *wg) Value(key any) any {
	return o.x.Value(key)
}

func (o *wg) NewWorker() error {
	o.w.Add(1)
	return nil
}

func (o *wg) NewWorkerTry() bool {
	return true
}

func (o *wg) DeferWorker() {
	o.w.Done()
}

func (o *wg) DeferMain() {
	if o.c != nil {
		o.c()
	}
}

func (o *wg) WaitAll() error {
	o.w.Wait()
	return nil
}

func (o *wg) Weighted() int64 {
	return -1
}

func (o *wg) New() semtps.Sem {
	var (
		x context.Context
		n context.CancelFunc
	)

	x, n = context.WithCancel(o)

	return &wg{
		c: n,
		x: x,
		w: sync.WaitGroup{},
	}
}
