/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

package ticker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	errpol "github.com/nabbar/golib/errors/pool"
	"github.com/nabbar/golib/runner"
)

const (
	// pollStopWait is the maximum time to wait for cleanup after stopping.
	// This duration uses exponential backoff to allow the ticker goroutine
	// to complete its cleanup operations.
	pollStopWait = 2 * time.Second

	// defaultDuration is the fallback duration used when an invalid duration
	// (less than 1 millisecond) is provided to New().
	defaultDuration = 30 * time.Second
)

// ErrInvalid indicates an operation was attempted on an invalid or nil ticker instance.
var ErrInvalid = errors.New("invalid instance")

// run is the internal implementation of the Ticker interface.
// It manages the lifecycle of a ticker-based periodic execution.
//
// Fields:
//   - m: Mutex protecting Start/Stop/Restart operations to prevent concurrent state changes
//   - e: Error pool for collecting errors from the ticker function
//   - t: Atomic storage for the start time (zero value indicates not running)
//   - n: Atomic storage for the cancellation function of the current context
//   - f: The user-provided function to execute on each tick
//   - d: The duration between ticks
type run struct {
	m sync.Mutex                       // mutex to start / stop
	e errpol.Pool                      // error collection pool
	t libatm.Value[time.Time]          // start time (zero = not running)
	n libatm.Value[context.CancelFunc] // context cancellation function

	f runner.FuncTicker // user function to execute
	d time.Duration     // tick interval
}

// Uptime returns the duration since the ticker was started.
// Returns 0 if the ticker is not currently running.
//
// This method is safe for concurrent use and uses atomic operations
// to read the start time without locks.
func (o *run) Uptime() time.Duration {
	if i := o.t.Load(); i.IsZero() {
		return 0
	} else {
		return time.Since(i)
	}
}

// IsRunning returns true if the ticker is currently running.
// It checks if the uptime is greater than zero.
//
// This method is safe for concurrent use.
func (o *run) IsRunning() bool {
	return o.Uptime() > 0
}

// Restart stops the ticker if running and immediately starts it again.
// This is equivalent to calling Stop() followed by Start(), but atomic.
//
// Parameters:
//   - ctx: The context for the new ticker instance. Must not be nil.
//
// Returns an error if:
//   - ctx is nil
//   - the stop or start operations fail
//
// The method is protected by a mutex to ensure atomicity.
// Any panic during restart is recovered and logged.
func (o *run) Restart(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/ticker/restart", r)
		}
	}()

	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	o.m.Lock()
	defer o.m.Unlock()

	// Cancel any existing context
	o.cancel()

	// Stop if running
	if e := o.deMuxStop(ctx); e != nil {
		return e
	}

	// Start fresh instance
	return o.deMuxStart(ctx)
}

// Stop stops the ticker if it is currently running.
// This method is idempotent - calling Stop() on an already stopped ticker is safe.
//
// Parameters:
//   - ctx: Context for the stop operation (not used for timeout, but required
//     by the Runner interface)
//
// Returns an error if:
//   - ctx is nil
//   - the stop operation fails
//
// The method waits for the ticker goroutine to complete cleanup using
// exponential backoff polling, up to pollStopWait duration.
// Any panic during stop is recovered and logged.
func (o *run) Stop(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/ticker/stop", r)
		}
	}()

	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	o.m.Lock()
	defer o.m.Unlock()

	// Already stopped - this is not an error
	if !o.IsRunning() {
		return nil
	}

	return o.deMuxStop(ctx)
}

// Start starts the ticker to begin executing the function at regular intervals.
// If the ticker is already running, it will be stopped and restarted.
//
// Parameters:
//   - ctx: The context that controls the ticker's lifetime. When this context
//     is cancelled, the ticker will stop automatically. Must not be nil.
//
// Returns an error if:
//   - ctx is nil
//   - the start operation fails
//
// The ticker function will be called repeatedly with the interval specified in New().
// Errors from the function are collected and can be retrieved via ErrorsLast() or ErrorsList().
// Any panic during start is recovered and logged.
//
// Thread-safety: This method is protected by a mutex and safe for concurrent use.
func (o *run) Start(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/ticker/start", r)
		}
	}()

	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	o.m.Lock()
	defer o.m.Unlock()

	// Stop any existing instance before starting new one
	if o.IsRunning() {
		_ = o.deMuxStop(ctx)
	}

	return o.deMuxStart(ctx)
}

// ErrorsLast returns the most recent error collected from the ticker function.
// Returns nil if no errors have occurred or if all function calls returned nil.
//
// This method is safe for concurrent use.
func (o *run) ErrorsLast() error {
	return o.e.Last()
}

