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

package startStop

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
	// pollStopTimeout is the timeout for the stop function call itself.
	// This ensures that a misbehaving stop function doesn't block indefinitely.
	pollStopTimeout = 5 * time.Second

	// pollStopWait is the maximum time to wait for cleanup after stopping.
	// This uses an exponential backoff polling strategy to detect when the
	// start function has completed its cleanup (uptime reset to zero).
	pollStopWait = 2 * time.Second
)

// ErrInvalid is returned when the runner instance or its functions are nil/invalid.
var ErrInvalid = errors.New("invalid instance")

// run is the internal implementation of the StartStop interface.
// It provides thread-safe lifecycle management for services with start/stop operations.
type run struct {
	// m protects concurrent Start/Stop operations to ensure serialization
	m sync.Mutex

	// e tracks all errors that occur during start/stop operations
	e errpol.Pool

	// t stores the service start time; zero value means not running
	t libatm.Value[time.Time]

	// n stores the current context cancel function for graceful shutdowns
	n libatm.Value[context.CancelFunc]

	// f is the user-provided start function (should block until service stops)
	f runner.FuncAction

	// s is the user-provided stop function (should gracefully shutdown the service)
	s runner.FuncAction
}

// Uptime returns the duration since the service was started.
// Returns 0 if the service is not currently running.
// This method is thread-safe and can be called concurrently.
func (o *run) Uptime() time.Duration {
	if i := o.t.Load(); i.IsZero() {
		return 0
	} else {
		return time.Since(i)
	}
}

// IsRunning returns true if the service is currently running.
// This is determined by checking if the uptime is greater than zero.
// This method is thread-safe and can be called concurrently.
func (o *run) IsRunning() bool {
	return o.Uptime() > 0
}

// Restart stops the runner and starts it again with a fresh state.
// It ensures the previous instance is fully stopped before starting a new one.
// The context is used for both the stop and start operations.
// Returns an error if either the stop or start operation fails.
func (o *run) Restart(ctx context.Context) error {
	defer func() {
		// Recover from any panic in the restart function
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/restart", r)
		}
	}()

	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	// Cancel the running context to signal shutdown
	o.cancel()

	if e := o.Stop(ctx); e != nil {
		return e
	}

	return o.Start(ctx)
}

// Stop gracefully stops the runner by cancelling its context and calling the stop function.
// It waits for the start function to complete cleanup before returning using an exponential
// backoff polling strategy. This method is idempotent - calling Stop on an already stopped
// runner is safe and returns nil immediately.
// Returns an error if the stop function fails (though errors are also tracked internally).
func (o *run) Stop(ctx context.Context) error {
	defer func() {
		// Recover from any panic in the stop function
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/stop", r)
		}
	}()

	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	// Fast path: if already stopped, return immediately without locking
	if !o.IsRunning() {
		return nil
	}

	// Lock to prevent concurrent stop operations
	o.m.Lock()
	defer o.m.Unlock()

	// Cancel the context to signal the start function to stop
	o.cancel()

	// Create a timeout context for the stop function to prevent hanging
	stopCtx, cancel := context.WithTimeout(ctx, pollStopTimeout)
	defer cancel()

	// Call the stop function asynchronously (it will also cancel context and clear uptime)
	go o.getFctStop(stopCtx)

	// Wait for the start function to complete its cleanup by polling the uptime.
	// The start function's defer will set uptime to zero when it finishes.
	// We use exponential backoff to reduce CPU usage while waiting.
	waitTime := time.Millisecond
	totalWait := time.Duration(0)

	for o.IsRunning() && totalWait < pollStopWait {
		time.Sleep(waitTime)
		totalWait += waitTime
		if waitTime < 10*time.Millisecond {
			waitTime *= 2
		}
	}

	return nil
}

// Start launches the start function in a goroutine with proper context management.
// If the runner is already running, it stops it first to ensure clean state.
// Returns nil immediately after launching; errors from the start function are tracked
// internally and can be retrieved via ErrorsLast()/ErrorsList().
// Multiple concurrent Start() calls are serialized by a mutex - each call will stop
// any previous instance and start a fresh one. Previous errors are cleared on each start.
func (o *run) Start(ctx context.Context) error {
	defer func() {
		// Recover from any panic in the start function
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/start", r)
		}
	}()

	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	// Stop any existing instance to ensure clean state transition
	if o.IsRunning() {
		_ = o.Stop(ctx)
	}

	o.m.Lock()
	defer o.m.Unlock()

	// Clear previous errors before starting
	o.e.Clear()

	// Create a new cancellable context for this start instance
	cx := o.newCancel(ctx)

	// Launch the start function asynchronously
	// Errors will be captured and stored in the error pool
	go o.getFctStart(cx)

	return nil
}

// ErrorsLast returns the most recent error that occurred during start/stop operations.
// Returns nil if no errors have occurred or if errors were cleared on the last Start().
// This method is thread-safe.
func (o *run) ErrorsLast() error {
	return o.e.Last()
}

// ErrorsList returns all errors that have occurred since the last Start() call.
// Errors are cleared each time Start() is called.
// Returns an empty slice if no errors have occurred.
// This method is thread-safe.
func (o *run) ErrorsList() []error {
	return o.e.Slice()
}

// getFctStart executes the start function with proper error handling and cleanup.
// This function is always called in a goroutine. It sets the start time before execution
// and clears it after completion (via defer). Any panics are recovered and logged.
func (o *run) getFctStart(ctx context.Context) {
	defer func() {
		// Recover from any panic in the start function
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/fctStart", r)
		}
	}()
	defer func() {
		// Clear the start time to indicate we're no longer running
		o.t.Store(time.Time{})

		// Cancel the context in case it wasn't already cancelled
		o.cancel()
	}()

	// Get the start function; use a placeholder that returns an error if not set
	var fct = o.f
	if fct == nil {
		fct = func(ctx context.Context) error {
			return fmt.Errorf("invalid start function")
		}
	}

	// Mark the start time
	o.t.Store(time.Now())

	// Execute the actual start function
	o.e.Add(fct(ctx))
}

// getFctStop executes the stop function with proper error handling and cleanup.
// This function is always called in a goroutine. It cancels the context, clears the
// start time, and calls the stop function. Any panics are recovered and logged.
func (o *run) getFctStop(ctx context.Context) {
	defer func() {
		// Recover from any panic in the stop function
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/fctStop", r)
		}
	}()

	// Cancel the context to ensure the start function stops
	o.cancel()

	// Clear the start time to indicate we're no longer running
	// This is what Stop() polls for to detect completion
	o.t.Store(time.Time{})

	// Get the stop function; use a placeholder that returns an error if not set
	var fct = o.s
	if fct == nil {
		fct = func(ctx context.Context) error {
			return fmt.Errorf("invalid stop function")
		}
	}

	o.e.Add(fct(ctx))
}

// cancel calls the stored cancel function to signal the running service to stop.
// This is thread-safe and idempotent - calling cancel multiple times is safe.
// If no cancel function is stored (not running), this is a no-op.
func (o *run) cancel() {
	if o == nil || o.n == nil {
		return
	} else if n := o.n.Load(); n != nil {
		n()
	}
}

// newCancel creates a new cancellable context and stores its cancel function.
// It also cancels any previously stored context to ensure clean state transitions.
// This ensures that old instances are properly cleaned up before starting new ones.
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
