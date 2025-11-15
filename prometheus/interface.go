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

// Package prometheus provides production-ready Prometheus metrics integration for Go applications
// with first-class support for the Gin web framework.
//
// This package offers dynamic metric management, thread-safe operations, and efficient metric
// collection with minimal overhead. It includes pre-configured web metrics, metric pool management,
// and seamless Gin middleware integration.
//
// # Key Features
//
//   - Dynamic metric registration and management at runtime
//   - Thread-safe operations with atomic state management
//   - Built-in Gin middleware and route handlers
//   - Support for Counter, Gauge, Histogram, and Summary metrics
//   - Path exclusion for health checks and internal endpoints
//   - Pre-configured web metrics via webmetrics subpackage
//
// # Basic Usage
//
//	prm := prometheus.New(libctx.NewContext)
//	prm.SetSlowTime(5) // 5 seconds threshold for slow requests
//
//	// Register metrics
//	prm.AddMetric(true, webmetrics.MetricRequestTotal("myapp"))
//	prm.AddMetric(true, webmetrics.MetricRequestURITotal("myapp"))
//
//	// Setup Gin
//	router := gin.Default()
//	router.Use(gin.HandlerFunc(prm.MiddleWareGin))
//	router.GET("/metrics", prm.ExposeGin)
//
// # Thread Safety
//
// All operations are thread-safe and can be called concurrently from multiple goroutines.
// The package uses atomic operations, synchronized pools, and proper locking mechanisms.
//
// See the subpackages for specialized functionality:
//   - metrics: Metric definitions and sdkcol
//   - pool: Thread-safe metric pool management
//   - types: Core type definitions and interfaces
//   - webmetrics: Pre-configured production-ready web metrics
//   - bloom: Bloom filter for unique value tracking
package prometheus

import (
	"context"
	"sync/atomic"

	ginsdk "github.com/gin-gonic/gin"
	libatm "github.com/nabbar/golib/atomic"
	libmet "github.com/nabbar/golib/prometheus/metrics"
	prmpol "github.com/nabbar/golib/prometheus/pool"
	prmsdk "github.com/prometheus/client_golang/prometheus"
	sdkcol "github.com/prometheus/client_golang/prometheus/collectors"
	sdkpht "github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// DefaultSlowTime is the default threshold in seconds for marking requests as slow.
	// Requests taking longer than this threshold will be tracked by slow request metrics.
	DefaultSlowTime = int32(5)
)

// FuncGetPrometheus is a function type that returns a Prometheus instance.
// This is used by metrics that need access to the Prometheus instance for
// dynamic configuration (e.g., duration buckets, slow time threshold).
type FuncGetPrometheus func() Prometheus

// FuncCollectMetrics is a function type for triggering metric collection.
// It accepts a context and optional metric names to collect.
type FuncCollectMetrics func(ctx context.Context, name ...string)

