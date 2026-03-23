/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

/*
Package status provides a comprehensive, thread-safe, and high-performance health check
and status monitoring system designed for production-grade services, particularly HTTP APIs.

It offers a robust framework for aggregating health metrics from various application
components (such as databases, caches, and external services) and exposing them through
a flexible and configurable HTTP endpoint, designed for seamless integration with the
Gin web framework.

# Key Features

  - **Advanced Health Aggregation**: Go beyond simple "up/down" checks with sophisticated
    validation logic. Define critical and non-critical dependencies using control modes
    like `Must`, `Should`, `AnyOf`, and `Quorum`.
  - **High-Performance Caching**: A built-in caching layer, powered by `atomic.Value`,
    provides lock-free reads of the system's health status. This minimizes performance
    impact and prevents "thundering herd" issues on downstream services during frequent
    health probes (e.g., from Kubernetes).
  - **Thread-Safety**: All public API methods and internal state mutations are designed
    to be safe for concurrent access, making the package suitable for highly parallel
    applications.
  - **Rich Metadata**: Enrich status responses with descriptive information, including
    global service details (description, links) and component-specific metadata,
    to provide context for operators and monitoring systems.
  - **Multi-Format Output**: Natively supports multiple response formats, including JSON (default),
    plain text (for simple parsers or command-line tools), and a structured map output,
    negotiated via HTTP headers or query parameters.
  - **Dynamic Configuration**: Load mandatory component definitions from configuration keys
    at runtime, allowing the health monitoring system to adapt to changes in the application's
    environment without requiring a restart.
  - **Filtering**: Supports filtering the status response to show only specific components
    or groups of components using wildcard patterns.
  - **Ecosystem Integration**: Designed to work seamlessly with `github.com/nabbar/golib/monitor`
    for the actual health checking logic and `github.com/gin-gonic/gin` for HTTP endpoint exposure.

# Architecture

The `status` package is architecturally designed to be modular and extensible. It separates
concerns into distinct sub-packages, each responsible for a specific part of the health
monitoring logic. This modularity enhances maintainability and allows for a clear separation
of concerns.

	status/
	├── control/       # Defines the validation logic and enum types for control modes (Must, Should, etc.).
	├── mandatory/     # Manages a single group of components that share a specific validation mode.
	├── listmandatory/ # Manages a collection of mandatory groups, enabling complex, multi-layered validation rules.
	├── interface.go   # Defines the main public interfaces (Status, Route, Info, Pool).
	├── model.go       # Contains the core implementation of the Status interface and the aggregation logic.
	├── config.go      # Defines configuration structures and validation logic.
	├── cache.go       # Implements the caching layer using atomic values for lock-free reads.
	└── route.go       # Provides the HTTP handler and middleware for Gin integration.

## Component Interaction Diagram

The following diagram illustrates the high-level components and their interactions within the
`status` ecosystem:

	┌──────────────────────────────────────────────────────┐
	│                  Status Package                      │
	│  HTTP Endpoint + Component Health Aggregation        │
	└──────────────┬────────────┬──────────────┬───────────┘
	               │            │              │
	      ┌────────▼───┐  ┌────▼─────┐  ┌────▼────────┐
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

# Data Flow and Logic

The request processing flow is optimized for performance, with a fast path for cached
responses and a slower path for live computations.

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

# Core Concepts

## Component Monitoring

The system aggregates health status from multiple sources. Each source is a "Monitor"
(defined in `github.com/nabbar/golib/monitor`) that performs the actual check (e.g.,
pinging a database or calling an external API). The `status` package consumes the results
of these monitors.

## Control Modes

Control modes dictate how the failure of a specific component affects the global
application status. This is the core of the validation logic.

  - **Ignore**: The component is monitored, but its status is completely ignored in the
    global calculation. Useful for experimental features or non-critical background jobs.
  - **Should**: Represents a non-critical dependency. Failure of a `Should` component
    results in a `WARN` global status, but not `KO`. This indicates a degraded but
    partially functional service.
  - **Must**: Represents a critical dependency. Failure of a `Must` component results
    in a `KO` global status, indicating a service outage.
  - **AnyOf**: Used for redundant groups (e.g., a cluster of services). The group is
    healthy if *at least one* component in the group is healthy.
  - **Quorum**: Used for consensus-based groups. The group is healthy if *more than 50%*
    of the components in the group are healthy.

## Caching Strategy

To prevent performance bottlenecks from frequent health checks, the status is cached.
  - **Duration**: The cache TTL is configurable, defaulting to 3 seconds.
  - **Mechanism**: The implementation uses `atomic.Value` to store the computed status
    and its timestamp. This allows for lock-free reads in the hot path (i.e., when serving
    a cached response), making it extremely fast and scalable. A write lock is only
    acquired when the cache is stale and a new status needs to be computed.

## Output Formats

The HTTP endpoint is flexible and can serve responses in various formats:
  - **JSON**: The default format, structured and easy to parse by machines.
  - **Text**: A human-readable format, useful for quick checks with command-line tools
    like `curl` and `grep`.
  - **Map Mode**: A variation of the JSON output where components are returned as a
    JSON map (keyed by component name) instead of a list. This can simplify parsing
    for clients that need to look up specific components.

# Usage

## Basic Setup

Here is a minimal example of how to set up the status monitoring system with Gin:

	import (
		"github.com/gin-gonic/gin"
		"github.com/nabbar/golib/context"
		"github.com/nabbar/golib/status"
		"github.com/nabbar/golib/monitor/pool"
		"github.com/nabbar/golib/monitor/types"
	)

	func main() {
		// Initialize the global context for the application.
		ctx := context.NewGlobal()

		// Create a new status manager.
		sts := status.New(ctx)
		sts.SetInfo("my-app", "v1.0.0", "build-hash")

		// Create and register the monitor pool. The status package will use this
		// pool to get the health of individual components.
		monPool := pool.New(ctx)
		sts.RegisterPool(func() montps.Pool { return monPool })

		// Setup Gin router.
		r := gin.Default()

		// Register the status middleware on a chosen endpoint.
		r.GET("/status", func(c *gin.Context) {
			sts.MiddleWare(c)
		})

		r.Run(":8080")
	}

## Configuration

You can customize the behavior of the status endpoint, including HTTP return codes
and mandatory component rules.

	cfg := status.Config{
		// Map internal status to HTTP status codes.
		ReturnCode: map[monsts.Status]int{
			monsts.OK:   200, // OK
			monsts.Warn: 207, // Multi-Status
			monsts.KO:   503, // Service Unavailable
		},
		// Define mandatory component groups.
		Component: []status.Mandatory{
			{
				Mode: control.Must,
				Keys: []string{"database"}, // Static definition: "database" must be up.
			},
			{
				Mode:       control.Should,
				ConfigKeys: []string{"cache-component"}, // Dynamic: load keys from a config key.
			},
		},
	}
	sts.SetConfig(cfg)

	// To use ConfigKeys, you must register a resolver function that can fetch
	// the component configuration by its key.
	sts.RegisterGetConfigCpt(func(key string) cfgtypes.Component {
		// Your logic to retrieve component by key from your config system.
		return myConfig.ComponentGet(key)
	})

# Programmatic Health Checks

In addition to the HTTP endpoint, the package provides methods for programmatic health
checks, which are useful for internal application logic, startup sequences, or custom probes.

## Live vs. Cached Checks

  - **Live Checks** (`IsHealthy()`, `IsStrictlyHealthy()`):
    These methods force a re-evaluation of all monitors, bypassing the cache. They provide
    the most up-to-date state but are more expensive. Use them when you need immediate
    confirmation of a state change, such as during application startup or before shutting down.

  - **Cached Checks** (`IsCacheHealthy()`, `IsCacheStrictlyHealthy()`):
    These methods return the cached result if it is still valid (within the TTL). They are
    extremely fast (<10ns) and thread-safe, making them ideal for high-frequency endpoints
    like `/health` or `/status`.

## Strict vs. Tolerant Checks

  - **Tolerant Check** (`IsHealthy`, `IsCacheHealthy`):
    Returns `true` if the global status is `OK` or `WARN`. This is suitable for **Readiness Probes**,
    where a degraded service (WARN) might still be able to serve some traffic.

  - **Strict Check** (`IsStrictlyHealthy`, `IsCacheStrictlyHealthy`):
    Returns `true` *only* if the global status is `OK`. This is suitable for **Liveness Probes**,
    where you might want to restart the service if it's not fully healthy.

## Checking Specific Components

You can also check the health of one or more specific components. The control logic
(Must/Should/etc.) associated with these components is still applied during the check.

	// Check if "database" and "cache" are healthy (OK or WARN).
	if sts.IsHealthy("database", "cache") {
		// Proceed with logic that requires the DB and Cache.
	}

# HTTP API Details

The exposed endpoint supports several query parameters and HTTP headers for content negotiation:

  - **Short Response**: `short=true` (query) or `X-Verbose: false` (header). Returns only the
    overall status without the detailed component list.
  - **Text Format**: `format=text` (query) or `Accept: text/plain` (header). Returns the
    response in plain text instead of JSON.
  - **Map Mode**: `map=true` (query) or `X-MapMode: true` (header). Returns components as a
    map (keyed by name) instead of a list in the JSON response.

## Filtering

The response can be filtered to include only specific components using a comma-separated list
of patterns. This is supported via:
  - **Query Parameter**: `filter=pattern1,pattern2`
  - **Header**: `X-Filter: pattern1,pattern2`

**Filtering Logic**:
 1. **Mandatory Groups**: The filter is first applied to the names of configured mandatory groups.
    If any group matches, only components within those groups are returned.
 2. **Monitor Names**: If no mandatory groups match, the filter is applied to individual monitor
    names in the pool.

**Patterns**: The patterns support standard shell-style wildcards (via `path.Match`):
  - `*`: Matches any sequence of non-separator characters.
  - `?`: Matches any single non-separator character.

**Example**:
  - `?filter=db-*`: Returns all components/groups starting with "db-".
  - `?filter=redis,mongo`: Returns components/groups named "redis" or "mongo".

# Status Aggregation Logic

The global health status is determined by aggregating the status of all monitored components,
factoring in their assigned control modes. The final status is the most severe status
encountered, following this hierarchy: `KO` > `WARN` > `OK`.

 1. **Initialization**: The global status starts at `OK`.
 2. **Iteration**: The system iterates through all registered monitors.
 3. **Mode Application**: For each component, its control mode is determined.
    - If a component is part of a `Must` group and its status is `KO`, the global status
    immediately becomes `KO`.
    - If a component is part of a `Must` or `Should` group and its status is `WARN`, the
    global status is elevated to `WARN` (if it was previously `OK`).
    - `AnyOf` and `Quorum` groups are evaluated based on their specific rules.
 4. **Finalization**: The final aggregated status is cached and returned.

The `IsHealthy()` method returns `true` for both `OK` and `WARN` states, reflecting readiness,
while `IsStrictlyHealthy()` returns `true` only for the `OK` state, reflecting liveness.
*/
package status
