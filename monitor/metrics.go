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
	"slices"
	"time"

	monsts "github.com/nabbar/golib/monitor/status"
	libprm "github.com/nabbar/golib/prometheus"
	librun "github.com/nabbar/golib/runner"
)

// RegisterMetricsName registers a list of metric names that will be used during Prometheus metrics collection.
//
// Parameters:
//   - names: A slice of strings representing the Prometheus metric identifiers for this monitor.
//
// Thread-Safety:
// This method is thread-safe and replaces any previously registered list of names.
// It includes a recovery mechanism to handle potential panics during storage in the internal context.
func (o *mon) RegisterMetricsName(names ...string) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/RegisterMetricsName", r)
		}
	}()

	o.x.Store(keyMetricsName, names)
}

// RegisterMetricsAddName appends new metric names to the existing list of registered Prometheus metric identifiers.
//
// Performance Detail:
// This method performs a Copy-On-Write (COW) operation by loading the existing list,
// merging it with the new names, ensuring no duplicates (using slices.Contains),
// and storing the new slice back in the internal context map.
//
// Thread-Safety:
// Safe for concurrent use across multiple goroutines.
func (o *mon) RegisterMetricsAddName(names ...string) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/RegisterMetricsAddName", r)
		}
	}()

	var n []string
	if i, l := o.x.Load(keyMetricsName); !l || i == nil {
		n = make([]string, 0)
	} else if v, k := i.([]string); !k {
		n = make([]string, 0)
	} else {
		n = v
	}

	for _, i := range names {
		if !slices.Contains(n, i) {
			n = append(n, i)
		}
	}

	o.x.Store(keyMetricsName, n)
}

// RegisterCollectMetrics associates a collection function with the monitor instance.
//
// Parameters:
//   - fct: A provider function (libprm.FuncCollectMetrics) that will be triggered after
//     each successful health check cycle to export current metrics to Prometheus.
//
// The registered function will receive the monitor's metrics (latency, status, rise/fall times)
// via the Collect* methods during the Prometheus scrape cycle.
func (o *mon) RegisterCollectMetrics(fct libprm.FuncCollectMetrics) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/RegisterCollectMetrics", r)
		}
	}()

	o.x.Store(keyMetricsFunc, fct)
}

// CollectLatency retrieves the execution duration of the most recent health check.
// This method is primarily intended for use by a Prometheus metrics collector function.
func (o *mon) CollectLatency() time.Duration {
	return o.Latency()
}

// CollectUpTime retrieves the total duration the monitored component has spent in the OK state.
// This method is primarily intended for use by a Prometheus metrics collector function.
func (o *mon) CollectUpTime() time.Duration {
	return o.Uptime()
}

// CollectDownTime retrieves the total duration the monitored component has spent in degraded states (Warn or KO).
// This method is primarily intended for use by a Prometheus metrics collector function.
func (o *mon) CollectDownTime() time.Duration {
	return o.Downtime()
}

// CollectRiseTime retrieves the total cumulative duration spent in "rising" transitions.
// This method is primarily intended for use by a Prometheus metrics collector function.
func (o *mon) CollectRiseTime() time.Duration {
	return o.l.RiseTime()
}

// CollectFallTime retrieves the total cumulative duration spent in "falling" transitions.
// This method is primarily intended for use by a Prometheus metrics collector function.
func (o *mon) CollectFallTime() time.Duration {
	return o.l.FallTime()
}

// CollectStatus returns a snapshot of the current health status and transition indicators (Rise/Fall).
// This method is primarily intended for use by a Prometheus metrics collector function.
func (o *mon) CollectStatus() (sts monsts.Status, rise bool, fall bool) {
	return o.Status(), o.IsRise(), o.IsFall()
}

// collectMetrics is an internal method that orchestrates the dispatch of monitor metrics to the
// registered Prometheus collector function.
//
// Dispatch Workflow:
//  1. Retrieves the registered metric names identifiers.
//  2. Retrieves the registered collector function.
//  3. If both are present, invokes the collector function with the provided context.
//
// Trigger:
// This method is automatically called at the end of each health check execution pipeline.
func (o *mon) collectMetrics(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/collectMetrics", r)
		}
	}()

	var (
		n []string
		f libprm.FuncCollectMetrics
	)

	// Attempt to load metric identifiers.
	if i, l := o.x.Load(keyMetricsName); !l || i == nil {
		return
	} else if v, k := i.([]string); !k {
		return
	} else {
		n = v
	}

	// Attempt to load the collection provider function.
	if i, l := o.x.Load(keyMetricsFunc); !l {
		return
	} else if v, k := i.(libprm.FuncCollectMetrics); !k {
		return
	} else {
		f = v
	}

	// Guard against empty registration.
	if len(n) < 1 || f == nil {
		return
	}

	// Execute the collector function for Prometheus reporting.
	f(ctx, n...)
}
