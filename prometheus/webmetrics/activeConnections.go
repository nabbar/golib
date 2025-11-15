/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

package webmetrics

import (
	"context"

	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
)

// MetricActiveConnections creates a Gauge metric that tracks the current number of active
// HTTP connections being processed by the server.
//
// # Metric Type
//
// Gauge - Represents the current state (number of concurrent requests). Can go up or down.
//
// # Metric Name
//
// {prefix}_active_connections (e.g., "gin_active_connections")
//
// # Labels
//
// No labels. Provides a simple count of concurrent requests.
//
// # Use Cases
//
//   - Monitor server load and concurrency
//   - Detect traffic spikes
//   - Capacity planning
//   - Auto-scaling triggers
//   - Identify bottlenecks
//   - Set connection limit alerts
//
// # Dashboard Queries
//
//	// Current active connections
//	gin_active_connections
//
//	// Average active connections over 5 minutes
//	avg_over_time(gin_active_connections[5m])
//
//	// Peak active connections in last hour
//	max_over_time(gin_active_connections[1h])
//
//	// Connection saturation (assuming 1000 max connections)
//	gin_active_connections / 1000 * 100
//
// # Implementation Note
//
// This metric should be incremented at the start of request processing and
// decremented when the request completes. Use with middleware like:
//
//	func MetricsMiddleware(metric prmmet.Metric) gin.HandlerFunc {
//	    return func(c *gin.Context) {
//	        metric.Inc(nil)           // Increment on request start
//	        defer metric.Dec(nil)     // Decrement on request end
//	        c.Next()
//	    }
//	}
//
// # Important
//
// Unlike other webmetrics which collect during the request, this metric requires
// special handling in middleware to properly track the connection lifecycle.
// The SetCollect function is intentionally empty as the metric is manipulated
// directly by the middleware.
//
// # Parameters
//
//   - prefixName: The prefix for the metric name. If empty, defaults to "gin"
//
// # Returns
//
//   - A configured Metric instance ready to be added to a Prometheus pool
//
// # Example
//
//	pool := prometheus.GetPool()
//	metric := webmetrics.MetricActiveConnections("myapp")
//	pool.Add(metric)
//
//	// In your Gin setup:
//	router.Use(func(c *gin.Context) {
//	    metric.Inc(nil)
//	    defer metric.Dec(nil)
//	    c.Next()
//	})
func MetricActiveConnections(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "active_connections"), prmtps.Gauge)
	met.SetDesc("Current number of active HTTP connections being processed")

	// This metric is managed by middleware, not via collection
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		// Intentionally empty - metric is incremented/decremented by middleware
	})

	return met
}
