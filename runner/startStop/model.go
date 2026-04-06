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
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	errpol "github.com/nabbar/golib/errors/pool"
	"github.com/nabbar/golib/runner"
)

const (
	// pollStopTimeout defines the maximum duration allowed for the execution of
	// the user-provided stop function. If the stop function exceeds this timeout,
	// the Stop() method will proceed to its next phase (waiting for cleanup),
	// ensuring that a hanging stop function doesn't block the caller forever.
	// This serves as a safety guardrail for the Stop() operation.
	pollStopTimeout = 5 * time.Second

	// pollStopWait defines the safety timeout when waiting for the 'start' function
	// goroutine to completely terminate and perform its cleanup (defer blocks).
	// This is a secondary timeout used in Stop() after the stop function has been called.
	// It ensures that even if the start function doesn't exit immediately after its
	// context is canceled, the Stop() method will eventually return control.
	pollStopWait = 2 * time.Second
)

// ErrInvalid is returned when a operation is attempted on an uninitialized or
// nil runner instance. It ensures that the package fails gracefully when
// API contract is violated.
var ErrInvalid = errors.New("invalid instance")

// run is the concrete implementation of the StartStop interface.
// It manages the state transitions, synchronization, and error collection
// for a single service lifecycle.
//
// The struct uses a combination of:
//   - sync.Mutex for serialized state changes (Start/Stop).
//   - atomic variables for high-performance, non-blocking state queries (IsRunning/Uptime).
//   - atomic storage for dynamic components like context cancellation and wait channels.
type run struct {
	// m is a mutual exclusion lock used to serialize Start() and Stop() calls.
	// This prevents race conditions when multiple goroutines attempt to change
	// the service state simultaneously.
	m sync.Mutex

	// e is an error pool that captures and stores errors returned by the start
	// and stop functions. It allows for asynchronous error collection and
	// maintains an ordered history of errors.
	e errpol.Pool

	// t holds the timestamp when the service was started. It uses an atomic
	// wrapper for thread-safe access without explicit locking for the Uptime() method.
	t libatm.Value[time.Time]

	// n holds the current context.CancelFunc. When the service starts, a new
	// cancellable context is created, and its cancel function is stored here.
	// Calling this function signals the 'start' function to terminate gracefully.
	n libatm.Value[context.CancelFunc]

	// w holds a channel (chan struct{}) that is closed when the 'start' function's
	// goroutine has completely finished execution and cleaned up its resources.
	// This allows the Stop() method to wait efficiently for full termination.
	w libatm.Value[chan struct{}]

	// r is an atomic boolean indicating whether the service is currently running.
	// This allows for high-performance state checks (IsRunning) without acquiring the mutex.
	r *atomic.Bool

	// f is the user-provided 'start' function. This function is the core of the
	// service and should typically contain a blocking loop or a long-running task.
	f runner.FuncAction

	// s is the user-provided 'stop' function. This function is responsible for
	// triggering the shutdown of whatever the 'start' function is doing.
	s runner.FuncAction
}

// Uptime calculates the time elapsed since the service was started.
// If the service is not running or hasn't been started yet, it returns 0.
// This method is non-blocking and thread-safe, utilizing atomic loads
// for high performance.
func (o *run) Uptime() time.Duration {
	// Fast check: if the atomic running flag is false, uptime is zero.
	if !o.IsRunning() {
		return 0
	}

	// Load the start time from the atomic value.
	i := o.t.Load()
	if i.IsZero() {
		return 0
	}

	// Calculate duration since the recorded start time.
	return time.Since(i)
}

// IsRunning returns the current status of the service.
// It returns true if the service has been started and hasn't finished its execution.
// This method uses atomic loads and is safe for concurrent use from any number of goroutines.
func (o *run) IsRunning() bool {
	// Safety check for nil receiver or uninitialized atomic boolean.
	if o == nil || o.r == nil {
		return false
	}
	// Atomic load ensures visibility across threads without a mutex.
	return o.r.Load()
}

