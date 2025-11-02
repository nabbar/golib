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

// MetricRequestTotal creates a Counter metric that tracks the total number of HTTP requests
// received by the server.
//
// # Metric Type
//
// Counter - A cumulative metric that only increases over time. Suitable for tracking
// total request counts.
//
// # Metric Name
//
// {prefix}_request_total (e.g., "gin_request_total" or "myapp_request_total")
//
// # Labels
//
// This metric has no labels. It provides a simple overall count of all requests,
// regardless of endpoint, method, or status code.
//
// # Use Cases
//
//   - Monitor overall server traffic
//   - Calculate request rate (requests/second) using rate() function
//   - Set up alerts for traffic spikes or drops
//   - Track cumulative request volume over time
//
// # Dashboard Queries
//
//	// Requests per second (5m average)
//	rate(gin_request_total[5m])
//
//	// Total requests in last hour
//	increase(gin_request_total[1h])
//
//	// Request rate comparison
//	rate(gin_request_total[5m]) / rate(gin_request_total[1h] offset 1h)
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
//	metric := webmetrics.MetricRequestTotal("myapp")
//	pool.Add(metric)
func MetricRequestTotal(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_total"), prmtps.Counter)
	met.SetDesc("Total number of HTTP requests received by the server")
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		_ = m.Inc(nil)
	})

	return met
}
