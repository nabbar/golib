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

// MetricRequestIPTotal creates a Counter metric that tracks the approximate number of unique
// client IP addresses that have accessed the server.
//
// # Metric Type
//
// Counter - Increments when a new unique IP address is detected.
//
// # Metric Name
//
// {prefix}_request_ip_total (e.g., "gin_request_ip_total")
//
// # Labels
//
// No labels. Provides a count of unique IP addresses.
//
// # Use Cases
//
//   - Track unique visitor/client count
//   - Monitor user base growth
//   - Detect unusual traffic patterns (bot attacks, DDoS)
//   - Estimate active user base
//   - Geographic distribution analysis (when combined with GeoIP)
//
// # Dashboard Queries
//
//	// Approximate unique IPs seen
//	gin_request_ip_total
//
//	// Rate of new unique IPs
//	rate(gin_request_ip_total[1h])
//
//	// Average requests per unique IP
//	rate(gin_request_total[5m]) / rate(gin_request_ip_total[5m])
//
// # Important Notes
//
// This metric uses a Bloom filter (DefaultBloom) for efficient unique IP tracking:
//   - Space-efficient: Does not store actual IP addresses
//   - Probabilistic: May have false positives (IP counted twice) but never false negatives
//   - Memory-bounded: Fixed memory usage regardless of unique IP count
//   - Privacy-friendly: IP addresses are hashed, not stored
//
// The Bloom filter trades perfect accuracy for memory efficiency, making it suitable
// for high-traffic production environments where tracking millions of unique IPs
// would otherwise be prohibitively expensive.
//
// # Accuracy
//
// The false positive rate depends on the Bloom filter configuration but is typically
// very low (<1%) for reasonable traffic volumes.
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
//	metric := webmetrics.MetricRequestIPTotal("myapp")
//	pool.Add(metric)
func MetricRequestIPTotal(prefixName string) prmmet.Metric {
	met := prmmet.NewMetrics(getDefaultPrefix(prefixName, "request_ip_total"), prmtps.Counter)
	met.SetDesc("Approximate count of unique client IP addresses")
	met.SetCollect(func(ctx context.Context, m prmmet.Metric) {
		var (
			c  *ginsdk.Context
			ok bool
		)

		if c, ok = ctx.(*ginsdk.Context); !ok {
			return
		}

		clientIP := c.ClientIP()

		if !DefaultBloom.Contains(m.GetName(), clientIP) {
			DefaultBloom.Add(m.GetName(), clientIP)
			_ = m.Inc(nil)
		}
	})

	return met
}