// Restart performs a sequential Stop() and Start() operation.
// It ensures that the current instance is fully stopped (including cleanup)
// before launching a new one. This is useful for reloading configurations
// or recovering from certain error states.
//
// Parameters:
//   - ctx: The context used for both stop and start operations. It handles
//     timeout and cancellation for the orchestration.
//
// Returns:
//   - error: The first error encountered during the stop or start process.
func (o *run) Restart(ctx context.Context) error {
	// Panic recovery to prevent a buggy user-function from crashing the whole process.
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/restart", r)
		}
	}()

	// Context validation.
	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	// 1. Signal cancellation to the running start function (early notification).
	o.cancel()

	// 2. Perform a graceful stop and wait for it to complete.
	// This will call the user-provided stop function.
	if e := o.Stop(ctx); e != nil {
		return e
	}

	// 3. Start the service again.
	// This will launch a new goroutine for the start function.
	return o.Start(ctx)
}

// Stop initiates a graceful shutdown of the service.
//
// Internal flow:
//  1. Check if the service is already stopped (early return).
//  2. Acquire the mutex to serialize stop operations and prevent concurrent Stop/Start races.
//  3. Signal the internal context cancellation (informs the start function to exit).
//  4. Execute the user-provided 'stop' function in a background goroutine with a timeout.
//  5. Release the mutex to allow status checks (IsRunning, Uptime) during the wait period.
//  6. Wait for the 'start' goroutine to signal its full termination via the 'w' channel,
//     subject to secondary timeouts for safety.
//
// This method returns nil if the service was already stopped or if it stops
// successfully within the allocated timeouts.
func (o *run) Stop(ctx context.Context) error {
	// Panic recovery for the internal Stop logic.
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/stop", r)
		}
	}()

	// Context validation.
	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	// Fast path check: avoid locking if the atomic flag says we're already stopped.
	if !o.IsRunning() {
		return nil
	}

	// Serialize access to state transition to prevent multiple concurrent stops.
	o.m.Lock()

	// Double check state after acquiring lock to handle potential race conditions.
	if !o.IsRunning() {
		o.m.Unlock()
		return nil
	}

	// Signal the 'start' function to begin its exit sequence immediately.
	o.cancel()

	// Retrieve the notification channel that will be closed when the start goroutine
	// has completely finished its execution and cleanup.
	wait := o.w.Load()

	// Prepare a context for the stop function with a safety timeout to prevent hanging.
	stCtx, stCnl := context.WithTimeout(ctx, pollStopTimeout)
	defer stCnl()

	// Execute the user-provided stop function asynchronously.
	// We do this in a goroutine so that we don't hold the mutex for the entire duration
	// of the stop call, which could be slow.
	go o.getFctStop(stCtx)

	// Release the lock so other status methods (IsRunning, Uptime) can be called
	// while we are waiting for termination.
	o.m.Unlock()

	// Synchronization point: wait for the service goroutine to actually exit.
	if wait != nil {
		select {
		case <-wait:
			// Success: The start function returned and finished all deferred cleanup.
		case <-ctx.Done():
			// Termination: The caller's context was cancelled or timed out.
		case <-time.After(pollStopWait):
			// Timeout: We reached the safety timeout waiting for the goroutine to exit.
		}
	}

	return nil
}

// Start launches the service.
// If the service is already running, Start() will first call Stop() to ensure
// a clean restart and reset of the internal state.
//
// Internal flow:
//  1. Stop any existing instance (if running).
//  2. Acquire mutex to serialize state transitions.
//  3. Reset the error history for the new execution cycle.
//  4. Create a new "done" channel for termination notification.
//  5. Create a new cancellable context for the service.
//  6. Launch the user-provided 'start' function in a new goroutine.
//
// Returns an error if the context is already cancelled or if a panic occurs.
func (o *run) Start(ctx context.Context) error {
	// Panic recovery for the internal Start logic.
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/start", r)
		}
	}()

	// Context validation.
	if ctx == nil {
		return fmt.Errorf("invalid nil context")
	}

	// Ensure we are in a clean, stopped state before starting a new instance.
	if o.IsRunning() {
		_ = o.Stop(ctx)
	}

	// Lock to ensure exclusive access during the initialization phase.
	o.m.Lock()
	defer o.m.Unlock()

	// Reset error history from previous runs to provide a clean slate for the new execution.
	o.e.Clear()

	// Create and store the channel used to signal when this specific run finishes.
	// This channel is closed by the getFctStart goroutine when it exits.
	done := make(chan struct{})
	o.w.Store(done)

	// Create a new context specific to this start/stop cycle.
	cx := o.newCancel()

	// Check if the provided context is already canceled before launching.
	if e := ctx.Err(); e != nil {
		close(done)
		return e
	}

	// Start the asynchronous execution in a dedicated goroutine.
	go o.getFctStart(cx, done)

	return nil
}

