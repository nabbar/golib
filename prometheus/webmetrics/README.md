# WebMetrics - Prometheus Metrics for Gin Web Servers

[![Go Reference](https://pkg.go.dev/badge/github.com/nabbar/golib/prometheus/webmetrics.svg)](https://pkg.go.dev/github.com/nabbar/golib/prometheus/webmetrics)

A comprehensive collection of pre-configured Prometheus metrics specifically designed for monitoring Gin-based web servers. This package provides production-ready metrics for tracking request rates, latencies, errors, bandwidth, and more.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Metrics Catalog](#metrics-catalog)
- [Quick Start](#quick-start)
- [Detailed Usage](#detailed-usage)
- [Dashboard Examples](#dashboard-examples)
- [Best Practices](#best-practices)
- [Performance Considerations](#performance-considerations)

## Overview

### What is WebMetrics?

WebMetrics is a specialized metrics collection package that integrates seamlessly with Gin web framework and Prometheus. It provides:

- **11 production-ready metrics** covering all aspects of HTTP server monitoring
- **Zero-configuration defaults** with customizable prefixes
- **Label-based dimensionality** for detailed analysis
- **Efficient implementations** using Bloom filters and optimized collectors
- **Complete GoDoc documentation** for all metrics

### Key Features

✅ **Traffic Monitoring**: Track request volumes, unique clients, and bandwidth usage  
✅ **Performance Analysis**: Measure latencies, identify slow endpoints  
✅ **Error Tracking**: Monitor error rates, status codes, and failure patterns  
✅ **Resource Utilization**: Track active connections and payload sizes  
✅ **Dashboard Ready**: Includes PromQL queries for common visualizations  

## Architecture

### How It Works

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│ Gin Request │────────▶│  Middleware  │────────▶│   Handler   │
└─────────────┘         └──────────────┘         └─────────────┘
                               │                         │
                               │                         │
                               ▼                         ▼
                        ┌──────────────┐         ┌──────────────┐
                        │   Metrics    │◀────────│  Collection  │
                        │   Context    │         │   Trigger    │
                        └──────────────┘         └──────────────┘
                               │
                               ▼
                        ┌──────────────┐
                        │  Prometheus  │
                        │   Registry   │
                        └──────────────┘
```

### Collection Flow

1. **Request Arrival**: Gin middleware captures request start time
2. **Context Enhancement**: Request data stored in Gin context
3. **Handler Execution**: Your application logic runs
4. **Metric Collection**: Metrics are collected after handler completes
5. **Prometheus Export**: Metrics available at `/metrics` endpoint

### Data Model

Each metric follows Prometheus best practices:

- **Metric Name**: `{prefix}_{metric_name}` (e.g., `gin_request_total`)
- **Labels**: Dimensional data for filtering and grouping
- **Type**: Counter, Gauge, or Histogram based on use case
- **Description**: Human-readable description of what is measured

## Metrics Catalog

### Overview Table

| Metric | Type | Labels | Purpose | Cardinality |
|--------|------|--------|---------|-------------|
| `request_total` | Counter | None | Total request count | Low (1) |
| `request_uri_total` | Counter | uri, method, code | Requests per endpoint | Medium |
| `request_duration` | Histogram | uri | Request latency distribution | Medium |
| `request_slow_total` | Histogram | uri, method, code | Slow requests tracking | Medium |
| `request_errors_total` | Counter | uri, method, code, error_type | Error tracking | Medium |
| `request_body_total` | Counter | None | Incoming data volume | Low (1) |
| `response_body_total` | Counter | None | Outgoing data volume | Low (1) |
| `response_size_bytes` | Histogram | uri, method | Response size distribution | Medium |
| `request_ip_total` | Counter | None | Unique client IPs | Low (1) |
| `status_code_total` | Counter | code, class | Status code distribution | Low (~50) |
| `active_connections` | Gauge | None | Current concurrent requests | Low (1) |

### Cardinality Guidelines

- **Low**: 1-100 unique label combinations
- **Medium**: 100-10,000 unique label combinations
- **High**: >10,000 unique label combinations (⚠️ avoid)

## Quick Start

### Basic Setup

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/nabbar/golib/prometheus"
    "github.com/nabbar/golib/prometheus/webmetrics"
)

func main() {
    // Initialize Prometheus
    prm := prometheus.New(prometheus.Config{
        SlowTime: 5, // 5 seconds threshold for slow requests
    })
    
    // Get metric pool
    pool := prm.GetPool()
    
    // Register all standard metrics
    pool.Add(webmetrics.MetricRequestTotal("myapp"))
    pool.Add(webmetrics.MetricRequestURITotal("myapp"))
    pool.Add(webmetrics.MetricRequestLatency("myapp", func() prometheus.Prometheus { return prm }))
    pool.Add(webmetrics.MetricRequestErrors("myapp"))
    pool.Add(webmetrics.MetricStatusCodeTotal("myapp"))
    
    // Setup Gin
    router := gin.Default()
    
    // Add Prometheus middleware
    router.Use(prm.Middleware())
    
    // Expose metrics endpoint
    router.GET("/metrics", gin.WrapH(prm.Handler()))
    
    // Your routes
    router.GET("/api/users", handleUsers)
    
    router.Run(":8080")
}
```

### Minimal Setup (Essential Metrics Only)

```go
// Minimal monitoring setup with 4 key metrics
pool.Add(webmetrics.MetricRequestTotal("myapp"))           // Traffic volume
pool.Add(webmetrics.MetricRequestURITotal("myapp"))        // Per-endpoint traffic
pool.Add(webmetrics.MetricRequestErrors("myapp"))          // Error tracking
pool.Add(webmetrics.MetricStatusCodeTotal("myapp"))        // Status codes
```

### Complete Setup (All Metrics)

```go
// Comprehensive monitoring with all available metrics
func setupMetrics(prm prometheus.Prometheus, prefix string) {
    pool := prm.GetPool()
    
    getFct := func() prometheus.Prometheus { return prm }
    
    // Traffic metrics
    pool.Add(webmetrics.MetricRequestTotal(prefix))
    pool.Add(webmetrics.MetricRequestURITotal(prefix))
    pool.Add(webmetrics.MetricRequestIPTotal(prefix))
    
    // Performance metrics
    pool.Add(webmetrics.MetricRequestLatency(prefix, getFct))
    pool.Add(webmetrics.MetricRequestSlow(prefix, getFct))
    
    // Error metrics
    pool.Add(webmetrics.MetricRequestErrors(prefix))
    pool.Add(webmetrics.MetricStatusCodeTotal(prefix))
    
    // Resource metrics
    pool.Add(webmetrics.MetricRequestBody(prefix))
    pool.Add(webmetrics.MetricResponseBody(prefix))
    pool.Add(webmetrics.MetricResponseSizeByEndpoint(prefix, getFct))
    
    // Connection metrics (requires special middleware)
    activeConns := webmetrics.MetricActiveConnections(prefix)
    pool.Add(activeConns)
    
    // Note: Active connections requires custom middleware
    // See "Active Connections Middleware" section below
}
```

## Detailed Usage

### Active Connections Middleware

The `active_connections` metric requires special handling:

```go
func setupActiveConnectionsMiddleware(metric prmmet.Metric) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Increment on request start
        metric.Inc(nil)
        
        // Decrement on request end (using defer)
        defer metric.Dec(nil)
        
        // Process request
        c.Next()
    }
}

// Usage
activeConns := webmetrics.MetricActiveConnections("myapp")
pool.Add(activeConns)
router.Use(setupActiveConnectionsMiddleware(activeConns))
```

### Custom Metric Prefixes

Use custom prefixes to distinguish between multiple services:

```go
// Service A
setupMetrics(prmA, "service_a")  // → service_a_request_total

// Service B  
setupMetrics(prmB, "service_b")  // → service_b_request_total

// Default (uses "gin" prefix)
setupMetrics(prm, "")            // → gin_request_total
```

### Filtering by Labels

Metrics with labels support filtering in Prometheus:

```go
// Example: request_uri_total has labels: uri, method, code

// All POST requests
gin_request_uri_total{method="POST"}

// All 404 errors
gin_request_uri_total{code="404"}

// Specific endpoint
gin_request_uri_total{uri="/api/users/:id"}

// Combination
gin_request_uri_total{uri="/api/orders", method="POST", code="201"}
```

## Dashboard Examples

### Traffic Overview Dashboard

```promql
# Total Request Rate (req/sec)
rate(gin_request_total[5m])

# Request Rate by Endpoint
sum by(uri) (rate(gin_request_uri_total[5m]))

# Top 10 Busiest Endpoints
topk(10, sum by(uri) (rate(gin_request_uri_total[5m])))

# Request Distribution by HTTP Method
sum by(method) (rate(gin_request_uri_total[5m]))
```

### Performance Dashboard

```promql
# Average Request Latency
rate(gin_request_duration_sum[5m]) / rate(gin_request_duration_count[5m])

# 95th Percentile Latency by Endpoint
histogram_quantile(0.95, sum by(uri, le) (rate(gin_request_duration_bucket[5m])))

# 99th Percentile Latency (Tail Latency)
histogram_quantile(0.99, sum by(uri, le) (rate(gin_request_duration_bucket[5m])))

# Slowest Endpoints (p99 latency)
topk(5, histogram_quantile(0.99, sum by(uri, le) (rate(gin_request_duration_bucket[5m]))))

# Slow Request Rate
sum(rate(gin_request_slow_total_count[5m]))

# Percentage of Slow Requests
sum(rate(gin_request_slow_total_count[5m])) / sum(rate(gin_request_total[5m])) * 100
```

### Error Monitoring Dashboard

```promql
# Total Error Rate
sum(rate(gin_request_errors_total[5m]))

# Error Rate by Type (Client vs Server)
sum by(error_type) (rate(gin_request_errors_total[5m]))

# 4xx Error Rate
sum(rate(gin_request_errors_total{error_type="client_error"}[5m]))

# 5xx Error Rate
sum(rate(gin_request_errors_total{error_type="server_error"}[5m]))

# Error Percentage
sum(rate(gin_request_errors_total[5m])) / sum(rate(gin_request_total[5m])) * 100

# Top 5 Endpoints with Most Errors
topk(5, sum by(uri) (rate(gin_request_errors_total[5m])))

# Most Common Error Codes
topk(10, sum by(code) (rate(gin_status_code_total{class=~"4xx|5xx"}[5m])))
```

### Success Rate & Availability

```promql
# Success Rate (2xx responses)
sum(rate(gin_status_code_total{class="2xx"}[5m])) / sum(rate(gin_status_code_total[5m])) * 100

# Availability (non-5xx responses)
(sum(rate(gin_status_code_total[5m])) - sum(rate(gin_status_code_total{class="5xx"}[5m]))) / 
sum(rate(gin_status_code_total[5m])) * 100

# SLO Compliance (99.9% target)
(sum(rate(gin_request_total[30d])) - sum(rate(gin_request_errors_total[30d]))) / 
sum(rate(gin_request_total[30d])) >= 0.999
```

### Bandwidth & Resource Usage

```promql
# Incoming Bandwidth (MB/s)
rate(gin_request_body_total[5m]) / 1024 / 1024

# Outgoing Bandwidth (MB/s)
rate(gin_response_body_total[5m]) / 1024 / 1024

# Total Bandwidth (MB/s)
(rate(gin_request_body_total[5m]) + rate(gin_response_body_total[5m])) / 1024 / 1024

# Average Request Size (bytes)
rate(gin_request_body_total[5m]) / rate(gin_request_total[5m])

# Average Response Size (bytes)
rate(gin_response_body_total[5m]) / rate(gin_request_total[5m])

# Average Response Size by Endpoint
rate(gin_response_size_bytes_sum[5m]) / rate(gin_response_size_bytes_count[5m])

# Current Active Connections
gin_active_connections

# Peak Active Connections (last hour)
max_over_time(gin_active_connections[1h])
```

### Client Tracking

```promql
# Approximate Unique Clients
gin_request_ip_total

# New Unique IPs per Hour
rate(gin_request_ip_total[1h])

# Average Requests per Unique IP
rate(gin_request_total[5m]) / rate(gin_request_ip_total[5m])
```

## Best Practices

### 1. Choose Appropriate Metric Types

- **Counter**: For cumulative values that only increase (requests, errors, bytes)
- **Gauge**: For values that go up and down (active connections, queue size)
- **Histogram**: For distributions (latency, response size)

### 2. Manage Cardinality

⚠️ **High cardinality can impact Prometheus performance**

**Good** (Low cardinality):
```go
// Uses route pattern: /users/:id
uri = c.FullPath()  // → "/users/:id"
```

**Bad** (High cardinality):
```go
// Uses actual path with ID
uri = c.Request.URL.Path  // → "/users/123", "/users/456", ...
// This creates thousands of unique label combinations!
```

### 3. Use Metric Naming Conventions

Follow Prometheus naming best practices:

- Use `_total` suffix for counters: `request_total`, `errors_total`
- Use base units: `_bytes`, `_seconds`
- Use `_duration` for latencies
- Avoid redundant prefixes: `gin_gin_request` ❌

### 4. Set Up Alerts

Example alert rules:

```yaml
groups:
  - name: webserver_alerts
    rules:
      # High error rate
      - alert: HighErrorRate
        expr: |
          sum(rate(gin_request_errors_total[5m])) / 
          sum(rate(gin_request_total[5m])) > 0.05
        for: 5m
        annotations:
          summary: "Error rate above 5%"
      
      # High latency
      - alert: HighLatency
        expr: |
          histogram_quantile(0.99, 
            sum by(le) (rate(gin_request_duration_bucket[5m]))
          ) > 2
        for: 10m
        annotations:
          summary: "P99 latency above 2 seconds"
      
      # Service down
      - alert: ServiceDown
        expr: up{job="myapp"} == 0
        for: 1m
        annotations:
          summary: "Service is down"
```

### 5. Dashboard Organization

Organize dashboards by audience:

**Operations Dashboard**:
- Request rate
- Error rate
- Latency percentiles
- Active connections

**Development Dashboard**:
- Endpoint performance
- Error breakdown
- Slow queries
- Resource usage

**Business Dashboard**:
- Traffic trends
- Availability SLA
- User growth (unique IPs)
- Geographic distribution

### 6. Retention and Aggregation

Configure appropriate retention in Prometheus:

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  
# Short-term high-resolution data
storage:
  tsdb:
    retention.time: 15d
    
# Long-term aggregated data (using recording rules)
rule_files:
  - recording_rules.yml
```

Example recording rules:

```yaml
groups:
  - name: aggregate_metrics
    interval: 5m
    rules:
      # Pre-aggregate common queries
      - record: job:gin_request_rate:5m
        expr: sum(rate(gin_request_total[5m]))
      
      - record: job:gin_error_rate:5m
        expr: sum(rate(gin_request_errors_total[5m]))
      
      - record: job:gin_latency_p95:5m
        expr: histogram_quantile(0.95, sum by(le) (rate(gin_request_duration_bucket[5m])))
```

## Performance Considerations

### Memory Usage

Each metric instance consumes memory:

- **Counter/Gauge**: ~200 bytes + (labels * 50 bytes)
- **Histogram**: ~2KB + (buckets * labels * 100 bytes)

**Example calculation** for 100 endpoints:
- `request_uri_total` (3 labels, ~20 status codes, ~5 methods): ~10KB
- `request_duration` (1 label, 10 buckets): ~100KB
- Total for all metrics: ~500KB - 2MB

### CPU Impact

Metric collection overhead:
- **Counter increment**: ~50ns
- **Histogram observe**: ~200ns
- **Label lookup**: ~100ns

For 1000 req/s: ~0.5-1% CPU overhead

### Optimization Tips

1. **Limit histogram buckets**: Use 8-12 buckets maximum
2. **Reduce label cardinality**: Avoid user IDs, timestamps in labels
3. **Use recording rules**: Pre-aggregate common queries
4. **Efficient Bloom filters**: For unique IP tracking
5. **Batch metrics scraping**: Default 15s interval is sufficient

## Troubleshooting

### High Memory Usage

**Symptom**: Prometheus memory grows continuously

**Solution**:
1. Check label cardinality: `promtool tsdb analyze /data/prometheus`
2. Remove high-cardinality labels
3. Use recording rules to aggregate data
4. Reduce retention period

### Missing Metrics

**Symptom**: Metrics not appearing in Prometheus

**Checklist**:
- ✅ Metric registered with pool: `pool.Add(metric)`
- ✅ Middleware configured: `router.Use(prm.Middleware())`
- ✅ Metrics endpoint exposed: `router.GET("/metrics", ...)`
- ✅ Prometheus scraping configured
- ✅ Check Prometheus targets page

### Inaccurate Latency

**Symptom**: Latency metrics seem wrong

**Solution**:
- Ensure middleware sets start time in context
- Check `librtr.GinContextStartUnixNanoTime` is set
- Verify time synchronization (NTP)

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please ensure:
- GoDoc documentation for new metrics
- Examples in README
- Follow existing naming conventions
- Include dashboard query examples

## Support

- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/nabbar/golib/prometheus/webmetrics)
- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)

---

**Built with ❤️ for production monitoring**
