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
 *
 */

/*
Package status provides a comprehensive, thread-safe health check and status monitoring system
designed for production-grade HTTP APIs.

It integrates seamlessly with the Gin web framework to expose status endpoints that aggregate
health metrics from various application components (databases, caches, external services)
using flexible and configurable validation strategies.

# Overview

This package is built to address the needs of microservices and distributed systems where
simple "up/down" checks are often insufficient. It offers:

  - **Flexible Validation**: Support for complex dependency rules via control modes (Must, Should, AnyOf, Quorum).
  - **High Performance**: Built-in caching mechanism using atomic operations to minimize lock contention and reduce load on downstream services during frequent health checks (e.g., Kubernetes probes).
  - **Thread-Safety**: All public API methods and internal state mutations are safe for concurrent access.
  - **Multi-Format Output**: Native support for JSON (default), plain text (for simple parsers), and structured map outputs.
  - **Dynamic Configuration**: Capability to load mandatory component definitions dynamically from configuration keys, allowing for runtime adaptability.
  - **Ecosystem Integration**: Designed to work hand-in-hand with `github.com/nabbar/golib/monitor` for the actual health checking logic.

# Architecture

The package is structured into modular subpackages to separate concerns and improve maintainability:

	status/
	├── control/       # Defines the validation logic and enum types for control modes (Must, Should, etc.).
	├── mandatory/     # Manages a single group of components that share a specific validation mode.
	├── listmandatory/ # Manages a collection of mandatory groups, enabling complex, multi-layered validation rules.
	├── interface.go   # Defines the main public interfaces (Status, Route, Info, Pool).
	├── model.go       # Contains the core implementation of the Status interface and the aggregation logic.
	├── config.go      # Defines configuration structures and validation logic.
	├── cache.go       # Implements the caching layer using atomic values for lock-free reads.
	└── route.go       # Provides the HTTP handler and middleware for Gin integration.

# Component Overview

The following diagram illustrates the high-level components and their interactions:

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

# Data Flow & Logic

The following diagram details the request processing flow, from the incoming HTTP request
to the final response, highlighting the caching and computation steps:

	[HTTP Request] (GET /status)
	      │
	      ▼
	[MiddleWare] (route.go)
	      │
	      ├─> Parse Query Params & Headers (short, format, map)
	      │   Determines verbosity and output format.
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
	      ├─> Format: JSON / Text
	      ├─> Verbosity: Full (details) / Short (status only)
	      ├─> Structure: List / Map
	      │
	      ▼
	[HTTP Response] (Status Code + Body)

# Key Features Detail

## Component Monitoring
The system aggregates health status from multiple sources.
Each source is a "Monitor" (defined in `github.com/nabbar/golib/monitor`) that performs the actual check (e.g., pinging a DB).

## Control Modes
Control modes dictate how the failure of a specific component affects the global application status:
  - **Ignore**: The component is monitored, but its status is completely ignored in the global calculation.
  - **Should**: Non-critical dependency. Failure results in a `WARN` global status, but not `KO`.
  - **Must**: Critical dependency. Failure results in a `KO` global status.
  - **AnyOf**: Redundancy check. The group is healthy if *at least one* component is healthy.
  - **Quorum**: Consensus check. The group is healthy if *more than 50%* of components are healthy.

## Caching
To prevent "thundering herd" problems or excessive load on dependencies during high-frequency
health checks (like Kubernetes liveness probes), the status is cached.
  - **Duration**: Configurable, defaults to 3 seconds.
  - **Mechanism**: Uses `atomic.Value` to store the result and timestamp, allowing for lock-free reads in the hot path.

## Output Formats
  - **JSON**: Standard format, easy to parse.
  - **Text**: Human-readable, useful for command-line tools (curl/grep).
  - **Map Mode**: Returns components as a JSON map (keyed by name) instead of a list, facilitating direct access by component name.

# Usage

## Basic Setup

	import (
		"github.com/gin-gonic/gin"
		"github.com/nabbar/golib/context"
		"github.com/nabbar/golib/status"
		"github.com/nabbar/golib/monitor/pool"
		"github.com/nabbar/golib/monitor/types"
	)

	func main() {
		// Initialize the global context
		ctx := context.NewGlobal()

		// Create the status manager
		sts := status.New(ctx)
		sts.SetInfo("my-app", "v1.0.0", "build-hash")

		// Create and register the monitor pool
		monPool := pool.New(ctx)
		sts.RegisterPool(func() montps.Pool { return monPool })

		// Setup Gin router
		r := gin.Default()

		// Register the status middleware
		r.GET("/status", func(c *gin.Context) {
			sts.MiddleWare(c)
		})

		r.Run(":8080")
	}

## Configuration

You can configure HTTP return codes and mandatory components.
This example shows how to mix static definitions with dynamic loading.

	cfg := status.Config{
		// Map internal status to HTTP status codes
		ReturnCode: map[monsts.Status]int{
			monsts.OK:   200,
			monsts.Warn: 207, // Multi-Status
			monsts.KO:   503, // Service Unavailable
		},
		// Define mandatory component groups
		Component: []status.Mandatory{
			{
				Mode: control.Must,
				Keys: []string{"database"}, // Static definition: "database" must be up
			},
			{
				Mode:       control.Should,
				ConfigKeys: []string{"cache-component"}, // Dynamic: load keys from config "cache-component"
			},
		},
	}
	sts.SetConfig(cfg)

	// Register a resolver for dynamic loading (ConfigKeys)
	sts.RegisterGetConfigCpt(func(key string) cfgtypes.Component {
		// Logic to retrieve component by key from your config system
		return myConfig.ComponentGet(key)
	})

# Programmatic Health Checks

The package provides several methods to check the system's health programmatically, useful for internal logic or custom probes.

## Live vs. Cached Checks

  - **Live Checks** (`IsHealthy()`, `IsStrictlyHealthy()`):
    These methods force a re-evaluation of all monitors. They provide the most up-to-date
    state but are more expensive. Use them when you need immediate confirmation of a state
    change (e.g., during startup or shutdown).

  - **Cached Checks** (`IsCacheHealthy()`, `IsCacheStrictlyHealthy()`):
    These methods return the cached result if it is still valid (within the TTL).
    They are extremely fast (<10ns) and thread-safe. Use them for high-frequency endpoints
    like `/health` or `/status`.

## Strict vs. Tolerant Checks

  - **Tolerant Check** (`IsHealthy`, `IsCacheHealthy`):
    Returns `true` if the global status is `OK` or `WARN`. This is suitable for **Readiness Probes**,
    where a degraded service (WARN) might still be able to serve some traffic.

  - **Strict Check** (`IsStrictlyHealthy`, `IsCacheStrictlyHealthy`):
    Returns `true` *only* if the global status is `OK`. This is suitable for **Liveness Probes**,
    where you might want to restart the service if it's not fully healthy (depending on your restart policy).

## Checking Specific Components

You can also check the health of one or more specific components. The control logic (Must/Should/etc.) associated
with these components is still applied during the check.

	// Check if "database" and "cache" are healthy
	if sts.IsHealthy("database", "cache") {
		// Proceed with logic requiring DB and Cache
	}

# HTTP API

The exposed endpoint supports several query parameters and headers for content negotiation:

  - `short=true` (query) or `X-Verbose: false` (header): Returns only the overall
    status without component details.
  - `format=text` (query) or `Accept: text/plain` (header): Returns plain text output.
  - `map=true` (query) or `X-MapMode: true` (header): Returns components as a map
    instead of a list.

# Subpackages

The package is modularized to separate concerns:

  - `control`: Defines validation modes (e.g., Must, Should) and their
    encoding/decoding logic. It is the brain of the validation strategy.
  - `mandatory`: Manages a single group of components associated with a specific
    validation mode. It is a thread-safe container for a list of component keys
    and their control mode.
  - `listmandatory`: Manages a collection of `mandatory` groups, allowing for the
    creation of complex validation rules across different sets of components.

# Control Modes & Logic

The health of the application is determined by aggregating the status of monitored
components according to their assigned control mode.

## Available Modes

  - `Ignore`: The component's status is completely ignored and does not affect the
    global status.
  - `Should`: The component is important but not critical. A failure (`KO`) or
    warning (`WARN`) in this component will result in a global `WARN` status, but
    not `KO`.
  - `Must`: The component is critical. A failure (`KO`) will result in a global
    `KO` status. A warning (`WARN`) will result in a global `WARN`.
  - `AnyOf`: Used for redundant groups (e.g., a cluster of services). The global
    status will be `KO` only if all components in the group are `KO`. Otherwise,
    the group's status is determined by the best-status member.
  - `Quorum`: Used for distributed consensus groups. The global status will be `KO`
    if more than 50% of the components in the group are `KO`.

## Status Transitions

The global status is calculated dynamically based on the worst-case scenario
allowed by the configuration.

 1. `OK`: All `Must` components are `OK`, `Quorum` is satisfied, and `AnyOf` has
    healthy candidates.
 2. `WARN`: A `Should` component is `KO`, or a `Must` component is `WARN`. The
    service is considered degraded but still operational.
 3. `KO`: A `Must` component is `KO`, `Quorum` is lost, or an `AnyOf` group has no
    healthy candidates. The service is considered unavailable.

Transitions occur automatically as component health changes. The `IsHealthy()` method
returns `true` for both `OK` and `WARN` states, while `IsStrictlyHealthy()` returns
`true` only for the `OK` state.
*/
package status
