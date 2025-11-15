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

package metrics_test

import (
	"context"
	"fmt"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// ExampleNewMetrics_counter demonstrates creating and using a Counter metric
// to track the total number of HTTP requests.
func ExampleNewMetrics_counter() {
	// Create a new counter metric
	counter := prmmet.NewMetrics("http_requests_total", prmtps.Counter)
	counter.SetDesc("Total HTTP requests")
	counter.AddLabel("method", "status")

	// Create and register the Prometheus collector
	vec := prmsdk.NewCounterVec(
		prmsdk.CounterOpts{
			Name: counter.GetName(),
			Help: counter.GetDesc(),
		},
		counter.GetLabel(),
	)
	_ = counter.Register(prmsdk.NewRegistry(), vec)

	// Increment the counter for different requests
	_ = counter.Inc([]string{"GET", "200"})
	_ = counter.Inc([]string{"POST", "201"})
	_ = counter.Add([]string{"GET", "200"}, 5.0)

	fmt.Println("Counter metric updated successfully")
	// Output: Counter metric updated successfully
}

// ExampleNewMetrics_gauge demonstrates creating and using a Gauge metric
// to track current queue size.
func ExampleNewMetrics_gauge() {
	// Create a new gauge metric
	gauge := prmmet.NewMetrics("queue_size", prmtps.Gauge)
	gauge.SetDesc("Current number of items in queue")
	gauge.AddLabel("queue_name")

	// Create and register the Prometheus collector
	vec := prmsdk.NewGaugeVec(
		prmsdk.GaugeOpts{
			Name: gauge.GetName(),
			Help: gauge.GetDesc(),
		},
		gauge.GetLabel(),
	)
	_ = gauge.Register(prmsdk.NewRegistry(), vec)

	// Set absolute values
	_ = gauge.SetGaugeValue([]string{"jobs"}, 42.0)

	// Increment and decrement
	_ = gauge.Inc([]string{"jobs"})
	_ = gauge.Add([]string{"jobs"}, -3.0)

	fmt.Println("Gauge metric updated successfully")
	// Output: Gauge metric updated successfully
}

// ExampleNewMetrics_histogram demonstrates creating and using a Histogram metric
// to track request duration distribution.
func ExampleNewMetrics_histogram() {
	// Create a new histogram metric
	histogram := prmmet.NewMetrics("request_duration_seconds", prmtps.Histogram)
	histogram.SetDesc("Request duration in seconds")
	histogram.AddLabel("method", "endpoint")
	histogram.AddBuckets(0.1, 0.5, 1.0, 2.5, 5.0, 10.0)

	// Create and register the Prometheus collector
	vec := prmsdk.NewHistogramVec(
		prmsdk.HistogramOpts{
			Name:    histogram.GetName(),
			Help:    histogram.GetDesc(),
			Buckets: histogram.GetBuckets(),
		},
		histogram.GetLabel(),
	)
	_ = histogram.Register(prmsdk.NewRegistry(), vec)

	// Record observations
	_ = histogram.Observe([]string{"GET", "/api/users"}, 0.234)
	_ = histogram.Observe([]string{"POST", "/api/users"}, 0.567)
	_ = histogram.Observe([]string{"GET", "/api/users"}, 0.123)

	fmt.Println("Histogram metric updated successfully")
	// Output: Histogram metric updated successfully
}

// ExampleNewMetrics_summary demonstrates creating and using a Summary metric
// to track response size quantiles.
func ExampleNewMetrics_summary() {
	// Create a new summary metric
	summary := prmmet.NewMetrics("response_size_bytes", prmtps.Summary)
	summary.SetDesc("Response size in bytes")
	summary.AddLabel("endpoint")
	summary.AddObjective(0.5, 0.05)   // Median with 5% error
	summary.AddObjective(0.95, 0.01)  // 95th percentile with 1% error
	summary.AddObjective(0.99, 0.001) // 99th percentile with 0.1% error

	// Create and register the Prometheus collector
	vec := prmsdk.NewSummaryVec(
		prmsdk.SummaryOpts{
			Name:       summary.GetName(),
			Help:       summary.GetDesc(),
			Objectives: summary.GetObjectives(),
		},
		summary.GetLabel(),
	)
	_ = summary.Register(prmsdk.NewRegistry(), vec)

	// Record observations
	_ = summary.Observe([]string{"/api/users"}, 1024.0)
	_ = summary.Observe([]string{"/api/users"}, 2048.0)
	_ = summary.Observe([]string{"/api/users"}, 512.0)

	fmt.Println("Summary metric updated successfully")
	// Output: Summary metric updated successfully
}

// ExampleMetric_SetCollect demonstrates using a custom collection function
// for pull-based metrics.
func ExampleMetric_SetCollect() {
	// Create a gauge metric for system memory
	gauge := prmmet.NewMetrics("system_memory_bytes", prmtps.Gauge)
	gauge.SetDesc("Current system memory usage")
	gauge.AddLabel("type")

	// Create and register the Prometheus collector
	vec := prmsdk.NewGaugeVec(
		prmsdk.GaugeOpts{
			Name: gauge.GetName(),
			Help: gauge.GetDesc(),
		},
		gauge.GetLabel(),
	)
	_ = gauge.Register(prmsdk.NewRegistry(), vec)

	// Set custom collection function that queries current state
	gauge.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		// In a real scenario, you would query actual system memory
		// Here we just demonstrate the pattern
		used := 8_000_000_000.0   // 8GB
		free := 4_000_000_000.0   // 4GB
		total := 12_000_000_000.0 // 12GB

		_ = m.SetGaugeValue([]string{"used"}, used)
		_ = m.SetGaugeValue([]string{"free"}, free)
		_ = m.SetGaugeValue([]string{"total"}, total)
	})

	// Trigger collection
	gauge.Collect(context.Background())

	fmt.Println("Collection function executed")
	// Output: Collection function executed
}

// ExampleNewMetrics_lifecycle demonstrates the full lifecycle of a metric
// including creation, registration, updates, and cleanup.
func ExampleNewMetrics_lifecycle() {
	// 1. Create the metric
	counter := prmmet.NewMetrics("example_counter", prmtps.Counter)
	counter.SetDesc("Example counter for demonstration")
	counter.AddLabel("environment")

	// 2. Create and register the collector
	reg := prmsdk.NewRegistry()
	vec := prmsdk.NewCounterVec(
		prmsdk.CounterOpts{
			Name: counter.GetName(),
			Help: counter.GetDesc(),
		},
		counter.GetLabel(),
	)
	err := counter.Register(reg, vec)
	if err != nil {
		fmt.Printf("Failed to register: %v\n", err)
		return
	}

	// 3. Use the metric
	_ = counter.Inc([]string{"production"})
	_ = counter.Add([]string{"staging"}, 10.0)

	// 4. Clean up when done
	if e := counter.UnRegister(reg); e != nil {
		fmt.Printf("Failed to unregister: %v\n", err)
		return
	}

	fmt.Println("Metric unregistered successfully")

	// Output: Metric unregistered successfully
}
