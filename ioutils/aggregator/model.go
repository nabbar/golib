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
 */

package aggregator

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	liblog "github.com/nabbar/golib/logger"
	"github.com/nabbar/golib/runner"
	librun "github.com/nabbar/golib/runner/startStop"
	libsem "github.com/nabbar/golib/semaphore"
)

// agg is the internal implementation of the Aggregator interface.
//
// It uses atomic values for thread-safe access to shared state and a mutex
// to protect the writer function from concurrent calls.
//
// Fields:
//   - x: Internal context for cancellation propagation
//   - n: Context cancel function
//   - l: Logger instance
//   - r: Runner for lifecycle management
//   - at: Async function timer interval
//   - am: Max concurrent async functions (-1 = unlimited)
//   - af: Async callback function
//   - st: Sync function timer interval
//   - sf: Sync callback function
//   - mw: Mutex protecting fw from concurrent calls
//   - fw: Writer function (Config.FctWriter)
//   - sh: Channel buffer size
//   - ch: Buffered channel for write operations
//   - op: Atomic boolean indicating if channel is open
type agg struct {
	x libatm.Value[context.Context]    // context control
	n libatm.Value[context.CancelFunc] // running control

	l libatm.Value[liblog.Logger]    // logger instance
	r libatm.Value[librun.StartStop] // runner instance

	at time.Duration             // ticker duration of asynchronous function
	am int                       // maximum asynchronous call in same time
	af func(ctx context.Context) // asynchronous function

	st time.Duration             // ticker duration of synchronous function
	sf func(ctx context.Context) // synchronous function

	mw sync.Mutex                        // mutex single call of fw
	fw func(p []byte) (n int, err error) // main function write
	sh int                               // size of buffered channel data
	ch libatm.Value[chan []byte]         // channel data
	op *atomic.Bool                      // channel is closing

	cd *atomic.Int64 // counter of message in buffered channel
	cw *atomic.Int64 // counter of waiting write to buffered channel

	sd *atomic.Int64 // size of message in buffered channel
	sw *atomic.Int64 // size of waiting write to buffered channel
}

// NbWaiting returns the number of Write() calls currently waiting to send data to the channel.
// See Aggregator.NbWaiting() for details.
func (o *agg) NbWaiting() int64 {
	return o.cw.Load()
}

// SizeWaiting returns the total size in bytes of blocked Write() calls.
// See Aggregator.SizeWaiting() for details.
func (o *agg) SizeWaiting() int64 {
	return o.sw.Load()
}

// NbProcessing returns the number of items buffered in the channel waiting to be processed.
// See Aggregator.NbProcessing() for details.
func (o *agg) NbProcessing() int64 {
	return o.cd.Load()
}

// SizeProcessing returns the total size in bytes of buffered data items.
// See Aggregator.SizeProcessing() for details.
func (o *agg) SizeProcessing() int64 {
	return o.sd.Load()
}

// run is the main processing loop that handles write operations and periodic callbacks.
//
// This function:
//  1. Checks for multi-start condition and returns ErrStillRunning if already running
//  2. Initializes the internal context and opens the write channel
//  3. Creates a semaphore for limiting concurrent async function calls
//  4. Enters the main select loop to process:
//     - Context cancellation (via Done channel)
//     - Async callback timer ticks
//     - Sync callback timer ticks
//     - Write data from the channel
//
// The function runs until the context is cancelled or an error occurs.
// All cleanup is handled in the deferred function.
//
// Parameters:
//   - ctx: Parent context for the processing loop
//
// Returns:
//   - error: ErrStillRunning if already running, or the context error on cancellation
func (o *agg) run(ctx context.Context) error {
	defer runner.RecoveryCaller("golib/ioutils/aggregator/run", recover())

	var (
		sem libsem.Semaphore

		tckAsc = time.NewTicker(o.at)
		tckSnc = time.NewTicker(o.st)
	)

	defer func() {
		if sem != nil {
			sem.DeferMain()
		}
		_ = o.Close()

		o.logInfo("stopping aggregator")
		tckSnc.Stop()
		tckAsc.Stop()
	}()

	// Check if already running - prevent multi-start
	if o.op.Load() {
		return ErrStillRunning
	}

	// Initialize context and open channel (which sets op to true)
	o.ctxNew(ctx)
	o.chanOpen()
	o.cntReset() // Reset counters on start

	sem = libsem.New(context.Background(), o.am, false)
	o.logInfo("starting aggregator")
	for o.Err() == nil {
		select {
		case <-o.Done():
			return o.Err()

		case <-tckAsc.C:
			o.callASyn(sem)

		case <-tckSnc.C:
			o.callSyn()

		case p, ok := <-o.chanData():
			o.cntDataDec(len(p))
			if !ok {
				continue
			} else if e := o.fctWrite(p); e != nil {
				o.logError("error writing data", e)
			}
		}
	}

	return o.Err()
}