// ErrorsLast returns the most recent error encountered during service operation.
// This is thread-safe and can be called at any time, though it is most useful
// after a Stop() or when IsRunning() returns false.
func (o *run) ErrorsLast() error {
	// Non-blocking retrieval from the error pool.
	return o.e.Last()
}

// ErrorsList returns the full history of errors encountered during the current
// or most recent execution cycle. The list is automatically cleared on every
// call to Start().
func (o *run) ErrorsList() []error {
	// Non-blocking retrieval from the error pool.
	return o.e.Slice()
}

// getFctStart manages the execution of the user-provided 'start' function.
// It is intended to be run in its own goroutine and handles all runtime
// orchestration for the service's primary task.
//
// Responsibilities:
//   - Sets the 'running' state (atomic flag) and start timestamp.
//   - Executes the 'start' function.
//   - Captures any returned error and adds it to the error pool.
//   - Ensures cleanup via defers: resetting state, cancelling context.
//   - Signals termination by closing the 'done' channel last.
func (o *run) getFctStart(ctx context.Context, done chan struct{}) {
	// Recovery from panics within the user-provided start function.
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/fctStart", r)
		}
	}()

	// The 'done' channel MUST be closed last to ensure all state changes (defers) are visible.
	// This signals to Stop() that it's safe to proceed.
	defer func() {
		if done != nil {
			close(done)
		}
	}()

	// Cleanup defer block: ensures state is reset when the function exits (success or error).
	defer func() {
		// Reset start time to zero value.
		o.t.Store(time.Time{})
		// Update atomic flag to indicate the service is no longer running.
		o.r.Store(false)
		// Ensure context cancellation is triggered.
		o.cancel()
	}()

	// Initialize runtime state before calling the user function.
	o.t.Store(time.Now())
	o.r.Store(true)

	// Fallback for missing start function to avoid nil dereference.
	var fct = o.f
	if fct == nil {
		fct = func(ctx context.Context) error {
			return fmt.Errorf("invalid start function")
		}
	}

	// Blocking call to user function. The core execution happens here.
	// Any returned error is automatically stored in the pool.
	o.e.Add(fct(ctx))
}

// getFctStop manages the execution of the user-provided 'stop' function.
// It is intended to be run in its own goroutine (launched by Stop()).
//
// Responsibilities:
//   - Ensures context cancellation is triggered.
//   - Immediately updates the running state to false.
//   - Executes the user-provided 'stop' function.
//   - Captures any error from the stop function into the error pool.
func (o *run) getFctStop(ctx context.Context) {
	// Recovery from panics within the user-provided stop function.
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/server/startstop/fctStop", r)
		}
	}()

	// Ensure context is cancelled to speed up 'start' function exit if not already done.
	o.cancel()

	// Update state immediately to reflect that shutdown has been initiated.
	o.t.Store(time.Time{})
	o.r.Store(false)

	// Fallback for missing stop function to avoid nil dereference.
	var fct = o.s
	if fct == nil {
		fct = func(ctx context.Context) error {
			return fmt.Errorf("invalid stop function")
		}
	}

	// Execute user stop logic and record any errors.
	o.e.Add(fct(ctx))
}

// cancel is a helper to safely call the current context's cancel function.
// It retrieves the cancel function from atomic storage and invokes it.
// This method is idempotent and safe for concurrent calls.
func (o *run) cancel() {
	// Safety check for uninitialized instance or atomic storage.
	if o == nil || o.n == nil {
		return
	} else if n := o.n.Load(); n != nil {
		// Invoke the stored cancel function.
		n()
	}
}

// newCancel creates a new cancelable context, stores its cancel function
// in atomic storage, and ensures any previous context is cancelled to avoid leaks.
//
// Returns the newly created context.
func (o *run) newCancel() context.Context {
	var (
		x context.Context
		n context.CancelFunc
	)

	// Create a fresh background context.
	// Note: It's detached from the Start() caller's context to survive the Start call's scope.
	x, n = context.WithCancel(context.Background())

	// Safety check for uninitialized instance.
	if o == nil || o.n == nil {
		n() // Cleanup immediately if instance is invalid.
		return x
	}

	// Swap the new cancel function into the atomic storage.
	oldCancel := o.n.Swap(n)

	// If there was an old cancel function (from a previous run), invoke it to ensure cleanup.
	if oldCancel != nil {
		oldCancel()
	}

	return x
}
