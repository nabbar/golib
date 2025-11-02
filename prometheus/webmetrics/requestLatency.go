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
	"time"

	ginsdk "github.com/gin-gonic/gin"
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	librtr "github.com/nabbar/golib/router"
)

// MetricRequestLatency creates a Histogram metric that measures the time taken to process
// HTTP requests, providing detailed latency distribution per endpoint.
//
// # Metric Type
//
// Histogram - Captures the distribution of request durations with configurable buckets.
// Provides _sum, _count, and _bucket metrics for calculating percentiles and averages.
//
// # Metric Name
//
// {prefix}_request_duration (e.g., "gin_request_duration")
//
// # Labels
//
//   - uri: The route pattern (e.g., "/api/users/:id")
//
// # Use Cases
//
//   - Monitor request latency distribution
//   - Calculate latency percentiles (p50, p90, p95, p99)
//   - Identify slow endpoints requiring optimization
//   - Set SLO/SLA targets for response times
//   - Detect performance degradation over time
//
// # Dashboard Queries
//
//	// 95th percentile latency by endpoint
//	histogram_quantile(0.95, sum by(uri, le) (rate(gin_request_duration_bucket[5m])))
//
//	// Average request duration
//	rate(gin_request_duration_sum[5m]) / rate(gin_request_duration_count[5m])
//
//	// Slowest endpoints (p99 latency)
//	topk(5, histogram_quantile(0.99, sum by(uri, le) (rate(gin_request_duration_bucket[5m]))))
//
//	// Requests exceeding 500ms SLA
//	sum(rate(gin_request_duration_bucket{le="0.5"}[5m])) / sum(rate(gin_request_duration_count[5m]))
//
// # Histogram Buckets
//
// Uses the duration buckets configured in the Prometheus instance (typically exponential
// buckets covering microseconds to seconds).
//
// # Prerequisites
//
// Requires that the Gin middleware sets the GinContextStartUnixNanoTime value in the
// context before request processing begins.
//
// # Parameters
//
//   - prefixName: The prefix for the metric name. If empty, defaults to "gin"
//   - fct: Function that returns the Prometheus instance (used to get duration buckets)
//
// # Returns
//
//   - A configured Metric instance, or nil if fct is nil or returns nil
//
// # Example
//
//	pool := prometheus.GetPool()
//	metric := webmetrics.MetricRequestLatency("myapp", func() prometheus.Prometheus { return prm })
//	if metric != nil {
//	    pool.Add(metric)
//	}
func MetricRequestLatency(prefixName string, fct libprm.FuncGetPrometheus) prmmet.Metric {
	var (
		met prmmet.Metric
		prm libprm.Prometheus
	)

	if fct == nil {
		return nil
	} else if prm = fct(); prm == nil {
		return nil
	}

	met = prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_duration"), prmtps.Histogram)
	met.SetDesc("HTTP request latency distribution in seconds")
	met.AddLabel("uri")
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		} else if start := c.GetInt64(librtr.GinContextStartUnixNanoTime); start == 0 {
			return
		} else if ts := time.Unix(0, start); ts.IsZero() {
			return
		} else {
			_ = m.Observe([]string{c.FullPath()}, time.Since(ts).Seconds())
		}
	})

	return met
}
