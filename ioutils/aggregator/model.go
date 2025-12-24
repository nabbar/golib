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
	r libatm.Value[librun.StartStop]   // runner instance

	le libatm.Value[func(msg string, err ...error)] // logger instance
	li libatm.Value[func(msg string, arg ...any)]   // logger instance

	at time.Duration             // ticker duration of asynchronous function
	am int                       // maximum asynchronous call in same time
	af func(ctx context.Context) // asynchronous function

	st time.Duration             // ticker duration of synchronous function
	sf func(ctx context.Context) // synchronous function

	lc sync.Mutex                        // mutex lifecycle operations
	fw func(p []byte) (n int, err error) // main function write
	sh int                               // size of buffered channel data
	ch libatm.Value[chan []byte]         // channel data
	op *atomic.Bool                      // channel is closing

	cd *atomic.Int64 // counter of message in buffered channel
	cw *atomic.Int64 // counter of waiting write to buffered channel

	sd *atomic.Int64 // size of message in buffered channel
	sw *atomic.Int64 // size of waiting write to buffered channel
}

type ctxKey string

const ckStartSignal ctxKey = "startSignal"

// SetLoggerError sets a custom error logging function for the aggregator.
//
// The provided function will be called whenever an error occurs during internal
// operations, such as write failures or context cancellation.
//
// Parameters:
//   - f: The error logging function. If nil, a no-op function is used.
//
// This method is safe for concurrent use.
//
// Example:
//
//	agg.SetLoggerError(func(msg string, err ...error) {
//	    log.Printf("ERROR: %s: %v", msg, err)
//	})
func (a *agg) SetLoggerError(f func(msg string, err ...error)) {
	if f == nil {
		a.le.Store(func(msg string, err ...error) {})
		return
	}

	a.le.Store(f)
}

// SetLoggerInfo sets a custom info logging function for the aggregator.
//
// The provided function will be called for informational messages during
// normal operations, such as start/stop events.
//
// Parameters:
//   - f: The info logging function. If nil, a no-op function is used.
//
// This method is safe for concurrent use.
//
// Example:
//
//	agg.SetLoggerInfo(func(msg string, arg ...any) {
//	    log.Printf("INFO: "+msg, arg...)
//	})
func (a *agg) SetLoggerInfo(f func(msg string, arg ...any)) {
	if f == nil {
		a.li.Store(func(msg string, arg ...any) {})
		return
	}

	a.li.Store(f)
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
//  3. Resets all counters to ensure clean state on start
//  4. Creates a semaphore for limiting concurrent async function calls
//  5. Enters the main select loop to process:
//     - Context cancellation (via Done channel)
//     - Async callback timer ticks
//     - Sync callback timer ticks
//     - Write data from the channel (decrements counters immediately on receive)
//
// The function runs until the context is cancelled or an error occurs.
// All cleanup is handled in the deferred function.
//
// Parameters:
//   - ctx: Parent context for the processing loop
//
// Returns:
//   - error: ErrStillRunning if already running, ErrInvalidInstance if writer
//     function is nil, or the context error on cancellation
func (o *agg) run(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/run", r)
		}
	}()

	var (
		sem libsem.Semaphore

		tckAsc = time.NewTicker(o.at)
		tckSnc = time.NewTicker(o.st)

		fctWrt  func(p []byte) error
		fctSyn  func()
		fctAsyn func(sem libsem.Semaphore)

		sig chan error
	)

	if v := ctx.Value(ckStartSignal); v != nil {
		if s, ok := v.(chan error); ok {
			sig = s
		}
	}

	// signal is a helper to notify Start of success (nil) or failure
	// It ensures the channel is sent to and closed exactly once.
	signal := func(err error) {
		if sig != nil {
			select {
			case sig <- err:
			default:
			}
			close(sig)
			sig = nil
		}
	}

	defer func() {
		// Cleanup: release semaphore, close aggregator, stop timers
		if sem != nil {
			sem.DeferMain()
		}
		o.cleanup()

		o.logInfo("stopping aggregator")
		tckSnc.Stop()
		tckAsc.Stop()
	}()

	// Check if function write is set
	if o.fw == nil {
		signal(ErrInvalidInstance)
		return ErrInvalidInstance
	}

	// Check if already running - prevent multi-start
	if o.op.Load() {
		signal(ErrStillRunning)
		return ErrStillRunning
	}

	// Initialize context and open channel (which sets op to true)
	o.ctxNew(ctx)
	o.chanOpen()
	o.cntReset() // Reset counters on start to ensure clean state

	signal(nil)

	fctWrt = func(p []byte) error {
		if len(p) < 1 {
			return nil
		} else {
			_, e := o.fw(p)
			return e
		}
	}

	fctAsyn = o.callASyn()
	fctSyn = o.callSyn()

	sem = libsem.New(context.Background(), o.am, false)
	o.logInfo("starting aggregator")

	for o.Err() == nil {
		select {
		case <-o.Done():
			return o.Err()

		case <-tckAsc.C:
			fctAsyn(sem)

		case <-tckSnc.C:
			fctSyn()

		case p, ok := <-o.chanData():
			// Decrement counter immediately when data is received from channel
			o.cntDataDec(len(p))
			if !ok {
				// Channel closed, skip this iteration
				continue
			} else {
				// Log write errors but continue processing
				o.logError("error writing data", fctWrt(p))
			}
		}
	}

	return o.Err()
}

