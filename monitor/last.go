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
	"fmt"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	errpol "github.com/nabbar/golib/errors/pool"
	monsts "github.com/nabbar/golib/monitor/status"
)

// lastRun tracks the state and metrics from health check executions.
// It maintains status transitions, timing metrics, and error information.
type lastRun struct {
	status  *atomic.Uint64          // Current health status (KO, Warn, OK)
	runtime libatm.Value[time.Time] // Timestamp of the last status change

	isRise *atomic.Bool // True if currently transitioning to better status
	isFall *atomic.Bool // True if currently transitioning to worse status

	cntRise *atomic.Uint64 // Count of consecutive successful checks during rise
	cntFall *atomic.Uint64 // Count of consecutive failed checks during fall

	uptime   *atomic.Int64 // Total time in OK status
	downtime *atomic.Int64 // Total time in KO or Warn status
	riseTime *atomic.Int64 // Total time spent rising to better status
	fallTime *atomic.Int64 // Total time spent falling to worse status
	latency  *atomic.Int64 // Duration of the last health check execution

	err errpol.Pool // Error from the last health check, nil if successful
}

// newLastRun creates a new lastRun instance with initial KO status.
// The initial state indicates no health check has run yet.
func newLastRun() *lastRun {
	l := &lastRun{
		status:   new(atomic.Uint64),
		runtime:  libatm.NewValue[time.Time](),
		isRise:   new(atomic.Bool),
		isFall:   new(atomic.Bool),
		cntRise:  new(atomic.Uint64),
		cntFall:  new(atomic.Uint64),
		uptime:   new(atomic.Int64),
		downtime: new(atomic.Int64),
		riseTime: new(atomic.Int64),
		fallTime: new(atomic.Int64),
		latency:  new(atomic.Int64),
		err:      errpol.New(),
	}

	l.status.Store(monsts.KO.Uint64())
	l.runtime.Store(time.Now())
	// l.isRise.Store(false)
	// l.isFall.Store(false)
	// l.cntRise.Store(0)
	// l.cntFall.Store(0)
	// l.uptime.Store(0)
	// l.downtime.Store(0)
	// l.riseTime.Store(0)
	// l.fallTime.Store(0)
	// l.latency.Store(0)
	l.err.Clear()
	l.err.Add(fmt.Errorf("no healcheck still run"))

	return l
}

func (o *lastRun) Runtime() time.Duration {
	if t := o.runtime.Load(); t.IsZero() {
		return time.Duration(0)
	} else {
		return time.Since(t)
	}
}

// Latency returns the duration of the last health check execution.
// This is thread-safe.
func (o *lastRun) Latency() time.Duration {
	return time.Duration(o.latency.Load())
}

// FallTime returns the total time spent in falling transitions.
// This is thread-safe.
func (o *lastRun) FallTime() time.Duration {
	return time.Duration(o.fallTime.Load())
}

// RiseTime returns the total time spent in rising transitions.
// This is thread-safe.
func (o *lastRun) RiseTime() time.Duration {
	return time.Duration(o.riseTime.Load())
}

// UpTime returns the total time the component has been in OK status.
// This is thread-safe.
func (o *lastRun) UpTime() time.Duration {
	return time.Duration(o.uptime.Load())
}

// DownTime returns the total time the component has been in KO or Warn status.
// This is thread-safe.
func (o *lastRun) DownTime() time.Duration {
	return time.Duration(o.downtime.Load())
}

// Status returns the current health status.
// This is thread-safe.
func (o *lastRun) Status() monsts.Status {
	return monsts.ParseUint64(o.status.Load())
}

// IsRise returns true if currently transitioning to better health status.
// This is thread-safe.
func (o *lastRun) IsRise() bool {
	return o.isRise.Load()
}

// IsFall returns true if currently transitioning to worse health status.
// This is thread-safe.
func (o *lastRun) IsFall() bool {
	return o.isFall.Load()
}

// Error returns the error from the last health check.
// Returns nil if the last check was successful. This is thread-safe.
func (o *lastRun) Error() error {
	return o.err.Last()
}

