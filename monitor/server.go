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
	// MaxPoolStart is the maximum time to wait for the monitor to start.
	MaxPoolStart = 3 * time.Second
	// MaxTickPooler is the polling interval when waiting for the monitor to start.
	MaxTickPooler = 5 * time.Millisecond
)

// Start begins the periodic health check execution.
// It initializes the runner and waits for it to start successfully.
// Returns an error if the monitor is invalid or fails to start within MaxPoolStart.
func (o *mon) Start(ctx context.Context) error {
	defer librun.RecoveryCaller("golib/monitor/Start", recover())
	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return ErrorInvalid.Error(nil)
	}

	var c = o.getCfg()

	r := runtck.New(c.intervalCheck, o.runFunc)
	e := r.Start(ctx)
	r = o.r.Swap(r)

	if r != nil && r.IsRunning() {
		_ = r.Stop(ctx)
	}

	if e != nil {
		return e
	} else {
		return o.poolIsRunning(ctx)
	}
}

// Stop halts the periodic health check execution.
// It gracefully shuts down the runner and cleans up resources.
// Returns an error if the monitor is invalid or the runner fails to stop.
func (o *mon) Stop(ctx context.Context) error {
	defer librun.RecoveryCaller("golib/monitor/Stop", recover())
	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return ErrorInvalid.Error(nil)
	} else {
		r := o.r.Swap(runtck.New(time.Duration(math.MaxInt64), nil))
		if r != nil && r.IsRunning() {
			return r.Stop(ctx)
		} else {
			return nil
		}
	}
}

// Restart stops and then starts the monitor.
// Returns an error if either operation fails.
func (o *mon) Restart(ctx context.Context) error {
	defer librun.RecoveryCaller("golib/monitor/Restart", recover())
	if e := o.Stop(ctx); e != nil {
		return e
	} else if e = o.Start(ctx); e != nil {
		return e
	}

	return nil
}

// IsRunning returns true if the monitor is currently executing health checks.
func (o *mon) IsRunning() bool {
	defer librun.RecoveryCaller("golib/monitor/IsRunning", recover())
	if o == nil || o.x == nil || o.i == nil || o.r == nil {
		return false
	} else if r := o.r.Load(); r == nil {
		return false
	} else {
		return r.IsRunning()
	}
}

// poolIsRunning polls until the runner is confirmed to be running or a timeout occurs.
// Returns an error if the context is cancelled or MaxPoolStart is exceeded.
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

// runFunc is the function executed on each health check tick.
// It runs the health check and adjusts the ticker interval based on status transitions.
func (o *mon) runFunc(ctx context.Context, tck *time.Ticker) error {
	var cfg = o.getCfg()

	o.check(ctx, cfg)

	if o.IsRise() {
		tck.Reset(cfg.intervalRise)
	} else if o.IsFall() {
		tck.Reset(cfg.intervalFall)
	} else {
		tck.Reset(cfg.intervalCheck)
	}

	return nil
}

// check executes a single health check using the registered health check function.
// It handles missing health check or configuration errors, and collects metrics after execution.
func (o *mon) check(ctx context.Context, cfg *runCfg) {
	var fct montps.HealthCheck

	if fct = o.getFct(); fct == nil {
		l := o.getLastCheck()
		l.setStatus(ErrorMissingHealthCheck.Error(nil), 0, cfg)
		o.x.Store(keyLastRun, l)
	} else if cfg == nil {
		l := o.getLastCheck()
		l.setStatus(ErrorValidatorError.Error(nil), 0, cfg)
		o.x.Store(keyLastRun, l)
	}

	m := newMiddleware(cfg, fct)
	m.Add(o.mdlStatus)
	m.Run(ctx)

	// store metrics to prometheus exporter
	o.collectMetrics(ctx)
}