// MetricsCollection defines the interface for managing Prometheus metrics.
// It provides methods to add, retrieve, delete, and list metrics in the pool.
type MetricsCollection interface {
	// GetMetric retrieves a metric instance by name from the metric pool.
	// Returns nil if the metric is not found.
	//
	// The method searches both API metrics (collected via Gin middleware) and
	// non-API metrics (collected manually or via ticker).
	//
	// Example:
	//  metric := prm.GetMetric("http_requests_total")
	//  if metric != nil {
	//      // Use metric
	//  }
	GetMetric(name string) libmet.Metric

	// AddMetric registers a metric instance in the Prometheus pool.
	//
	// Parameters:
	//   - isAPI: If true, the metric is collected via Gin middleware after each request.
	//            If false, the metric must be collected manually via CollectMetrics.
	//   - metric: The metric instance to register. Must have a unique name and valid configuration.
	//
	// Returns an error if:
	//   - The metric has no collect function
	//   - A histogram has no buckets configured
	//   - A summary has no objectives configured
	//   - The metric type is invalid
	//
	// Example:
	//  err := prm.AddMetric(true, webmetrics.MetricRequestTotal("myapp"))
	//  if err != nil {
	//      log.Fatal(err)
	//  }
	AddMetric(isAPI bool, metric libmet.Metric) error

	// DelMetric unregisters and removes a metric from the pool.
	// The metric will no longer be collected or exported.
	// It is safe to call with a non-existent metric name.
	//
	// Example:
	//  prm.DelMetric("temporary_metric")
	DelMetric(name string) error

	// ListMetric returns a slice of all registered metric names.
	// This includes both API metrics and non-API metrics.
	//
	// Example:
	//  names := prm.ListMetric()
	//  for _, name := range names {
	//      fmt.Println("Metric:", name)
	//  }
	ListMetric() []string

	// ClearMetric removes and unregisters all metrics from the specified pools.
	//
	// This method clears metrics from one or both metric pools (API and non-API).
	// Each metric is unregistered from Prometheus before being removed from the pool.
	//
	// Parameters:
	//   - Api: If true, clears all non-API metrics (manually collected metrics)
	//   - Web: If true, clears all API/Web metrics (Gin middleware collected metrics)
	//
	// Returns a slice of errors encountered during the clearing process.
	// An empty slice indicates all metrics were cleared successfully.
	//
	// Use cases:
	//   - Cleaning up before application shutdown
	//   - Resetting metrics between test runs
	//   - Removing all metrics when reconfiguring the monitoring setup
	//
	// Example:
	//  // Clear only API metrics
	//  errors := prm.ClearMetric(false, true)
	//
	//  // Clear all metrics
	//  errors := prm.ClearMetric(true, true)
	//  if len(errors) > 0 {
	//      log.Printf("Errors during clear: %v", errors)
	//  }
	ClearMetric(Api, Web bool) []error
}

// GinRoute defines the interface for Gin web framework integration.
// It provides middleware and route handlers for Prometheus metrics.
type GinRoute interface {
	// Expose handles the /metrics endpoint for Prometheus scraping.
	// It serves the metrics in Prometheus text format.
	//
	// This method accepts a generic context and will use ExposeGin if
	// the context is a Gin context.
	//
	// Example:
	//  router.GET("/metrics", func(c *gin.Context) {
	//      prm.Expose(c)
	//  })
	Expose(ctx context.Context)

	// ExposeGin is the Gin-specific metrics endpoint handler.
	// It serves metrics in Prometheus text format for scraping.
	//
	// Example:
	//  router.GET("/metrics", prm.ExposeGin)
	ExposeGin(c *ginsdk.Context)

	// MiddleWare is a generic middleware handler for metric collection.
	// It accepts a context and will use MiddleWareGin if the context is a Gin context.
	//
	// The middleware:
	//   - Sets the request start time in the context
	//   - Sets the request path (including query parameters)
	//   - Calls the next handler
	//   - Triggers metric collection after the handler completes
	//
	// Example:
	//  router.Use(func(c *gin.Context) {
	//      prm.MiddleWare(c)
	//  })
	MiddleWare(ctx context.Context)

	// MiddleWareGin is the Gin-specific middleware for Prometheus metric collection.
	//
	// The middleware:
	//   1. Captures request start time (if not already set)
	//   2. Captures request path with query parameters
	//   3. Calls c.Next() to execute the handler
	//   4. Triggers metric collection (unless path is excluded)
	//
	// Example:
	//  router.Use(gin.HandlerFunc(prm.MiddleWareGin))
	MiddleWareGin(c *ginsdk.Context)

	// ExcludePath adds path patterns to exclude from metric collection.
	// Paths starting with the given prefixes will not trigger metric collection.
	//
	// This is useful for:
	//   - Health check endpoints (/health, /ready)
	//   - Internal endpoints (/debug, /internal)
	//   - The metrics endpoint itself (/metrics)
	//
	// Leading slashes are added automatically if missing.
	//
	// Example:
	//  prm.ExcludePath("/health", "/ready", "/metrics")
	ExcludePath(startWith ...string)
}