// callASyn returns a function that invokes the async callback if configured.
//
// The returned function is called periodically by the async timer in run().
// When invoked, it attempts to acquire a semaphore slot, and if successful,
// launches the async callback in a new goroutine. If the semaphore is full
// (max workers reached), the call is skipped to avoid blocking the main loop.
//
// The function respects Config.AsyncMax to limit concurrent async executions.
// Panics in the async callback are recovered and logged via RecoveryCaller.
//
// Returns a no-op function if:
//   - The aggregator is not running (op.Load() == false)
//   - No async function is configured (af == nil)
//   - The context is not initialized (x.Load() == nil)
func (o *agg) callASyn() func(sem libsem.Semaphore) {

	if !o.op.Load() {
		return func(sem libsem.Semaphore) {}
	} else if o.af == nil {
		return func(sem libsem.Semaphore) {}
	} else if o.x.Load() == nil {
		return func(sem libsem.Semaphore) {}
	}

	return func(sem libsem.Semaphore) {
		if !sem.NewWorkerTry() {
			// Semaphore full, skip this async call to avoid blocking
			return
		}

		// Launch async function in new goroutine
		go func() {
			defer func() {
				if r := recover(); r != nil {
					runner.RecoveryCaller("golib/ioutils/aggregator/callasyn", r)
				}
			}()

			defer sem.DeferWorker()

			o.af(o.x.Load())
		}()
	}
}

// callSyn returns a function that invokes the sync callback if configured.
//
// The returned function is called periodically by the sync timer in run().
// Unlike async callbacks, this function is called synchronously (blocking)
// on each timer tick, so it should complete quickly to avoid delaying the
// main processing loop and subsequent write operations.
//
// Panics in the sync callback are recovered and logged via RecoveryCaller.
//
// Returns a no-op function if:
//   - The aggregator is not running (op.Load() == false)
//   - No sync function is configured (sf == nil)
//   - The context is not initialized (x.Load() == nil)
func (o *agg) callSyn() func() {
	if !o.op.Load() {
		return func() {}
	} else if o.sf == nil {
		return func() {}
	} else if o.x.Load() == nil {
		return func() {}
	}

	return func() {
		defer func() {
			if r := recover(); r != nil {
				runner.RecoveryCaller("golib/ioutils/aggregator/callsyn", r)
			}
		}()

		o.sf(o.x.Load())
	}
}

// cntDataInc increments the processing counters when data enters the buffer.
// It tracks both the number of items and total bytes in the channel.
// This is called in Write() after the write is accepted but before sending to the channel.
func (o *agg) cntDataInc(i int) {
	o.cd.Add(1)
	o.sd.Add(int64(i))
}

// cntDataDec decrements the processing counters when data is consumed from the buffer.
// It ensures counters never go negative by resetting to 0 if needed.
// This is called immediately when data is received from the channel in the run loop,
// before the actual write operation, to maintain accurate real-time metrics.
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

// cntWaitInc increments the waiting counters when a Write() call blocks.
// It tracks both the number of blocked writes and total bytes waiting.
func (o *agg) cntWaitInc(i int) {
	o.cw.Add(1)
	o.sw.Add(int64(i))
}

// cntWaitDec decrements the waiting counters when a blocked Write() proceeds.
// It ensures counters never go negative by resetting to 0 if needed.
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

// cntReset resets all counters to zero.
// Called when the aggregator starts to ensure clean state.
func (o *agg) cntReset() {
	o.cd.Store(0)
	o.sd.Store(0)
	o.cw.Store(0)
	o.sw.Store(0)
}

// cleanup releases internal resources (context and channel).
// It is used by run() and Close() to ensure resources are freed without deadlock.
func (o *agg) cleanup() {
	o.ctxClose()
	o.chanClose()
}
