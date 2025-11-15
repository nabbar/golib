# Monitor Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://golang.org/)

Production-ready health monitoring system for Go applications with automatic status transitions, configurable thresholds, and comprehensive metrics tracking.

> **AI Disclaimer**: AI tools are used solely to assist with testing, documentation, and bug fixes under human supervision, in compliance with EU AI Act Article 50.4.

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
- [Configuration](#configuration)
- [Status Transitions](#status-transitions)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Contributing](#contributing)
- [Future Enhancements](#future-enhancements)
- [License](#license)

---

## Overview

The monitor package provides a sophisticated health monitoring system for production Go applications. It implements automatic health check execution with intelligent status transitions, hysteresis to prevent flapping, and comprehensive metrics collection.

### Design Philosophy

1. **Reliability First**: Hysteresis-based transitions prevent status flapping during temporary issues
2. **Observability**: Track latency, uptime, downtime, and state transitions for complete visibility
3. **Flexibility**: Configurable intervals, thresholds, and extensible middleware chain
4. **Thread-Safe**: Fine-grained locking and atomic operations for concurrent access
5. **Composable**: Independent subpackages (info, status, pool, types) work together seamlessly

### Value Proposition

- **Prevent Alert Fatigue**: Hysteresis prevents flapping during transient failures
- **Adaptive Monitoring**: Automatically adjusts check frequency based on component health
- **Production Ready**: Thread-safe, tested, and battle-proven in production
- **Observable**: Complete visibility into health status and performance metrics
- **Scalable**: Efficiently manage hundreds of monitors with pool management

---

## Key Features

- **Three-State Model**: OK → Warn → KO transitions with configurable thresholds
- **Adaptive Intervals**: Different check frequencies for normal, rising, and falling states
- **Comprehensive Metrics**: Latency, uptime, downtime, rise/fall times
- **Thread-Safe**: Concurrent access safe with fine-grained locking
- **Pool Management**: Group and manage multiple monitors with batch operations
- **Prometheus Integration**: Built-in metrics export
- **Middleware Chain**: Extensible health check pipeline
- **Dynamic Metadata**: Runtime-generated component information
- **Shell Commands**: CLI-style operational control

---

## Installation

```bash
go get github.com/nabbar/golib/monitor
```

---

## Architecture

### Package Structure

```
monitor/
├── monitor          # Core health monitoring
├── pool/            # Monitor pool management
├── info/            # Component metadata
├── status/          # Status enumeration
└── types/           # Type definitions
```

### Component Hierarchy

```
┌────────────────────────────────────────┐
│         Monitor Package                 │
│    Health Check Monitoring System       │
└──────┬──────┬────────┬─────────┬───────┘
       │      │        │         │
   ┌───▼──┐ ┌─▼───┐ ┌─▼────┐ ┌──▼──────┐
   │ Pool │ │Info │ │Status│ │  Types  │
   └──────┘ └─────┘ └──────┘ └─────────┘
```

### Status Transition Model

```
       ┌──────────────┐
       │      KO      │  ← Component unhealthy
       └──────┬───────┘
              │ riseCountKO successes
              ▼
       ┌──────────────┐
       │     Warn     │  ← Component degraded
       └──────┬───────┘
              │ riseCountWarn successes
              ▼
       ┌──────────────┐
       │      OK      │  ← Component healthy
       └──────┬───────┘
              │ fallCountWarn failures
              ▼
       (returns to Warn, then KO)
```

---

## Quick Start

### Basic Monitor

```go
import (
    "context"
    "time"
    "github.com/nabbar/golib/monitor"
    "github.com/nabbar/golib/monitor/info"
    "github.com/nabbar/golib/monitor/types"
    "github.com/nabbar/golib/duration"
)

// Create monitor
inf, _ := info.New("database-monitor")
mon, _ := monitor.New(context.Background, inf)

// Configure
cfg := types.Config{
    Name:          "postgres",
    CheckTimeout:  duration.ParseDuration(5 * time.Second),
    IntervalCheck: duration.ParseDuration(30 * time.Second),
    FallCountKO:   3,
    RiseCountKO:   3,
}
mon.SetConfig(context.Background(), cfg)

// Register health check
mon.SetHealthCheck(func(ctx context.Context) error {
    return db.PingContext(ctx)
})

// Start
mon.Start(context.Background())
defer mon.Stop(context.Background())

// Query
fmt.Printf("Status: %s\n", mon.Status())
```

### Monitor Pool

```go
import "github.com/nabbar/golib/monitor/pool"

pool := pool.New(ctxFunc)

// Add monitors
pool.MonitorAdd(createDBMonitor())
pool.MonitorAdd(createAPIMonitor())

// Register metrics
pool.RegisterMetrics(promFunc, logFunc)
defer pool.UnregisterMetrics()

// Start all
pool.Start(ctx)
defer pool.Stop(ctx)
```

---

## Performance

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Monitor Creation | 1.2 µs | 2.1 KB | 18 allocs |
| Health Check | 15 µs | 448 B | 5 allocs |
| Status Transition | 800 ns | 0 B | 0 allocs |
| Metrics Collection | 2.5 µs | 0 B | 0 allocs |
| Pool.Start (10 monitors) | 85 µs | 8 KB | 120 allocs |

---

## Use Cases

### 1. Microservice Health Monitoring
Monitor multiple services with automatic transitions and metrics collection.

### 2. Database Connection Pooling
Track database health with adaptive intervals for faster issue detection.

### 3. External Service Dependencies
Monitor third-party API availability with configurable timeouts.

### 4. Kubernetes Probes
Integrate with liveness and readiness probes.

### 5. Custom Middleware
Extend health checks with logging, metrics, or custom logic.

---

## Subpackages

### monitor (Core)
Core health check monitoring with status transitions, metrics, and lifecycle management.

**GoDoc**: [pkg.go.dev/github.com/nabbar/golib/monitor](https://pkg.go.dev/github.com/nabbar/golib/monitor)

### pool
Manage multiple monitors as a group with batch operations and Prometheus integration.

**Documentation**: [pool/README.md](./pool/README.md)

### info
Dynamic metadata management with caching and lazy evaluation.

**Documentation**: [info/README.md](./info/README.md)

### status
Type-safe status enumeration (OK, Warn, KO) with multi-format encoding.

### types
Shared interfaces, configuration types, and error codes.

---

## Configuration

```go
type Config struct {
    Name          string            // Component name
    CheckTimeout  duration.Duration // Health check timeout (min: 5s)
    IntervalCheck duration.Duration // Normal check interval (min: 1s)
    IntervalFall  duration.Duration // Interval when falling (min: 1s)
    IntervalRise  duration.Duration // Interval when rising (min: 1s)
    FallCountKO   int              // Failures for Warn→KO (min: 1)
    FallCountWarn int              // Failures for OK→Warn (min: 1)
    RiseCountKO   int              // Successes for KO→Warn (min: 1)
    RiseCountWarn int              // Successes for Warn→OK (min: 1)
}
```

**Best Practices**:
- `CheckTimeout` < `IntervalCheck` (prevent overlapping checks)
- Use shorter `IntervalFall` for faster issue detection
- Set counts ≥ 2 to prevent flapping

---

## Status Transitions

### Transition Rules

| From | To | Condition | Resets |
|------|-----|-----------|--------|
| KO | Warn | `riseCountKO` consecutive successes | Fall counters |
| Warn | OK | `riseCountWarn` consecutive successes | Fall counters |
| OK | Warn | `fallCountWarn` consecutive failures | Rise counters |
| Warn | KO | `fallCountKO` consecutive failures | Rise counters |

### Example Sequence

Configuration: `FallCountWarn:2, FallCountKO:3, RiseCountKO:3, RiseCountWarn:2`

```
Check 1: ✓ → OK
Check 2: ✗ → OK (1 failure)
Check 3: ✗ → Warn (2 failures, threshold reached)
Check 4: ✗ → Warn (1 KO failure)
Check 5: ✗ → Warn (2 KO failures)
Check 6: ✗ → KO (3 KO failures, threshold reached)
Check 7-9: ✓✓✓ → Warn (3 successes, KO threshold reached)
Check 10-11: ✓✓ → OK (2 successes, Warn threshold reached)
```

---

## Best Practices

### Configuration
- Set `CheckTimeout` < `IntervalCheck`
- Use faster `IntervalFall` for issue detection
- Configure counts ≥ 2 to prevent flapping

### Health Checks
- Respect context timeout
- Return specific errors
- Keep checks lightweight
- Handle transient failures

### Lifecycle
- Always call `Stop()` when done
- Use `defer` for cleanup
- Check `IsRunning()` before operations

### Pool Management
- Use pools for related monitors
- Call `UnregisterMetrics()` on shutdown
- Register Prometheus metrics early

---

## API Reference

### Monitor Interface

```go
type Monitor interface {
    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    IsRunning() bool
    
    // Configuration
    SetConfig(ctx context.Context, cfg Config) error
    GetConfig() Config
    
    // Health Check
    SetHealthCheck(hc HealthCheck)
    RegisterMiddleware(mw Middleware)
    
    // Status & Metrics
    Status() status.Status
    Latency() time.Duration
    Uptime() time.Duration
    Downtime() time.Duration
    
    // Info
    InfoGet() Info
    InfoMap() map[string]interface{}
    
    // Encoding
    MarshalText() ([]byte, error)
    MarshalJSON() ([]byte, error)
}
```

### Pool Interface

```go
type Pool interface {
    // Monitor Management
    MonitorAdd(mon Monitor) error
    MonitorGet(name string) Monitor
    MonitorDel(name string)
    MonitorList() []string
    
    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Restart(ctx context.Context) error
    
    // Metrics
    RegisterMetrics(prom, log func) error
    UnregisterMetrics()
    
    // Shell
    GetShellCommand(ctx context.Context) []Command
}
```

---

## Testing

**Test Suite**: 595 specs across 4 packages with 86.1% overall coverage

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
monitor/                122 specs    68.5% coverage   0.23s
monitor/info/           139 specs   100.0% coverage   0.12s
monitor/pool/           153 specs    76.2% coverage  11.78s
monitor/status/         181 specs    98.4% coverage   0.02s
```

**Quality Assurance**
- ✅ Zero data races (verified with `-race`)
- ✅ Thread-safe concurrent operations
- ✅ Comprehensive edge case testing
- ✅ Time-dependent behavior validation

See [TESTING.md](./TESTING.md) for detailed testing documentation.

---

## Contributing

Contributions welcome! Please follow these guidelines:

### Code Standards
- Write tests for new features
- Update documentation
- Add GoDoc comments for public APIs
- Run `go fmt` and `go vet`
- Test with race detector (`-race`)

### AI Usage Policy
- **DO NOT** use AI tools to generate package code or core logic
- **DO** use AI to assist with:
  - Writing and improving tests
  - Documentation and comments
  - Debugging and bug fixes
  
All AI-assisted work must be reviewed and validated by a human maintainer.

### Pull Request Process
1. Fork the repository
2. Create a feature branch
3. Write tests (coverage > 70%)
4. Update documentation
5. Run full test suite with race detection
6. Submit PR with clear description

---

## Future Enhancements

Potential improvements under consideration:

- **Circuit Breaker Pattern**: Automatic service isolation during failures
- **Distributed Monitoring**: Cluster-wide health coordination
- **Historical Metrics**: Long-term trend analysis
- **Custom Exporters**: Support for other metrics systems (StatsD, InfluxDB)
- **Health Check Templates**: Predefined checks for common services
- **Dynamic Thresholds**: Adaptive thresholds based on historical data

Contributions and suggestions are welcome!

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

---

## Resources

- **Issues**: [GitHub Issues](https://github.com/nabbar/golib/issues)
- **Documentation**: [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/monitor)
- **Testing Guide**: [TESTING.md](TESTING.md)
- **Contributing**: [CONTRIBUTING.md](../../CONTRIBUTING.md)

**Related Packages**:
- [context](https://github.com/nabbar/golib/tree/main/context) - Context management
- [runner](https://github.com/nabbar/golib/tree/main/runner) - Ticker and lifecycle management
- [prometheus](https://github.com/nabbar/golib/tree/main/prometheus) - Metrics export
- [status](https://github.com/nabbar/golib/tree/main/status) - Status aggregation

---

**Version**: Go 1.18+ on Linux, macOS, Windows  
**Maintained By**: Monitor Package Contributors
