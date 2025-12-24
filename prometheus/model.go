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
	"net/http"
	"strings"
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	prmpol "github.com/nabbar/golib/prometheus/pool"
	prmsdk "github.com/prometheus/client_golang/prometheus"
)

// prom is the internal implementation of the Prometheus interface.
// It manages metric pools, configuration, and provides thread-safe operations
// for metric collection and management.
type prom struct {
	// exc stores path prefixes to exclude from metric collection
	// Protected by libatm.Value for thread-safe access
	exc libatm.Value[[]string]

	// slw is the slow request threshold in seconds
	// Uses atomic.Int32 for thread-safe reads and writes
	slw *atomic.Int32

	// rqd stores request duration buckets for histogram metrics
	// Protected by libatm.Value for thread-safe access
	rqd libatm.Value[[]float64]

	// web is the metric pool for Gin/API metrics
	// Metrics in this pool are collected after each HTTP request via middleware
	web prmpol.MetricPool

	// oth is the metric pool for non-API metrics
	// Metrics in this pool must be collected manually or via ticker
	oth prmpol.MetricPool

	// hdl is the HTTP handler for the /metrics endpoint
	// Provided by prometheus/promhttp package
	hdl http.Handler

	// reg is the Prometheus registerer for metric registration and unregistration
	// All metrics are registered to this instance
	reg prmsdk.Registerer
}

// SetSlowTime configures the slow request threshold in seconds.
// Requests taking longer than this threshold will be tracked by slow request metrics.
//
// This method is thread-safe and can be called concurrently.
//
// Default: 5 seconds (DefaultSlowTime)
//
// Example:
//
//	prm.SetSlowTime(10) // Mark requests slower than 10 seconds
func (m *prom) SetSlowTime(slowTime int32) {
	m.slw.Store(slowTime)
}

// GetSlowTime returns the current slow request threshold in seconds.
//
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	threshold := prm.GetSlowTime()
//	log.Printf("Slow request threshold: %d seconds", threshold)
func (m *prom) GetSlowTime() int32 {
	return m.slw.Load()
}

// SetDuration appends additional duration buckets for histogram metrics.
// These buckets are used by request latency histograms.
//
// This method is thread-safe and can be called concurrently.
//
// Default buckets: [0.1, 0.3, 1.2, 5, 10] (in seconds)
//
// Note: This method appends to existing buckets rather than replacing them.
// To replace buckets, you would need to create a new Prometheus instance.
//
// Example:
//
//	// Add additional buckets for very fast or very slow requests
//	prm.SetDuration([]float64{0.01, 0.05, 15, 30})
func (m *prom) SetDuration(duration []float64) {
	p := append(make([]float64, 0), m.rqd.Load()...)
	m.rqd.Store(append(p, duration...))
}

// GetDuration returns the current duration buckets for histogram metrics.
//
// This method is thread-safe and can be called concurrently.
//
// Example:
//
//	buckets := prm.GetDuration()
//	log.Printf("Current buckets: %v", buckets)
func (m *prom) GetDuration() []float64 {
	return m.rqd.Load()
}

// ExcludePath adds path patterns to exclude from metric collection.
// Paths starting with the given prefixes will not trigger metric collection.
//
// This method is thread-safe and can be called concurrently.
//
// Leading slashes are added automatically if missing.
// Empty strings are ignored.
//
// Use cases:
//   - Health check endpoints (/health, /ready)
//   - Internal endpoints (/debug, /internal)
//   - The metrics endpoint itself (/metrics)
//   - Static assets (/static, /assets)
//
// Example:
//
//	prm.ExcludePath("/health", "/ready", "/metrics")
//	prm.ExcludePath("favicon.ico") // Automatically becomes "/favicon.ico"
func (m *prom) ExcludePath(startWith ...string) {
	r := make([]string, 0, len(startWith))
	for _, s := range startWith {
		if len(s) < 1 {
			continue
		} else if !strings.HasPrefix(s, "/") {
			s = "/" + s
		}
		r = append(r, s)
	}

	p := append(make([]string, 0), m.exc.Load()...)
	m.exc.Store(append(p, r...))
}

// isExclude checks if a given path should be excluded from metric collection.
// Returns true if the path starts with any of the excluded prefixes.
//
// This is an internal method used by the middleware to determine if metrics
// should be collected for a request.
//
// The path is normalized by adding a leading slash if missing.
func (m *prom) isExclude(path string) bool {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	for _, p := range m.exc.Load() {
		if len(p) > 0 && strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}
