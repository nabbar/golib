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
	"strconv"

	ginsdk "github.com/gin-gonic/gin"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
)

// MetricRequestURITotal creates a Counter metric that tracks the total number of requests
// broken down by URI path, HTTP method, and response status code.
//
// # Metric Type
//
// Counter - Increments for every request, with labels providing dimensional data.
//
// # Metric Name
//
// {prefix}_request_uri_total (e.g., "gin_request_uri_total")
//
// # Labels
//
//   - uri: The route pattern (e.g., "/api/users/:id")
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - code: HTTP status code (200, 404, 500, etc.)
//
// # Use Cases
//
//   - Track request distribution across endpoints
//   - Monitor success/error rates per endpoint
//   - Identify most frequently accessed routes
//   - Analyze endpoint-specific traffic patterns
//   - Alert on unusual status codes for specific endpoints
//
// # Dashboard Queries
//
//	// Top 5 most requested endpoints
//	topk(5, sum by(uri) (rate(gin_request_uri_total[5m])))
//
//	// 4xx error rate by endpoint
//	sum by(uri) (rate(gin_request_uri_total{code=~"4.."}[5m]))
//
//	// Success rate (2xx) percentage
//	sum(rate(gin_request_uri_total{code=~"2.."}[5m])) / sum(rate(gin_request_uri_total[5m])) * 100
//
//	// Requests per method
//	sum by(method) (rate(gin_request_uri_total[5m]))
//
// # Important Note
//
// Use c.FullPath() which returns the route pattern (e.g., "/users/:id") rather than
// c.Request.URL.Path which would return actual values (e.g., "/users/123"), preventing
// cardinality explosion in Prometheus.
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
//	metric := webmetrics.MetricRequestURITotal("myapp")
//	pool.Add(metric)
func MetricRequestURITotal(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_uri_total"), prmtps.Counter)
	met.SetDesc("Total number of requests by URI, method, and status code")
	met.AddLabel("uri", "method", "code")
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		}

		_ = m.Inc([]string{c.FullPath(), c.Request.Method, strconv.Itoa(c.Writer.Status())})
	})

	return met
}
