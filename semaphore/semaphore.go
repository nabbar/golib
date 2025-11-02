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

package semaphore

import semtps "github.com/nabbar/golib/semaphore/types"

// NewWorker acquires a worker slot, blocking until one is available.
// Returns an error if the context is cancelled before acquisition.
func (o *sem) NewWorker() error {
	return o.s.NewWorker()
}

// NewWorkerTry attempts to acquire a worker slot without blocking.
// Returns true if successful, false if no slots are available.
func (o *sem) NewWorkerTry() bool {
	return o.s.NewWorkerTry()
}

// DeferWorker releases a worker slot.
// Should be called with defer after NewWorker() or NewWorkerTry() succeeds.
func (o *sem) DeferWorker() {
	o.s.DeferWorker()
}

// DeferMain shuts down the MPB progress container (if enabled) and releases all resources.
// Should be called with defer in the main goroutine.
func (o *sem) DeferMain() {
	if o.isMbp() {
		o.m.Shutdown()
	}

	o.s.DeferMain()
}

// WaitAll blocks until all worker slots are available (all workers completed).
func (o *sem) WaitAll() error {
	return o.s.WaitAll()
}

// Weighted returns the maximum number of concurrent workers allowed.
// Returns -1 for unlimited concurrency.
func (o *sem) Weighted() int64 {
	return o.s.Weighted()
}

// Clone creates a new semaphore with independent worker management
// but shared MPB progress container (if progress was enabled).
func (o *sem) Clone() Semaphore {
	return &sem{
		s: o.s.New(),
		m: o.m,
	}
}

// New creates a new independent base semaphore with the same concurrency limit.
func (o *sem) New() semtps.Sem {
	return o.s.New()
}
