# Status Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Comprehensive health check and status monitoring system for HTTP APIs with flexible control modes, caching, and multi-format output support.

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
  - [control - Validation Modes](#control-subpackage)
  - [mandatory - Component Groups](#mandatory-subpackage)
  - [listmandatory - Group Management](#listmandatory-subpackage)
- [Configuration](#configuration)
- [HTTP API](#http-api)
- [Best Practices](#best-practices)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

This package provides production-ready health check and status monitoring for Go HTTP APIs. It integrates with the Gin web framework to expose status endpoints that aggregate component health checks with configurable validation strategies.

### Design Philosophy

1. **Flexible Validation**: Multiple control modes (Must, Should, AnyOf, Quorum) for different component criticality levels
2. **Performance-Focused**: Built-in caching with atomic operations reduces overhead for frequent health checks
3. **Thread-Safe**: Full concurrency support with proper synchronization primitives
4. **Multi-Format**: JSON and plain text output for different consumption scenarios
5. **Integration-Ready**: Works seamlessly with github.com/nabbar/golib/monitor for component monitoring

---

## Key Features

- **Component Monitoring**: Aggregate health from multiple monitored components
- **Control Modes**: Flexible validation strategies (Ignore, Should, Must, AnyOf, Quorum)
- **Caching**: 3-second default cache with atomic operations for high-frequency checks
- **Multi-Format Output**: JSON (default) and plain text via query parameters or headers
- **Verbosity Control**: Short (status only) or full (with component details) responses
- **Thread-Safe**: Atomic operations and mutex protection for concurrent access
- **Configurable HTTP Codes**: Customize return codes for OK (200), Warn (207), KO (500) states
- **Version Tracking**: Include application name, version, and build information
- **Gin Integration**: Drop-in middleware for Gin web framework

---

## Installation

```bash
go get github.com/nabbar/golib/status
```

---

## Architecture

### Package Structure

The package is organized into focused subpackages with clear responsibilities:

```
status/
├── control/             # Validation mode definitions
│   ├── interface.go     # Mode type and constants
│   ├── encode.go        # Marshaling support (JSON, YAML, TOML, CBOR)
│   └── format.go        # String parsing and formatting
├── mandatory/           # Single component group management
│   ├── interface.go     # Mandatory interface
│   └── model.go         # Thread-safe implementation
├── listmandatory/       # Multiple group management
│   ├── interface.go     # ListMandatory interface
│   └── model.go         # Collection handling
├── interface.go         # Main Status interface
├── model.go             # Core implementation
├── config.go            # Configuration structures
├── cache.go             # Status caching
├── route.go             # HTTP endpoint handler
└── encode.go            # Response marshaling
```

### Component Overview

```
┌──────────────────────────────────────────────────────┐
│                  Status Package                      │
│  HTTP Endpoint + Component Health Aggregation       │
└──────────────┬────────────┬──────────────┬──────────┘
               │            │              │
      ┌────────▼───┐  ┌────▼─────┐  ┌────▼────────┐
      │  control   │  │mandatory │  │listmandatory│
      │            │  │          │  │             │
      │ Validation │  │  Group   │  │ Collection  │
      │   Modes    │  │ Manager  │  │  Manager    │
      └────────────┘  └──────────┘  └─────────────┘
               │            │              │
               └────────────┴──────────────┘
                            │
                  ┌─────────▼──────────┐
                  │  monitor/types     │
                  │  Component Health  │
                  └────────────────────┘
```

| Component | Purpose | Thread-Safe | Marshaling |
|-----------|---------|-------------|------------|
| **`status`** | Main coordinator, HTTP endpoint | ✅ | JSON, Text |
| **`control`** | Validation mode definitions | ✅ | JSON, YAML, TOML, CBOR |
| **`mandatory`** | Component group with mode | ✅ | N/A |
| **`listmandatory`** | Multiple group management | ✅ | N/A |

### Data Flow

```
[HTTP Request]
      ↓
[MiddleWare] → Parse query params/headers
      ↓
[getStatus] → Walk monitor pool
      ↓
[Control Mode Logic]
  ├─ Ignore: Skip component
  ├─ Should: Warn only (no failure)
  ├─ Must: Must be healthy
  ├─ AnyOf: At least one healthy
  └─ Quorum: Majority (>50%) healthy
      ↓
[Encode] → JSON or Text format
      ↓
[HTTP Response] → With status code
```

---

## Performance

### Caching

The package implements efficient caching to reduce overhead:

- **Default Cache**: 3 seconds (configurable)
- **Atomic Operations**: Lock-free reads via `atomic.Value`
- **Thread-Safe**: Safe for concurrent health checks
- **Cache Methods**:
  - `IsCacheHealthy()`: Check >= Warn (accepts warnings)
  - `IsCacheStrictlyHealthy()`: Check == OK (strict)

**Performance Impact**: Cached checks complete in <10ns (no component walking)

### Benchmark Results

```
BenchmarkNew-12                    13,963,723      77.33 ns/op      72 B/op    4 allocs/op
BenchmarkSetMode-12               135,494,511       9.02 ns/op       0 B/op    0 allocs/op
BenchmarkGetMode-12               174,430,036       6.77 ns/op       0 B/op    0 allocs/op
BenchmarkKeyAdd-12                 23,573,498      45.56 ns/op      40 B/op    2 allocs/op
BenchmarkKeyHas-12                100,000,000      10.04 ns/op       0 B/op    0 allocs/op
BenchmarkKeyList-12                21,103,474      57.20 ns/op      80 B/op    1 allocs/op
BenchmarkConcurrentReads-12        37,712,851      32.36 ns/op      48 B/op    1 allocs/op
BenchmarkConcurrentWrites-12       10,351,329     118.60 ns/op      40 B/op    2 allocs/op
BenchmarkMixedOperations-12        31,084,405      38.72 ns/op      15 B/op    0 allocs/op
```

*Measured on AMD Ryzen 9 7900X3D, Linux, Go 1.21+*

### Concurrency

- **Atomic State**: `atomic.Int32`, `atomic.Int64`, `atomic.Value` for cache
- **Mutex Protection**: `sync.RWMutex` for configuration and pool access
- **Zero Races**: Verified with `go test -race` (no data races detected)

---

## Use Cases

This package is designed for scenarios requiring robust health monitoring:

**Microservices Health Checks**
- Aggregate health from databases, caches, queues, external APIs
- Configure critical (Must) vs optional (Should) dependencies
- Return appropriate HTTP codes for load balancers and orchestrators

**Kubernetes/Docker Health Probes**
- Liveness probe: Use `IsStrictlyHealthy()` for restart signals
- Readiness probe: Use `IsHealthy()` to tolerate warnings
- Startup probe: Check with cached status for efficiency

**API Gateway Integration**
- Expose `/health` and `/status` endpoints
- JSON for programmatic consumption
- Text format for quick visual inspection

**Monitoring Systems**
- Integrate with Prometheus, Datadog, New Relic
- Cache reduces load on monitoring components
- Detailed component status for diagnostics

**Distributed Systems**
- **AnyOf** mode: Redis cluster with multiple nodes (any healthy = OK)
- **Quorum** mode: Database replicas (majority must be healthy)
- **Must** mode: Core dependencies (all must be healthy)

---

## Quick Start

### Basic Status Endpoint

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/nabbar/golib/context"
    "github.com/nabbar/golib/status"
    "github.com/nabbar/golib/monitor/pool"
    "github.com/nabbar/golib/monitor/info"
)

func main() {
    // Create context
    ctx := context.NewGlobal()
    
    // Create status instance
    sts := status.New(ctx)
    
    // Set application info
    sts.SetInfo("my-api", "v1.0.0", "abc123")
    
    // Create and register monitor pool
    monPool := pool.New(ctx)
    sts.RegisterPool(func() montps.Pool { return monPool })
    
    // Add a component monitor
    dbMonitor := info.New(func(context.Context) (monsts.Status, string, error) {
        // Check database health
        if dbHealthy() {
            return monsts.OK, "Database connected", nil
        }
        return monsts.KO, "Database connection failed", nil
    })
    monPool.MonitorAdd(dbMonitor)
    
    // Set up Gin router
    r := gin.Default()
    r.GET("/status", func(c *gin.Context) {
        sts.MiddleWare(c)
    })
    
    r.Run(":8080")
}

func dbHealthy() bool {
    // Your health check logic
    return true
}
```

### With Configuration

```go
import (
    "net/http"
    "github.com/nabbar/golib/status"
    "github.com/nabbar/golib/status/control"
    monsts "github.com/nabbar/golib/monitor/status"
)

func setupStatus() status.Status {
    sts := status.New(ctx)
    sts.SetInfo("my-api", "v1.0.0", "abc123")
    
    // Configure HTTP return codes and mandatory components
    cfg := status.Config{
        ReturnCode: map[monsts.Status]int{
            monsts.OK:   http.StatusOK,           // 200
            monsts.Warn: http.StatusMultiStatus,  // 207
            monsts.KO:   http.StatusServiceUnavailable, // 503
        },
        MandatoryComponent: []status.Mandatory{
            {
                Mode: control.Must,
                Keys: []string{"database", "cache"},
            },
            {
                Mode: control.Should,
                Keys: []string{"email-service"},
            },
            {
                Mode: control.AnyOf,
                Keys: []string{"redis-1", "redis-2", "redis-3"},
            },
        },
    }
    sts.SetConfig(cfg)
    
    return sts
}
```

### Health Check Methods

```go
// In your application
sts := setupStatus()

// Check if healthy (tolerates warnings)
if sts.IsHealthy() {
    log.Println("Service is healthy")
}

// Check specific components
if sts.IsHealthy("database", "cache") {
    log.Println("Core components healthy")
}

// Strict check (no warnings allowed)
if sts.IsStrictlyHealthy() {
    log.Println("Service is perfectly healthy")
}

// Cached checks (efficient for frequent calls)
if sts.IsCacheHealthy() {
    // Uses cached status (3s default)
}
```

---

## Subpackages

### `control` Subpackage

Validation mode definitions for component health evaluation.

**Available Modes**

```go
package control

type Mode uint8

const (
    Ignore Mode = iota  // No validation (component ignored)
    Should              // Warning only (doesn't cause failure)
    Must                // Must be healthy (causes failure if not)
    AnyOf               // At least one in group must be healthy
    Quorum              // Majority (>50%) must be healthy
)
```

**Mode Behavior**

| Mode | Component Status | Overall Impact |
|------|------------------|----------------|
| **Ignore** | Any | No impact (skipped) |
| **Should** | KO | → Warn (not KO) |
| **Should** | Warn | → Warn |
| **Should** | OK | No impact |
| **Must** | KO | → KO |
| **Must** | Warn | → Warn |
| **Must** | OK | No impact |
| **AnyOf** | All KO | → KO |
| **AnyOf** | At least 1 OK | No impact |
| **AnyOf** | Only Warn | → Warn |
| **Quorum** | ≤50% OK+Warn | → KO |
| **Quorum** | >50% OK+Warn | No impact or → Warn |

**String Parsing**

```go
import "github.com/nabbar/golib/status/control"

// Parse from string (case-insensitive)
mode := control.Parse("must")      // → Must
mode = control.Parse("ANYOF")      // → AnyOf
mode = control.Parse("invalid")    // → Ignore (default)

// Format to string
str := control.Must.String()       // → "Must"
```

**Marshaling Support**

The `Mode` type implements multiple encoding interfaces:

```go
// JSON
data, _ := json.Marshal(control.Must)  // → "Must"
var mode control.Mode
json.Unmarshal([]byte(`"must"`), &mode)

// YAML
data, _ := yaml.Marshal(control.Must)
yaml.Unmarshal([]byte("must"), &mode)

// TOML
data, _ := mode.MarshalTOML()

// CBOR (binary)
data, _ := mode.MarshalCBOR()

// Plain text
data, _ := mode.MarshalText()
```

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/status/control) for complete API.

---

### `mandatory` Subpackage

Manages a single component group with an associated validation mode.

**Features**
- Thread-safe key management
- Atomic mode operations
- Lock-free reads for high performance

**API Example**

```go
import (
    "github.com/nabbar/golib/status/mandatory"
    "github.com/nabbar/golib/status/control"
)

// Create new group
m := mandatory.New()

// Set validation mode
m.SetMode(control.Must)

// Add component keys
m.KeyAdd("database", "cache", "queue")

// Check if key exists
if m.KeyHas("database") {
    fmt.Println("Database is in mandatory group")
}

// Get current mode
mode := m.GetMode()  // → Must

// List all keys
keys := m.KeyList()  // → ["database", "cache", "queue"]

// Remove keys
m.KeyDel("queue")
```

**Thread Safety**

```go
// Safe for concurrent use
var wg sync.WaitGroup
m := mandatory.New()

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        m.KeyAdd(fmt.Sprintf("component-%d", id))
    }(i)
}
wg.Wait()

fmt.Println(m.KeyList())  // All 100 components added safely
```

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/status/mandatory) for complete API.

---

### `listmandatory` Subpackage

Manages multiple mandatory groups as a collection.

**Features**
- Thread-safe collection operations
- Iterator pattern with Walk method
- Automatic cleanup of invalid entries

**API Example**

```go
import (
    "github.com/nabbar/golib/status/listmandatory"
    "github.com/nabbar/golib/status/mandatory"
    "github.com/nabbar/golib/status/control"
)

// Create list
list := listmandatory.New()

// Create and add groups
coreGroup := mandatory.New()
coreGroup.SetMode(control.Must)
coreGroup.KeyAdd("database", "cache")
list.Add(coreGroup)

optionalGroup := mandatory.New()
optionalGroup.SetMode(control.Should)
optionalGroup.KeyAdd("email", "sms")
list.Add(optionalGroup)

redisCluster := mandatory.New()
redisCluster.SetMode(control.AnyOf)
redisCluster.KeyAdd("redis-1", "redis-2", "redis-3")
list.Add(redisCluster)

// Get count
count := list.Len()  // → 3

// Walk through groups
list.Walk(func(m mandatory.Mandatory) bool {
    fmt.Printf("Mode: %s, Keys: %v\n", m.GetMode(), m.KeyList())
    return true  // Continue iteration
})

// Remove a group
list.Del(optionalGroup)
```

See [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/status/listmandatory) for complete API.

---

## Configuration

### Config Structure

```go
type Config struct {
    // HTTP status codes for each health state
    ReturnCode map[monsts.Status]int
    
    // Component groups with validation modes
    MandatoryComponent []Mandatory
}

type Mandatory struct {
    Mode control.Mode  // Validation mode
    Keys []string      // Component names
}
```

### Default Values

If `ReturnCode` is empty, defaults are:
- `monsts.OK` → 200 (http.StatusOK)
- `monsts.Warn` → 207 (http.StatusMultiStatus)
- `monsts.KO` → 500 (http.StatusInternalServerError)

### Configuration Examples

**Standard Web API**
```go
cfg := status.Config{
    ReturnCode: map[monsts.Status]int{
        monsts.OK:   200,  // OK
        monsts.Warn: 200,  // Treat warnings as OK
        monsts.KO:   503,  // Service Unavailable
    },
    MandatoryComponent: []status.Mandatory{
        {Mode: control.Must, Keys: []string{"database"}},
        {Mode: control.Should, Keys: []string{"cache"}},
    },
}
```

**Kubernetes Health Probes**
```go
// Liveness: Strict (restart if any issue)
livenessCfg := status.Config{
    ReturnCode: map[monsts.Status]int{
        monsts.OK:   200,
        monsts.Warn: 500,  // Treat warnings as failure
        monsts.KO:   500,
    },
    MandatoryComponent: []status.Mandatory{
        {Mode: control.Must, Keys: []string{"core"}},
    },
}

// Readiness: Tolerant (accept warnings)
readinessCfg := status.Config{
    ReturnCode: map[monsts.Status]int{
        monsts.OK:   200,
        monsts.Warn: 200,  // Accept warnings
        monsts.KO:   503,
    },
}
```

**Distributed System**
```go
cfg := status.Config{
    MandatoryComponent: []status.Mandatory{
        // Core database: must be healthy
        {Mode: control.Must, Keys: []string{"postgres"}},
        
        // Redis cluster: any node OK
        {Mode: control.AnyOf, Keys: []string{
            "redis-master", "redis-replica-1", "redis-replica-2",
        }},
        
        // Kafka cluster: quorum required
        {Mode: control.Quorum, Keys: []string{
            "kafka-1", "kafka-2", "kafka-3",
        }},
        
        // Optional services
        {Mode: control.Should, Keys: []string{"email", "sms"}},
    },
}
```

---

## HTTP API

### Endpoint Configuration

```go
r := gin.Default()

// Simple endpoint
r.GET("/status", func(c *gin.Context) {
    sts.MiddleWare(c)
})

// With generic context
r.GET("/health", func(c *gin.Context) {
    sts.Expose(c)
})
```

### Query Parameters

| Parameter | Values | Description |
|-----------|--------|-------------|
| `short` | `true`, `1` | Return only overall status (no component details) |
| `format` | `text`, `json` | Output format (default: JSON) |

### HTTP Headers

| Header | Values | Description |
|--------|--------|-------------|
| `X-Verbose` | `false` | Return short output (same as `short=true`) |
| `Accept` | `text/plain`, `application/json` | Content negotiation |

### Response Formats

**JSON Response (default)**

```json
{
  "name": "my-api",
  "release": "v1.0.0",
  "hash": "abc123",
  "date_build": "2024-11-13T08:00:00Z",
  "status": "OK",
  "message": "",
  "component": {
    "database": {
      "status": "OK",
      "message": "Database connected",
      "options": {}
    },
    "cache": {
      "status": "OK",
      "message": "Redis operational",
      "options": {}
    }
  }
}
```

**JSON Response (short)**

```json
{
  "name": "my-api",
  "release": "v1.0.0",
  "hash": "abc123",
  "date_build": "2024-11-13T08:00:00Z",
  "status": "OK",
  "message": ""
}
```

**Text Response**

```
OK: my-api (v1.0.0 - abc123)
```

**Text Response (verbose)**

```
OK: my-api (v1.0.0 - abc123)
  database: OK - Database connected
  cache: OK - Redis operational
```

### HTTP Status Codes

The HTTP status code in the response depends on configuration:

| Health Status | Default Code | Meaning |
|---------------|--------------|---------|
| `OK` | 200 | All components healthy |
| `Warn` | 207 | Some warnings present |
| `KO` | 500 | One or more critical components failed |

Customize codes via `Config.ReturnCode` (see [Configuration](#configuration)).

### Usage Examples

**cURL**

```bash
# JSON (default)
curl http://localhost:8080/status

# JSON short
curl http://localhost:8080/status?short=true

# Text format
curl http://localhost:8080/status?format=text

# Text format with details
curl -H "Accept: text/plain" http://localhost:8080/status

# Short via header
curl -H "X-Verbose: false" http://localhost:8080/status
```

**Go Client**

```go
resp, err := http.Get("http://localhost:8080/status")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

if resp.StatusCode == http.StatusOK {
    log.Println("Service is healthy")
}

var status struct {
    Name    string `json:"name"`
    Release string `json:"release"`
    Status  string `json:"status"`
}
json.NewDecoder(resp.Body).Decode(&status)
```

---

## Best Practices

### Component Organization

**Group by Criticality**

```go
cfg := status.Config{
    MandatoryComponent: []status.Mandatory{
        // Critical: Must be healthy for service to function
        {Mode: control.Must, Keys: []string{
            "database",
            "auth-service",
        }},
        
        // Important: Warnings acceptable, but not failures
        {Mode: control.Should, Keys: []string{
            "cache",
            "search-index",
        }},
        
        // Redundant: Any one instance is sufficient
        {Mode: control.AnyOf, Keys: []string{
            "worker-1", "worker-2", "worker-3",
        }},
    },
}
```

### Health Check Implementation

**✅ Good Practices**

```go
// 1. Fast health checks (<100ms)
func checkDatabase() (monsts.Status, string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        return monsts.KO, "Database unreachable", err
    }
    return monsts.OK, "Database connected", nil
}

// 2. Use cached status for efficiency
if sts.IsCacheHealthy() {
    // Serve traffic (uses 3s cache)
}

// 3. Distinguish between warnings and failures
func checkCache() (monsts.Status, string, error) {
    if err := cache.Ping(); err != nil {
        // Cache down but not critical
        return monsts.Warn, "Cache unavailable", err
    }
    return monsts.OK, "Cache operational", nil
}

// 4. Provide meaningful messages
return monsts.KO, fmt.Sprintf("Connection pool exhausted: %d/%d", active, max), nil
```

**❌ Bad Practices**

```go
// 1. Slow health checks (blocks thread)
func checkBad() (monsts.Status, string, error) {
    time.Sleep(5 * time.Second) // Too slow!
    return monsts.OK, "", nil
}

// 2. Ignoring cache benefits
for {
    if sts.IsHealthy() { // Recalculates every time
        // Heavy operation repeated unnecessarily
    }
    time.Sleep(100 * time.Millisecond)
}

// 3. Generic error messages
return monsts.KO, "Error", err // Not helpful for debugging

// 4. Silent failures
func checkSilent() (monsts.Status, string, error) {
    db.Ping() // Ignoring error
    return monsts.OK, "", nil
}
```

### Error Handling

**Always Handle Errors**

```go
// ✅ Good
if err := sts.MonitorAdd(monitor); err != nil {
    log.Printf("Failed to add monitor: %v", err)
    return err
}

// ❌ Bad
sts.MonitorAdd(monitor) // Ignoring error
```

### Kubernetes Integration

**Liveness Probe** (Restart on failure)

```yaml
livenessProbe:
  httpGet:
    path: /status?short=true
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

**Readiness Probe** (Remove from service on failure)

```yaml
readinessProbe:
  httpGet:
    path: /status?short=true
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 2
```

**Startup Probe** (Initial health check)

```yaml
startupProbe:
  httpGet:
    path: /status?short=true
    port: 8080
  initialDelaySeconds: 0
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 30  # Allow 150s for startup
```

### Monitoring Integration

**Prometheus Metrics**

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    healthStatus = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "app_health_status",
            Help: "Health status (0=KO, 1=Warn, 2=OK)",
        },
        []string{"component"},
    )
)