// ErrorsList returns all errors collected from the ticker function.
// The slice contains errors in the order they were collected.
// Returns an empty slice if no errors have occurred.
//
// Note: This returns a snapshot of errors at the time of the call.
// The ticker function may add more errors after this call returns.
//
// This method is safe for concurrent use.
func (o *run) ErrorsList() []error {
	return o.e.Slice()
}

// deMuxStop performs the internal stop operation without holding the mutex.
// It cancels the ticker's context and waits for the goroutine to complete cleanup.
//
// The method uses exponential backoff polling to wait for the ticker goroutine
// to finish, starting with 1ms and doubling up to 10ms intervals.
// It gives up after pollStopWait (2 seconds) to prevent indefinite blocking.
//
// This method must be called while holding o.m lock.
func (o *run) deMuxStop(ctx context.Context) error {
	// Cancel the context to signal the ticker goroutine to stop
	o.cancel()

	// Wait for the ticker goroutine to complete its cleanup
	// by checking if the uptime has been cleared (polling with exponential backoff)
	waitTime := time.Millisecond
	totalWait := time.Duration(0)

	for o.IsRunning() && totalWait < pollStopWait {
		time.Sleep(waitTime)
		totalWait += waitTime
		// Exponential backoff with max of 10ms per sleep
		if waitTime < 10*time.Millisecond {
			waitTime *= 2
		}
	}

	return nil
}

// deMuxStart performs the internal start operation without holding the mutex.
// It clears previous errors, starts a new goroutine for the ticker, and waits
// for the goroutine to initialize.
//
// The goroutine:
//   - Creates a time.Ticker with the configured duration
//   - Records the start time atomically
//   - Loops until context is cancelled, executing the user function on each tick
//   - Cleans up resources (stops ticker, clears start time) on exit
//   - Recovers from any panics to prevent process crashes
//
// This method must be called while holding o.m lock.
func (o *run) deMuxStart(ctx context.Context) error {
	// Clear previous errors before starting fresh
	o.e.Clear()

	// Launch ticker goroutine with a new cancellable context
	go func(x context.Context) {
		var tck = time.NewTicker(o.d)

		defer func() {
			// Recover from any panic to prevent process crash
			if r := recover(); r != nil {
				runner.RecoveryCaller("golib/server/ticker/deMuxStart", r)
			}
		}()

		defer func() {
			// Always clean up resources
			tck.Stop()
			o.cancel()
			// Clear start time to signal we're no longer running
			o.t.Store(time.Time{})
		}()

		// Record start time atomically
		o.t.Store(time.Now())

		// Main ticker loop
		for {
			select {
			case <-x.Done():
				// Context cancelled - exit gracefully
				return
			case <-tck.C:
				// Tick received - execute user function
				o.getFunction(x, tck)
			}
		}
	}(o.newCancel(ctx))

	// Wait for goroutine to initialize (start time to be set)
	// This ensures Start() returns only after the ticker is actually running
	for ctx.Err() == nil && !o.IsRunning() {
		time.Sleep(3 * time.Millisecond)
	}

	return nil
}

// getFunction executes the user-provided ticker function and collects any error.
// Panics in the user function are recovered to prevent the ticker from crashing.
//
// Parameters:
//   - ctx: The ticker's context (will be cancelled on Stop)
//   - tck: The underlying time.Ticker
func (o *run) getFunction(ctx context.Context, tck *time.Ticker) {
	defer func() {
		// Recover from any panic in the user function
		// This ensures one bad tick doesn't kill the entire ticker
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/ticker/getFunction", r)
		}
	}()

	// Execute the user function and collect any error
	// Nil errors are handled gracefully by the error pool
	o.e.Add(o.f(ctx, tck))
}

// cancel calls the stored cancel function to cancel the current running context.
// This is thread-safe and idempotent.
func (o *run) cancel() {
	if o == nil || o.n == nil {
		return
	} else if n := o.n.Load(); n != nil {
		n()
	}
}

// newCancel creates a new cancellable context and stores its cancel function.
// It also cancels any previously stored context to ensure clean state transitions.
// Returns the new context that will be passed to the start function.
func (o *run) newCancel(ctx context.Context) context.Context {
	if o == nil || o.n == nil {
		// Fallback for invalid state - return a context that expires immediately
		var n context.CancelFunc

		ctx, n = context.WithCancel(context.Background())
		n()

		return ctx
	}

	// Create a new cancellable context from the provided context
	x, n := context.WithCancel(ctx)

	// Store the new cancel function and retrieve the old one
	oldCancel := o.n.Swap(n)

	// Cancel the old context if it exists to ensure clean transition
	if oldCancel != nil {
		oldCancel()
	}

	return x
}