// Collect defines the interface for metric collection operations.
type Collect interface {
	// Collect triggers collection of all metrics.
	// This is a convenience method that calls CollectMetrics with the provided context.
	libmet.Collect

	// CollectMetrics triggers metric collection for specified metrics or all metrics.
	//
	// Parameters:
	//   - ctx: The context for metric collection. If it's a Gin context, API metrics
	//          are collected. Otherwise, non-API metrics are collected.
	//   - name: Optional list of metric names to collect. If empty, all metrics are collected.
	//
	// The collection runs concurrently for all metrics using a semaphore to control
	// concurrency and prevent resource exhaustion.
	//
	// Example:
	//  // Collect all metrics
	//  prm.CollectMetrics(ctx)
	//
	//  // Collect specific metrics
	//  prm.CollectMetrics(ctx, "http_requests_total", "http_duration_seconds")
	CollectMetrics(ctx context.Context, name ...string)
}

// Prometheus is the main interface for Prometheus metrics integration.
// It combines metric management, Gin routing, collection, and configuration.
type Prometheus interface {
	GinRoute
	MetricsCollection
	Collect

	// SetSlowTime configures the slow request threshold in seconds.
	// Requests taking longer than this threshold will be tracked by slow request metrics.
	//
	// Default: 5 seconds (DefaultSlowTime)
	//
	// Example:
	//  prm.SetSlowTime(10) // Set to 10 seconds
	SetSlowTime(slowTime int32)

	// GetSlowTime returns the current slow request threshold in seconds.
	//
	// Example:
	//  threshold := prm.GetSlowTime()
	//  fmt.Printf("Slow time threshold: %d seconds\n", threshold)
	GetSlowTime() int32

	// SetDuration appends additional duration buckets for histogram metrics.
	// These buckets are used by request latency histograms.
	//
	// Default buckets: [0.1, 0.3, 1.2, 5, 10] (in seconds)
	//
	// Note: This appends to existing buckets rather than replacing them.
	//
	// Example:
	//  prm.SetDuration([]float64{0.05, 0.5, 2.5, 15})
	SetDuration(duration []float64)

	// GetDuration returns the current duration buckets for histogram metrics.
	//
	// Example:
	//  buckets := prm.GetDuration()
	//  fmt.Printf("Duration buckets: %v\n", buckets)
	GetDuration() []float64
}

// New creates a new Prometheus instance with default configuration.
//
// Parameters:
//   - ctx: A function that returns a context.Context. This is used for metric collection
//     and pool operations. Typically, you should use libctx.NewContext from the context package.
//
// Returns:
//   - A Prometheus instance ready for metric registration and collection
//
// Default configuration:
//   - Slow time threshold: 5 seconds
//   - Duration buckets: [0.1, 0.3, 1.2, 5, 10] seconds
//   - Two metric pools: one for API metrics (Gin), one for non-API metrics
//
// Example:
//
//	import libctx "github.com/nabbar/golib/context"
//
//	prm := prometheus.New(libctx.NewContext)
//	prm.SetSlowTime(10)
//	prm.AddMetric(true, webmetrics.MetricRequestTotal("myapp"))
//
//	router := gin.Default()
//	router.Use(gin.HandlerFunc(prm.MiddleWareGin))
//	router.GET("/metrics", prm.ExposeGin)
func New(ctx context.Context) Prometheus {
	r := prmsdk.NewRegistry()
	r.MustRegister(sdkcol.NewProcessCollector(sdkcol.ProcessCollectorOpts{}))
	r.MustRegister(sdkcol.NewGoCollector())

	p := &prom{
		exc: libatm.NewValue[[]string](),
		slw: new(atomic.Int32),
		rqd: libatm.NewValue[[]float64](),
		web: prmpol.New(ctx, r),
		oth: prmpol.New(ctx, r),
		hdl: sdkpht.HandlerFor(r, sdkpht.HandlerOpts{}),
		reg: r,
	}

	p.slw.Store(DefaultSlowTime)
	p.rqd.Store([]float64{0.1, 0.3, 1.2, 5, 10})

	return p
}
