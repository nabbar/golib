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
	"time"

	"github.com/nabbar/golib/runner"
	librun "github.com/nabbar/golib/runner/startStop"
)

// Start begins processing writes in a background goroutine.
//
// This method:
//  1. Creates a new processing goroutine that reads from the write channel
//  2. Initializes timers for async and sync callbacks (if configured)
//  3. Opens the internal channel for accepting writes
//
// The aggregator must be started before it can accept Write operations.
// Calling Start on an already-running aggregator returns ErrStillRunning.
//
// The provided context is used as the parent context for the processing goroutine.
// When this context is cancelled, the aggregator stops processing and exits.
//
// Parameters:
//   - ctx: Context for cancellation. If nil, context.Background() is used.
//
// Returns:
//   - error: nil on success, ErrStillRunning if already started, or other error.
//
// Example:
//
//	agg, _ := aggregator.New(ctx, cfg, logger)
//	if err := agg.Start(ctx); err != nil {
//	    return err
//	}
//	defer agg.Close()
//
// Note: Start returns immediately. The processing goroutine starts asynchronously.
// A small delay (10ms) is added to allow the goroutine to initialize before returning.
func (o *agg) Start(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/start", r)
		}
	}()

	r := o.getRunner()
	if r == nil {
		r = o.newRunner()
	}

	e := r.Start(ctx)
	o.setRunner(r)

	time.Sleep(10 * time.Millisecond)
	return e
}

// Stop gracefully stops the aggregator's processing goroutine.
//
// This method:
//  1. Signals the processing goroutine to stop
//  2. Waits for the goroutine to exit (respecting the context deadline)
//  3. Does not close the write channel (use Close for full cleanup)
//
// After Stop, the aggregator can be restarted with Start().
// Stop is idempotent and can be called multiple times safely.
//
// Parameters:
//   - ctx: Context with deadline for graceful shutdown. The method blocks
//     until the processing goroutine exits or the context is cancelled.
//
// Returns:
//   - error: nil on success, or context error if timeout occurs.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	if err := agg.Stop(ctx); err != nil {
//	    log.Printf("stop failed: %v", err)
//	}
func (o *agg) Stop(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/stop", r)
		}
	}()

	r := o.getRunner()
	if r == nil {
		return nil
	}

	e := r.Stop(ctx)
	o.setRunner(r)

	time.Sleep(10 * time.Millisecond)
	return e
}

// Restart stops and then starts the aggregator.
//
// This is equivalent to calling Stop() followed by Start() with a small delay
// between them to ensure clean shutdown before restart.
//
// Restart is useful for:
//   - Applying configuration changes
//   - Recovering from errors
//   - Periodic restarts for resource cleanup
//
// Parameters:
//   - ctx: Context for the restart operation (used for both Stop and Start)
//
// Returns:
//   - error: nil on success, or error from Stop/Start
//
// Example:
//
//	if err := agg.Restart(ctx); err != nil {
//	    log.Printf("restart failed: %v", err)
//	}
func (o *agg) Restart(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/restart", r)
		}
	}()

	if e := o.Stop(ctx); e != nil {
		return e
	}

	time.Sleep(10 * time.Millisecond)

	return o.Start(ctx)
}

// IsRunning returns true if the aggregator is currently processing writes.
//
// This method checks both the runner state and the internal channel state,
// and automatically fixes any inconsistencies between them.
//
// IsRunning is useful for:
//   - Checking if the aggregator is ready to accept writes
//   - Monitoring aggregator health
//   - Implementing retry logic
//
// Returns:
//   - bool: true if the processing goroutine is running, false otherwise
//
// Example:
//
//	if !agg.IsRunning() {
//	    if err := agg.Start(ctx); err != nil {
//	        return err
//	    }
//	}
//	agg.Write(data)
func (o *agg) IsRunning() bool {
	defer func() {
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/isrunning", r)
		}
	}()

	r := o.getRunner()

	// If state is inconsistent, fix it without calling Close() to avoid recursion
	if r == nil {
		if o.op.Load() {
			o.chanClose()
			o.ctxClose()
		}
		return false
	}

	// Synchronize status between runner and channel state
	// Fix inconsistencies without changing the authoritative state
	if r.IsRunning() {
		if o.op.Load() {
			// Both runner and channel agree: running
			return true
		} else {
			// Runner says running but channel is closed: stop runner
			x, n := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer n()
			_ = o.Stop(x)
			return false
		}
	} else {
		if o.op.Load() {
			// Runner stopped but channel still open: close channel
			o.chanClose()
			o.ctxClose()
			return false
		} else {
			// Both agree: not running
			return false
		}
	}
}

// Uptime returns the duration since the aggregator was started.
//
// Returns 0 if the aggregator is not running or has never been started.
//
// Returns:
//   - time.Duration: Time since Start() was called, or 0 if not running
//
// Example:
//
//	if uptime := agg.Uptime(); uptime > 24*time.Hour {
//	    log.Printf("aggregator has been running for %v", uptime)
//	}
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

// ErrorsLast returns the most recent error from the processing goroutine.
//
// This includes errors from:
//   - The writer function (Config.FctWriter)
//   - Async/sync callbacks
//   - Internal processing errors
//
// Returns:
//   - error: The last error that occurred, or nil if no errors
//
// Example:
//
//	if err := agg.ErrorsLast(); err != nil {
//	    log.Printf("last error: %v", err)
//	}
func (o *agg) ErrorsLast() error {
	r := o.getRunner()
	if r == nil {
		return nil
	}
	return r.ErrorsLast()
}

// ErrorsList returns all errors that have occurred since the aggregator started.
//
// The list is maintained by the underlying runner and may be limited in size.
// See github.com/nabbar/golib/runner/startStop for details.
//
// Returns:
//   - []error: Slice of all errors, or nil if no errors or not running
//
// Example:
//
//	errs := agg.ErrorsList()
//	for _, err := range errs {
//	    log.Printf("error: %v", err)
//	}
func (o *agg) ErrorsList() []error {
	r := o.getRunner()
	if r == nil {
		return nil
	}
	return r.ErrorsList()
}

// newRunner creates a new StartStop runner with the aggregator's run and closeRun functions.
// The runner manages the lifecycle of the processing goroutine.
func (o *agg) newRunner() librun.StartStop {
	return librun.New(o.run, o.closeRun)
}

// getRunner returns the current runner instance.
// Returns nil if no runner has been created yet.
func (o *agg) getRunner() librun.StartStop {
	return o.r.Load()
}

// setRunner stores the runner instance, creating a new one if nil.
// This ensures the aggregator always has a valid runner.
func (o *agg) setRunner(r librun.StartStop) {
	if r == nil {
		r = o.newRunner()
	}
	o.r.Store(r)
}
