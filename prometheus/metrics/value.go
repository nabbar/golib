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
	"fmt"

	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// SetGaugeValue sets the absolute value for a Gauge type metric.
//
// Unlike Inc and Add which modify the current value, SetGaugeValue replaces
// the metric value entirely. This is useful for metrics that represent a current
// state (e.g., temperature, memory usage, queue size).
//
// The labelValues must match the labels defined when creating the metric, in the same order.
// The metric must be registered before calling this method.
//
// Parameters:
//   - labelValues: slice of label values matching the metric's label names in order
//   - value: the value to set (can be any float64, including negative values)
//
// Returns an error if:
//   - the metric type is None (not initialized)
//   - the metric is not a Gauge type
//   - the metric has not been registered
//   - the registered collector is not a GaugeVec
//
// Thread-safe: uses atomic load operation.
//
// Example:
//
//	gauge := NewMetrics("queue_size", prmtps.Gauge)
//	gauge.AddLabel("queue")
//	gauge.Register(...)
//	gauge.SetGaugeValue([]string{"jobs"}, 42.0)  // Set queue_size{queue="jobs"} = 42
func (m *metrics) SetGaugeValue(labelValues []string, value float64) error {
	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	} else if m.t != prmtps.Gauge {
		return fmt.Errorf("metric '%s' not Gauge type", m.n)
	} else if v := m.v.Load(); v == nil {
		return fmt.Errorf("metric '%s' not registred", m.n)
	} else if g, k := v.(*prmsdk.GaugeVec); !k {
		return fmt.Errorf("metric '%s' not GaugeVec type", m.n)
	} else {
		g.WithLabelValues(labelValues...).Set(value)
		m.v.Store(g)
	}

	return nil
}

// Inc increments the metric value by 1.
//
// This is the most common operation for Counter metrics (e.g., counting requests,
// errors, or events). It can also be used with Gauge metrics when you need to
// increment by one.
//
// For Counter metrics, this increases the monotonic counter (which never decreases).
// For Gauge metrics, this increases the current value (which can later be decreased).
//
// The metric must be registered before calling this method.
//
// Parameters:
//   - labelValues: slice of label values matching the metric's label names in order
//
// Returns an error if:
//   - the metric type is None (not initialized)
//   - the metric is not a Counter or Gauge type
//   - the metric has not been registered
//   - the registered collector is not a CounterVec or GaugeVec
//
// Thread-safe: uses atomic load operation.
//
// Example:
//
//	counter := NewMetrics("requests_total", prmtps.Counter)
//	counter.AddLabel("method")
//	counter.Register(...)
//	counter.Inc([]string{"GET"})  // Increment requests_total{method="GET"} by 1
func (m *metrics) Inc(labelValues []string) error {
	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	} else if m.t != prmtps.Counter && m.t != prmtps.Gauge {
		return fmt.Errorf("metric '%s' not Counter or Gauge type", m.n)
	} else if v := m.v.Load(); v == nil {
		return fmt.Errorf("metric '%s' not registred", m.n)
	} else if g, k := v.(*prmsdk.GaugeVec); k {
		g.WithLabelValues(labelValues...).Inc()
		m.v.Store(g)
	} else if c, k := v.(*prmsdk.CounterVec); k {
		c.WithLabelValues(labelValues...).Inc()
		m.v.Store(c)
	} else {
		return fmt.Errorf("metric '%s' not CounterVec or GaugeVec type", m.n)
	}

	return nil
}

