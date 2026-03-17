# Status Package

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://golang.org/)

Package status provides a comprehensive, thread-safe, and high-performance health check and status monitoring system designed for production-grade services, particularly HTTP APIs.

It offers a robust framework for aggregating health metrics from various application components (such as databases, caches, and external services) and exposing them through a flexible and configurable HTTP endpoint, designed for seamless integration with the Gin web framework.

---

## Table of Contents

- [Key Features](#key-features)
- [Architecture](#architecture)
- [Data Flow and Logic](#data-flow-and-logic)
- [Core Concepts](#core-concepts)
- [Use Cases](#use-cases)
- [Usage](#usage)
- [Programmatic Health Checks](#programmatic-health-checks)
- [HTTP API Details](#http-api-details)
- [Subpackages](#subpackages)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Contributing](#contributing)
- [Resources](#resources)
- [AI Transparency Notice](#ai-transparency-notice)
- [License](#license)

---

## Key Features

- **Advanced Health Aggregation**: Go beyond simple "up/down" checks with sophisticated validation logic. Define critical and non-critical dependencies using control modes like `Must`, `Should`, `AnyOf`, and `Quorum`.
- **High-Performance Caching**: A built-in caching layer, powered by `atomic.Value`, provides lock-free reads of the system's health status. This minimizes performance impact and prevents "thundering herd" issues on downstream services during frequent health probes (e.g., from Kubernetes).
- **Thread-Safety**: All public API methods and internal state mutations are designed to be safe for concurrent access, making the package suitable for highly parallel applications.
- **Multi-Format Output**: Natively supports multiple response formats, including JSON (default), plain text (for simple parsers or command-line tools), and a structured map output, negotiated via HTTP headers or query parameters.
- **Dynamic Configuration**: Load mandatory component definitions from configuration keys at runtime, allowing the health monitoring system to adapt to changes in the application's environment without requiring a restart.
- **Filtering**: Supports filtering the status response to show only specific components or groups of components using wildcard patterns.
- **Ecosystem Integration**: Designed to work seamlessly with `github.com/nabbar/golib/monitor` for the actual health checking logic and `github.com/gin-gonic/gin` for HTTP endpoint exposure.

---

## Architecture

The `status` package is architecturally designed to be modular and extensible. It separates concerns into distinct sub-packages, each responsible for a specific part of the health monitoring logic. This modularity enhances maintainability and allows for a clear separation of concerns.

```
status/
├── control/       # Defines the validation logic and enum types for control modes (Must, Should, etc.).
├── mandatory/     # Manages a single group of components that share a specific validation mode.
├── listmandatory/ # Manages a collection of mandatory groups, enabling complex, multi-layered validation rules.
├── interface.go   # Defines the main public interfaces (Status, Route, Info, Pool).
├── model.go       # Contains the core implementation of the Status interface and the aggregation logic.
├── config.go      # Defines configuration structures and validation logic.
├── cache.go       # Implements the caching layer using atomic values for lock-free reads.
└── route.go       # Provides the HTTP handler and middleware for Gin integration.
```

### Component Interaction Diagram

The following diagram illustrates the high-level components and their interactions within the `status` ecosystem:

```
┌──────────────────────────────────────────────────────┐
│                  Status Package                      │
│  HTTP Endpoint + Component Health Aggregation        │
└──────────────┬────────────┬──────────────┬───────────┘
               │            │              │
      ┌────────▼───┐  ┌─────▼────┐  ┌──────▼──────┐
      │  control   │  │mandatory │  │listmandatory│
      │            │  │          │  │             │
      │ Validation │  │  Group   │  │  Group      │
      │   Modes    │  │ Manager  │  │ Collection  │
      └────────────┘  └──────────┘  └─────────────┘
               │            │              │
               └────────────┴──────────────┘
                            │
                  ┌─────────▼──────────┐
                  │  monitor/types     │
                  │  Component Health  │
                  └────────────────────┘
```

---

## Data Flow and Logic

The request processing flow is optimized for performance, with a fast path for cached responses and a slower path for live computations.

```
[HTTP Request] (GET /status)
      │
      ▼
[MiddleWare] (route.go)
      │
      ├─> Parse Query Params & Headers (short, format, map, filter)
      │   Determines verbosity, output format, and filters.
      │
      ▼
[Status Computation] (model.go)
      │
      ├─> Check Cache (cache.go)
      │     │
      │     ├─> Valid? ───> Return Cached Status (Fast Path)
      │     │               (Atomic read, < 10ns)
      │     │
      │     └─> Invalid? ─┐ (Slow Path)
	  │                   │
      │           [Walk Monitor Pool] (pool.go)
      │           Iterate over all registered monitors.
      │                   │
      │                   ▼
      │           [Apply Control Modes] (control/mandatory)
      │           Evaluate health based on configured rules.
      │                   │
      │             ┌─────┴─────┐
      │             │           │
      │        [Must/Should] [AnyOf/Quorum]
      │             │           │
      │             ▼           ▼
      │        Check Indiv.   Check Group
      │        Component      Logic (Thresholds)
      │             │           │
      │             └─────┬─────┘
      │                   │
      │                   ▼
      │           [Aggregate Status]
      │           Determine Global Status (OK / WARN / KO)
      │                   │
      │                   ▼
      └─<── Update Cache ─┘
            (Atomic write)
      │
      ▼
[Response Encoding] (encode.go)
      │
      ├─> Apply Filters (if any)
      ├─> Format: JSON / Text
      ├─> Verbosity: Full (details) / Short (status only)
      ├─> Structure: List / Map
      │
      ▼
[HTTP Response] (Status Code + Body)
```

---

## Core Concepts

### Component Monitoring

The system aggregates health status from multiple sources. Each source is a "Monitor" (defined in `github.com/nabbar/golib/monitor`) that performs the actual check (e.g., pinging a database or calling an external API). The `status` package consumes the results of these monitors.

### Control Modes

Control modes dictate how the failure of a specific component affects the global application status. This is the core of the validation logic.

- **Ignore**: The component is monitored, but its status is completely ignored in the global calculation. Useful for experimental features or non-critical background jobs.
- **Should**: Represents a non-critical dependency. Failure of a `Should` component results in a `WARN` global status, but not `KO`. This indicates a degraded but partially functional service.
- **Must**: Represents a critical dependency. Failure of a `Must` component results in a `KO` global status, indicating a service outage.
- **AnyOf**: Used for redundant groups (e.g., a cluster of services). The group is healthy if *at least one* component in the group is healthy.
- **Quorum**: Used for consensus-based groups. The group is healthy if *more than 50%* of the components in the group are healthy.

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

### Caching Strategy

To prevent performance bottlenecks from frequent health checks, the status is cached.
- **Duration**: The cache TTL is configurable, defaulting to 3 seconds.
- **Mechanism**: The implementation uses `atomic.Value` to store the computed status and its timestamp. This allows for lock-free reads in the hot path (i.e., when serving a cached response), making it extremely fast and scalable. A write lock is only acquired when the cache is stale and a new status needs to be computed.

### Output Formats

The HTTP endpoint is flexible and can serve responses in various formats:
- **JSON**: The default format, structured and easy to parse by machines.
- **Text**: A human-readable format, useful for quick checks with command-line tools like `curl` and `grep`.
- **Map Mode**: A variation of the JSON output where components are returned as a JSON map (keyed by component name) instead of a list. This can simplify parsing for clients that need to look up specific components.

---

## Use Cases

This package is designed for scenarios requiring robust health monitoring:

- **Microservices Health Checks**: Aggregate health from databases, caches, queues, external APIs. Configure critical (`Must`) vs optional (`Should`) dependencies. Return appropriate HTTP codes for load balancers and orchestrators.
- **Kubernetes/Docker Health Probes**:
  - **Liveness probe**: Use `IsStrictlyHealthy()` for restart signals.
  - **Readiness probe**: Use `IsHealthy()` to tolerate warnings.
  - **Startup probe**: Check with cached status for efficiency.
- **API Gateway Integration**: Expose `/health` and `/status` endpoints. JSON for programmatic consumption, text format for quick visual inspection.
- **Monitoring Systems**: Integrate with Prometheus, Datadog, New Relic. Cache reduces load on monitoring components. Detailed component status for diagnostics.
- **Distributed Systems**:
  - `AnyOf` mode: Redis cluster with multiple nodes (any healthy = OK).
  - `Quorum` mode: Database replicas (majority must be healthy).
  - `Must` mode: Core dependencies (all must be healthy).

---

## Usage

### Basic Setup

Here is a minimal example of how to set up the status monitoring system with Gin:

```go
import (
	"github.com/gin-gonic/gin"
	"github.com/nabbar/golib/context"
	"github.com/nabbar/golib/status"
	"github.com/nabbar/golib/monitor/pool"
	"github.com/nabbar/golib/monitor/types"
	"github.com/nabbar/golib/monitor/info"
	monsts "github.com/nabbar/golib/monitor/status"
	"context"
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
	dbMonitor := info.New(func(ctx context.Context) (monsts.Status, string, error) {
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

### Configuration

You can customize the behavior of the status endpoint, including HTTP return codes and mandatory component rules.

```go
import (
    "net/http"
    "github.com/nabbar/golib/status"
    "github.com/nabbar/golib/status/control"
    monsts "github.com/nabbar/golib/monitor/status"
    cfgtypes "github.com/nabbar/golib/config/types" // For dynamic config
)

func setupStatus() status.Status {
    ctx := context.NewGlobal() // Assuming ctx is available
    sts := status.New(ctx)
    sts.SetInfo("my-api", "v1.0.0", "abc123")
    
    // Configure HTTP return codes and mandatory components
    cfg := status.Config{
        ReturnCode: map[monsts.Status]int{
            monsts.OK:   http.StatusOK,           // 200
            monsts.Warn: http.StatusMultiStatus,  // 207
            monsts.KO:   http.StatusServiceUnavailable, // 503
        },
        Component: []status.Mandatory{
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

    // Example for dynamic component loading
    // sts.RegisterGetConfigCpt(func(key string) cfgtypes.Component {
    //     // Your logic to retrieve component by key from your config system.
    //     return myConfig.ComponentGet(key)
    // })
    
    return sts
}
```

---

## Programmatic Health Checks

In addition to the HTTP endpoint, the package provides methods for programmatic health checks, which are useful for internal application logic, startup sequences, or custom probes.

### Live vs. Cached Checks

- **Live Checks** (`IsHealthy()`, `IsStrictlyHealthy()`):
  These methods force a re-evaluation of all monitors, bypassing the cache. They provide the most up-to-date state but are more expensive. Use them when you need immediate confirmation of a state change, such as during application startup or before shutting down.

- **Cached Checks** (`IsCacheHealthy()`, `IsCacheStrictlyHealthy()`):
  These methods return the cached result if it is still valid (within the TTL). They are extremely fast (<10ns) and thread-safe, making them ideal for high-frequency endpoints like `/health` or `/status`.

### Strict vs. Tolerant Checks

- **Tolerant Check** (`IsHealthy`, `IsCacheHealthy`):
  Returns `true` if the global status is `OK` or `WARN`. This is suitable for **Readiness Probes**, where a degraded service (WARN) might still be able to serve some traffic.

- **Strict Check** (`IsStrictlyHealthy`, `IsCacheStrictlyHealthy`):
  Returns `true` *only* if the global status is `OK`. This is suitable for **Liveness Probes**, where you might want to restart the service if it's not fully healthy.

### Checking Specific Components

You can also check the health of one or more specific components. The control logic (Must/Should/etc.) associated with these components is still applied during the check.

```go
// Assuming 'sts' is your initialized status.Status instance
// Check if "database" and "cache" are healthy (OK or WARN).
if sts.IsHealthy("database", "cache") {
	// Proceed with logic that requires the DB and Cache.
}
```

---

## HTTP API Details

The exposed endpoint supports several query parameters and HTTP headers for content negotiation:

- **Short Response**: `short=true` (query) or `X-Verbose: false` (header). Returns only the overall status without the detailed component list.
- **Text Format**: `format=text` (query) or `Accept: text/plain` (header). Returns the response in plain text instead of JSON.
- **Map Mode**: `map=true` (query) or `X-MapMode: true` (header). Returns components as a map (keyed by name) instead of a list in the JSON response.

### Filtering

The response can be filtered to include only specific components using a comma-separated list of patterns. This is supported via:
- **Query Parameter**: `filter=pattern1,pattern2`
- **Header**: `X-Filter: pattern1,pattern2`

**Filtering Logic**:
1. **Mandatory Groups**: The filter is first applied to the names of configured mandatory groups. If any group matches, only components within those groups are returned.
2. **Monitor Names**: If no mandatory groups match, the filter is applied to individual monitor names in the pool.

**Patterns**: The patterns support standard shell-style wildcards (via `path.Match`):
- `*`: Matches any sequence of non-separator characters.
- `?`: Matches any single non-separator character.

**Example**:
- `?filter=db-*`: Returns all components/groups starting with "db-".
- `?filter=redis,mongo`: Returns components/groups named "redis" or "mongo".

---

## Subpackages

The `status` package is composed of several subpackages, each with a specific role:

- **`control`**: Defines the `Mode` type and associated functions for specifying how component health influences the overall application status. It provides mechanisms for parsing and marshaling these modes.
  - [GoDoc for control](https://pkg.go.dev/github.com/nabbar/golib/status/control)
- **`mandatory`**: Manages a single group of component keys that share a common validation `Mode`. It provides thread-safe operations for adding, removing, and querying component keys within a group.
  - [GoDoc for mandatory](https://pkg.go.dev/github.com/nabbar/golib/status/mandatory)
- **`listmandatory`**: Manages a collection of `mandatory` groups. It allows for complex, multi-layered validation rules by organizing multiple groups, each with its own mode and set of components.
  - [GoDoc for listmandatory](https://pkg.go.dev/github.com/nabbar/golib/status/listmandatory)

---

## API Reference

This section provides a high-level overview of the public API of the `status` package and its subpackages. For detailed documentation, refer to the [GoDoc](https://pkg.go.dev/github.com/nabbar/golib/status) links provided in the [Subpackages](#subpackages) section.

### `status` Package

The main `status` package provides the core `Status` interface and its constructor `New`.

#### `Status` Interface

The `Status` interface combines `Route`, `Info`, and `Pool` interfaces, along with methods for configuration and health checks.

- `New(ctx context.Context) Status`: Creates a new `Status` instance.
- `RegisterGetConfigCpt(fct FuncGetCfgCpt)`: Registers a function for dynamic component configuration lookup.
- `SetConfig(cfg Config)`: Applies the health check configuration.
- `GetConfig() Config`: Retrieves the current configuration.
- `IsHealthy(name ...string) bool`: Performs a live, tolerant health check.
- `IsStrictlyHealthy(name ...string) bool`: Performs a live, strict health check.
- `IsCacheHealthy() bool`: Performs a cached, tolerant health check.
- `IsCacheStrictlyHealthy() bool`: Performs a cached, strict health check.
- `MarshalText() ([]byte, error)`: Marshals the status to plain text.
- `MarshalJSON() ([]byte, error)`: Marshals the status to JSON.

#### `Config` and `Mandatory` Structs

These structs are used to configure the health check system.

```go
// Config defines the complete configuration for the status system.
type Config struct {
	// ReturnCode maps health statuses (OK, Warn, KO) to HTTP status codes.
	ReturnCode map[monsts.Status]int `mapstructure:"return-code"`

	// Info contains global descriptive metadata about the service (e.g., description, links).
	Info map[string]interface{} `mapstructure:"info"`

	// Component defines the list of mandatory component groups.
	Component []Mandatory `mapstructure:"component"`
}

// Mandatory defines a group of components that share a control mode and metadata.
type Mandatory struct {
	// Mode defines how the group affects the overall status (e.g., Must, Should).
	Mode stsctr.Mode `mapstructure:"mode"`

	// Name is a unique identifier for the group.
	Name string `mapstructure:"name"`

	// Info contains descriptive metadata about the group (e.g., description, runbook link).
	Info map[string]interface{} `mapstructure:"info"`

	// Keys is a list of static monitor names belonging to this group.
	Keys []string `mapstructure:"keys"`

	// ConfigKeys allows for dynamic resolution of monitor names from a config component.
	ConfigKeys []string `mapstructure:"configKeys,omitempty"`
}
```

#### `Route` Interface (embedded in `Status`)

Handles HTTP routing and response rendering.

- `Expose(ctx context.Context)`: Generic handler for status requests, compatible with `context.Context`.
- `MiddleWare(c *gin.Context)`: Gin middleware for processing status requests.
- `SetErrorReturn(f func() liberr.ReturnGin)`: Registers a custom error formatter for HTTP responses.

#### `Info` Interface (embedded in `Status`)

Manages application version and build information.

- `SetInfo(name, release, hash string)`: Manually sets application name, release, and build hash.
- `SetVersion(vers libver.Version)`: Sets application information from a `version.Version` object.

#### `Pool` Interface (embedded in `Status`)

Manages the collection of monitors.

- `RegisterPool(fct montps.FuncPool)`: Registers a function to provide the monitor pool.
- `MonitorAdd(mon montps.Monitor) error`: Adds a monitor to the pool.
- `MonitorGet(name string) montps.Monitor`: Retrieves a monitor by name.
- `MonitorSet(mon montps.Monitor) error`: Adds or updates a monitor.
- `MonitorDel(name string)`: Removes a monitor by name.
- `MonitorList() []string`: Returns a list of all monitor names.
- `MonitorWalk(fct func(name string, val montps.Monitor) bool, validName ...string)`: Iterates over monitors.

### `control` Subpackage

Defines the `Mode` type for validation strategies.

- `Mode`: An `uint8` type representing different control modes (`Ignore`, `Should`, `Must`, `AnyOf`, `Quorum`).
- `Parse(s string) Mode`: Converts a string to a `Mode` (case-insensitive).
- `ParseBytes(p []byte) Mode`: Converts a byte slice to a `Mode`.
- `ParseUint64(p uint64) Mode`: Converts a `uint64` to a `Mode`.
- `ParseUint32(p uint32) Mode`: Converts a `uint32` to a `Mode`.
- `ParseUint16(p uint16) Mode`: Converts a `uint16` to a `Mode`.
- `ParseUint8(p uint8) Mode`: Converts a `uint8` to a `Mode`.
- `ParseInt64(p int64) Mode`: Converts an `int64` to a `Mode`.
- `ParseInt32(p int32) Mode`: Converts an `int32` to a `Mode`.
- `ParseInt16(p int16) Mode`: Converts an `int16` to a `Mode`.
- `ParseInt8(p int8) Mode`: Converts an `int8` to a `Mode`.

### `mandatory` Subpackage

Defines the `Mandatory` interface for managing a single group of components.

- `New() Mandatory`: Creates a new `Mandatory` instance.
- `SetMode(m stsctr.Mode)`: Sets the validation mode for the group.
- `GetMode() stsctr.Mode`: Retrieves the current validation mode.
- `SetName(s string)`: Sets the name of the group.
- `GetName() string`: Retrieves the name of the group.
- `KeyHas(key string) bool`: Checks if a component key is in the group.
- `KeyAdd(keys ...string)`: Adds component keys to the group.
- `KeyDel(keys ...string)`: Removes component keys from the group.
- `KeyList() []string`: Returns a list of all component keys in the group.

### `listmandatory` Subpackage

Defines the `ListMandatory` interface for managing a collection of mandatory groups.

- `New(m ...stsmdt.Mandatory) ListMandatory`: Creates a new `ListMandatory` instance, optionally with initial groups.
- `Len() int`: Returns the number of mandatory groups in the list.
- `Walk(fct func(k string, m stsmdt.Mandatory) bool)`: Iterates through all mandatory groups.
- `Add(m ...stsmdt.Mandatory)`: Adds one or more mandatory groups to the list.
- `Del(m stsmdt.Mandatory)`: Removes a mandatory group by content matching.
- `DelKey(s string)`: Removes a mandatory group by its name.
- `GetMode(key string) stsctr.Mode`: Returns the validation mode for a specific component key from the first matching group.
- `SetMode(key string, mod stsctr.Mode)`: Updates the validation mode for the first group containing a specific key.
- `GetList() []stsmdt.Mandatory`: Returns a slice of all mandatory groups.

---

## Testing

The package is extensively tested to ensure reliability and thread-safety.

- **Total Specifications**: 383
- **Total Assertions**: 723
- **Average Coverage**: 88.55%
- **Race Detection**: All tests are passing with the `-race` flag.

See [TESTING.md](TESTING.md) for detailed testing documentation, including how to run tests and interpret the results.

---

## Contributing

Contributions are welcome! Please follow these guidelines:

- **Code Style**: Follow existing code style and patterns.
- **Testing**: Maintain or improve test coverage (≥80%). All contributions must pass `go test -race`.
- **Documentation**: Update `README.md` and `doc.go` for new features.
- **Pull Requests**: Provide a clear description of changes and reference any related issues.

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for detailed guidelines.

---

## Resources

This section provides links to external, authoritative documentation related to the concepts implemented in this package.

- **Health Check Patterns**:
  - [Kubernetes Probes Documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/): The official documentation from Kubernetes, which defines the industry-standard Liveness, Readiness, and Startup probe patterns that this package is designed to support.
- **API Health Check Standard**:
  - [IETF Draft: Health Check Response Format for HTTP APIs](https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check): An ongoing effort to standardize the JSON response format for health check endpoints. This package is conceptually aligned with this draft.
- **Site Reliability Engineering (SRE) Concepts**:
  - [Google SRE Book - Monitoring Distributed Systems](https://sre.google/sre-book/monitoring-distributed-systems/): The canonical resource for understanding the role of monitoring in reliability. It explains the concepts of SLI, SLO, and SLA, for which this package provides the foundational data (SLIs).
- **HTTP Status Codes**:
  - [MDN Web Docs: HTTP response status codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status): A clear and authoritative reference for HTTP status codes, including `200 OK`, `503 Service Unavailable`, and `207 Multi-Status`, which are commonly used in health check responses.

---

## AI Transparency Notice

In accordance with Article 50.4 of the EU AI Act, AI assistance has been used for testing, documentation, and bug fixing under human supervision.

---

## License

MIT License - See [LICENSE](../../LICENSE) file for details.