// fctWrite calls the configured writer function with mutex protection.
//
// This ensures that Config.FctWriter is never called concurrently, even though
// multiple goroutines may be calling Write() simultaneously.
//
// Parameters:
//   - p: Data to write
//
// Returns:
//   - error: nil on success, ErrInvalidInstance if no writer configured, or writer error
func (o *agg) fctWrite(p []byte) error {
	o.mw.Lock()
	defer o.mw.Unlock()

	if len(p) < 1 {
		return nil
	} else if o.fw == nil {
		return ErrInvalidInstance
	} else {
		_, e := o.fw(p)
		return e
	}
}

// callASyn invokes the async callback function if configured.
//
// The function is called in a new goroutine and is limited by the semaphore
// to prevent too many concurrent async calls (respecting Config.AsyncMax).
//
// If the semaphore is full (max workers reached), the call is skipped.
// This prevents blocking the main processing loop.
//
// Parameters:
//   - sem: Semaphore for limiting concurrent workers
func (o *agg) callASyn(sem libsem.Semaphore) {
	defer runner.RecoveryCaller("golib/ioutils/aggregator/callasyn", recover())

	if !o.op.Load() {
		return
	} else if o.af == nil {
		return
	} else if o.x.Load() == nil {
		return
	} else if !sem.NewWorkerTry() {
		return
	} else if e := sem.NewWorker(); e != nil {
		o.logError("aggregator failed to start new async worker", e)
		return
	} else {
		go func() {
			defer sem.DeferWorker()
			o.af(o.x.Load())
		}()
	}
}

// callSyn invokes the sync callback function if configured.
//
// This function is called synchronously (blocking) on the timer tick.
// It should complete quickly to avoid delaying write processing.
func (o *agg) callSyn() {
	defer runner.RecoveryCaller("golib/ioutils/aggregator/callsyn", recover())

	if !o.op.Load() {
		return
	} else if o.sf == nil {
		return
	} else if o.x.Load() == nil {
		return
	}
	o.sf(o.x.Load())
}

func (o *agg) cntDataInc(i int) {
	o.cd.Add(1)
	o.sd.Add(int64(i))
}

func (o *agg) cntDataDec(i int) {
	o.cd.Add(-1)
	if j := o.cd.Load(); j < 0 {
		o.cd.Store(0)
	}
	o.sd.Add(int64(-i))
	if j := o.sd.Load(); j < 0 {
		o.sd.Store(0)
	}
}

func (o *agg) cntWaitInc(i int) {
	o.cw.Add(1)
	o.sw.Add(int64(i))
}

func (o *agg) cntWaitDec(i int) {
	o.cw.Add(-1)
	if j := o.cw.Load(); j < 0 {
		o.cw.Store(0)
	}
	o.sw.Add(int64(-i))
	if j := o.sw.Load(); j < 0 {
		o.sw.Store(0)
	}
}

func (o *agg) cntReset() {
	o.cd.Store(0)
	o.sd.Store(0)
	o.cw.Store(0)
	o.sw.Store(0)
}
