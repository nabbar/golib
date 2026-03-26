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
	libdur "github.com/nabbar/golib/duration"
	errpol "github.com/nabbar/golib/errors/pool"
	monsts "github.com/nabbar/golib/monitor/status"
)

// lastRun is an internal structure used to track the state and metrics resulting from health check executions.
//
// Performance Optimization:
// This structure is designed for high-performance, lock-free access. It uses atomic primitives
// for all fields to allow concurrent status reads without CPU-intensive mutex locking.
type lastRun struct {
	// status stores the current health status (OK, Warn, KO) as a uint64 representation of monsts.Status.
	status *atomic.Uint64

	// runtime stores the timestamp of the last status change or last health check execution.
	runtime libatm.Value[time.Time]

	// isRise indicates if the monitored component is currently in a "rising" phase (transitioning to a better status).
	isRise *atomic.Bool

	// isFall indicates if the monitored component is currently in a "falling" phase (transitioning to a worse status).
	isFall *atomic.Bool

	// cntRise tracks the number of consecutive successful checks during a rise transition.
	cntRise *atomic.Uint64

	// cntFall tracks the number of consecutive failed checks during a fall transition.
	cntFall *atomic.Uint64

	// uptime accumulates the total duration (in nanoseconds) spent in the OK status.
	uptime *atomic.Int64

	// downtime accumulates the total duration (in nanoseconds) spent in non-OK statuses (KO or Warn).
	downtime *atomic.Int64

	// riseTime accumulates the total duration (in nanoseconds) spent in rising transitions.
	riseTime *atomic.Int64

	// fallTime accumulates the total duration (in nanoseconds) spent in falling transitions.
	fallTime *atomic.Int64

	// latency stores the execution duration of the very last health check (in nanoseconds).
	latency *atomic.Int64

	// err is a thread-safe pool of errors encountered during the last health check execution.
	err errpol.Pool
}

// newLastRun initializes and returns a new lastRun instance.
//
// Default State:
//   - Initial status: KO (indicates no check has run yet).
//   - Error pool: contains a "no healthcheck still run" message.
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
	l.err.Clear()
	l.err.Add(fmt.Errorf("no healcheck still run"))

	return l
}

// Runtime returns the duration elapsed since the last recorded status update or health check execution.
// It calculates the time difference between the current time and the atomically stored runtime timestamp.
func (o *lastRun) Runtime() time.Duration {
	if t := o.runtime.Load(); t.IsZero() {
		return time.Duration(0)
	} else {
		return time.Since(t)
	}
}

// Latency returns the execution duration of the most recent health check.
// This operation is lock-free and thread-safe.
func (o *lastRun) Latency() time.Duration {
	return time.Duration(o.latency.Load())
}

// FallTime returns the cumulative duration the component has spent in a "falling" state.
// This operation is lock-free and thread-safe.
func (o *lastRun) FallTime() time.Duration {
	return time.Duration(o.fallTime.Load())
}

// RiseTime returns the cumulative duration the component has spent in a "rising" state.
// This operation is lock-free and thread-safe.
func (o *lastRun) RiseTime() time.Duration {
	return time.Duration(o.riseTime.Load())
}

// UpTime returns the cumulative duration the component has been in the OK status.
// This operation is lock-free and thread-safe.
func (o *lastRun) UpTime() time.Duration {
	return time.Duration(o.uptime.Load())
}

// DownTime returns the cumulative duration the component has been in either Warn or KO status.
// This operation is lock-free and thread-safe.
func (o *lastRun) DownTime() time.Duration {
	return time.Duration(o.downtime.Load())
}

// Status retrieves the current health status of the monitored component.
// It parses the atomically stored uint64 back into a monsts.Status.
// This operation is lock-free and thread-safe.
func (o *lastRun) Status() monsts.Status {
	return monsts.ParseUint64(o.status.Load())
}

// IsRise reports whether the component is currently transitioning toward a better health state (Rise phase).
// This operation is lock-free and thread-safe.
func (o *lastRun) IsRise() bool {
	return o.isRise.Load()
}

// IsFall reports whether the component is currently transitioning toward a worse health state (Fall phase).
// This operation is lock-free and thread-safe.
func (o *lastRun) IsFall() bool {
	return o.isFall.Load()
}

// Error returns the last error captured during a health check execution.
// If the last check was successful, it might return nil.
func (o *lastRun) Error() error {
	return o.err.Last()
}

// LatencyString returns a string representation of the last health check latency,
// truncated to millisecond precision for readability.
func (o *lastRun) LatencyString() string {
	return libdur.Duration(o.latency.Load()).TruncateMilliseconds().String()
}

// latencyMS returns the last health check latency in milliseconds as a uint64.
func (o *lastRun) latencyMS() uint64 {
	return o.int64ToUint64(libdur.Duration(o.latency.Load()).TruncateMilliseconds().Milliseconds())
}

