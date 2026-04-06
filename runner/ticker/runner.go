/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

package ticker

import (
	"context"
	"time"
)

// Uptime provides the continuous duration for which the runner has been in the 'running' state.
//
// If the runner is not currently running (e.g., it is stopped, starting, or stopping),
// it returns 0. The uptime is reset to 0 every time the runner is stopped or restarted.
//
// Returns:
//   - time.Duration: The elapsed time since the last successful start.
func (o *run) Uptime() time.Duration {
	// Only return uptime if the runner's main loop is active.
	if !o.IsRunning() {
		return 0
	}

	// Calculate the difference between now and the recorded start time 't'.
	if i := o.t.Load(); !i.IsZero() {
		return time.Since(i)
	}

	return 0
}

// IsRunning provides an atomic check of whether the runner's main loop is active.
//
// It returns true only if the Finite State Machine (FSM) is in the 'running' state.
// This method is thread-safe and can be called from multiple goroutines simultaneously.
//
// Returns:
//   - bool: true if the runner is active and executing the ticker function at intervals.
func (o *run) IsRunning() bool {
	// State 'running' is set in deMuxStart after the background goroutine is ready.
	return o.getState() == running
}

// Restart performs a sequential stop and start of the runner.
//
// This is useful for reloading configurations or resetting the error history.
// If the runner is already running, it is first stopped gracefully.
//
// Parameters:
//   - ctx: A context that governs the maximum time allowed for both the stop and start operations.
//     If nil, a default timeout context (5 seconds) is created.
//
// Returns:
//   - error: Nil on success, or an error if the operation timed out or failed.
func (o *run) Restart(ctx context.Context) error {
	// Internally, Start(ctx) handles stopping an already running instance.
	return o.Start(ctx)
}

// Stop gracefully shuts down the ticker runner and its background goroutines.
//
// It acquires a mutex to ensure that no other Start or Stop operation is concurrently
// executing. It then signals the background loop to terminate and waits for it
// to reach the 'stopped' state.
//
// Parameters:
//   - ctx: A context to limit the time we wait for the runner to stop. If nil, a default
//     timeout context of 5 seconds is used. If the context expires before the runner
//     has fully stopped, an error is returned.
//
// Returns:
//   - error: Nil if the runner stopped successfully, or the context's error if it timed out.
func (o *run) Stop(ctx context.Context) error {
	// Provide a default 5-second timeout if the user did not provide a context.
	if ctx == nil {
		var cnl context.CancelFunc
		ctx, cnl = context.WithTimeout(context.Background(), 5*time.Second)
		defer cnl()
	}

	// Use a mutex to prevent race conditions during state transitions.
	o.m.Lock()
	defer o.m.Unlock()

	// Initiate the internal stop sequence (cancels context and sets FSM state).
	o.deMuxStop()

	// Poll the state until 'stopped' or context timeout.
	for {
		if o.getState() == stopped {
			// Successfully reached the stopped state.
			return nil
		}
		select {
		case <-ctx.Done():
			// Context timeout or cancellation reached.
			return ctx.Err()
		default:
			// Small sleep to avoid high CPU usage during the wait loop.
			time.Sleep(pollState)
		}
	}
}

// Start initiates the runner sequence, creating a background goroutine for the ticker.
//
// If the runner is already in a non-stopped state (e.g., 'running' or 'starting'),
// it will first attempt to stop it gracefully before starting it again.
// It acquires a mutex to ensure that only one Start or Stop operation is active at a time.
//
// Parameters:
//   - ctx: A context to limit the time we wait for the runner to reach the 'running' state.
//     If nil, a default timeout context of 5 seconds is used.
//
// Returns:
//   - error: Nil if the runner reached the 'running' state, or an error (e.g., context deadline exceeded).
func (o *run) Start(ctx context.Context) error {
	// Provide a default 5-second timeout if no context is given.
	if ctx == nil {
		var cnl context.CancelFunc
		ctx, cnl = context.WithTimeout(context.Background(), 5*time.Second)
		defer cnl()
	}

	// Protect the start sequence with a mutex to prevent multiple starts or mixed start/stops.
	o.m.Lock()
	defer o.m.Unlock()

	// If the runner is already active, stop it before starting a new instance.
	if o.getState() != stopped {
		o.deMuxStop()

		// Wait until the current instance has fully stopped.
		for o.getState() != stopped {
			select {
			case <-ctx.Done():
				// Return error if we can't stop within the given context.
				return ctx.Err()
			default:
				time.Sleep(pollState)
			}
		}
	}

	// Initiate the internal start sequence (resets timers and launches goroutine).
	o.deMuxStart()

	// Wait until the background goroutine transitions to the 'running' state.
	for {
		if o.getState() == running {
			// Successfully reached the running state.
			return nil
		}

		select {
		case <-ctx.Done():
			// Context timeout or cancellation reached during startup.
			return ctx.Err()
		default:
			// Small sleep to avoid high CPU usage during the wait loop.
			time.Sleep(pollState)
		}
	}
}

// ErrorsLast retrieves the most recent error returned by the ticker function.
//
// It returns nil if no error has occurred since the last start or restart.
// This is part of the liberr.Errors interface implementation.
//
// Returns:
//   - error: The last error recorded in the error pool.
func (o *run) ErrorsLast() error {
	return o.e.Last()
}

// ErrorsList retrieves all errors encountered by the ticker function during the current run.
//
// This is part of the liberr.Errors interface implementation. The list is cleared
// automatically whenever the ticker is started or restarted.
//
// Returns:
//   - []error: A slice containing all collected errors in chronological order.
func (o *run) ErrorsList() []error {
	return o.e.Slice()
}
