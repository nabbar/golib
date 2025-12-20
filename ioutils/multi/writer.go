/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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

package multi

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// writeWrapper wraps an io.Writer to add latency tracking for adaptive mode.
// It measures write duration and triggers mode switching callbacks when
// latency thresholds are crossed after sampling enough operations.
type writeWrapper struct {
	io.Writer
	cnt *atomic.Int64 // Sample counter for write operations
	lat *atomic.Int64 // Cumulative latency in nanoseconds
	smp int64         // Sample size before triggering callback
	fct func(int64)   // Callback function with mean latency
}

// Write delegates to the wrapped Writer and measures write latency.
// The measured latency is used for adaptive mode switching.
func (w *writeWrapper) Write(p []byte) (n int, err error) {
	start := time.Now()
	defer func() {
		w.recLat(time.Since(start).Nanoseconds())
	}()
	return w.Writer.Write(p)
}

// Lat returns the cumulative latency in nanoseconds.
func (w *writeWrapper) Lat() int64 {
	return w.lat.Load()
}

// Smp returns the current sample count.
func (w *writeWrapper) Smp() int64 {
	return w.cnt.Load()
}

// recLat records latency measurement and triggers callback when sample threshold is reached.
func (w *writeWrapper) recLat(lat int64) {
	if w.smp < 1 {
		return
	}

	w.cnt.Add(1)
	w.lat.Add(lat)

	// Reset statistics after sample window
	if c := w.cnt.Load(); c >= w.smp*2 {
		if w.fct != nil {
			w.fct(w.lat.Load() / c)
		}
		w.cnt.Store(0)
		w.lat.Store(0)
	}
}

// newWriteSeq creates a sequential write wrapper using io.MultiWriter.
// Returns a wrapper with io.Discard if no writers provided, direct writer if one,
// or io.MultiWriter for multiple writers.
func newWriteSeq(sample int64, fct func(int64), w ...io.Writer) *writeWrapper {
	if len(w) == 0 {
		return &writeWrapper{
			Writer: io.Discard,
			cnt:    new(atomic.Int64),
			lat:    new(atomic.Int64),
			smp:    0,
			fct:    nil,
		}
	} else if len(w) == 1 {
		return &writeWrapper{
			Writer: w[0],
			cnt:    new(atomic.Int64),
			lat:    new(atomic.Int64),
			smp:    sample,
			fct:    fct,
		}
	} else {
		return &writeWrapper{
			Writer: io.MultiWriter(w...),
			cnt:    new(atomic.Int64),
			lat:    new(atomic.Int64),
			smp:    sample,
			fct:    fct,
		}
	}
}

// newWritePar creates a parallel write wrapper using wrtParallel for concurrent writes.
// Uses parallel writing only if data size exceeds minSize threshold.
// Falls back to sequential for small writes to avoid goroutine overhead.
func newWritePar(sample int64, fct func(int64), minSize int, w ...io.Writer) *writeWrapper {
	if len(w) == 0 {
		return &writeWrapper{
			Writer: io.Discard,
			cnt:    new(atomic.Int64),
			lat:    new(atomic.Int64),
			smp:    0,
			fct:    nil,
		}
	} else if len(w) == 1 {
		return &writeWrapper{
			Writer: w[0],
			cnt:    new(atomic.Int64),
			lat:    new(atomic.Int64),
			smp:    sample,
			fct:    fct,
		}
	} else {
		return &writeWrapper{
			Writer: &wrtParallel{
				w: w,
				s: minSize,
			},
			cnt: new(atomic.Int64),
			lat: new(atomic.Int64),
			smp: sample,
			fct: fct,
		}
	}
}

// wrtParallel implements parallel writes to multiple io.Writer destinations.
// Small writes (below threshold) are done sequentially to avoid goroutine overhead.
// Large writes spawn goroutines for each writer and collect the first error if any.
type wrtParallel struct {
	w []io.Writer // Writers to broadcast to
	s int         // Minimum size threshold for parallel execution
}

// Write broadcasts data to all writers. Uses parallel goroutines for large writes,
// sequential writes for small data. Returns len(p) on success or first error encountered.
func (o *wrtParallel) Write(p []byte) (n int, err error) {
	if len(p) < o.s {
		for _, w := range o.w {
			if _, e := w.Write(p); e != nil {
				return len(p), e
			}
		}
		return len(p), nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(o.w))

	for _, w := range o.w {
		wg.Add(1)
		go func(w io.Writer) {
			defer wg.Done()
			if _, e := w.Write(p); e != nil {
				select {
				case errCh <- e:
				default:
				}
			}
		}(w)
	}

	wg.Wait()
	close(errCh)

	// Return first error if any
	if e := <-errCh; e != nil {
		return len(p), e
	}

	return len(p), nil
}
