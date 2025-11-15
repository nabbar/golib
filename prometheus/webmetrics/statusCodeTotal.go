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

// MetricStatusCodeTotal creates a Counter metric that tracks the distribution of HTTP
// status codes, providing a high-level view of response patterns.
//
// # Metric Type
//
// Counter - Increments for every response, categorized by status code.
//
// # Metric Name
//
// {prefix}_status_code_total (e.g., "gin_status_code_total")
//
// # Labels
//
//   - code: HTTP status code (200, 201, 400, 404, 500, etc.)
//   - class: Status code class ("1xx", "2xx", "3xx", "4xx", "5xx")
//
// # Use Cases
//
//   - Monitor success rate (2xx responses)
//   - Track redirect patterns (3xx responses)
//   - Identify common errors (404, 500, etc.)
//   - Calculate availability metrics
//   - Detect unusual response patterns
//   - Create SLO dashboards
//
// # Dashboard Queries
//
//	// Success rate (2xx responses percentage)
//	sum(rate(gin_status_code_total{class="2xx"}[5m])) / sum(rate(gin_status_code_total[5m])) * 100
//
//	// Most common status codes
//	topk(10, sum by(code) (rate(gin_status_code_total[5m])))
//
//	// Error rate by class
//	sum by(class) (rate(gin_status_code_total{class=~"4xx|5xx"}[5m]))
//
//	// 404 Not Found rate
//	rate(gin_status_code_total{code="404"}[5m])
//
//	// 500 Internal Server Error rate
//	rate(gin_status_code_total{code="500"}[5m])
//
//	// Availability (non-5xx percentage)
//	(sum(rate(gin_status_code_total[5m])) - sum(rate(gin_status_code_total{class="5xx"}[5m]))) /
//	sum(rate(gin_status_code_total[5m])) * 100
//
// # Status Code Classes
//
//   - 1xx: Informational responses
//   - 2xx: Success
//   - 3xx: Redirection
//   - 4xx: Client errors
//   - 5xx: Server errors
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
//	metric := webmetrics.MetricStatusCodeTotal("myapp")
//	pool.Add(metric)
func MetricStatusCodeTotal(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "status_code_total"), prmtps.Counter)
	met.SetDesc("Total number of HTTP responses by status code")
	met.AddLabel("code", "class")
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		}

		statusCode := c.Writer.Status()
		codeStr := strconv.Itoa(statusCode)

		// Determine status code class
		class := "unknown"
		switch {
		case statusCode >= 100 && statusCode < 200:
			class = "1xx"
		case statusCode >= 200 && statusCode < 300:
			class = "2xx"
		case statusCode >= 300 && statusCode < 400:
			class = "3xx"
		case statusCode >= 400 && statusCode < 500:
			class = "4xx"
		case statusCode >= 500 && statusCode < 600:
			class = "5xx"
		}

		_ = m.Inc([]string{codeStr, class})
	})

	return met
}
