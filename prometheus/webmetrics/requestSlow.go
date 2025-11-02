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
	"fmt"
	"strconv"
	"time"

	ginsdk "github.com/gin-gonic/gin"
	libprm "github.com/nabbar/golib/prometheus"
	prmmet "github.com/nabbar/golib/prometheus/metrics"
	prmtps "github.com/nabbar/golib/prometheus/types"
	librtr "github.com/nabbar/golib/router"
)

// MetricRequestSlow creates a Histogram metric that tracks requests exceeding a configured
// slow request threshold, providing insights into performance issues.
//
// # Metric Type
//
// Histogram - Although named "total", this is implemented as a histogram to track the
// distribution of slow requests. It increments only for requests exceeding the threshold.
//
// # Metric Name
//
// {prefix}_request_slow_total (e.g., "gin_request_slow_total")
//
// # Labels
//
//   - uri: The route pattern (e.g., "/api/users/:id")
//   - method: HTTP method (GET, POST, etc.)
//   - code: HTTP status code
//
// # Use Cases
//
//   - Identify consistently slow endpoints
//   - Monitor SLA/SLO compliance
//   - Detect performance regressions
//   - Prioritize optimization efforts
//   - Alert on degraded performance
//
// # Dashboard Queries
//
//	// Slow requests per second by endpoint
//	sum by(uri) (rate(gin_request_slow_total_count[5m]))
//
//	// Percentage of slow requests
//	sum(rate(gin_request_slow_total_count[5m])) / sum(rate(gin_request_total[5m])) * 100
//
//	// Endpoints with most slow requests
//	topk(5, sum by(uri) (rate(gin_request_slow_total_count[5m])))
//
//	// Slow requests by status code
//	sum by(code) (rate(gin_request_slow_total_count[5m]))
//
// # Slow Request Threshold
//
// The threshold for what constitutes a "slow" request is configured via the Prometheus
// instance's GetSlowTime() method. The threshold is evaluated dynamically on each request,
// allowing runtime adjustments.
//
// # Prerequisites
//
// Requires that the Gin middleware sets the GinContextStartUnixNanoTime value in the
// context before request processing begins.
//
// # Parameters
//
//   - prefixName: The prefix for the metric name. If empty, defaults to "gin"
//   - fct: Function that returns the Prometheus instance (used to get slow time threshold and duration buckets)
//
// # Returns
//
//   - A configured Metric instance, or nil if fct is nil or returns nil
//
// # Example
//
//	pool := prometheus.GetPool()
//	metric := webmetrics.MetricRequestSlow("myapp", func() prometheus.Prometheus { return prm })
//	if metric != nil {
//	    pool.Add(metric)
//	}
func MetricRequestSlow(prefixName string, fct libprm.FuncGetPrometheus) prmmet.Metric {
	var (
		met prmmet.Metric
		prm libprm.Prometheus
	)

	if fct == nil {
		return nil
	} else if prm = fct(); prm == nil {
		return nil
	}

	met = prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_slow_total"), prmtps.Histogram)
	met.SetDesc(fmt.Sprintf("Requests exceeding slow threshold (threshold: %ds)", prm.GetSlowTime()))
	met.AddLabel("uri", "method", "code")
	met.AddBuckets(prm.GetDuration()...)
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		} else if fct == nil {
			return
		} else if prm = fct(); prm == nil {
			return
		} else if start := c.GetInt64(librtr.GinContextStartUnixNanoTime); start == 0 {
			return
		} else if ts := time.Unix(0, start); ts.IsZero() {
			return
		} else if int32(time.Since(ts).Seconds()) > prm.GetSlowTime() {
			_ = m.Inc([]string{c.FullPath(), c.Request.Method, strconv.Itoa(c.Writer.Status())})
		}
	})

	return met
}
