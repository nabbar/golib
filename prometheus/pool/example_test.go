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

package pool_test

import (
	"context"
	"fmt"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmpool "github.com/nabbar/golib/prometheus/pool"
	prmtps "github.com/nabbar/golib/prometheus/types"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// ExampleNew demonstrates creating a new metric pool.
func ExampleNew() {
	// Create a new pool
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	if pool != nil {
		fmt.Println("Pool created successfully")
	}
	// Output: Pool created successfully
}

// ExampleMetricPool_Add demonstrates adding metrics to a pool.
func ExampleMetricPool_Add() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// Create a counter metric
	counter := prmmet.NewMetrics("requests_total", prmtps.Counter)
	counter.SetDesc("Total number of requests")
	counter.AddLabel("method", "status")

	// Set a collection function (required for Add)
	counter.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		// Custom collection logic here
	})

	// Add to pool (automatically registers with Prometheus)
	err := pool.Add(counter)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Metric added successfully")
	// Output: Metric added successfully
}

// ExampleMetricPool_Get demonstrates retrieving metrics from a pool.
func ExampleMetricPool_Get() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// Create and add a metric
	gauge := prmmet.NewMetrics("temperature", prmtps.Gauge)
	gauge.SetDesc("Current temperature")
	gauge.AddLabel("location")
	gauge.SetCollect(func(ctx context.Context, m prmmet.Metric) {})

	_ = pool.Add(gauge)

	// Retrieve the metric
	retrieved := pool.Get("temperature")
	if retrieved != nil {
		metricType := "unknown"
		switch retrieved.GetType() {
		case prmtps.Counter:
			metricType = "Counter"
		case prmtps.Gauge:
			metricType = "Gauge"
		case prmtps.Histogram:
			metricType = "Histogram"
		case prmtps.Summary:
			metricType = "Summary"
		}
		fmt.Printf("Metric found: %s (type: %s)\n", retrieved.GetName(), metricType)
	}

	// Try to get non-existent metric
	notFound := pool.Get("does_not_exist")
	if notFound == nil {
		fmt.Println("Metric not found returns nil")
	}

	// Output:
	// Metric found: temperature (type: Gauge)
	// Metric not found returns nil
}

// ExampleMetricPool_List demonstrates listing all metrics in a pool.
func ExampleMetricPool_List() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// Add multiple metrics
	for i, name := range []string{"metric_a", "metric_b", "metric_c"} {
		m := prmmet.NewMetrics(name, prmtps.Counter)
		m.SetDesc(fmt.Sprintf("Metric %d", i))
		m.SetCollect(func(ctx context.Context, m prmmet.Metric) {})
		_ = pool.Add(m)
	}

	// List all metrics
	list := pool.List()
	fmt.Printf("Total metrics: %d\n", len(list))
	fmt.Println("Metrics contain 'metric_a':", contains(list, "metric_a"))

	// Output:
	// Total metrics: 3
	// Metrics contain 'metric_a': true
}

// ExampleMetricPool_Walk demonstrates iterating over metrics in a pool.
func ExampleMetricPool_Walk() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// Add metrics with very unique names to avoid conflicts
	types := []prmtps.MetricType{prmtps.Counter, prmtps.Gauge, prmtps.Histogram}
	for i, t := range types {
		m := prmmet.NewMetrics(fmt.Sprintf("example_walk_metric_%d_%d", i, 99999), t)
		m.SetDesc(fmt.Sprintf("Metric %d", i))
		if t == prmtps.Histogram {
			m.AddBuckets(prmsdk.DefBuckets...)
		}
		m.SetCollect(func(ctx context.Context, m prmmet.Metric) {})
		if err := pool.Add(m); err != nil {
			// Skip if already exists
			continue
		}
	}

	// Walk all metrics and collect types
	typeCount := make(map[string]int)
	totalCount := 0
	pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
		totalCount++
		metricType := "unknown"
		switch val.GetType() {
		case prmtps.Counter:
			metricType = "Counter"
		case prmtps.Gauge:
			metricType = "Gauge"
		case prmtps.Histogram:
			metricType = "Histogram"
		case prmtps.Summary:
			metricType = "Summary"
		}
		typeCount[metricType]++
		return true
	})

	fmt.Printf("Found Counter: %d\n", typeCount["Counter"])
	fmt.Printf("Found Gauge: %d\n", typeCount["Gauge"])
	fmt.Printf("Found Histogram: %d\n", typeCount["Histogram"])
	fmt.Printf("Total visited: %d\n", totalCount)
	// Output:
	// Found Counter: 1
	// Found Gauge: 1
	// Found Histogram: 1
	// Total visited: 3
}

// ExampleMetricPool_Walk_limit demonstrates walking specific metrics.
func ExampleMetricPool_Walk_limit() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// Add metrics with unique names
	for i := 1; i <= 4; i++ {
		name := fmt.Sprintf("limit_metric_%d", i)
		m := prmmet.NewMetrics(name, prmtps.Counter)
		m.SetDesc("Test metric")
		m.SetCollect(func(ctx context.Context, m prmmet.Metric) {})
		_ = pool.Add(m)
	}

	// Walk only specific metrics
	visited := 0
	pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
		visited++
		return true
	}, "limit_metric_1", "limit_metric_3")

	fmt.Printf("Visited specific metrics: %d\n", visited)
	fmt.Printf("Total metrics in pool: %d\n", len(pool.List()))
	// Output:
	// Visited specific metrics: 2
	// Total metrics in pool: 4
}

// ExampleMetricPool_Del demonstrates removing metrics from a pool.
func ExampleMetricPool_Del() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// Add a metric
	m := prmmet.NewMetrics("temporary_metric", prmtps.Counter)
	m.SetDesc("Temporary metric")
	m.SetCollect(func(ctx context.Context, m prmmet.Metric) {})
	_ = pool.Add(m)

	fmt.Println("Before delete:", len(pool.List()))

	// Delete the metric (also unregisters from Prometheus)
	pool.Del("temporary_metric")

	fmt.Println("After delete:", len(pool.List()))

	// Try to get deleted metric
	retrieved := pool.Get("temporary_metric")
	fmt.Println("Metric after delete:", retrieved)

	// Output:
	// Before delete: 1
	// After delete: 0
	// Metric after delete: <nil>
}

// ExampleMetricPool_lifecycle demonstrates the full lifecycle of metrics in a pool.
func ExampleMetricPool_lifecycle() {
	pool := prmpool.New(context.Background(), prmsdk.NewRegistry())

	// 1. Create a metric
	counter := prmmet.NewMetrics("http_requests", prmtps.Counter)
	counter.SetDesc("HTTP request counter")
	counter.AddLabel("method")
	counter.SetCollect(func(ctx context.Context, m prmmet.Metric) {})

	// 2. Add to pool (automatically registers)
	if err := pool.Add(counter); err != nil {
		fmt.Printf("Add failed: %v\n", err)
		return
	}
	fmt.Println("1. Metric added")

	// 3. Retrieve and use
	m := pool.Get("http_requests")
	if m != nil {
		fmt.Println("2. Metric retrieved")
	}

	// 4. List all metrics
	fmt.Printf("3. Pool contains %d metrics\n", len(pool.List()))

	// 5. Clean up when done
	pool.Del("http_requests")
	fmt.Println("4. Metric removed")

	// Output:
	// 1. Metric added
	// 2. Metric retrieved
	// 3. Pool contains 1 metrics
	// 4. Metric removed
}

// Helper function for examples
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
