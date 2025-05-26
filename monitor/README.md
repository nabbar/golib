# Documentation - `github.com/nabbar/golib/monitor`

This documentation provides an overview and usage guide for the `github.com/nabbar/golib/monitor` package and its subpackages. The package is designed to help developers implement, manage, and monitor health checks for various components in their applications, with advanced features such as metrics collection, status management, and pooling of monitors.

## Subpackages
- [`monitor/pool`](#monitorpool-subpackage-documentation): Manages a collection of health monitors as a group. See the [monitor/pool documentation](#monitorpool-subpackage-documentation) for details.
- [`monitor`](#monitor-package-documentation): Core logic for defining, running, and monitoring health checks. See the [monitor documentation](#monitor-package-documentation) for details.
- [`monitor/info`](#monitorinfo-subpackage-documentation): Provides types and utilities for component metadata. See the [monitor/info documentation](#monitorinfo-subpackage-documentation) for details.
- [`monitor/status`](#monitorstatus-subpackage-documentation): Defines status types and utilities for health status management. See the [monitor/status documentation](#monitorstatus-subpackage-documentation) for details.


---

## monitor/pool Subpackage Documentation

The `monitor/pool` subpackage provides a system to manage and operate a collection of health monitors as a group. It enables dynamic addition, removal, lifecycle control, metrics aggregation, and operational shell commands for all monitors in the pool. All operations are thread-safe and suitable for concurrent environments.

---

### Features

- **Dynamic Monitor Management**: Add, get, set, delete, list, and walk through monitors in the pool.
- **Lifecycle Control**: Start, stop, and restart all or selected monitors.
- **Metrics Aggregation**: Collect and export metrics (latency, uptime, downtime, status, SLI, etc.) for all monitors.
- **Shell Command Integration**: Expose operational commands for listing, controlling, and querying monitors.
- **Prometheus & Logger Integration**: Register Prometheus and logger functions for observability.
- **Thread-Safe**: All operations are safe for concurrent use.

---

### Main Concepts

#### Pool Creation

Create a new pool by providing a context function. The pool can be further configured with Prometheus and logger integrations.

```go
import (
    "github.com/nabbar/golib/monitor/pool"
    "github.com/nabbar/golib/context"
)

p := pool.New(context.NewFuncContext())
```

#### Monitor Management

- **Add**: Add a new monitor to the pool.
- **Get**: Retrieve a monitor by name.
- **Set**: Update or replace a monitor in the pool.
- **Delete**: Remove a monitor by name.
- **List**: List all monitor names.
- **Walk**: Iterate over all monitors, optionally filtering by name.

```go
p.MonitorAdd(monitor)
mon := p.MonitorGet("name")
p.MonitorSet(monitor)
p.MonitorDel("name")
names := p.MonitorList()
p.MonitorWalk(func(name string, mon Monitor) bool { /* ... */ })
```

#### Lifecycle Management

- **Start**: Start all monitors in the pool.
- **Stop**: Stop all monitors.
- **Restart**: Restart all monitors.
- **IsRunning**: Check if any monitor is running.
- **Uptime**: Get the maximum uptime among all monitors.

```go
err := p.Start(ctx)
err := p.Stop(ctx)
err := p.Restart(ctx)
running := p.IsRunning()
uptime := p.Uptime()
```

#### Metrics Collection

- **InitMetrics**: Register Prometheus and logger functions, and initialize metrics.
- **TriggerCollectMetrics**: Periodically trigger metrics collection for all monitors.
- **Metrics Export**: Metrics include latency, uptime, downtime, rise/fall times, status, and SLI rates.

```go
p.InitMetrics(prometheusFunc, loggerFunc)
go p.TriggerCollectMetrics(ctx, time.Minute)
```

#### Shell Command Integration

The pool exposes shell commands for operational control:

- `list`: List all monitors.
- `info`: Show detailed info for monitors.
- `start`, `stop`, `restart`: Control monitor lifecycle.
- `status`: Show status and messages for monitors.

Retrieve available shell commands:

```go
cmds := p.GetShellCommand(ctx)
```

---

### Encoding and Export

- **MarshalText**: Export the pool as a human-readable text.
- **MarshalJSON**: Export the pool as a JSON object with monitor statuses.

```go
txt, err := p.MarshalText()
jsn, err := p.MarshalJSON()
```

---

### Notes

- All operations are thread-safe and suitable for concurrent use.
- Designed for Go 1.18+.
- Integrates with Prometheus and logging systems for observability.
- Can be used as a standalone pool or as part of a larger monitoring system.

---

## monitor Package Documentation

The `monitor` package provides the core logic for defining, running, and monitoring health checks for individual components in Go applications. It offers a flexible, thread-safe abstraction for health monitoring, status management, metrics collection, and integration with logging and Prometheus.

---

### Features

- **Monitor abstraction**: Encapsulates health check logic, status, metrics, and configuration.
- **Custom health checks**: Register any function as a health check.
- **Status management**: Tracks status (`OK`, `Warn`, `KO`), rise/fall transitions, and error messages.
- **Metrics collection**: Latency, uptime, downtime, rise/fall times, and status.
- **Flexible configuration**: Control check intervals, timeouts, and thresholds.
- **Logger integration**: Pluggable logging for each monitor.
- **Cloning**: Clone monitors with new contexts.
- **Prometheus integration**: Register and collect custom metrics.
- **Thread-safe**: All operations are safe for concurrent use.

---

### Main Concepts

#### Monitor Creation

Create a monitor by providing a context and an `Info` object describing the monitored component.

```go
import (
    "github.com/nabbar/golib/monitor"
    "github.com/nabbar/golib/monitor/info"
)

inf, _ := info.New("MyComponent")
mon, err := monitor.New(nil, inf)
if err != nil {
    // handle error
}
```

#### Health Check Registration

Assign a health check function to the monitor. This function will be called periodically according to the configured intervals.

```go
mon.SetHealthCheck(func(ctx context.Context) error {
    // custom health check logic
    return nil // or return an error if unhealthy
})
```

#### Configuration

Configure the monitor with check intervals, timeouts, and thresholds for status transitions (rise/fall counts).

```go
cfg := monitor.Config{
    Name:          "MyMonitor",
    CheckTimeout:  5 * time.Second,
    IntervalCheck: 10 * time.Second,
    IntervalFall:  10 * time.Second,
    IntervalRise:  10 * time.Second,
    FallCountKO:   2,
    FallCountWarn: 1,
    RiseCountKO:   1,
    RiseCountWarn: 2,
    Logger:        /* logger options */,
}
mon.SetConfig(nil, cfg)
```

#### Status and Metrics

- **Status**: The monitor tracks its current status and transitions (rise/fall).
- **Metrics**: Latency, uptime, downtime, rise/fall times, and status are tracked and can be exported.

```go
status := mon.Status()      // Current status (OK, Warn, KO)
latency := mon.Latency()    // Last check latency
uptime := mon.Uptime()      // Total uptime
downtime := mon.Downtime()  // Total downtime
```

#### Lifecycle

- **Start**: Begin periodic health checks.
- **Stop**: Stop health checks.
- **Restart**: Restart the monitor.
- **IsRunning**: Check if the monitor is active.

```go
err := mon.Start(ctx)
defer mon.Stop(ctx)
```

#### Encoding and Export

Monitors can be encoded as text or JSON for reporting and integration.

```go
txt, _ := mon.MarshalText()
jsn, _ := mon.MarshalJSON()
```

---

### Metrics Integration

- Register custom metric names and collection functions for Prometheus integration.
- Collect latency, uptime, downtime, rise/fall times, and status.

```go
mon.RegisterMetricsName("my_metric")
mon.RegisterCollectMetrics(func(ctx context.Context, names ...string) {
    // custom Prometheus collection logic
})
```

---

### Error Handling

- Custom error codes for empty parameters, missing health checks, invalid config, logger errors, and timeouts.
- Errors are returned as descriptive error types.

---

### Notes

- All operations are thread-safe and suitable for concurrent use.
- Designed for Go 1.18+.
- Integrates with logging and Prometheus for observability.
- Can be used standalone or as part of a monitor pool.

---

### Example

```go
import (
    "github.com/nabbar/golib/monitor"
    "github.com/nabbar/golib/monitor/info"
    "context"
    "time"
)

inf, _ := info.New("API")
mon, _ := monitor.New(nil, inf)
mon.SetHealthCheck(func(ctx context.Context) error {
    // check API health
    return nil
})
cfg := monitor.Config{
    Name:          "API",
    CheckTimeout:  5 * time.Second,
    IntervalCheck: 30 * time.Second,
    // ...
}
mon.SetConfig(nil, cfg)
mon.Start(context.Background())
defer mon.Stop(context.Background())
```

---

## monitor/info Subpackage Documentation

The `monitor/info` subpackage provides types and utilities to describe and manage metadata for monitored components. It enables dynamic registration and retrieval of component names and additional information, supporting both static and runtime-generated data.

---

### Features

- Register custom functions to provide the component name and additional info dynamically.
- Store and retrieve metadata as key-value pairs.
- Thread-safe operations for concurrent environments.
- Encode info as string, bytes, text, or JSON for reporting and integration.

---

### Main Types

#### Info Interface

Defines the contract for managing component metadata:

- `RegisterName(FuncName)`: Register a function to provide the component name.
- `RegisterInfo(FuncInfo)`: Register a function to provide additional info as a map.
- `Name() string`: Retrieve the current name (from registered function or stored value).
- `Info() map[string]interface{}`: Retrieve the current info map (from registered function or stored values).

#### FuncName and FuncInfo

- `FuncName`: `func() (string, error)` — Function type to provide a name.
- `FuncInfo`: `func() (map[string]interface{}, error)` — Function type to provide info.

#### Encode Interface

- `String() string`: Returns a human-readable string representation.
- `Bytes() []byte`: Returns a byte slice representation.

---

### Usage

#### Creating an Info Object

```go
import "github.com/nabbar/golib/monitor/info"

inf, err := info.New("DefaultName")
if err != nil {
    // handle error
}
```

#### Registering Dynamic Name and Info

```go
inf.RegisterName(func() (string, error) {
    return "DynamicName", nil
})

inf.RegisterInfo(func() (map[string]interface{}, error) {
    return map[string]interface{}{
        "version": "1.0.0",
        "env":     "production",
    }, nil
})
```

#### Retrieving Name and Info

```go
name := inf.Name()
meta := inf.Info()
```

#### Encoding

- `inf.MarshalText()` and `inf.MarshalJSON()` provide text and JSON representations for integration and reporting.

---

### Notes

- If no dynamic function is registered, the default name and info are used.
- The subpackage is thread-safe and suitable for concurrent use.
- Designed for extensibility and integration with the main monitor system.

---

## monitor/status Subpackage Documentation

The `monitor/status` subpackage defines the status types and utilities for managing and encoding the health status of monitored components. It provides a simple, extensible way to represent, parse, and serialize status values for health checks and monitoring systems.

---

### Features

- Defines standard status values: `KO`, `Warn`, `OK`
- Conversion between status and string, integer, or float representations
- Parsing from string or integer to status
- JSON marshaling and unmarshaling support
- Thread-safe and lightweight

---

### Main Types

#### Status Type

Represents the health status as an unsigned 8-bit integer with three possible values:

- `KO` (default, value 0): Component is not operational
- `Warn` (value 1): Component is in a warning state
- `OK` (value 2): Component is healthy

##### Methods

- `String() string`: Returns the string representation (`"OK"`, `"Warn"`, or `"KO"`)
- `Int() int64`: Returns the integer value of the status
- `Float() float64`: Returns the float value of the status
- `MarshalJSON() ([]byte, error)`: Serializes the status as a JSON string
- `UnmarshalJSON([]byte) error`: Parses the status from a JSON string or integer

---

### Constructors

- `NewFromString(sts string) Status`: Parses a status from a string (`"OK"`, `"Warn"`, or any other string for `KO`)
- `NewFromInt(sts int64) Status`: Parses a status from an integer (returns `OK`, `Warn`, or `KO`)

---

### Usage Example

```go
import "github.com/nabbar/golib/monitor/status"

var s status.Status

s = status.NewFromString("OK")   // s == status.OK
s = status.NewFromInt(1)         // s == status.Warn

str := s.String()                // "Warn"
i := s.Int()                     // 1
f := s.Float()                   // 1.0

data, _ := s.MarshalJSON()       // "\"Warn\""
_ = s.UnmarshalJSON([]byte("\"OK\"")) // s == status.OK
```

---

### Notes

- If parsing fails or the value is unknown, the status defaults to `KO`.
- The status type is designed for easy integration with monitoring, alerting, and reporting systems.
- Supports both string and numeric representations for flexibility in configuration and serialization.

---

## Usage Example

```go
import (
    "github.com/nabbar/golib/monitor"
    "github.com/nabbar/golib/monitor/pool"
    // ... other imports
)

// Create a monitor
mon, err := monitor.New(ctx, info)
mon.SetHealthCheck(myHealthCheckFunc)
mon.SetConfig(ctx, myConfig)

// Create a pool and add monitors
p := pool.New(ctx)
p.MonitorAdd(mon)
p.Start(ctx)

// Collect metrics periodically
go p.TriggerCollectMetrics(ctx, time.Minute)
```

---

## Summary

- Use `monitor/pool` to manage multiple monitors as a group.
- Use `monitor` to define and run individual health checks.
- Integrate with logging and Prometheus for observability.
- Extend with `info` and `status` subpackages for richer metadata and status handling.

