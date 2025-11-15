# Prometheus Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Production-ready Prometheus metrics integration for Go applications with Gin web framework support, dynamic metric management, and thread-safe operations.

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Installation](#installation)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Performance](#performance)
- [Use Cases](#use-cases)
- [Subpackages](#subpackages)
  - [metrics - Metric Definitions](#metrics-subpackage)
  - [pool - Metric Pool Management](#pool-subpackage)
  - [types - Core Type Definitions](#types-subpackage)
  - [webmetrics - Pre-configured Web Metrics](#webmetrics-subpackage)
  - [bloom - Bloom Filter Implementation](#bloom-subpackage)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This library provides production-ready Prometheus metrics integration for Go applications, with first-class support for the Gin web framework. It emphasizes dynamic metric management, thread safety, and efficient metric collection with minimal overhead.

### Design Philosophy

1. **Dynamic Management**: Register, unregister, and update metrics at runtime
2. **Thread-Safe**: Atomic operations and mutex protection for concurrent operations
3. **Gin-First**: Seamless integration with Gin middleware and routing
4. **Type-Safe**: Strong typing for all metric types (Counter, Gauge, Histogram, Summary)
5. **Modular**: Independent subpackages that work together seamlessly
6. **Performance**: Minimal overhead with efficient metric collection

---

## Key Features

- **Dynamic Metric Management**: Add, remove, and update metrics at runtime without restarts
- **Thread-Safe Operations**: Atomic state management and synchronized metric collection
- **Gin Integration**: Built-in middleware, route handlers, and context support
- **Multiple Metric Types**:
  - **Counter**: Cumulative values (requests, errors, bytes transferred)
  - **Gauge**: Values that can increase or decrease (active connections, queue size)
  - **Histogram**: Distributions with buckets (request latency, response size)
  - **Summary**: Streaming quantiles (precise percentiles)
- **Path Exclusion**: Exclude specific paths from monitoring (e.g., health checks)
- **Pre-configured Metrics**: 11 production-ready web metrics via `webmetrics` subpackage
- **Bloom Filter Support**: Efficient unique value tracking (e.g., unique IP addresses)
- **Standard Interfaces**: Compatible with standard `prometheus/client_golang`

---

## Installation

```bash
go get github.com/nabbar/golib/prometheus
```

---

## Architecture

### Package Structure

The package is organized into specialized subpackages:

```
prometheus/
├── metrics/             # Metric definitions and collectors
├── pool/                # Thread-safe metric pool management
├── types/               # Core type definitions and interfaces
├── webmetrics/          # Pre-configured web server metrics
├── bloom/               # Bloom filter for unique tracking
├── interface.go         # Main Prometheus interface
├── model.go             # Implementation of Prometheus interface
├── pool.go              # Metric pool operations
└── route.go             # Gin route handlers and middleware
```

### Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                  Prometheus Package                      │
│  New(), Middleware(), Expose(), Metric Management       │
└──────────────┬──────────────┬──────────────┬────────────┘
               │              │              │
      ┌────────▼─────┐  ┌────▼─────┐  ┌────▼──────┐
      │   metrics    │  │   pool   │  │   types   │
      │              │  │          │  │           │
      │  Counter     │  │ Add/Get  │  │ Metric    │
      │  Gauge       │  │ Del/List │  │ MetricType│
      │  Histogram   │  │ Walk     │  │ Register  │
      │  Summary     │  │          │  │           │
      └──────────────┘  └──────────┘  └───────────┘
               │
      ┌────────▼──────────┐
      │   webmetrics      │
      │                   │
      │  11 Pre-configured│
      │  Production Metrics│
      └───────────────────┘
```

| Component | Purpose | Thread-Safe | Coverage |
|-----------|---------|-------------|----------|
| **`metrics`** | Metric definition and collection | ✅ | 95.5% |
| **`pool`** | Metric pool management | ✅ | 72.5% |
| **`types`** | Type definitions and registration | ✅ | 100% |
| **`webmetrics`** | Pre-configured web metrics | ✅ | 0% (simple constructors) |
| **`bloom`** | Bloom filter for unique tracking | ✅ | 94.7% |
| **Root** | Main interface and middleware | ✅ | 90.9% |

### Metric Flow

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│ HTTP Request│────────▶│  Middleware  │────────▶│   Handler   │
└─────────────┘         └──────────────┘         └─────────────┘
                               │                         │
                               │ (set start time)        │
                               ▼                         │
                        ┌──────────────┐                 │
                        │ Gin Context  │                 │
                        │ - Start Time │                 │
                        │ - Path Info  │                 │
                        └──────────────┘                 │
                               │                         │
                               │◀────────────────────────┘
                               │ (after handler)
                               ▼
                        ┌──────────────┐
                        │   Collect    │
                        │   Metrics    │
                        └──────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │  Prometheus  │
                        │   Registry   │
                        └──────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │ /metrics     │
                        │  Endpoint    │
                        └──────────────┘
```

---

## Performance

### Memory Efficiency

- **Metric Storage**: ~200 bytes per counter/gauge, ~2KB per histogram (with 10 buckets)
- **Pool Overhead**: O(n) where n is number of registered metrics
- **Concurrent Safe**: Lock-free reads, synchronized writes

### Thread Safety

All operations are thread-safe through:

- **Atomic Operations**: `atomic.Int32` for scalar values
- **Value Protection**: `libatm.Value` for complex types (slices, maps)
- **Mutex-Free Reads**: Lock-free metric access
- **Concurrent Collection**: Parallel metric collection with semaphore control

### Throughput Benchmarks

| Operation | Overhead | Memory | Notes |
|-----------|----------|--------|-------|
| Counter Inc | ~50ns | O(1) | Atomic operation |
| Histogram Observe | ~200ns | O(1) | Bucket lookup |
| Metric Collection | ~500ns | O(1) | Per metric |
| Middleware | ~1µs | O(1) | Per request |

For 1000 req/s: ~0.1% CPU overhead

---

## Use Cases

This library is designed for scenarios requiring Prometheus monitoring:

**Web APIs**
- Monitor request rates, latencies, and error rates
- Track endpoint performance and SLA compliance
- Identify slow endpoints and bottlenecks
- Monitor active connections and resource usage

**Microservices**
- Service-level metrics with custom prefixes
- Cross-service performance comparison
- Error tracking and alerting
- Distributed tracing correlation

**Production Monitoring**
- Real-time dashboards with Grafana
- Automated alerting based on thresholds
- Capacity planning with historical data
- Incident investigation and debugging

**Performance Testing**
- Load testing metrics and analysis
- Latency distribution analysis
- Identify performance regressions
- Benchmark different configurations

---

## Quick Start

### Basic Setup

```go
package main

import (
    "github.com/gin-gonic/gin"
    libctx "github.com/nabbar/golib/context"
    "github.com/nabbar/golib/prometheus"
    "github.com/nabbar/golib/prometheus/webmetrics"
)

func main() {
    // Create Prometheus instance
    prm := prometheus.New(libctx.NewContext)
    
    // Configure slow request threshold (5 seconds)
    prm.SetSlowTime(5)
    
    // Register essential metrics
    prm.AddMetric(true, webmetrics.MetricRequestTotal("myapp"))
    prm.AddMetric(true, webmetrics.MetricRequestURITotal("myapp"))
    prm.AddMetric(true, webmetrics.MetricStatusCodeTotal("myapp"))
    
    // Setup Gin
    router := gin.Default()
    
    // Add Prometheus middleware (collects metrics)
    router.Use(gin.HandlerFunc(prm.MiddleWareGin))
    
    // Expose metrics endpoint
    router.GET("/metrics", prm.ExposeGin)
    
    // Your API routes
    router.GET("/api/users", handleUsers)
    router.POST("/api/orders", handleOrders)
    
    router.Run(":8080")
}

func handleUsers(c *gin.Context) {
    c.JSON(200, gin.H{"users": []string{"alice", "bob"}})
}

func handleOrders(c *gin.Context) {
    c.JSON(201, gin.H{"order_id": "12345"})
}
```

### Complete Setup with All Metrics

```go
package main

import (
    "github.com/gin-gonic/gin"
    libctx "github.com/nabbar/golib/context"
    "github.com/nabbar/golib/prometheus"
    "github.com/nabbar/golib/prometheus/webmetrics"
)

func setupMetrics(prm prometheus.Prometheus) {
    prefix := "myapp"
    getFct := func() prometheus.Prometheus { return prm }
    
    // Traffic metrics
    prm.AddMetric(true, webmetrics.MetricRequestTotal(prefix))
    prm.AddMetric(true, webmetrics.MetricRequestURITotal(prefix))
    prm.AddMetric(true, webmetrics.MetricRequestIPTotal(prefix))
    
    // Performance metrics
    prm.AddMetric(true, webmetrics.MetricRequestLatency(prefix, getFct))
    prm.AddMetric(true, webmetrics.MetricRequestSlow(prefix, getFct))
    
    // Error metrics
    prm.AddMetric(true, webmetrics.MetricRequestErrors(prefix))
    prm.AddMetric(true, webmetrics.MetricStatusCodeTotal(prefix))
    
    // Resource metrics
    prm.AddMetric(true, webmetrics.MetricRequestBody(prefix))
    prm.AddMetric(true, webmetrics.MetricResponseBody(prefix))
    prm.AddMetric(true, webmetrics.MetricResponseSizeByEndpoint(prefix, getFct))
    
    // Active connections (requires custom middleware)
    activeConns := webmetrics.MetricActiveConnections(prefix)
    prm.AddMetric(true, activeConns)
    
    // Configure request duration buckets
    prm.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})
}

func main() {
    prm := prometheus.New(libctx.NewContext)
    prm.SetSlowTime(5)
    
    setupMetrics(prm)
    
    // Exclude health check endpoints from metrics
    prm.ExcludePath("/health", "/ready", "/metrics")
    
    router := gin.Default()
    router.Use(gin.HandlerFunc(prm.MiddleWareGin))
    router.GET("/metrics", prm.ExposeGin)
    
    // Health endpoints (excluded from metrics)
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    // API routes (monitored)
    api := router.Group("/api/v1")
    {
        api.GET("/users", handleUsers)
        api.POST("/users", createUser)
        api.GET("/users/:id", getUser)
    }
    
    router.Run(":8080")
}
```

### Custom Metric Registration

```go
package main

import (
    "context"
    libctx "github.com/nabbar/golib/context"
    "github.com/nabbar/golib/prometheus"
    prmmet "github.com/nabbar/golib/prometheus/metrics"
    prmtps "github.com/nabbar/golib/prometheus/types"
)

func main() {
    prm := prometheus.New(libctx.NewContext)
    
    // Create custom counter
    counter := prmmet.NewMetrics("custom_events_total", prmtps.Counter)
    counter.SetDesc("Total number of custom events")
    counter.AddLabel("event_type", "severity")
    counter.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
        // Collection logic here
        metric.Inc(map[string]string{
            "event_type": "login",
            "severity": "info",
        })
    })
    
    // Create custom histogram
    histogram := prmmet.NewMetrics("task_duration_seconds", prmtps.Histogram)
    histogram.SetDesc("Task processing duration in seconds")
    histogram.AddLabel("task_type")
    histogram.AddBuckets(0.1, 0.5, 1.0, 2.5, 5.0, 10.0)
    histogram.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
        // Collection logic here
        metric.Observe(map[string]string{"task_type": "import"}, 1.5)
    })
    
    // Register custom metrics
    prm.AddMetric(false, counter)    // false = not API metric
    prm.AddMetric(false, histogram)
    
    // Manually trigger collection
    prm.CollectMetrics(context.Background())
}
```

---

## Subpackages

### `metrics` Subpackage

Metric definitions with collection logic and Prometheus collector integration.

**Features**
- Type-safe metric creation (Counter, Gauge, Histogram, Summary)
- Label support for dimensional metrics
- Collection function callbacks
- Prometheus collector registration
- Helper methods (Inc, Dec, Set, Observe)

**API Example**

```go
import prmmet "github.com/nabbar/golib/prometheus/metrics"
import prmtps "github.com/nabbar/golib/prometheus/types"

// Create counter
counter := prmmet.NewMetrics("http_requests_total", prmtps.Counter)
counter.SetDesc("Total HTTP requests")
counter.AddLabel("method", "status")
counter.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
    // Increment counter
    metric.Inc(map[string]string{"method": "GET", "status": "200"})
})

// Create histogram
latency := prmmet.NewMetrics("http_duration_seconds", prmtps.Histogram)
latency.SetDesc("HTTP request latency")
latency.AddLabel("endpoint")
latency.AddBuckets(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)
latency.SetCollect(func(ctx context.Context, metric prmmet.Metric) {
    // Observe duration
    metric.Observe(map[string]string{"endpoint": "/api/users"}, 0.123)
})
```

**Metric Types**
- **Counter**: Monotonically increasing values (use `Inc`, `Add`)
- **Gauge**: Values that go up/down (use `Inc`, `Dec`, `Set`, `Add`, `Sub`)
- **Histogram**: Distributions with buckets (use `Observe`)
- **Summary**: Streaming quantiles (use `Observe`)

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/prometheus/metrics) for complete API.

---

### `pool` Subpackage

Thread-safe metric pool management with efficient storage and retrieval.

**Features**
- Add/Get/Delete operations
- List all registered metrics
- Walk through metrics with callbacks
- Thread-safe with sync.Map
- Context-aware operations

**API Example**

```go
import prmpool "github.com/nabbar/golib/prometheus/pool"
import libctx "github.com/nabbar/golib/context"

// Create pool
pool := prmpool.New(libctx.NewContext)

// Add metric
err := pool.Add(metric)

// Retrieve metric
m := pool.Get("metric_name")

// List all metrics
names := pool.List()

// Walk through metrics
pool.Walk(func(p prmpool.MetricPool, key string, val prmmet.Metric) bool {
    fmt.Printf("Metric: %s\n", key)
    return true // continue walking
})

// Delete metric
pool.Del("metric_name")
```

**Thread Safety**: All operations are safe for concurrent use

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/prometheus/pool) for complete API.

---

### `types` Subpackage

Core type definitions, interfaces, and Prometheus collector registration.

**Features**
- `Metric` interface definition
- `MetricType` enum (Counter, Gauge, Histogram, Summary)
- Prometheus collector creation
- Type validation
- GoDoc-documented examples

**API Example**

```go
import prmtps "github.com/nabbar/golib/prometheus/types"

// Metric types
var (
    None      prmtps.MetricType = 0  // Invalid/uninitialized
    Counter   prmtps.MetricType = 1  // Cumulative
    Gauge     prmtps.MetricType = 2  // Up/down
    Histogram prmtps.MetricType = 3  // Distribution
    Summary   prmtps.MetricType = 4  // Quantiles
)

// Register metric with Prometheus
collector, err := metricType.Register(metric)
if err != nil {
    log.Fatal(err)
}
prometheus.MustRegister(collector)
```

**Type-Specific Requirements**
- **Counter/Gauge**: No special requirements
- **Histogram**: Must call `AddBuckets()` before registration
- **Summary**: Must call `SetObjectives()` before registration

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/prometheus/types) for complete API.

---

### `webmetrics` Subpackage

Pre-configured production-ready metrics for web servers.

**11 Production Metrics**
1. `request_total` - Total request count
2. `request_uri_total` - Requests per endpoint
3. `request_duration` - Request latency distribution
4. `request_slow_total` - Slow requests tracking
5. `request_errors_total` - Error tracking
6. `request_body_total` - Incoming bandwidth
7. `response_body_total` - Outgoing bandwidth
8. `response_size_bytes` - Response size distribution
9. `request_ip_total` - Unique client IPs
10. `status_code_total` - Status code distribution
11. `active_connections` - Current active requests

**Quick Setup**

```go
import "github.com/nabbar/golib/prometheus/webmetrics"

prefix := "myapp"
getFct := func() prometheus.Prometheus { return prm }

// Essential metrics
prm.AddMetric(true, webmetrics.MetricRequestTotal(prefix))
prm.AddMetric(true, webmetrics.MetricRequestURITotal(prefix))
prm.AddMetric(true, webmetrics.MetricStatusCodeTotal(prefix))
prm.AddMetric(true, webmetrics.MetricRequestErrors(prefix))

// Performance metrics
prm.AddMetric(true, webmetrics.MetricRequestLatency(prefix, getFct))
prm.AddMetric(true, webmetrics.MetricRequestSlow(prefix, getFct))

// Bandwidth metrics
prm.AddMetric(true, webmetrics.MetricRequestBody(prefix))
prm.AddMetric(true, webmetrics.MetricResponseBody(prefix))
prm.AddMetric(true, webmetrics.MetricResponseSizeByEndpoint(prefix, getFct))

// Client tracking
prm.AddMetric(true, webmetrics.MetricRequestIPTotal(prefix))

// Active connections (requires special middleware)
prm.AddMetric(true, webmetrics.MetricActiveConnections(prefix))
```

See [webmetrics/README.md](webmetrics/README.md) for detailed documentation, PromQL queries, and dashboard examples.

---

### `bloom` Subpackage

Space-efficient probabilistic data structure for unique value tracking.

**Features**
- Fast membership testing
- Configurable false positive rate
- Memory efficient (uses bit array)
- Thread-safe operations
- No false negatives

**Use Case**: Track unique IP addresses without storing all IPs

**API Example**

```go
import "github.com/nabbar/golib/prometheus/bloom"

// Create bloom filter (1000 items, 1% false positive rate)
bf := bloom.New(1000, 0.01)

// Add items
bf.Add([]byte("192.168.1.1"))
bf.Add([]byte("10.0.0.1"))

// Check membership
if bf.Test([]byte("192.168.1.1")) {
    // Probably present (might be false positive)
}

// Get approximate count
count := bf.Count()
```

**Performance**: O(k) where k is number of hash functions (~3-7 for 1% FPR)

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/prometheus/bloom) for complete API.

---

## Best Practices

**Choose Appropriate Metric Types**
```go
// ✅ Good: Counter for cumulative values
requests := prmmet.NewMetrics("requests_total", prmtps.Counter)

// ✅ Good: Gauge for current state
activeConns := prmmet.NewMetrics("active_connections", prmtps.Gauge)

// ✅ Good: Histogram for distributions
latency := prmmet.NewMetrics("request_duration_seconds", prmtps.Histogram)
latency.AddBuckets(0.1, 0.3, 1.2, 5, 10)

// ❌ Bad: Gauge for cumulative values (will reset)
requests := prmmet.NewMetrics("requests_total", prmtps.Gauge)
```

**Manage Label Cardinality**
```go
// ✅ Good: Low cardinality (bounded label values)
counter.AddLabel("method", "status", "endpoint")  // ~5 × ~50 × ~100 = 25K combinations

// ❌ Bad: High cardinality (unbounded label values)
counter.AddLabel("user_id", "timestamp")  // Millions of combinations!
```

**Use Path Patterns, Not Actual Paths**
```go
// ✅ Good: Use route pattern
endpoint := c.FullPath()  // "/users/:id"

// ❌ Bad: Use actual path
endpoint := c.Request.URL.Path  // "/users/12345" (creates thousands of series)
```

**Exclude Health Checks**
```go
// Prevent metrics pollution from health checks
prm.ExcludePath("/health", "/ready", "/metrics", "/favicon.ico")
```

**Set Appropriate Histogram Buckets**
```go
// ✅ Good: Buckets appropriate for your use case
// For API latency (milliseconds to seconds)
latency.AddBuckets(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10)

// For file sizes (bytes to megabytes)
size.AddBuckets(1024, 10240, 102400, 1048576, 10485760)

// ❌ Bad: Too many buckets (memory waste)
latency.AddBuckets(0.001, 0.002, 0.003, ..., 100)  // 100+ buckets!
```

**Always Handle Errors**
```go
// ✅ Good: Check errors
if err := prm.AddMetric(true, metric); err != nil {
    log.Printf("Failed to add metric: %v", err)
}

// ❌ Bad: Ignore errors
prm.AddMetric(true, metric)
```

**Clean Up Metrics**
```go
// Remove unused metrics to free memory
defer prm.DelMetric("temporary_metric")
```

---

## Testing

**Test Suite**: 426 specs across 5 packages with 91.6% overall coverage

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...
```

**Test Results**

```
prometheus/                137 specs    90.9% coverage   1.38s
prometheus/bloom/           24 specs    94.7% coverage   0.13s
prometheus/metrics/        179 specs    95.5% coverage   0.03s
prometheus/pool/            74 specs    72.5% coverage   0.02s
prometheus/types/           36 specs   100.0% coverage   0.01s
```

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Comprehensive edge case testing
- ✅ Integration tests with Gin

See [TESTING.md](TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

**Code Contributions**
- Do not use AI to generate package implementation code
- AI may assist with tests, documentation, and bug fixing
- All contributions must pass `go test -race`
- Maintain or improve test coverage (≥80%)
- Follow existing code style and patterns

**Documentation**
- Update README.md for new features
- Add GoDoc comments for all public APIs
- Provide examples for common use cases
- Keep TESTING.md synchronized with test changes

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex test scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results and coverage
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Metric Features**
- Custom aggregation functions
- Metric templating and cloning
- Metric groups for bulk operations
- Metric versioning for breaking changes
- Automatic metric expiration

**Performance**
- Metric caching layer
- Batch metric collection
- Async metric updates
- Memory-optimized storage
- Custom metric serialization

**Integration**
- OpenTelemetry bridge
- StatsD protocol support
- InfluxDB line protocol
- CloudWatch metrics export
- Datadog integration

**Observability**
- Built-in tracing integration
- Metric correlation with logs
- Anomaly detection
- Automatic alerting rules
- Performance profiling

**Developer Experience**
- Metric discovery API
- Real-time metric preview
- Metric validation CLI
- Dashboard templates
- Grafana integration

Suggestions and contributions are welcome via GitHub issues.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/prometheus)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Web Metrics**: [webmetrics/README.md](webmetrics/README.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Prometheus**: [prometheus.io](https://prometheus.io/)
- **Gin Framework**: [gin-gonic.com](https://gin-gonic.com/)
