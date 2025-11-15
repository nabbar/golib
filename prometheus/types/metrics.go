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

// Package types provides core type definitions for Prometheus metrics integration.
//
// This package defines the fundamental interfaces and types used throughout the
// github.com/nabbar/golib/prometheus ecosystem. It serves as the contract between
// metric implementations and the Prometheus client library.
//
// # Core Components
//
// The package provides:
//   - Metric interface: Contract for all metric implementations
//   - MetricType: Enumeration of supported Prometheus metric types
//   - Registration helpers: Methods to convert metrics into Prometheus collectors
//
// # Related Packages
//
//   - github.com/nabbar/golib/prometheus/metrics: Concrete metric implementations
//   - github.com/nabbar/golib/prometheus/pool: Metric pool management
//   - github.com/nabbar/golib/prometheus/webmetrics: Pre-configured web server metrics
//
// # Example Usage
//
// This package is typically used indirectly through higher-level packages:
//
//	import (
//	    "github.com/nabbar/golib/prometheus/metrics"
//	    "github.com/nabbar/golib/prometheus/types"
//	)
//
//	// Create a metric using the types
//	metric := metrics.NewMetrics("my_counter", types.Counter)
//	metric.SetDesc("Example counter metric")
//
//	// The metric type determines how it's registered with Prometheus
//	collector, err := metric.GetType().Register(metric)
package types

// Metric defines the interface that all Prometheus metrics must implement.
//
// This interface provides the contract for metric implementations, ensuring they
// can provide all necessary information for registration with Prometheus and
// proper metric collection.
//
// # Metric Properties
//
// Each metric must provide:
//   - A unique name following Prometheus naming conventions
//   - A type (Counter, Gauge, Histogram, or Summary)
//   - A human-readable description
//   - Optional labels for dimensional data
//   - Type-specific configuration (buckets for histograms, objectives for summaries)
//
// # Implementation
//
// This interface is implemented by github.com/nabbar/golib/prometheus/metrics.Metric.
// Users typically don't implement this interface directly but use the provided
// implementation.
//
// # Thread Safety
//
// Implementations of this interface should be safe for concurrent use, as metrics
// are often accessed from multiple goroutines simultaneously.
//
// # Example
//
//	// Using a concrete implementation
//	metric := metrics.NewMetrics("http_requests_total", types.Counter)
//	metric.SetDesc("Total HTTP requests")
//	metric.AddLabel("method", "status")
//
//	// Access via interface
//	var m types.Metric = metric
//	fmt.Println(m.GetName())   // "http_requests_total"
//	fmt.Println(m.GetType())   // types.Counter
type Metric interface {
	// GetName returns the metric name.
	//
	// The name must follow Prometheus naming conventions:
	//   - Use snake_case (e.g., "http_request_duration_seconds")
	//   - Use base units (seconds, bytes, not milliseconds or megabytes)
	//   - Include unit suffix where applicable (_seconds, _bytes, _total)
	//   - Must be unique within a Prometheus registry
	//
	// Example: "http_requests_total", "process_cpu_seconds_total"
	GetName() string

	// GetType returns the Prometheus metric type.
	//
	// Supported types:
	//   - Counter: Monotonically increasing value (e.g., request count)
	//   - Gauge: Value that can go up or down (e.g., memory usage)
	//   - Histogram: Distribution of values with buckets (e.g., request duration)
	//   - Summary: Distribution with configurable quantiles (e.g., response size)
	//
	// The type determines how Prometheus stores and queries the metric.
	GetType() MetricType

	// GetDesc returns the metric description.
	//
	// The description should:
	//   - Be human-readable and concise
	//   - Explain what the metric measures
	//   - Include units if not in the metric name
	//   - Be suitable for display in Prometheus UI and Grafana
	//
	// Example: "Total number of HTTP requests", "Current memory usage in bytes"
	GetDesc() string

	// GetLabel returns the list of label names for this metric.
	//
	// Labels provide dimensional data for filtering and grouping:
	//   - Each label name becomes a dimension in Prometheus
	//   - Label values are provided during metric collection
	//   - Use labels for bounded cardinality (avoid user IDs, timestamps)
	//
	// Returns an empty slice if the metric has no labels.
	//
	// Example: []string{"method", "status"} for HTTP metrics
	//
	// Warning: High cardinality (many unique label combinations) can impact
	// Prometheus performance. Keep label value sets bounded.
	GetLabel() []string

	// GetBuckets returns the histogram bucket boundaries.
	//
	// Required only for Histogram metrics. Buckets define the ranges for
	// grouping observations:
	//   - Each bucket counts observations less than or equal to its value
	//   - Buckets must be in ascending order
	//   - Common patterns: exponential, linear buckets
	//
	// Returns nil for non-histogram metrics.
	//
	// Example: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	// for duration in seconds.
	//
	// See prometheus.DefBuckets, prometheus.ExponentialBuckets, and
	// prometheus.LinearBuckets for bucket generation helpers.
	GetBuckets() []float64

	// GetObjectives returns the summary quantile objectives.
	//
	// Required only for Summary metrics. Objectives define which quantiles
	// to calculate and their allowed error:
	//   - Key: Quantile (0.5 = median, 0.95 = 95th percentile, 0.99 = 99th percentile)
	//   - Value: Allowed error (e.g., 0.01 = ±1%)
	//
	// Returns nil for non-summary metrics.
	//
	// Example: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	// means calculate p50 (±5%), p90 (±1%), and p99 (±0.1%).
	//
	// Note: Summaries are computationally more expensive than histograms and
	// cannot be aggregated across instances. Prefer histograms when possible.
	GetObjectives() map[float64]float64
}
