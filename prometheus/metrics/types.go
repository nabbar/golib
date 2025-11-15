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

// SetDesc sets the description (help text) for the metric.
//
// The description provides context about what the metric measures and is displayed
// in Prometheus UI and API responses. It should be descriptive but concise.
//
// This should be set before registering the metric with Prometheus.
// Thread-safe: uses atomic store operation.
func (m *metrics) SetDesc(desc string) {
	m.d.Store(desc)
}

// GetDesc returns the metric description.
//
// Returns the currently set description, or an empty string if none has been set.
// Thread-safe: uses atomic load operation.
func (m *metrics) GetDesc() string {
	return m.d.Load()
}

// AddLabel adds one or more label names to the metric.
//
// Labels define the dimensions of the metric, allowing you to differentiate
// between different instances of the same metric (e.g., by HTTP method or status code).
//
// Important notes:
//   - Labels must be added before registering the metric
//   - The order of labels is preserved and must match the order of label values when updating
//   - Label names should follow Prometheus conventions (lowercase with underscores)
//
// Thread-safe: uses atomic read-modify-write pattern.
//
// Example:
//
//	m.AddLabel("method", "status")
//	// Later when updating:
//	m.Inc([]string{"GET", "200"})  // method=GET, status=200
func (m *metrics) AddLabel(label ...string) {
	p := append(make([]string, 0), m.l.Load()...)
	m.l.Store(append(p, label...))
}

// GetLabel returns the list of label names for this metric.
//
// Returns a slice of label names in the order they were added.
// Thread-safe: uses atomic load operation.
func (m *metrics) GetLabel() []string {
	return m.l.Load()
}

// AddBuckets adds histogram buckets to the metric.
//
// Buckets define the observation boundaries for Histogram metrics. Each bucket
// counts observations less than or equal to its upper bound.
//
// Only applicable for Histogram type metrics. Must be called before registering
// the metric. Buckets are cumulative and should be in ascending order.
//
// Thread-safe: uses atomic read-modify-write pattern.
//
// Example:
//
//	m.AddBuckets(0.1, 0.5, 1.0, 5.0)  // Create 4 buckets
//
// For bucket design best practices, see:
// https://prometheus.io/docs/practices/histograms/
func (m *metrics) AddBuckets(bucket ...float64) {
	p := append(make([]float64, 0), m.b.Load()...)
	m.b.Store(append(p, bucket...))
}

// GetBuckets returns the list of histogram buckets.
//
// Returns a slice of bucket upper bounds in the order they were added.
// Thread-safe: uses atomic load operation.
func (m *metrics) GetBuckets() []float64 {
	return m.b.Load()
}

// AddObjective adds a quantile objective for Summary metrics.
//
// Objectives define which quantiles (percentiles) to track and their allowed error margins.
// For example, AddObjective(0.5, 0.05) tracks the median (50th percentile) with ±5% error.
//
// Parameters:
//   - key: The quantile to track (must be between 0 and 1)
//   - value: The allowed error margin (typically 0.01 to 0.1)
//
// Only applicable for Summary type metrics. Must be called before registration.
// If the same quantile is added twice, the last value overwrites the previous one.
//
// Thread-safe: uses atomic read-modify-write pattern.
//
// Example:
//
//	m.AddObjective(0.5, 0.05)   // Track median ±5%
//	m.AddObjective(0.95, 0.01)  // Track 95th percentile ±1%
//	m.AddObjective(0.99, 0.001) // Track 99th percentile ±0.1%
func (m *metrics) AddObjective(key, value float64) {
	r := make(map[float64]float64)
	for k, v := range m.o.Load() {
		r[k] = v
	}
	r[key] = value
	m.o.Store(r)
}

// GetObjectives returns the map of quantile objectives for Summary metrics.
//
// The returned map has quantiles (0-1) as keys and allowed error margins as values.
// Thread-safe: uses atomic load operation.
func (m *metrics) GetObjectives() map[float64]float64 {
	return m.o.Load()
}