// Add adds the given value to the metric.
//
// This operation allows you to increment or decrement a metric by a specific amount.
// It's commonly used for batch operations or when you know the delta value.
//
// For Counter metrics:
//   - The value must be non-negative (>= 0)
//   - Prometheus will panic if you try to add a negative value
//   - Use for accumulating totals (e.g., bytes transferred, items processed)
//
// For Gauge metrics:
//   - The value can be positive (increment) or negative (decrement)
//   - Use for tracking values that can go up or down (e.g., temperature changes, cache size)
//
// The metric must be registered before calling this method.
//
// Parameters:
//   - labelValues: slice of label values matching the metric's label names in order
//   - value: the value to add (must be >= 0 for Counter, any value for Gauge)
//
// Returns an error if:
//   - the metric type is None (not initialized)
//   - the metric is not a Counter or Gauge type
//   - the metric has not been registered
//   - the registered collector is not a CounterVec or GaugeVec
//
// Thread-safe: uses atomic load operation.
//
// Example:
//
//	counter := NewMetrics("bytes_sent_total", prmtps.Counter)
//	counter.AddLabel("endpoint")
//	counter.Register(...)
//	counter.Add([]string{"/api"}, 1024.0)  // Add 1024 bytes
//
//	gauge := NewMetrics("temperature", prmtps.Gauge)
//	gauge.AddLabel("sensor")
//	gauge.Register(...)
//	gauge.Add([]string{"room1"}, -2.5)  // Decrease by 2.5 degrees
func (m *metrics) Add(labelValues []string, value float64) error {
	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	} else if m.t != prmtps.Counter && m.t != prmtps.Gauge {
		return fmt.Errorf("metric '%s' not Counter or Gauge type", m.n)
	} else if v := m.v.Load(); v == nil {
		return fmt.Errorf("metric '%s' not registred", m.n)
	} else if g, k := v.(*prmsdk.GaugeVec); k {
		g.WithLabelValues(labelValues...).Add(value)
		m.v.Store(g)
	} else if c, k := v.(*prmsdk.CounterVec); k {
		c.WithLabelValues(labelValues...).Add(value)
		m.v.Store(c)
	} else {
		return fmt.Errorf("metric '%s' not CounterVec or GaugeVec type", m.n)
	}

	return nil
}

// Observe adds a single observation to the metric.
//
// This operation is used to record individual measurements or durations.
// It's commonly used for tracking request durations, response sizes, or any
// value distribution that you want to analyze.
//
// For Histogram metrics:
//   - The observation is counted in the appropriate bucket (based on AddBuckets configuration)
//   - Provides count, sum, and bucket counts for calculating averages and percentiles
//   - Good for server-side percentile calculation with bounded memory
//
// For Summary metrics:
//   - The observation is used to calculate streaming quantiles (based on AddObjective configuration)
//   - Provides exact quantiles but uses more memory and CPU
//   - Good for client-side percentile calculation
//
// The metric must be registered before calling this method.
//
// Parameters:
//   - labelValues: slice of label values matching the metric's label names in order
//   - value: the observed value (typically a duration in seconds or a size in bytes)
//
// Returns an error if:
//   - the metric type is None (not initialized)
//   - the metric is not a Histogram or Summary type
//   - the metric has not been registered
//   - the registered collector is not a HistogramVec or SummaryVec
//
// Thread-safe: uses atomic load operation.
//
// Example:
//
//	histogram := NewMetrics("request_duration_seconds", prmtps.Histogram)
//	histogram.AddLabel("method", "endpoint")
//	histogram.AddBuckets(0.1, 0.5, 1.0, 5.0)
//	histogram.Register(...)
//	// Record a request that took 0.234 seconds
//	histogram.Observe([]string{"GET", "/api"}, 0.234)
//
// For choosing between Histogram and Summary, see:
// https://prometheus.io/docs/practices/histograms/
func (m *metrics) Observe(labelValues []string, value float64) error {
	if m.t == prmtps.None {
		return fmt.Errorf("metric '%s' not existed", m.n)
	} else if m.t != prmtps.Histogram && m.t != prmtps.Summary {
		return fmt.Errorf("metric '%s' not Histogram or Summary type", m.n)
	} else if v := m.v.Load(); v == nil {
		return fmt.Errorf("metric '%s' not registred", m.n)
	} else if h, k := v.(*prmsdk.HistogramVec); k {
		h.WithLabelValues(labelValues...).Observe(value)
		m.v.Store(h)
	} else if s, k := v.(*prmsdk.SummaryVec); k {
		s.WithLabelValues(labelValues...).Observe(value)
		m.v.Store(s)
	} else {
		return fmt.Errorf("metric '%s' not CounterVec or GaugeVec type", m.n)
	}

	return nil
}
