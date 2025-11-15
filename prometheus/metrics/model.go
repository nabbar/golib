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

package metrics

import (
	"context"
	"fmt"

	libatm "github.com/nabbar/golib/atomic"
	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// metrics is the internal implementation of the Metric interface.
// It provides a thread-safe wrapper around Prometheus metric collectors using atomic operations
// from github.com/nabbar/golib/atomic for lock-free concurrent access.
//
// Each metric is uniquely identified by its name and type. The metric configuration
// (description, labels, buckets, objectives) should be set before registration.
// Once registered, the metric can accept value updates through Inc, Add, SetGaugeValue, or Observe.
//
// All fields except 't' and 'n' use atomic.Value to ensure thread-safe access without locks.
type metrics struct {
	t prmtps.MetricType                 // metric type (immutable after creation)
	n string                            // metric name (immutable after creation)
	d libatm.Value[string]              // metric description/help text
	l libatm.Value[[]string]            // label names for dimensional data
	b libatm.Value[[]float64]           // histogram bucket boundaries
	o libatm.Value[map[float64]float64] // summary quantile objectives
	f libatm.Value[FuncCollect]         // custom collection function
	v libatm.Value[prmsdk.Collector]    // registered Prometheus collector
}

// SetCollect sets a custom collection function for this metric.
//
// If a nil function is provided, a no-op function is stored instead to ensure
// Collect() can always be called safely without nil checks.
//
// The collection function will be invoked during metric scraping to allow
// on-demand metric value updates (pull-based metrics).
//
// Thread-safe: uses atomic store operation.
func (m *metrics) SetCollect(fct FuncCollect) {
	if fct == nil {
		fct = func(ctx context.Context, m Metric) {}
	}

	m.f.Store(fct)
}

// GetCollect returns the currently set collection function.
//
// Returns the registered FuncCollect, or nil if none has been set.
// Thread-safe: uses atomic load operation.
func (m *metrics) GetCollect() FuncCollect {
	return m.f.Load()
}

// GetName returns the metric name.
//
// The name is immutable after metric creation and can be safely accessed
// concurrently without synchronization.
func (m *metrics) GetName() string {
	return m.n
}

// GetType returns the metric type (Counter, Gauge, Histogram, Summary, or None).
//
// The type is immutable after metric creation and can be safely accessed
// concurrently without synchronization.
func (m *metrics) GetType() prmtps.MetricType {
	return m.t
}

// Collect invokes the custom collection function if one has been set.
//
// This method is typically called by Prometheus during metric scraping.
// The context is passed to the collection function to support cancellation
// and timeout handling.
//
// If no collection function has been registered (via SetCollect), this is a no-op.
// Thread-safe: uses atomic load operation to safely access the function pointer.
func (m *metrics) Collect(ctx context.Context) {
	if f := m.f.Load(); f != nil {
		f(ctx, m)
	}
}

// Register registers the metric collector with the Prometheus default registry.
//
// If a collector is already registered for this metric, it will be unregistered
// first to allow re-registration. This is useful when you need to update the
// metric configuration (e.g., change labels or buckets).
//
// Returns an error if registration with Prometheus fails (e.g., duplicate metric name).
// Thread-safe: uses atomic operations for collector access.
func (m *metrics) Register(reg prmsdk.Registerer, vec prmsdk.Collector) error {
	v := m.v.Load()
	if v != nil {
		reg.Unregister(v)
	}

	e := reg.Register(vec)
	m.v.Store(vec)

	return e
}

// UnRegister removes the metric from the Prometheus default registry.
//
// Returns true if the metric was successfully unregistered, false if it was
// never registered or already unregistered.
//
// This is useful for cleaning up metrics that are no longer needed, preventing
// stale data from being exported.
//
// Thread-safe: uses atomic load operation.
func (m *metrics) UnRegister(reg prmsdk.Registerer) error {
	v := m.v.Load()
	if v != nil {
		if !reg.Unregister(v) {
			return fmt.Errorf("cannot unregister metric: %s", m.n)
		} else {
			return nil
		}
	}

	return fmt.Errorf("invalid nil metric: %s", m.n)
}
