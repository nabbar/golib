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

	ginsdk "github.com/gin-gonic/gin"
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
)

// MetricResponseSizeByEndpoint creates a Histogram metric that tracks the distribution of
// response body sizes per endpoint, providing insights into payload characteristics.
//
// # Metric Type
//
// Histogram - Captures the distribution of response sizes with configurable buckets.
//
// # Metric Name
//
// {prefix}_response_size_bytes (e.g., "gin_response_size_bytes")
//
// # Labels
//
//   - uri: The route pattern (e.g., "/api/users/:id")
//   - method: HTTP method (GET, POST, etc.)
//
// # Use Cases
//
//   - Identify endpoints with large payloads
//   - Optimize response sizes for bandwidth efficiency
//   - Detect payload anomalies
//   - Capacity planning for bandwidth
//   - Cost analysis for data transfer
//   - API design decisions (pagination, field filtering)
//
// # Dashboard Queries
//
//	// Average response size by endpoint
//	rate(gin_response_size_bytes_sum[5m]) / rate(gin_response_size_bytes_count[5m])
//
//	// 95th percentile response size
//	histogram_quantile(0.95, sum by(uri, le) (rate(gin_response_size_bytes_bucket[5m])))
//
//	// Largest endpoints by response size
//	topk(5, rate(gin_response_size_bytes_sum[5m]) / rate(gin_response_size_bytes_count[5m]))
//
//	// Total bandwidth by endpoint (bytes/sec)
//	sum by(uri) (rate(gin_response_size_bytes_sum[5m]))
//
//	// Responses over 1MB
//	sum(rate(gin_response_size_bytes_bucket{le="1048576"}[5m])) / sum(rate(gin_response_size_bytes_count[5m]))
//
// # Histogram Buckets
//
// Uses size-appropriate buckets configured in the Prometheus instance, typically
// covering ranges from bytes to megabytes.
//
// # Parameters
//
//   - prefixName: The prefix for the metric name. If empty, defaults to "gin"
//   - fct: Function that returns the Prometheus instance (used to get size buckets)
//
// # Returns
//
//   - A configured Metric instance, or nil if fct is nil or returns nil
//
// # Example
//
//	pool := prometheus.GetPool()
//	metric := webmetrics.MetricResponseSizeByEndpoint("myapp", func() prometheus.Prometheus { return prm })
//	if metric != nil {
//	    pool.Add(metric)
//	}
func MetricResponseSizeByEndpoint(prefixName string, fct libprm.FuncGetPrometheus) prmmet.Metric {
	var (
		met prmmet.Metric
		prm libprm.Prometheus
	)

	if fct == nil {
		return nil
	} else if prm = fct(); prm == nil {
		return nil
	}

	met = prmmet.NewMetrics(getDefaultPrefix(prefixName, "response_size_bytes"), prmtps.Histogram)
	met.SetDesc("HTTP response size distribution in bytes by endpoint")
	met.AddLabel("uri", "method")

	// Use size buckets: 100B, 1KB, 10KB, 100KB, 1MB, 10MB, 100MB
	met.AddBuckets(100, 1024, 10240, 102400, 1048576, 10485760, 104857600)

	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		}

		size := c.Writer.Size()
		if size > 0 {
			_ = m.Observe([]string{c.FullPath(), c.Request.Method}, float64(size))
		}
	})

	return met
}
