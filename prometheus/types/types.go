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

package types

import (
	"fmt"

	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// MetricType represents the type of a Prometheus metric.
//
// Prometheus supports four fundamental metric types, each designed for specific
// use cases. The type determines how the metric is stored, aggregated, and queried.
//
// # Metric Types
//
// Counter: For cumulative values that only increase (or reset to zero)
//   - Use for: request counts, error counts, bytes transferred
//   - Operations: increment, add
//   - PromQL: rate(), increase(), irate()
//
// Gauge: For values that can go up or down
//   - Use for: memory usage, active connections, queue size, temperature
//   - Operations: set, increment, decrement
//   - PromQL: direct value, avg_over_time(), min_over_time(), max_over_time()
//
// Histogram: For distributions with pre-defined buckets
//   - Use for: request duration, response size
//   - Provides: count, sum, and buckets
//   - PromQL: histogram_quantile(), rate(<metric>_sum), rate(<metric>_count)
//   - Pros: Aggregatable across instances, efficient
//   - Cons: Fixed bucket boundaries
//
// Summary: For distributions with streaming quantiles
//   - Use for: request duration, response size (when histogram doesn't fit)
//   - Provides: count, sum, and quantiles
//   - PromQL: direct quantile values from <metric>{quantile="0.X"}
//   - Pros: Accurate quantiles
//   - Cons: Cannot aggregate across instances, computationally expensive
//
// # Choosing a Type
//
// Use Counter when:
//   - Value only increases (requests, errors, bytes sent)
//   - You want to calculate rates
//
// Use Gauge when:
//   - Value can increase or decrease (memory, connections, queue depth)
//   - You want current value or averages
//
// Use Histogram when:
//   - You need percentiles (p50, p95, p99)
//   - You can define buckets upfront
//   - You need to aggregate across multiple instances
//   - Performance matters
//
// Use Summary when:
//   - You need precise quantiles
//   - You cannot define buckets upfront
//   - You don't need cross-instance aggregation
//
// # Example
//
//	// Counter for HTTP requests
//	requestCounter := metrics.NewMetrics("http_requests_total", types.Counter)
//
//	// Gauge for active connections
//	activaConns := metrics.NewMetrics("active_connections", types.Gauge)
//
//	// Histogram for request duration
//	latency := metrics.NewMetrics("request_duration_seconds", types.Histogram)
//	latency.AddBuckets(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)
//
//	// Summary for response size
//	size := metrics.NewMetrics("response_size_bytes", types.Summary)
//	size.SetObjectives(map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001})
type MetricType int

const (
	// None represents an uninitialized or invalid metric type.
	//
	// This value should not be used for actual metrics. It exists as a zero value
	// for the MetricType enum and can be used to detect uninitialized metrics.
	//
	// Any attempt to register a metric with type None will fail.
	None MetricType = iota

	// Counter represents a cumulative metric that only increases.
	//
	// A counter is a monotonically increasing value that starts at zero and can only
	// increase (or be reset to zero on restart). Counters are ideal for tracking
	// cumulative counts of events.
	//
	// Common uses:
	//   - Total number of HTTP requests
	//   - Total number of errors
	//   - Total bytes transferred
	//   - Total number of tasks processed
	//
	// Operations:
	//   - Inc(): Increment by 1
	//   - Add(n): Increment by n (n must be positive)
	//
	// PromQL patterns:
	//   - rate(metric[5m]): Requests per second
	//   - increase(metric[1h]): Total increase over 1 hour
	//   - irate(metric[5m]): Instantaneous rate
	//
	// Example:
	//   http_requests_total{method="GET",status="200"} 12345
	Counter

	// Gauge represents a metric that can arbitrarily go up or down.
	//
	// A gauge represents a single numerical value that can increase or decrease.
	// Gauges are typically used for measured values that can fluctuate.
	//
	// Common uses:
	//   - Current memory usage
	//   - Number of active connections
	//   - Queue depth
	//   - Temperature readings
	//   - CPU usage percentage
	//
	// Operations:
	//   - Set(v): Set to specific value
	//   - Inc(): Increment by 1
	//   - Dec(): Decrement by 1
	//   - Add(v): Add v (can be negative)
	//   - Sub(v): Subtract v
	//
	// PromQL patterns:
	//   - metric: Current value
	//   - avg_over_time(metric[5m]): Average value
	//   - max_over_time(metric[1h]): Maximum value
	//   - min_over_time(metric[1h]): Minimum value
	//
	// Example:
	//   active_connections{server="web-1"} 42
	Gauge

	// Histogram represents a distribution of values with predefined buckets.
	//
	// A histogram samples observations (usually request durations or response sizes)
	// and counts them in configurable buckets. It also provides a sum of all observed
	// values and a total count.
	//
	// Metrics generated:
	//   - <name>_bucket{le="X"}: Cumulative count of observations <= X
	//   - <name>_sum: Sum of all observed values
	//   - <name>_count: Total number of observations
	//
	// Common uses:
	//   - HTTP request duration
	//   - Response size distribution
	//   - Query execution time
	//   - Job processing duration
	//
	// Operations:
	//   - Observe(v): Record an observation
	//
	// PromQL patterns:
	//   - histogram_quantile(0.95, rate(metric_bucket[5m])): 95th percentile
	//   - rate(metric_sum[5m]) / rate(metric_count[5m]): Average value
	//   - rate(metric_count[5m]): Operations per second
	//
	// Bucket configuration:
	//   Buckets must be defined when creating the metric:
	//   metric.AddBuckets(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)
	//
	// Advantages:
	//   - Can calculate quantiles across multiple instances
	//   - Efficient storage and computation
	//   - Aggregatable
	//
	// Example:
	//   http_request_duration_seconds_bucket{le="0.1"} 95000
	//   http_request_duration_seconds_bucket{le="0.5"} 99500
	//   http_request_duration_seconds_bucket{le="+Inf"} 100000
	//   http_request_duration_seconds_sum 28000
	//   http_request_duration_seconds_count 100000
	Histogram

	// Summary represents a distribution with streaming quantiles.
	//
	// A summary samples observations and calculates configurable quantiles over
	// a sliding time window. It's similar to a histogram but computes quantiles
	// on the client side.
	//
	// Metrics generated:
	//   - <name>{quantile="X"}: Pre-calculated quantile value
	//   - <name>_sum: Sum of all observed values
	//   - <name>_count: Total number of observations
	//
	// Common uses:
	//   - Request latencies (when precise quantiles needed)
	//   - Response sizes (when bucket ranges unknown)
	//
	// Operations:
	//   - Observe(v): Record an observation
	//
	// PromQL patterns:
	//   - metric{quantile="0.95"}: Pre-calculated 95th percentile
	//   - rate(metric_sum[5m]) / rate(metric_count[5m]): Average value
	//
	// Objectives configuration:
	//   Quantiles and their error margins must be defined:
	//   metric.SetObjectives(map[float64]float64{
	//       0.5: 0.05,   // median with ±5% error
	//       0.9: 0.01,   // 90th percentile with ±1% error
	//       0.99: 0.001, // 99th percentile with ±0.1% error
	//   })
	//
	// Disadvantages:
	//   - Cannot aggregate quantiles across instances
	//   - More computationally expensive than histograms
	//   - Quantiles are calculated per instance
	//
	// Recommendation:
	//   Prefer Histogram over Summary in most cases. Use Summary only when:
	//     - You cannot define buckets upfront
	//     - You need very precise quantiles for a single instance
	//     - Cross-instance aggregation is not needed
	//
	// Example:
	//   http_request_duration_seconds{quantile="0.5"} 0.15
	//   http_request_duration_seconds{quantile="0.9"} 0.35
	//   http_request_duration_seconds{quantile="0.99"} 0.75
	//   http_request_duration_seconds_sum 28000
	//   http_request_duration_seconds_count 100000
	Summary
)

// Register converts a Metric into a Prometheus Collector.
//
// This method creates the appropriate Prometheus collector (CounterVec, GaugeVec,
// HistogramVec, or SummaryVec) based on the metric's type. The collector can then
// be registered with a Prometheus registry.
//
// # Type-Specific Behavior
//
// Counter:
//   - Creates a CounterVec with the metric's name, description, and labels
//   - No additional configuration required
//
// Gauge:
//   - Creates a GaugeVec with the metric's name, description, and labels
//   - No additional configuration required
//
// Histogram:
//   - Creates a HistogramVec with buckets from GetBuckets()
//   - Returns error if buckets are not configured
//   - Error message: "metric 'X' is histogram type, cannot lose bucket param"
//
// Summary:
//   - Creates a SummaryVec with objectives from GetObjectives()
//   - Returns error if objectives are not configured
//   - Error message: "metric 'X' is summary type, cannot lose objectives param"
//
// None:
//   - Returns error: "metric type is not compatible"
//
// # Parameters
//
//   - met: The metric to register. Must provide name, description, labels, and
//     type-specific configuration (buckets for histogram, objectives for summary)
//
// # Returns
//
//   - prmsdk.Collector: The created Prometheus collector, ready for registration
//   - error: Non-nil if the metric type is invalid or required configuration is missing
//
// # Example
//
//	// Create and register a histogram
//	metric := metrics.NewMetrics("http_request_duration_seconds", types.Histogram)
//	metric.SetDesc("HTTP request latency")
//	metric.AddLabel("method", "status")
//	metric.AddBuckets(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)
//
//	collector, err := metric.GetType().Register(metric)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	prometheus.MustRegister(collector)
//
// # Common Errors
//
//   - "metric type is not compatible": Metric type is None or invalid
//   - "cannot lose bucket param": Histogram metric has no buckets configured
//   - "cannot lose objectives param": Summary metric has no objectives configured
//
// # Thread Safety
//
// This method is safe to call concurrently. The returned collector is also
// thread-safe and can be used from multiple goroutines.
func (m MetricType) Register(met Metric) (prmsdk.Collector, error) {
	switch met.GetType() {
	case Counter:
		return m.counterHandler(met)
	case Gauge:
		return m.gaugeHandler(met)
	case Histogram:
		return m.histogramHandler(met)
	case Summary:
		return m.summaryHandler(met)
	}

	return nil, fmt.Errorf("metric type is not compatible")
}

// counterHandler creates a Prometheus CounterVec from a Counter metric.
//
// This internal handler is called by Register() when the metric type is Counter.
// It creates a CounterVec that can track multiple time series distinguished by labels.
//
// # Parameters
//
//   - met: The metric providing name, description, and label names
//
// # Returns
//
//   - A CounterVec configured with the metric's properties
//   - Always returns nil error (kept for interface consistency)
//
// # Generated Metric
//
// The created counter will:
//   - Have the specified name and help text
//   - Support the specified label dimensions
//   - Start at 0 for each label combination
//   - Only allow increment operations
//
// # Example Output
//
//	# HELP http_requests_total Total number of HTTP requests
//	# TYPE http_requests_total counter
//	http_requests_total{method="GET",status="200"} 12345
//	http_requests_total{method="POST",status="201"} 6789
func (m MetricType) counterHandler(met Metric) (prmsdk.Collector, error) {
	return prmsdk.NewCounterVec(
		prmsdk.CounterOpts{Name: met.GetName(), Help: met.GetDesc()},
		met.GetLabel(),
	), nil
}

// gaugeHandler creates a Prometheus GaugeVec from a Gauge metric.
//
// This internal handler is called by Register() when the metric type is Gauge.
// It creates a GaugeVec that can track multiple time series distinguished by labels.
//
// # Parameters
//
//   - met: The metric providing name, description, and label names
//
// # Returns
//
//   - A GaugeVec configured with the metric's properties
//   - Always returns nil error (kept for interface consistency)
//
// # Generated Metric
//
// The created gauge will:
//   - Have the specified name and help text
//   - Support the specified label dimensions
//   - Start at 0 for each label combination
//   - Allow set, increment, and decrement operations
//
// # Example Output
//
//	# HELP active_connections Current number of active connections
//	# TYPE active_connections gauge
//	active_connections{server="web-1"} 42
//	active_connections{server="web-2"} 38
func (m MetricType) gaugeHandler(met Metric) (prmsdk.Collector, error) {
	return prmsdk.NewGaugeVec(
		prmsdk.GaugeOpts{Name: met.GetName(), Help: met.GetDesc()},
		met.GetLabel(),
	), nil
}

// histogramHandler creates a Prometheus HistogramVec from a Histogram metric.
//
// This internal handler is called by Register() when the metric type is Histogram.
// It creates a HistogramVec with the specified bucket boundaries for tracking
// value distributions.
//
// # Parameters
//
//   - met: The metric providing name, description, labels, and bucket configuration
//
// # Returns
//
//   - A HistogramVec configured with the metric's properties and buckets
//   - Error if the metric has no buckets configured
//
// # Validation
//
// Returns error if GetBuckets() returns an empty slice, as histograms require
// bucket boundaries to function. The error message includes the metric name:
// "metric 'X' is histogram type, cannot lose bucket param"
//
// # Generated Metrics
//
// The created histogram will generate three metric families:
//   - <name>_bucket{le="X"}: Cumulative count for each bucket
//   - <name>_sum: Sum of all observed values
//   - <name>_count: Total number of observations
//
// # Bucket Configuration
//
// Buckets should be configured before registration:
//   - Use exponential buckets for latency: 0.005, 0.01, 0.025, 0.05, 0.1, ...
//   - Use linear buckets for sizes: 100, 1000, 10000, 100000, ...
//   - Always include appropriate range for your use case
//   - Buckets must be in ascending order
//
// # Example Output
//
//	# HELP http_request_duration_seconds HTTP request latency
//	# TYPE http_request_duration_seconds histogram
//	http_request_duration_seconds_bucket{method="GET",le="0.005"} 850
//	http_request_duration_seconds_bucket{method="GET",le="0.01"} 920
//	http_request_duration_seconds_bucket{method="GET",le="0.025"} 990
//	http_request_duration_seconds_bucket{method="GET",le="+Inf"} 1000
//	http_request_duration_seconds_sum{method="GET"} 12.5
//	http_request_duration_seconds_count{method="GET"} 1000
func (m MetricType) histogramHandler(met Metric) (prmsdk.Collector, error) {
	if len(met.GetBuckets()) == 0 {
		return nil, fmt.Errorf("metric '%s' is histogram type, cannot lose bucket param", met.GetName())
	}

	return prmsdk.NewHistogramVec(
		prmsdk.HistogramOpts{Name: met.GetName(), Help: met.GetDesc(), Buckets: met.GetBuckets()},
		met.GetLabel(),
	), nil
}

// summaryHandler creates a Prometheus SummaryVec from a Summary metric.
//
// This internal handler is called by Register() when the metric type is Summary.
// It creates a SummaryVec with the specified quantile objectives for streaming
// quantile calculation.
//
// # Parameters
//
//   - met: The metric providing name, description, labels, and objectives configuration
//
// # Returns
//
//   - A SummaryVec configured with the metric's properties and objectives
//   - Error if the metric has no objectives configured
//
// # Validation
//
// Returns error if GetObjectives() returns an empty map, as summaries require
// quantile objectives to function. The error message includes the metric name:
// "metric 'X' is summary type, cannot lose objectives param"
//
// # Generated Metrics
//
// The created summary will generate metrics for each configured quantile plus totals:
//   - <name>{quantile="X"}: Pre-calculated quantile value
//   - <name>_sum: Sum of all observed values
//   - <name>_count: Total number of observations
//
// # Objectives Configuration
//
// Objectives define which quantiles to track and their error tolerance:
//   - Key: Quantile value (0.0 to 1.0)
//   - Value: Allowed error (e.g., 0.01 = ±1%)
//
// Example:
//
//	map[float64]float64{
//	    0.5: 0.05,   // Median with ±5% error
//	    0.9: 0.01,   // 90th percentile with ±1% error
//	    0.99: 0.001, // 99th percentile with ±0.1% error
//	}
//
// # Performance Note
//
// Summaries use more CPU and memory than histograms because they maintain
// a sliding window of observations. Use histograms unless you specifically need
// the features that summaries provide.
//
// # Example Output
//
//	# HELP http_request_duration_seconds HTTP request latency
//	# TYPE http_request_duration_seconds summary
//	http_request_duration_seconds{method="GET",quantile="0.5"} 0.12
//	http_request_duration_seconds{method="GET",quantile="0.9"} 0.35
//	http_request_duration_seconds{method="GET",quantile="0.99"} 0.75
//	http_request_duration_seconds_sum{method="GET"} 150.5
//	http_request_duration_seconds_count{method="GET"} 1000
func (m MetricType) summaryHandler(met Metric) (prmsdk.Collector, error) {
	if len(met.GetObjectives()) == 0 {
		return nil, fmt.Errorf("metric '%s' is summary type, cannot lose objectives param", met.GetName())
	}

	return prmsdk.NewSummaryVec(
		prmsdk.SummaryOpts{Name: met.GetName(), Help: met.GetDesc(), Objectives: met.GetObjectives()},
		met.GetLabel(),
	), nil
}