func updateMetrics(sts status.Status) {
    sts.MonitorWalk(func(name string, mon montps.Monitor) bool {
        status := mon.Status()
        healthStatus.WithLabelValues(name).Set(float64(status.Int()))
        return true
    })
}
```

### Performance Optimization

**Use Cached Methods**

```go
// High-frequency checks (every request)
func middleware(c *gin.Context) {
    if !sts.IsCacheHealthy() {
        c.AbortWithStatus(503)
        return
    }
    c.Next()
}

// Background monitoring
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        if !sts.IsHealthy() {
            alerting.Trigger("service-unhealthy")
        }
    }
}()
```

**Component Monitoring Best Practices**

- Keep checks lightweight (<100ms)
- Use timeouts to prevent hanging
- Cache expensive checks in the monitor itself
- Return specific error messages
- Use appropriate control modes (Must vs Should)

---

## Testing

**Test Suite**: 306 specs across 4 packages with 85.6% overall coverage

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# With race detection (recommended)
CGO_ENABLED=1 go test -race ./...

# Benchmarks
go test -bench=. -benchmem ./mandatory/
```

### Test Results

```
status/                120 specs    85.6% coverage   10.7s
status/control/        102 specs    95.0% coverage   0.01s
status/listmandatory/   29 specs    75.4% coverage   0.5s
status/mandatory/       55 specs    76.1% coverage   0.1s
```