// setStatus updates the status based on the health check result.
// It handles status transitions, counters, and time tracking.
// This is thread-safe.
func (o *lastRun) setStatus(err error, dur time.Duration, cfg *runCfg) {
	o.latency.Store(int64(dur))

	if err != nil {
		o.err.Clear()
		o.err.Add(err)
		o.setStatusFall(cfg)
	} else {
		o.err.Clear()
		o.setStatusRise(cfg)
	}
}

// setStatusFall handles status degradation logic.
// It increments fall counters and transitions status based on thresholds.
func (o *lastRun) setStatusFall(cfg *runCfg) {
	if cfg == nil {
		return
	}

	sts := o.Status()
	dur := int64(o.Runtime())
	o.runtime.Store(time.Now())

	o.cntRise.Store(0)
	o.cntFall.Add(1)

	switch sts {
	case monsts.OK:
		if o.cntFall.Load() >= uint64(cfg.fallCountWarn) {
			o.cntFall.Store(0)
			o.status.Store(monsts.Warn.Uint64())
		} else {
			o.status.Store(monsts.OK.Uint64())
		}

		o.isFall.Store(true)
		o.isRise.Store(false)
		o.fallTime.Add(dur)
		o.uptime.Add(dur)

	case monsts.Warn:
		if o.cntFall.Load() >= uint64(cfg.fallCountKO) {
			o.isFall.Store(false)
			o.cntFall.Store(0)
			o.status.Store(monsts.KO.Uint64())
		} else {
			o.isFall.Store(true)
			o.status.Store(monsts.Warn.Uint64())
		}

		o.isRise.Store(false)
		o.fallTime.Add(dur)
		o.downtime.Add(dur)

	default:
		o.cntFall.Store(0)
		o.isFall.Store(false)
		o.isRise.Store(false)
		o.status.Store(monsts.KO.Uint64())
		o.downtime.Add(dur)
	}
}

// setStatusRise handles status improvement logic.
// It increments rise counters and transitions status based on thresholds.
func (o *lastRun) setStatusRise(cfg *runCfg) {
	if cfg == nil {
		return
	}

	sts := o.Status()
	dur := int64(o.Runtime())
	o.runtime.Store(time.Now())

	o.cntFall.Store(0)
	o.cntRise.Add(1)

	switch sts {
	case monsts.KO:
		if o.cntRise.Load() >= uint64(cfg.riseCountKO) {
			o.cntRise.Store(0)
			o.status.Store(monsts.Warn.Uint64())
		} else {
			o.status.Store(monsts.KO.Uint64())
		}

		o.isFall.Store(false)
		o.isRise.Store(true)
		o.riseTime.Add(dur)
		o.downtime.Add(dur)

	case monsts.Warn:
		if o.cntRise.Load() >= uint64(cfg.riseCountWarn) {
			o.cntRise.Store(0)
			o.isRise.Store(false)
			o.status.Store(monsts.OK.Uint64())
		} else {
			o.isRise.Store(true)
			o.status.Store(monsts.Warn.Uint64())
		}

		o.isFall.Store(false)
		o.riseTime.Add(dur)
		o.downtime.Add(dur)

	default:
		o.cntRise.Store(0)
		o.isFall.Store(false)
		o.isRise.Store(false)
		o.status.Store(monsts.OK.Uint64())
		o.uptime.Add(dur)
	}
}

// Clone creates a deep copy of the lastRun instance.
// This is used to ensure atomic updates to the monitor status.
func (o *lastRun) Clone() *lastRun {
	n := newLastRun()

	n.status.Store(o.status.Load())
	n.runtime.Store(o.runtime.Load())

	n.isRise.Store(o.isRise.Load())
	n.isFall.Store(o.isFall.Load())

	n.cntRise.Store(o.cntRise.Load())
	n.cntFall.Store(o.cntFall.Load())

	n.uptime.Store(o.uptime.Load())
	n.downtime.Store(o.downtime.Load())
	n.riseTime.Store(o.riseTime.Load())
	n.fallTime.Store(o.fallTime.Load())
	n.latency.Store(o.latency.Load())

	// Copy errors
	n.err.Add(o.err.Slice()...)

	return n
}
