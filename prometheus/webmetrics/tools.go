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

// Package webmetrics provides Prometheus metrics specifically designed for Gin web servers.
//
// This package offers a comprehensive set of pre-configured metrics to monitor HTTP server
// performance, including request rates, latencies, response sizes, error rates, and more.
// All metrics are designed to work seamlessly with Gin's context and middleware system.
//
// # Architecture
//
// The package is built on top of github.com/nabbar/golib/prometheus/metrics and provides
// factory functions that create pre-configured Metric instances. Each metric is designed
// to extract relevant information from Gin's context during request processing.
//
// # Usage with Gin
//
// These metrics are typically registered with a Prometheus pool and collected via middleware:
//
//	import (
//	    "github.com/gin-gonic/gin"
//	    "github.com/nabbar/golib/prometheus"
//	    "github.com/nabbar/golib/prometheus/webmetrics"
//	)
//
//	func setupMetrics(prm prometheus.Prometheus) {
//	    pool := prm.GetPool()
//
//	    // Register standard web metrics
//	    pool.Add(webmetrics.MetricRequestTotal("myapp"))
//	    pool.Add(webmetrics.MetricRequestLatency("myapp", func() prometheus.Prometheus { return prm }))
//	    pool.Add(webmetrics.MetricRequestURITotal("myapp"))
//	    // ... add more metrics as needed
//	}
//
// # Metric Naming Convention
//
// All metrics follow the pattern: {prefix}_{metric_name}
// - Default prefix is "gin" if not specified
// - Custom prefix can be provided (e.g., "myapp" results in "myapp_request_total")
//
// # Bloom Filter for Unique Tracking
//
// Some metrics (like IP tracking) use a Bloom filter to efficiently track unique values
// without storing all historical data. The DefaultBloom filter is shared across metrics.
//
// For more details on individual metrics, see their respective function documentation.
package webmetrics

import (
	"fmt"

	prmblm "github.com/nabbar/golib/prometheus/bloom"
)

// DefaultBloom is a shared Bloom filter used by metrics that need to track unique values
// (such as unique IP addresses) without storing complete historical data.
//
// The Bloom filter provides probabilistic membership testing with minimal memory usage,
// making it ideal for high-cardinality data like IP addresses in production environments.
//
// This filter is safe for concurrent use and is shared across all webmetrics instances.
var DefaultBloom = prmblm.New()

// getDefaultPrefix constructs the full metric name by combining the provided prefix
// with the metric name.
//
// Parameters:
//   - pfx: The prefix for the metric (e.g., "myapp"). If empty, defaults to "gin"
//   - name: The metric name (e.g., "request_total")
//
// Returns:
//   - The fully qualified metric name (e.g., "gin_request_total" or "myapp_request_total")
//
// This function ensures consistent naming across all webmetrics, following Prometheus
// naming conventions with underscores separating components.
func getDefaultPrefix(pfx, name string) string {
	if len(pfx) < 1 {
		pfx = "gin"
	}

	return fmt.Sprintf("%s_%s", pfx, name)
}