### Quality Assurance

- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Comprehensive edge case coverage
- ✅ Benchmark performance validation

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
- Add examples for common use cases
- Keep TESTING.md synchronized with test changes

**Testing**
- Write tests for all new features
- Test edge cases and error conditions
- Verify thread safety with race detector
- Add comments explaining complex scenarios

**Pull Requests**
- Provide clear description of changes
- Reference related issues
- Include test results
- Update documentation

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Future Enhancements

Potential improvements for future versions:

**Control Modes**
- Custom validation logic via function callbacks
- Weighted quorum (different components have different weights)
- Time-based validation (require health for N consecutive checks)

**Caching**
- Per-component cache duration
- Cache invalidation hooks
- Configurable cache strategies (LRU, TTL, etc.)

**Monitoring**
- Built-in metrics export (Prometheus format)
- Health check history and trends
- Circuit breaker integration for failing components

**Output Formats**
- XML output support
- GraphQL endpoint support
- Structured logging output

**Advanced Features**
- Dependency graphs (component A depends on B)
- Health check scheduling (different intervals per component)
- Multi-region health aggregation
- gRPC health check protocol support

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
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/status)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)
- **Monitor Package**: [github.com/nabbar/golib/monitor](https://pkg.go.dev/github.com/nabbar/golib/monitor)
- **Gin Framework**: [gin-gonic/gin](https://github.com/gin-gonic/gin)

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Status Package Contributors