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
	"runtime"
	"time"

	"github.com/nabbar/golib/runner"
	librun "github.com/nabbar/golib/runner/startStop"
)

// Start initiates the background processing goroutine and enables the data aggregation pipeline.
//
// Lifecycle Orchestration:
//  1. Mutex-Free Strategy: Unlike traditional implementations, this method relies on the
//     underlying 'librun.StartStop' component to manage thread-safety and concurrency during
//     the startup phase. This architectural choice prevents complex deadlocks that can occur
//     when multiple mutexes are held during lifecycle transitions.
//  2. Idempotent Initialization: The method first retrieves the current runner. If none exists,
//     a fresh instance is created and stored atomically. If the aggregator is already in
//     the 'Running' state, the runner internally handles the request as a no-op.
//  3. Background Execution: It invokes the runner's Start() method, which spawns the 'run'
//     logic in a dedicated, managed goroutine that handles I/O serialization.
//
// Operational Prerequisites:
//   - A valid writer function must have been provided during New() via Config.FctWriter.
//   - The aggregator must be successfully started before any Write() calls can be processed.
//
// Parameters:
//   - ctx: Parent context used to derive the operational context for the run loop.
//     If this context is cancelled, the aggregator will perform an orderly shutdown.
//
// Returns:
//   - error: Returns nil if the startup signal was successfully dispatched, or an error
//     from the runner if initialization failed.
//
// Example:
//
//	agg, _ := aggregator.New(ctx, cfg)
//	if err := agg.Start(ctx); err != nil {
//	    log.Fatalf("Orderly startup failed: %v", err)
//	}
func (o *agg) Start(ctx context.Context) error {
	defer func() {
		// Recovery from panics during startup.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/start", r)
		}
	}()

	// Atomic load/creation of the lifecycle runner.
	r := o.getRunner()
	if r == nil {
		r = o.newRunner()
		o.setRunner(r)
	}

	// Dispatch the start signal to the runner.
	return r.Start(ctx)
}

// Stop executes a graceful termination of the aggregator's background processing goroutine.
//
// Shutdown Sequence:
//  1. Signaling: Commands the underlying runner to stop. This action cancels the operational
//     context, signaling the 'run' loop to stop accepting new data and finish its current batch.
//  2. Resource Decommissioning: If no runner is currently assigned but internal flags indicate
//     that resources (like the channel) are still active, it manually invokes cleanup() to
//     avoid memory leaks or stale channel pointers.
//  3. Thread-Safety: The operation is managed by the runner's internal state machine, ensuring
//     that multiple concurrent Stop() calls are handled safely and consistently.
//
// Post-Condition:
// After a successful Stop(), the aggregator transitions back to the 'Stopped' state. It remains
// configured and can be re-initialized by calling Start() again.
//
// Parameters:
//   - ctx: Context with a timeout or deadline, defining the maximum duration to wait for
//     the background goroutine to exit gracefully and finish writing buffered data.
//
// Returns:
//   - error: nil if the stop was successful, or a context error if the deadline was exceeded.
//
// Example:
//
//	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := agg.Stop(stopCtx); err != nil {
//	    log.Printf("Graceful shutdown failed: %v", err)
//	}
func (o *agg) Stop(ctx context.Context) error {
	defer func() {
		// Recovery from panics during shutdown.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/stop", r)
		}
	}()

	r := o.getRunner()
	if r == nil {
		// Cleanup resources if the runner was never created but the flag is set.
		if o.op.Load() {
			o.cleanup()
		}
		return nil
	}

	// Signal the runner to terminate the processing goroutine.
	return r.Stop(ctx)
}