// fallTimeString returns a string representation of the total fall time, truncated to seconds.
func (o *lastRun) fallTimeString() string {
	return libdur.Duration(o.fallTime.Load()).TruncateSeconds().String()
}

// fallTimeEpoc returns the total fall time in seconds as a uint64.
func (o *lastRun) fallTimeEpoc() uint64 {
	return o.int64ToUint64(libdur.Duration(o.fallTime.Load()).TruncateSeconds().Seconds())
}

// riseTimeString returns a string representation of the total rise time, truncated to seconds.
func (o *lastRun) riseTimeString() string {
	return libdur.Duration(o.riseTime.Load()).TruncateSeconds().String()
}

// riseTimeEpoc returns the total rise time in seconds as a uint64.
func (o *lastRun) riseTimeEpoc() uint64 {
	return o.int64ToUint64(libdur.Duration(o.riseTime.Load()).TruncateSeconds().Seconds())
}

// upTimeString returns a string representation of the total uptime, truncated to seconds.
func (o *lastRun) upTimeString() string {
	return libdur.Duration(o.uptime.Load()).TruncateSeconds().String()
}

// upTimeEpoc returns the total uptime in seconds as a uint64.
func (o *lastRun) upTimeEpoc() uint64 {
	return o.int64ToUint64(libdur.Duration(o.uptime.Load()).TruncateSeconds().Seconds())
}

// downTimeString returns a string representation of the total downtime, truncated to seconds.
func (o *lastRun) downTimeString() string {
	return libdur.Duration(o.downtime.Load()).TruncateSeconds().String()
}

// downTimeEpoc returns the total downtime in seconds as a uint64.
func (o *lastRun) downTimeEpoc() uint64 {
	return o.int64ToUint64(libdur.Duration(o.downtime.Load()).TruncateSeconds().Seconds())
}

// int64ToUint64 is a utility helper that converts an int64 to an uint64.
// It returns the absolute value to ensure a positive representation.
func (o *lastRun) int64ToUint64(i int64) uint64 {
	if i > 0 {
		return uint64(i)
	}

	return uint64(-i)
}

// setStatus updates the internal state of the lastRun based on the outcome of a health check.
//
// Parameters:
//   - err: The result of the health check diagnostic. Nil means success.
//   - dur: The measured latency of the health check execution.
//   - cfg: The configuration snapshot containing transition thresholds.
//
// This method handles the core state machine logic:
//  1. Stores latency.
//  2. Updates error pool.
//  3. Decides if a Rise or Fall transition should occur based on current state and consecutive results.
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

// setStatusFall implements the logic for health degradation (transition toward KO).
//
// Logic:
//   - If current status is OK: transition to Warn if failures >= fallCountWarn.
//   - If current status is Warn: transition to KO if failures >= fallCountKO.
//   - Accumulates fallTime and uptime/downtime accordingly.
func (o *lastRun) setStatusFall(cfg *runCfg) {
	if cfg == nil {
		return
	}

	sts := o.Status()
	dur := int64(o.Runtime())
	o.runtime.Store(time.Now())

	// Reset rise counter and increment fall counter upon failure.
	o.cntRise.Store(0)
	o.cntFall.Add(1)

	switch sts {
	case monsts.OK:
		// Threshold check for Warn transition.
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
		// Threshold check for KO transition.
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
		// Already KO.
		o.cntFall.Store(0)
		o.isFall.Store(false)
		o.isRise.Store(false)
		o.status.Store(monsts.KO.Uint64())
		o.downtime.Add(dur)
	}
}

// setStatusRise implements the logic for health improvement (transition toward OK).
//
// Logic:
//   - If current status is KO: transition to Warn if successes >= riseCountKO.
//   - If current status is Warn: transition to OK if successes >= riseCountWarn.
//   - Accumulates riseTime and uptime/downtime accordingly.
func (o *lastRun) setStatusRise(cfg *runCfg) {
	if cfg == nil {
		return
	}

	sts := o.Status()
	dur := int64(o.Runtime())
	o.runtime.Store(time.Now())

	// Reset fall counter and increment rise counter upon success.
	o.cntFall.Store(0)
	o.cntRise.Add(1)

	switch sts {
	case monsts.KO:
		// Threshold check for Warn transition.
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
		// Threshold check for OK transition.
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
		// Already OK.
		o.cntRise.Store(0)
		o.isFall.Store(false)
		o.isRise.Store(false)
		o.status.Store(monsts.OK.Uint64())
		o.uptime.Add(dur)
	}
}

// Clone creates and returns a deep copy of the lastRun instance.
// This is useful for capturing consistent snapshots for exporters or debuggers.
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

	// Thread-safe copy of the error slice.
	n.err.Add(o.err.Slice()...)

	return n
}
