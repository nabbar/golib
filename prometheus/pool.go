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

package prometheus

import (
	"context"

	ginsdk "github.com/gin-gonic/gin"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmpol "github.com/nabbar/golib/prometheus/pool"
	librun "github.com/nabbar/golib/runner"
	libsem "github.com/nabbar/golib/semaphore"
)

// GetMetric retrieves a metric instance by name from the metric pool.
// Returns nil if the metric is not found.
//
// The method searches both API metrics (collected via Gin middleware) and
// non-API metrics (collected manually or via ticker).
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	metric := prm.GetMetric("http_requests_total")
//	if metric != nil {
//	    log.Printf("Metric found: %s", metric.GetName())
//	} else {
//	    log.Printf("Metric not found")
//	}
func (m *prom) GetMetric(name string) libmet.Metric {
	if v := m.web.Get(name); v != nil {
		return v
	} else if v = m.oth.Get(name); v != nil {
		return v
	}

	return nil
}

// AddMetric registers a metric instance in the Prometheus pool.
//
// Parameters:
//   - isAPI: If true, the metric is added to the API pool and collected via Gin middleware
//     after each request. If false, the metric is added to the non-API pool and
//     must be collected manually via CollectMetrics.
//   - metric: The metric instance to register. Must have a unique name and valid configuration.
//
// Returns an error if:
//   - The metric has no collect function configured
//   - A histogram metric has no buckets configured
//   - A summary metric has no objectives configured
//   - The metric type is None or invalid
//   - Registration with Prometheus fails
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// API metric (collected automatically)
//	err := prm.AddMetric(true, webmetrics.MetricRequestTotal("myapp"))
//
//	// Non-API metric (manual collection)
//	err := prm.AddMetric(false, customMetric)
//	if err != nil {
//	    log.Fatalf("Failed to add metric: %v", err)
//	}
func (m *prom) AddMetric(isAPI bool, metric libmet.Metric) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/prometheus/AddMetric", r)
		}
	}()

	if isAPI {
		return m.web.Add(metric)
	} else {
		return m.oth.Add(metric)
	}
}

// DelMetric unregisters and removes a metric from both metric pools.
// The metric will no longer be collected or exported to Prometheus.
//
// This operation is idempotent - it is safe to call with a non-existent metric name.
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	prm.DelMetric("temporary_metric")
//	// Metric is now unregistered and will not appear in /metrics
func (m *prom) DelMetric(name string) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/prometheus/DelMetric", r)
		}
	}()

	if m.web.Get(name) != nil {
		return m.web.Del(name)
	} else if m.oth.Get(name) != nil {
		return m.oth.Del(name)
	}
	return nil
}

// ListMetric returns a slice of all registered metric names.
// This includes both API metrics and non-API metrics.
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	names := prm.ListMetric()
//	log.Printf("Registered metrics: %v", names)
//	for _, name := range names {
//	    metric := prm.GetMetric(name)
//	    log.Printf("  - %s (%s)", name, metric.GetType())
//	}
func (m *prom) ListMetric() []string {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/prometheus/ListMetric", r)
		}
	}()

	var res = make([]string, 0)
	res = append(res, m.web.List()...)
	res = append(res, m.oth.List()...)
	return res
}

// ClearMetric removes and unregisters all metrics from the specified pools.
//
// This method clears metrics from one or both metric pools (API and non-API).
// Each metric is unregistered from Prometheus before being removed from the pool.
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Parameters:
//   - Api: If true, clears all non-API metrics (manually collected metrics from 'oth' pool)
//   - Web: If true, clears all API/Web metrics (Gin middleware collected metrics from 'web' pool)
//
// Returns a slice of errors encountered during unregistration.
// If both pools are cleared, errors from both are concatenated.
//
// Example:
//
//	// Clear all metrics before shutdown
//	errors := prm.ClearMetric(true, true)
//	for _, err := range errors {
//	    log.Printf("Clear error: %v", err)
//	}
func (m *prom) ClearMetric(Api, Web bool) []error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/prometheus/Clear", r)
		}
	}()

	var e = make([]error, 0)

	if Web {
		e = append(e, m.web.Clear()...)
	}
	if Api {
		e = append(e, m.oth.Clear()...)
	}

	return e
}

// Collect triggers collection of all metrics.
// This is a convenience method that calls CollectMetrics with the provided context.
//
// Example:
//
//	prm.Collect(context.Background())
func (m *prom) Collect(ctx context.Context) {
	m.CollectMetrics(ctx)
}

// CollectMetrics triggers metric collection for specified metrics or all metrics.
//
// Parameters:
//   - ctx: The context for metric collection. If it's a Gin context (*gin.Context),
//     API metrics are collected. Otherwise, non-API metrics are collected.
//   - name: Optional list of metric names to collect. If empty, all metrics in the
//     appropriate pool (API or non-API) are collected.
//
// The collection runs concurrently for all metrics using a semaphore to control
// concurrency and prevent resource exhaustion. Each metric's Collect function is
// called in a separate goroutine.
//
// Thread-safe: Can be called concurrently from multiple goroutines.
//
// Example:
//
//	// Collect all API metrics (from Gin context)
//	prm.CollectMetrics(ginContext)
//
//	// Collect all non-API metrics (from standard context)
//	prm.CollectMetrics(context.Background())
//
//	// Collect specific metrics
//	prm.CollectMetrics(ctx, "custom_counter", "custom_gauge")
func (m *prom) CollectMetrics(ctx context.Context, name ...string) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/prometheus/CollectMetrics", r)
		}
	}()

	var (
		ok bool
		s  libsem.Semaphore
	)

	s = libsem.New(ctx, 0, false)
	defer s.DeferMain()

	if _, ok = ctx.(*ginsdk.Context); ok {
		m.web.Walk(m.runCollect(ctx, s), name...)
	} else {
		m.oth.Walk(m.runCollect(ctx, s), name...)
	}

	_ = s.WaitAll()
}

// runCollect returns a walk function for concurrent metric collection.
// This is an internal method used by CollectMetrics to collect metrics in parallel.
//
// The returned function:
//   - Acquires a semaphore slot before starting collection
//   - Runs the metric's Collect function in a goroutine
//   - Updates the metric in the pool after collection
//   - Releases the semaphore slot when done
//
// Returns false if semaphore acquisition fails, which stops the walk.
func (m *prom) runCollect(ctx context.Context, sem libsem.Semaphore) prmpol.FuncWalk {
	return func(pool prmpol.MetricPool, key string, val libmet.Metric) bool {
		if e := sem.NewWorker(); e != nil {
			return false
		}

		go func() {
			defer func() {
				if r := recover(); r != nil {
					librun.RecoveryCaller("golib/prometheus/runCollect/"+key, r)
				}
			}()

			defer sem.DeferWorker()
			val.Collect(ctx)
			pool.Set(key, val)
		}()

		return true
	}
}
