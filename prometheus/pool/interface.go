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

// Package pool provides a thread-safe registry for managing Prometheus metrics.
//
// This package offers a centralized way to store, retrieve, and manage Prometheus
// metrics throughout an application lifecycle. It provides automatic registration,
// de-registration, and lifecycle management of metrics.
//
// # Basic Usage
//
// Creating a pool and adding metrics:
//
//	pool := pool.New(contextFunc)
//
//	metric := metrics.NewMetrics("http_requests_total", types.Counter)
//	metric.SetDesc("Total HTTP requests")
//	metric.AddLabel("method", "status")
//	metric.SetCollect(func(ctx context.Context, m metrics.Metric) {
//	    // Custom collection logic
//	})
//
//	err := pool.Add(metric)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Thread Safety
//
// All pool operations are thread-safe and can be called concurrently from multiple
// goroutines. The pool uses github.com/nabbar/golib/context for internal storage.
//
// For more details on metrics, see: https://github.com/nabbar/golib/tree/main/prometheus/metrics
// For context management, see: https://github.com/nabbar/golib/tree/main/context
package pool

import (
	"context"

	libctx "github.com/nabbar/golib/context"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// FuncGet is a function type for retrieving a metric by name.
// This is typically used for custom metric access patterns.
type FuncGet func(name string) libmet.Metric

// FuncAdd is a function type for adding a metric to a collection.
// Returns an error if the metric cannot be added.
type FuncAdd func(metric libmet.Metric) error

// FuncSet is a function type for setting/replacing a metric at a specific key.
type FuncSet func(key string, metric libmet.Metric)

// FuncWalk is a function type for iterating over metrics in the pool.
// The function receives the pool reference, metric key, and metric value.
// Return false to stop the iteration early, true to continue.
type FuncWalk func(pool MetricPool, key string, val libmet.Metric) bool

// MetricPool is the interface for managing a collection of Prometheus metrics.
//
// It provides thread-safe operations for storing, retrieving, and managing metrics
// throughout the application lifecycle. All methods are safe for concurrent use.
type MetricPool interface {
	// Get retrieves a metric by its name from the pool.
	//
	// Returns nil if the metric doesn't exist or if the stored value
	// is not a valid Metric instance.
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Get(name string) libmet.Metric

	// Add registers a new metric in the pool and with Prometheus.
	//
	// This method performs several validation checks:
	//   - The metric name cannot be empty
	//   - The metric must have a collection function set
	//   - The metric must be successfully registered with Prometheus
	//
	// If any validation fails, an error is returned and the metric is not added.
	//
	// Returns an error if:
	//   - The metric name is empty
	//   - The collect function is nil
	//   - Prometheus registration fails (e.g., duplicate metric name)
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Add(metric libmet.Metric) error

	// Set directly stores a metric in the pool without validation or registration.
	//
	// Unlike Add, this method:
	//   - Does not validate the metric
	//   - Does not register the metric with Prometheus
	//   - Allows using a key different from the metric name
	//   - Can replace an existing metric at the same key
	//
	// Use Add for normal metric registration. Use Set for advanced scenarios
	// where you need custom keys or manual registration control.
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Set(key string, metric libmet.Metric)

	// Del removes a metric from the pool and unregisters it from Prometheus.
	//
	// If the metric exists, it will be:
	//   - Removed from the pool
	//   - Unregistered from Prometheus registry
	//
	// This is a no-op if the metric doesn't exist or is not a valid Metric.
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	Del(key string) error

	// List returns a slice containing the names of all metrics in the pool.
	//
	// The returned slice is a snapshot at the time of the call and will not
	// reflect subsequent changes to the pool.
	//
	// Returns an empty slice (never nil) if the pool is empty.
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	List() []string

	// Walk iterates over metrics in the pool, calling the provided function for each.
	//
	// The iteration continues until either:
	//   - All metrics have been visited
	//   - The function returns false
	//   - All specified limit keys have been visited (if limit is provided)
	//
	// Parameters:
	//   - fct: Function to call for each metric
	//   - limit: Optional list of specific metric keys to visit
	//
	// Returns true if all metrics were visited, false if stopped early.
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	//
	// Example:
	//
	//	// Visit all metrics
	//	pool.Walk(func(p MetricPool, key string, val libmet.Metric) bool {
	//	    fmt.Printf("Metric: %s, Type: %s\n", key, val.GetType())
	//	    return true
	//	})
	//
	//	// Visit only specific metrics
	//	pool.Walk(func(p MetricPool, key string, val libmet.Metric) bool {
	//	    val.Collect(context.Background())
	//	    return true
	//	}, "metric1", "metric2")
	Walk(fct FuncWalk, limit ...string)

	// Clear removes all metrics from the pool and unregisters them from Prometheus.
	//
	// This method iterates through all metrics in the pool, unregisters each one
	// from the Prometheus registry, and then removes it from the pool.
	//
	// Returns a slice of errors encountered during unregistration. An empty slice
	// indicates all metrics were successfully cleared.
	//
	// Use cases:
	//   - Cleaning up before application shutdown
	//   - Resetting metrics between test runs
	//   - Clearing metrics when reconfiguring monitoring
	//
	// Thread-safe: can be called concurrently from multiple goroutines.
	//
	// Example:
	//
	//	errors := pool.Clear()
	//	if len(errors) > 0 {
	//	    for _, err := range errors {
	//	        log.Printf("Error clearing metric: %v", err)
	//	    }
	//	}
	Clear() []error
}

// New creates a new MetricPool instance.
//
// The pool uses the provided context function for internal storage management.
// The context function should return a valid context that will be used throughout
// the pool's lifetime.
//
// Parameters:
//   - ctx: Function that returns the context for the pool
//
// Returns a new MetricPool ready to use.
//
// Example:
//
//	pool := pool.New(func() context.Context {
//	    return context.Background()
//	})
func New(ctx context.Context, reg prmsdk.Registerer) MetricPool {
	return &pool{
		p: libctx.New[string](ctx),
		r: reg,
	}
}
