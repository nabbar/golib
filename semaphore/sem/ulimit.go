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

// wg is an unlimited concurrency implementation using sync.WaitGroup.
// Used when nbrSimultaneous < 0.
type wg struct {
	c context.CancelFunc // Context cancellation function
	x context.Context    // Context for goroutine lifecycle

	w sync.WaitGroup // WaitGroup for tracking goroutines
}

// Deadline implements context.Context interface.
func (o *wg) Deadline() (deadline time.Time, ok bool) {
	return o.x.Deadline()
}

// Done implements context.Context interface.
func (o *wg) Done() <-chan struct{} {
	return o.x.Done()
}

// Err implements context.Context interface.
func (o *wg) Err() error {
	return o.x.Err()
}

// Value implements context.Context interface.
func (o *wg) Value(key any) any {
	return o.x.Value(key)
}

// NewWorker adds a worker to the WaitGroup.
// Always succeeds (no limit in WaitGroup mode).
func (o *wg) NewWorker() error {
	o.w.Add(1)
	return nil
}

// NewWorkerTry always returns true in WaitGroup mode (no limit).
// Adds a worker to the WaitGroup for consistency with weighted semaphore behavior.
func (o *wg) NewWorkerTry() bool {
	o.w.Add(1)
	return true
}

// DeferWorker marks a worker as done in the WaitGroup.
// Should be called with defer after NewWorker().
func (o *wg) DeferWorker() {
	o.w.Done()
}

// DeferMain cancels the context.
// Should be called with defer in the main goroutine.
func (o *wg) DeferMain() {
	if o.c != nil {
		o.c()
	}
}

// WaitAll blocks until all workers have called DeferWorker().
func (o *wg) WaitAll() error {
	o.w.Wait()
	return nil
}

// Weighted returns -1 to indicate unlimited concurrency.
func (o *wg) Weighted() int64 {
	return -1
}

// New creates a new independent WaitGroup-based semaphore.
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