// Restart performs an atomic stop and start sequence to refresh the aggregator's state.
//
// Implementation Strategy:
//  1. Graceful Exit: Invokes the full Stop() sequence and waits for completion.
//  2. Scheduler Yield: Calls 'runtime.Gosched()' between the two phases to allow the Go
//     scheduler to finalize the cleanup and goroutine scheduling of the previous run loop.
//  3. Clean Startup: Executes the Start() sequence to initialize a fresh runner and run loop.
//
// Use Cases:
//   - Clearing the internal data pipeline after an intermittent failure.
//   - Resetting periodic timers for synchronous and asynchronous callbacks.
//
// Parameters:
//   - ctx: Context controlling the overall duration of the combined stop and start sequence.
//
// Returns:
//   - error: nil on success, or the first error encountered during the combined sequence.
func (o *agg) Restart(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/restart", r)
		}
	}()

	// Execute an orderly stop first.
	if e := o.Stop(ctx); e != nil {
		return e
	}

	// Yield control back to the scheduler for a cleaner state transition.
	runtime.Gosched()

	// Re-initialize and start the data pipeline.
	return o.Start(ctx)
}

// IsRunning provides a real-time status check of the aggregator's activity.
//
// Technical Insight:
// This method queries the underlying runner instance, which serves as the source of truth
// for the background goroutine's state. It includes additional sanity checks to
// detect inconsistent states between the runner and internal flags.
//
// State Synchronization:
// If the runner reports it's running but the aggregator hasn't successfully initialized
// its internal state (flag 'op' is false), this method handles the inconsistency
// by attempting to stop the zombie runner and reporting 'false'.
//
// Returns:
//   - bool: true if the processing goroutine is currently active, false otherwise.
func (o *agg) IsRunning() bool {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/isrunning", r)
		}
	}()

	r := o.getRunner()

	if r == nil {
		// Cleanup resources if operational but no runner exists (inconsistent state).
		if o.op.Load() {
			o.cleanup()
		}
		return false
	}

	if r.IsRunning() {
		if o.op.Load() {
			return true // Healthy running state.
		} else if r.Uptime() > time.Second {
			// Runner is active for more than 1s but the internal flag is false: inconsistent state.
			// Trigger a non-blocking stop.
			x, n := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer n()
			_ = r.Stop(x)
			return false
		} else {
			// Runner might be in its very early startup phase.
			return false
		}
	} else {
		// Runner is NOT active.
		if o.op.Load() {
			// Internal flag was still true: perform resource cleanup to sync state.
			o.cleanup()
			return false
		} else {
			return false
		}
	}
}

// Uptime returns the duration for which the current aggregator run has been active.
//
// Operational Detail:
// The duration is calculated from the moment the runner successfully entered the 'Running'
// state. If the aggregator is stopped or hasn't been started, this method returns 0.
//
// Returns:
//   - time.Duration: Time since the last successful Start() invocation.
func (o *agg) Uptime() time.Duration {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/uptime", r)
		}
	}()

	r := o.getRunner()
	if r == nil {
		return 0
	}
	return r.Uptime()
}

// ErrorsLast retrieves the most recent error captured during background processing.
//
// Monitoring Utility:
// It provides immediate access to the last failure encountered by the FctWriter or
// during callback execution, facilitating real-time alerting and diagnostics.
//
// Returns:
//   - error: The last recorded error instance from the runner, or nil if no errors occurred.
func (o *agg) ErrorsLast() error {
	r := o.getRunner()
	if r == nil {
		return nil
	}
	return r.ErrorsLast()
}

// ErrorsList provides a comprehensive history of all errors captured during the current runner's lifecycle.
//
// Diagnostic Insight:
// This list is useful for post-mortem analysis and identifying patterns of failure
// (e.g., persistent I/O issues or recurring callback panics).
//
// Returns:
//   - []error: A slice containing all captured errors, or nil if no errors exist.
func (o *agg) ErrorsList() []error {
	r := o.getRunner()
	if r == nil {
		return nil
	}
	errs := r.ErrorsList()
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// newRunner constructs a new specialized 'librun.StartStop' component configured to
// manage the aggregator's lifecycle and background execution loop.
func (o *agg) newRunner() librun.StartStop {
	// 'o.run' is the main logic, 'o.closeRun' is the cleanup hook for the runner.
	return librun.New(o.run, o.closeRun)
}

// getRunner retrieves the currently active runner instance from atomic storage.
func (o *agg) getRunner() librun.StartStop {
	return o.r.Load()
}

// setRunner atomically persists the provided runner instance into the aggregator's state.
func (o *agg) setRunner(r librun.StartStop) {
	if r != nil {
		o.r.Store(r)
	}
}
