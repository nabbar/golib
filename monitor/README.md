# Monitor Package

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](../../LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://go.dev/doc/install)
[![Coverage](https://img.shields.io/badge/Coverage-84.7%25-brightgreen)](TESTING.md)

The **Monitor Package** is a high-performance, production-ready health monitoring system for Go applications. It provides a robust framework for tracking the operational status of internal components and external dependencies using an intelligent state machine with built-in hysteresis and lock-free metrics reporting.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Performance](#performance)
- [Subpackages](#subpackages)
    - [info](#info)
    - [pool](#pool)
    - [status](#status)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Best Practices](#best-practices)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [Resources](#resources)

---

## Overview

The monitor package is designed to provide "observation without interference". It allows developers to register periodic health checks that automatically transition through health states (OK, Warn, KO) based on configurable failure and recovery thresholds.

### Design Philosophy

1. **Lock-Free Hot Path**: Reading the status or metrics of a monitor is architected using atomic operations, ensuring zero contention even under thousands of concurrent requests.
2. **Dampened Transitions**: Hysteresis logic prevents "alert flapping" by requiring consecutive successes or failures before triggering a state change.
3. **Context-Aware**: Every health check execution is bounded by a context timeout, ensuring that hanging diagnostics do not block the system.
4. **Middleware-First**: Execution is wrapped in a LIFO stack, allowing for easy injection of tracing, logging, or custom metrics logic.

### Key Features

- ✅ **Three-State Machine**: Full lifecycle tracking (OK ↔ Warn ↔ KO).
- ✅ **Adaptive Ticker**: Dynamically adjusts polling frequency during transition phases (Rise/Fall).
- ✅ **Atomic Metrics**: High-precision tracking of Latency, Uptime, Downtime, and Transition times.
- ✅ **Prometheus Integration**: Built-in dispatching logic for automated metrics exporting.
- ✅ **Metadata Management**: Dynamic runtime information through the `info` subpackage.
- ✅ **Zero-Allocation Reads**: Optimized memory path for high-frequency status polling.

### Key Benefits

- **vs Standard Tickers**: Provides a complete state machine and metrics container out-of-the-box, rather than just a periodic trigger.
- **vs Basic Maps**: Thread-safety is guaranteed through atomic primitives rather than global mutexes, offering superior scalability on multi-core systems.

---

## Architecture

### Package Structure

```
monitor/
├── monitor.go               # Implementation of the core Monitor orchestrator
├── interface.go             # Public interface definitions and factory
├── model.go                 # Internal structures and atomic containers
├── last.go                  # High-performance metrics & status storage
├── server.go                # Ticker runner and periodic execution logic
├── internalConfig.go        # Configuration normalization and validation
├── middleware.go            # Execution pipeline implementation
├── encode.go                # JSON/Text serialization logic
├── doc.go                   # GoDoc package documentation
│
├── info/                    # Metadata management sub-package
├── pool/                    # Group management and batch operations
├── status/                  # Status enumeration and multi-format parsing
└── types/                   # Cross-package shared interfaces
```

### Package Architecture

The monitor uses a **Split-State Architecture**. Configuration and Metadata are stored in thread-safe but high-level containers, while operational metrics are stored in a dedicated `lastRun` structure using low-level `sync/atomic` primitives.

```
[ Monitor Instance ]
       |
       +--- [ Config Context ] ---> (Logger, Ticker Intervals, Thresholds)
       |
       +--- [ Metadata Container ] ---> (Atomic Name, Version, Custom Data)
       |
       +--- [ Background Runner ] ---> (Ticker Goroutine)
       |
       +--- [ Performance Metrics ] ---> (Atomic Status, Latency, Uptime Counters)
```

### Dataflow

The periodic check cycle follows a structured pipeline:

```
1. Ticker Tick --------> 2. Interval Resolver ----> 3. Middleware Stack
                               |                          |
   (Adjusts speed if           |                          |-- [ mdlStatus ] (Start Time)
    Rising or Falling)         |                          |-- [ User Function ] (Diagnostic)
                               |                          |-- [ mdlStatus ] (Set Result)
                               |                          |
4. Metrics Export <---- 5. State Transition <-------------+
      |                 (Update Counters & Status)
      v
[ Prometheus / Logs ]
```

---

## Performance

The monitor is optimized for zero-contention on the read path. The following benchmarks were captured on an Intel Core i7-4700HQ.

| Operation           | Performance | Memory     | Efficiency          |
|---------------------|-------------|------------|---------------------|
| **Status Read**     | ~3.14 ns/op | 0 B/op     | **Zero Garbage**    |
| **Latency Read**    | ~2.23 ns/op | 0 B/op     | **Zero Garbage**    |
| **Concurrent Read** | ~0.85 ns/op | 0 B/op     | **Linear Scaling**  |
| **Check Execution** | ~15.0 µs/op | 448 B/op   | Low Overhead        |
| **Configuration**   | ~49.4 µs/op | 24.8 KB/op | Administrative Path |

**Note**: Status and Metric reads are lock-free and do not produce pressure on the Garbage Collector.

---

## Subpackages

### info
Dynamic metadata management. It allows attaching functions to retrieve runtime data (like version or git hash) only when requested.
- **Documentation**: [info/README.md](./info/README.md)

### pool
Manages monitor groups. Provides batch control (Start/Stop all) and aggregated Prometheus exporters.
- **Documentation**: [pool/README.md](./pool/README.md)

### status
Type-safe status enumeration. Handles conversions and multi-format marshalling (JSON/YAML/TOML).
- **Documentation**: [status/README.md](./status/README.md)

---

## Use Cases

### 1. External API Resilience
Monitor third-party services with "dampened" transitions to avoid false alarms on transient glitches.

```go
cfg := types.Config{
    FallCountWarn: 3, // Ignore isolated failures
    IntervalCheck: duration.ParseDuration("30s"),
}
```

### 2. High-Frequency Telemetry
Feed Prometheus scrapers or liveness probes using the lock-free read path without impacting diagnostic performance.

```go
// Atomic read (~3ns) - No impact on system latency
status := mon.Status() 
```

---

## Quick Start

```go
import (
    "context"
    "github.com/nabbar/golib/monitor"
    "github.com/nabbar/golib/monitor/info"
    "github.com/nabbar/golib/monitor/types"
    "github.com/nabbar/golib/duration"
)

func main() {
    inf, _ := info.New("api-service")
    mon, _ := monitor.New(context.Background(), inf)
    
    _ = mon.SetConfig(context.Background(), types.Config{
        IntervalCheck: duration.ParseDuration("10s"),
        FallCountKO: 3,
    })
    
    mon.SetHealthCheck(func(ctx context.Context) error {
        return db.PingContext(ctx)
    })
    
    _ = mon.Start(context.Background())
    defer mon.Stop(context.Background())
}
```

---

## Best Practices

### ✅ DO
- **Use `Eventually` in tests**: Since monitoring is asynchronous, use non-blocking matchers.
- **Respect Context**: Ensure your diagnostic function honors the `ctx` provided to handle timeouts.
- **Register Metrics Early**: Association with Prometheus should be done during initialization.

### ❌ DON'T
- **Don't use `time.Sleep`**: The monitor orchestrator already handles intervals.
- **Don't block the Read Path**: The package provides atomic counters; do not wrap them in heavy mutex-guarded logic.

---

## API Reference

### 1. Primary Factory

| Function | Parameters                    | Returns            | Description                                 |
|----------|-------------------------------|--------------------|---------------------------------------------|
| `New`    | `ctx` (Context), `nfo` (Info) | `(Monitor, error)` | Initializes a thread-safe monitor instance. |

### 2. Monitor Interface

The `Monitor` interface aggregates multiple specialized behaviors.

#### Lifecycle Methods
| Method      | Parameters | Returns | Description                                                         |
|-------------|------------|---------|---------------------------------------------------------------------|
| `Start`     | `ctx`      | `error` | Launches the background ticker. Waits for operational confirmation. |
| `Stop`      | `ctx`      | `error` | Halts the background ticker and waits for current check completion. |
| `Restart`   | `ctx`      | `error` | Performs a synchronized full Stop followed by a Start cycle.        |
| `IsRunning` | -          | `bool`  | Thread-safe check of the background runner status.                  |

#### Configuration & Core Logic
| Method           | Parameters            | Returns            | Description                                                             |
|------------------|-----------------------|--------------------|-------------------------------------------------------------------------|
| `SetConfig`      | `ctx`, `cfg` (Config) | `error`            | Validates and applies runtime parameters and logging options.           |
| `GetConfig`      | -                     | `Config`           | Returns a deep-copy snapshot of the current effective configuration.    |
| `SetHealthCheck` | `fct` (HealthCheck)   | -                  | Registers the function responsible for the component diagnostic.        |
| `GetHealthCheck` | -                     | `HealthCheck`      | Retrieves the currently registered diagnostic function.                 |
| `Clone`          | `ctx`                 | `(Monitor, error)` | Deep copy of the monitor instance, inheriting state and running status. |

#### Status & State (MonitorStatus)
| Method     | Returns         | Performance | Description                                                           |
|------------|-----------------|-------------|-----------------------------------------------------------------------|
| `Status`   | `status.Status` | ~3ns        | Atomic retrieval of current health (OK/Warn/KO).                      |
| `Latency`  | `time.Duration` | ~2ns        | Atomic duration of the last executed health check.                    |
| `Uptime`   | `time.Duration` | ~2ns        | Total cumulative duration spent in the OK health status.              |
| `Downtime` | `time.Duration` | ~2ns        | Total cumulative duration spent in Warn or KO statuses.               |
| `Message`  | `string`        | -           | Returns the last error or status message captured during execution.   |
| `IsRise`   | `bool`          | -           | Reports if the monitor is currently recovering from a degraded state. |
| `IsFall`   | `bool`          | -           | Reports if the monitor is currently degrading toward a failure state. |

#### Metrics & Prometheus (MonitorMetrics)
| Method                   | Parameters    | Description                                                                 |
|--------------------------|---------------|-----------------------------------------------------------------------------|
| `RegisterMetricsName`    | `...string`   | Defines the Prometheus metric identifiers for this monitor instance.        |
| `RegisterMetricsAddName` | `...string`   | Appends new identifiers to the existing list (handles de-duplication).      |
| `RegisterCollectMetrics` | `FuncCollect` | Associates a provider function for metrics extraction during scrape cycles. |

#### Metadata Management (MonitorInfo)
| Method     | Returns          | Description                                                   |
|------------|------------------|---------------------------------------------------------------|
| `InfoName` | `string`         | Atomic retrieval of the monitor descriptive name.             |
| `InfoMap`  | `map[string]any` | Dynamic retrieval of component metadata (version, env, etc.). |
| `InfoUpd`  | -                | Thread-safe update of the monitor metadata implementation.    |

### 3. Config Structure (`types.Config`)

| Field           | Type       | Default     | Description                                                   |
|-----------------|------------|-------------|---------------------------------------------------------------|
| `Name`          | `string`   | "not named" | Unique identifier for logging and metrics.                    |
| `CheckTimeout`  | `Duration` | 5s          | Maximum allowed execution time for a single HealthCheck.      |
| `IntervalCheck` | `Duration` | 1s          | Normal polling frequency when the status is stable.           |
| `IntervalFall`  | `Duration` | 1s          | Polling frequency adjustment during degradation (Fall phase). |
| `IntervalRise`  | `Duration` | 1s          | Polling frequency adjustment during recovery (Rise phase).    |
| `FallCountKO`   | `uint8`    | 1           | Consecutive failures required to transition from Warn to KO.  |
| `FallCountWarn` | `uint8`    | 1           | Consecutive failures required to transition from OK to Warn.  |
| `RiseCountKO`   | `uint8`    | 1           | Consecutive successes required to transition from KO to Warn. |
| `RiseCountWarn` | `uint8`    | 1           | Consecutive successes required to transition from Warn to OK. |
| `Logger`        | `Options`  | -           | Integrated structured logging configuration.                  |

---

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Code Quality**
  - Follow Go best practices and idioms
  - Maintain or improve code coverage (target: >80%)
  - Pass all tests including race detector
  - Use `gofmt`, `golangci-lint` and `gosec`

2. **AI Usage Policy**
  - ❌ **AI must NEVER be used** to generate package code or core functionality
  - ✅ **AI assistance is limited to**:
    - Testing (writing and improving tests)
    - Debugging (troubleshooting and bug resolution)
    - Documentation (comments, README, TESTING.md)
  - All AI-assisted work must be reviewed and validated by humans

3. **Testing**
  - Add tests for new features
  - Use Ginkgo v2 / Gomega for test framework
  - Ensure zero race conditions
  - Maintain coverage above 80%

4. **Documentation**
  - Update GoDoc comments for public APIs
  - Add examples for new features
  - Update README.md and TESTING.md if needed

5. **Pull Request Process**
  - Fork the repository
  - Create a feature branch
  - Write clear commit messages
  - Ensure all tests pass
  - Update documentation
  - Submit PR with description of changes

---

## Resources

### Documentation
- **[TESTING.md](TESTING.md)**: Exhaustive test inventory, performance benchmarks, and CPU/Memory profiling data.

### Monitoring Standards & Industry References
- **[Google SRE Book: Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/)**: Core theoretical framework for service monitoring, explaining the difference between symptoms and causes, and the "Four Golden Signals".
- **[Kubernetes Probes Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)**: Industry standard for container orchestration probes. This package's transition logic (Fall/Rise counts) is designed to align with Kubernetes liveness and readiness probe behaviors.
- **[HAProxy Health Check Configuration](https://www.haproxy.com/blog/how-to-enable-health-checks-in-haproxy/)**: Operational reference for load-balancer health monitoring, detailing intervals and thresholds for high-availability systems.
- **[Traefik Backend Health Monitoring](https://doc.traefik.io/traefik/routing/services/#health-check)**: Configuration standards for dynamic proxy health checks used in modern microservices architectures.
- **[Prometheus Metric Naming Best Practices](https://prometheus.io/docs/practices/naming/)**: Essential guide for ensuring that metrics registered via `RegisterMetricsName` follow standard conventions for dashboards and alerting.

### Summary
These resources provide the context for why this package exists. The **Monitor Package** provides the Go implementation of the **Hysteresis** and **State Machine** patterns used by orchestrators like Kubernetes, with the **Atomic-Read** performance required for SRE-grade telemetry.

---

## AI Transparency

In compliance with EU AI Act Article 50.4: AI assistance was used for performance profiling, test inventory generation, and documentation synchronization under human supervision. Core monitoring logic is human-designed and validated.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.

Copyright (c) 2022-2025 Nicolas JUHEL
