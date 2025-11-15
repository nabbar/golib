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

import semtps "github.com/nabbar/golib/semaphore/types"

// NewWorker acquires a worker slot from the semaphore.
// Blocks until a slot is available or context is cancelled.
func (o *bar) NewWorker() error {
	return o.s.NewWorker()
}

// NewWorkerTry attempts to acquire a worker slot without blocking.
// Returns true if successful, false if no slots are available.
func (o *bar) NewWorkerTry() bool {
	return o.s.NewWorkerTry()
}

// DeferWorker releases a worker slot and increments the progress bar.
// Should be called with defer after NewWorker() succeeds.
func (o *bar) DeferWorker() {
	o.Inc(1)
	o.s.DeferWorker()
}

// DeferMain completes the progress bar and releases all semaphore resources.
// Should be called with defer in the main goroutine.
func (o *bar) DeferMain() {
	if o.isMPB() {
		o.Complete()
		if !o.Completed() {
			o.b.Abort(o.d)
		}
	}

	o.s.DeferMain()
}

// WaitAll blocks until all workers have completed.
func (o *bar) WaitAll() error {
	return o.s.WaitAll()
}

// Weighted returns the maximum number of simultaneous workers allowed.
func (o *bar) Weighted() int64 {
	return o.s.Weighted()
}

// New creates a new independent semaphore from this bar's semaphore.
func (o *bar) New() semtps.Sem {
	return o.s.New()
}
