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
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
)

// MetricRequestBody creates a Counter metric that tracks the cumulative size of all request
// bodies received by the server, measured in bytes.
//
// # Metric Type
//
// Counter - Accumulates the total bytes of all incoming request bodies.
//
// # Metric Name
//
// {prefix}_request_body_total (e.g., "gin_request_body_total")
//
// # Labels
//
// No labels. Provides an aggregate view of inbound data volume.
//
// # Use Cases
//
//   - Monitor bandwidth consumption for incoming requests
//   - Detect unusual data upload patterns
//   - Capacity planning for network and storage
//   - Track data ingestion rates
//   - Cost analysis for data transfer
//
// # Dashboard Queries
//
//	// Incoming data rate (MB/s)
//	rate(gin_request_body_total[5m]) / 1024 / 1024
//
//	// Total data received in last hour (GB)
//	increase(gin_request_body_total[1h]) / 1024 / 1024 / 1024
//
//	// Average request body size (bytes)
//	rate(gin_request_body_total[5m]) / rate(gin_request_total[5m])
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
//	metric := webmetrics.MetricRequestBody("myapp")
//	pool.Add(metric)
func MetricRequestBody(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_body_total"), prmtps.Counter)
	met.SetDesc("Cumulative size of all HTTP request bodies in bytes")
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		}

		if c.Request.ContentLength >= 0 {
			_ = m.Add(nil, float64(c.Request.ContentLength))
		}
	})

	return met
}
