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

// MetricRequestErrors creates a Counter metric that tracks HTTP errors (4xx and 5xx responses)
// broken down by endpoint, method, and status code.
//
// # Metric Type
//
// Counter - Increments for every error response (4xx or 5xx).
//
// # Metric Name
//
// {prefix}_request_errors_total (e.g., "gin_request_errors_total")
//
// # Labels
//
//   - uri: The route pattern (e.g., "/api/users/:id")
//   - method: HTTP method (GET, POST, PUT, DELETE, etc.)
//   - code: HTTP status code (400, 404, 500, 503, etc.)
//   - error_type: "client_error" (4xx) or "server_error" (5xx)
//
// # Use Cases
//
//   - Monitor application error rates
//   - Track client vs server errors separately
//   - Identify problematic endpoints
//   - Set up error rate alerts (SLO/SLA monitoring)
//   - Debug production issues
//   - Track error trends over time
//
// # Dashboard Queries
//
//	// Total error rate (errors/sec)
//	sum(rate(gin_request_errors_total[5m]))
//
//	// Error rate by error type
//	sum by(error_type) (rate(gin_request_errors_total[5m]))
//
//	// Top 5 endpoints with most errors
//	topk(5, sum by(uri) (rate(gin_request_errors_total[5m])))
//
//	// Client error (4xx) rate
//	sum(rate(gin_request_errors_total{error_type="client_error"}[5m]))
//
//	// Server error (5xx) rate
//	sum(rate(gin_request_errors_total{error_type="server_error"}[5m]))
//
//	// Error percentage
//	sum(rate(gin_request_errors_total[5m])) / sum(rate(gin_request_total[5m])) * 100
//
//	// Most common error codes
//	topk(10, sum by(code) (rate(gin_request_errors_total[5m])))
//
// # Important for SLOs
//
// This metric is crucial for Service Level Objectives (SLOs):
//   - Error budget tracking
//   - Availability calculations
//   - Incident detection and response
//
// Example SLO: "99.9% of requests should succeed" translates to:
//
//	(sum(rate(gin_request_total[30d])) - sum(rate(gin_request_errors_total[30d]))) /
//	sum(rate(gin_request_total[30d])) >= 0.999
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
//	metric := webmetrics.MetricRequestErrors("myapp")
//	pool.Add(metric)
func MetricRequestErrors(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_errors_total"), prmtps.Counter)
	met.SetDesc("Total number of HTTP error responses (4xx and 5xx)")
	met.AddLabel("uri", "method", "code", "error_type")
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		}

		statusCode := c.Writer.Status()

		// Only track errors (4xx and 5xx)
		if statusCode < 400 {
			return
		}

		errorType := "client_error" // 4xx
		if statusCode >= 500 {
			errorType = "server_error" // 5xx
		}

		_ = m.Inc([]string{c.FullPath(), c.Request.Method, strconv.Itoa(statusCode), errorType})
	})

	return met
}
