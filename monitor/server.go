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

package monitor

import (
	"context"
	"math"
	"time"

	montps "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/runner"
	runtck "github.com/nabbar/golib/runner/ticker"
)

const (
	// MaxPoolStart defines the maximum duration the Start method will wait for the internal ticker to become active.
	MaxPoolStart = 3 * time.Second

	// MaxTickPooler defines the interval at which the Start method checks if the internal ticker has successfully started.
	MaxTickPooler = 5 * time.Millisecond
)

// Start initiates the periodic health check execution cycle.
//
// Lifecycle Details:
//  1. Replaces the current ticker runner with a new one using the configured intervalCheck.
//  2. Gracefully stops the previous ticker runner if it was active.
//  3. Blocks until the new runner is confirmed as operational (poolIsRunning) or MaxPoolStart is reached.
//
// Thread-Safety:
// This method uses atomic Swap and Load operations to ensure consistent state transitions
// even when called concurrently.
//
// Returns:
//   - ErrorInvalid: If the monitor internal state is corrupted.
//   - ErrorTimeout: If the ticker fails to start within MaxPoolStart.
//   - error: Any error encountered during ticker startup.
func (o *mon) Start(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/Start", r)
		}
	}()

	// Invariants check.
	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return ErrorInvalid.Error(nil)
	}

	var c = o.getCfg()

	// Initialize new ticker with the primary check interval.
	r := runtck.New(c.intervalCheck, o.runFunc)
	e := r.Start(ctx)

	// Thread-safe replacement of the ticker runner.
	r = o.r.Swap(r)

	// Cleanup: Stop previous ticker if it was running.
	if r != nil && r.IsRunning() {
		_ = r.Stop(ctx)
	}

	if e != nil {
		return e
	} else {
		// Wait for operational confirmation.
		return o.poolIsRunning(ctx)
	}
}

// Stop gracefully terminates the periodic health check execution cycle.
//
// Workflow:
// Replaces the active runner with a dummy/inactive ticker (infinite interval) and stops
// the previous runner's background goroutine.
//
// Returns:
//   - ErrorInvalid: If the monitor is not properly initialized.
//   - error: Any error encountered during ticker termination.
func (o *mon) Stop(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/Stop", r)
		}
	}()

	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return ErrorInvalid.Error(nil)
	} else {
		// Replace current runner with an inactive one to halt checks immediately.
		r := o.r.Swap(runtck.New(time.Duration(math.MaxInt64), nil))
		if r != nil && r.IsRunning() {
			return r.Stop(ctx)
		} else {
			return nil
		}
	}
}

// Restart is a convenience method that performs a full stop followed by a start.
// This is useful when the configuration has changed significantly.
func (o *mon) Restart(ctx context.Context) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/Restart", r)
		}
	}()

	if e := o.Stop(ctx); e != nil {
		return e
	} else if e = o.Start(ctx); e != nil {
		return e
	}

	return nil
}

// IsRunning reports whether the monitor's background ticker is currently active and executing checks.
func (o *mon) IsRunning() bool {
	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return false
	} else if r := o.r.Load(); r == nil {
		return false
	} else {
		return r.IsRunning()
	}
}

// poolIsRunning implements a polling loop that waits until the background runner is 'running'.
// It uses MaxTickPooler for polling frequency and respects both the ctx cancellation and MaxPoolStart.
func (o *mon) poolIsRunning(ctx context.Context) error {
	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return ErrorInvalid.Error(nil)
	} else if r := o.r.Load(); r != nil && r.IsRunning() {
		return nil
	}

	var (
		tck = time.NewTicker(MaxTickPooler)
		tms = time.Now()
	)

	defer tck.Stop()

	for {
		select {
		case <-tck.C:
			if o.IsRunning() {
				return nil
			} else if time.Since(tms) >= MaxPoolStart {
				return ErrorTimeout.Error(nil)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// runFunc is the core worker function executed by the ticker runner on each interval tick.
//
// Logic:
//  1. Triggers the diagnostic 'check' logic.
//  2. Inspects current state (Rise/Fall) and dynamically adjusts the ticker's next interval.
//  3. If in Rise/Fall phase, it resets the ticker to the corresponding specific interval
//     (intervalRise or intervalFall). Otherwise, it resets to the standard intervalCheck.
func (o *mon) runFunc(ctx context.Context, tck librun.TickUpdate) error {
	var cfg = o.getCfg()

	o.check(ctx, cfg)

	// Dynamic Interval Adjustment:
	// This ensures the monitor reacts faster (or slower) during status transitions.
	if o.IsRise() {
		tck.Reset(cfg.intervalRise)
	} else if o.IsFall() {
		tck.Reset(cfg.intervalFall)
	} else {
		tck.Reset(cfg.intervalCheck)
	}

	return nil
}

// check performs a single execution of the health check diagnostic.
//
// Execution Chain:
//  1. Retrieves the diagnostic function and current configuration.
//  2. Wraps them into a middleware chain (middleWare).
//  3. Adds the mdlStatus middleware to the chain for latency tracking and status transition logic.
//  4. Executes the chain via m.Run().
//  5. Dispatches metrics to exporters at the end of execution.
func (o *mon) check(ctx context.Context, cfg *runCfg) {
	var fct montps.HealthCheck

	// Security Check: diagnostic function must be defined.
	if fct = o.getFct(); fct == nil {
		o.l.setStatus(ErrorMissingHealthCheck.Error(nil), 0, cfg)
		return
	} else if cfg == nil {
		o.l.setStatus(ErrorValidatorError.Error(nil), 0, cfg)
		return
	}

	// Prepare and execute the middleware pipeline.
	m := newMiddleware(cfg, fct)
	m.Add(o.mdlStatus)
	m.Run(ctx)

	// Prometheus Metrics Dispatch:
	o.collectMetrics(ctx)
}
