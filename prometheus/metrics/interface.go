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

// Package metrics provides a thread-safe wrapper around Prometheus metrics with full lifecycle management.
//
// This package simplifies the creation, configuration, and management of Prometheus metrics
// by providing a unified interface for Counter, Gauge, Histogram, and Summary metric types.
// All operations are thread-safe through the use of atomic operations from github.com/nabbar/golib/atomic.
//
// # Basic Usage
//
// Creating and registering a metric:
//
//	m := metrics.NewMetrics("http_requests_total", prometheus.Counter)
//	m.SetDesc("Total HTTP requests")
//	m.AddLabel("method", "status")
//	m.Register(prometheus.NewCounterVec(...))
//	m.Inc([]string{"GET", "200"})
//
// # Metric Types
//
// The package supports all Prometheus metric types:
//   - Counter: A cumulative metric that only increases
//   - Gauge: A metric that can go up and down
//   - Histogram: Samples observations and counts them in configurable buckets
//   - Summary: Similar to Histogram but calculates configurable quantiles
//
// # Thread Safety
//
// All metric operations are thread-safe and use atomic operations for lock-free concurrent access.
// This ensures high performance even under heavy concurrent load.
//
// For more details on Prometheus concepts, see: https://prometheus.io/docs/concepts/metric_types/
// For the golib atomic package, see: https://github.com/nabbar/golib/tree/main/atomic
package metrics

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// FuncGetMetric is a function type that returns a Metric instance.
// This is typically used in factory patterns or metric registry implementations.
type FuncGetMetric func() Metric

// FuncGetMetricName is a function type that returns a Metric instance by name.
// This is typically used for metric lookup operations in a registry.
type FuncGetMetricName func(name string) Metric

// FuncCollectMetrics is a function type that returns a Collect interface.
// This is used to retrieve collectors for metric aggregation.
type FuncCollectMetrics func() Collect

// FuncCollectMetricsByName is a function type that returns a Collect interface by name.
// This allows selective collection of metrics from a registry.
type FuncCollectMetricsByName func(name string) Collect

// Collect defines the interface for collecting metric data.
// Implementations should gather current metric values and prepare them for export.
type Collect interface {
	// Collect triggers the collection of metric data with the given context.
	// The context can be used to propagate cancellation signals and deadlines.
	Collect(ctx context.Context)
}

// Metric defines the interface for a Prometheus metric with full lifecycle management.
// It extends prmtps.Metric with additional functionality for value updates, registration,
// and custom collection logic.
type Metric interface {
	prmtps.Metric
	Collect

	// SetGaugeValue sets the value for a Gauge type metric with the specified label values.
	// Returns an error if the metric is not a Gauge type or doesn't exist.
	SetGaugeValue(labelValues []string, value float64) error

	// Inc increments the metric value by 1. Only valid for Counter and Gauge types.
	// Returns an error if the metric type doesn't support this operation.
	Inc(labelValues []string) error

	// Add adds the given value to the metric. Only valid for Counter and Gauge types.
	// For Counter, the value must be non-negative. For Gauge, negative values are allowed.
	// Returns an error if the metric type doesn't support this operation.
	Add(labelValues []string, value float64) error

	// Observe adds a single observation to the metric. Only valid for Histogram and Summary types.
	// Returns an error if the metric type doesn't support this operation.
	Observe(labelValues []string, value float64) error

	// Register registers the metric collector with Prometheus.
	// It will unregister any existing collector first to allow re-registration.
	// Returns an error if registration fails.
	Register(reg prmsdk.Registerer, vec prmsdk.Collector) error

	// UnRegister removes the metric from Prometheus registry.
	// Returns true if the metric was successfully unregistered, false otherwise.
	UnRegister(reg prmsdk.Registerer) error

	// SetDesc sets the description/help text for the metric.
	SetDesc(desc string)

	// AddLabel adds one or more label names to the metric.
	// Labels must be added before registering the metric.
	AddLabel(label ...string)

	// AddBuckets adds histogram buckets. Only applicable for Histogram type metrics.
	// Buckets must be added before registering the metric.
	AddBuckets(bucket ...float64)

	// AddObjective adds a quantile objective for Summary type metrics.
	// The key is the quantile (0-1), and value is the allowed error margin.
	// Objectives must be added before registering the metric.
	AddObjective(key, value float64)

	// SetCollect sets a custom collection function that will be called during metric collection.
	SetCollect(fct FuncCollect)

	// GetCollect returns the currently set collection function, or nil if none is set.
	GetCollect() FuncCollect
}

// FuncCollect is a function type for custom metric collection logic.
// It receives a context and the metric instance to allow updating metric values during collection.
//
// This is useful for pull-based metrics where values are computed on-demand rather than
// being continuously updated. The function is called automatically during metric scraping.
//
// Example:
//
//	m.SetCollect(func(ctx context.Context, metric Metric) {
//	    // Query current state
//	    value := getCurrentValue()
//	    // Update the metric
//	    metric.SetGaugeValue([]string{"label"}, value)
//	})
type FuncCollect func(ctx context.Context, m Metric)

// NewMetrics creates a new Metric instance with the specified name and type.
//
// The metric is initialized with empty configuration and uses atomic operations
// for thread-safe access. You must configure the metric (description, labels,
// buckets/objectives) before registering it with Prometheus.
//
// Parameters:
//   - name: The metric name (should follow Prometheus naming conventions)
//   - metricType: The type of metric (Counter, Gauge, Histogram, or Summary)
//
// Returns a new Metric instance ready to be configured.
//
// Example:
//
//	// Create a counter metric
//	counter := NewMetrics("requests_total", prmtps.Counter)
//	counter.SetDesc("Total number of requests")
//	counter.AddLabel("method", "status")
//
// For metric naming best practices, see:
// https://prometheus.io/docs/practices/naming/
func NewMetrics(name string, metricType prmtps.MetricType) Metric {
	return &metrics{
		t: metricType,
		n: name,
		d: libatm.NewValue[string](),
		l: libatm.NewValue[[]string](),
		b: libatm.NewValue[[]float64](),
		o: libatm.NewValue[map[float64]float64](),
		f: libatm.NewValue[FuncCollect](),
		v: libatm.NewValue[prmsdk.Collector](),
	}
}
