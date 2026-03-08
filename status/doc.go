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
Package status provides a comprehensive health check and status monitoring system
for HTTP APIs. It integrates with the Gin web framework to expose status endpoints
that aggregate component health checks with flexible validation strategies.

# Overview

This package is designed for production-ready microservices, offering:
  - Flexible Validation: Multiple control modes (Must, Should, AnyOf, Quorum).
  - Performance: Built-in caching with atomic operations to reduce load.
  - Thread-Safety: Full concurrency support for all operations.
  - Multi-Format Output: JSON (default), plain text, and map-based formats.
  - Dynamic Configuration: Load mandatory components from configuration keys.
  - Integration: Works seamlessly with github.com/nabbar/golib/monitor.

# Architecture

The package is organized into focused subpackages with clear responsibilities:

	status/
	├── control/       # Validation mode definitions and logic.
	├── mandatory/     # Management for a single group of components.
	├── listmandatory/ # Management for a collection of component groups.
	├── interface.go   # Main Status interface.
	├── model.go       # Core implementation of the Status interface.
	├── config.go      # Configuration structures for the status system.
	├── cache.go       # Caching mechanism for health status.
	└── route.go       # HTTP endpoint handler for Gin.

# Component Overview

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

The following diagram illustrates how the status package processes a request,
computes the health status, and returns the response.

	[HTTP Request] (GET /status)
	      │
	      ▼
	[MiddleWare] (route.go)
	      │
	      ├─> Parse Query Params & Headers (short, format, map)
	      │
	      ▼
	[Status Computation] (model.go)
	      │
	      ├─> Check Cache (cache.go)
	      │     │
	      │     ├─> Valid? ───> Return Cached Status
	      │     │
	      │     └─> Invalid? ─┐
	      │                   │
	      │           [Walk Monitor Pool] (pool.go)
	      │                   │
	      │                   ▼
	      │           [Apply Control Modes] (control/mandatory)
	      │                   │
	      │             ┌─────┴─────┐
	      │             │           │
	      │        [Must/Should] [AnyOf/Quorum]
	      │             │           │
	      │             ▼           ▼
	      │        Check Indiv.   Check Group
	      │        Component      Logic
	      │             │           │
	      │             └─────┬─────┘
	      │                   │
	      │                   ▼
	      │           [Aggregate Status]
	      │           (OK / WARN / KO)
	      │                   │
	      │                   ▼
	      └─<── Update Cache ─┘
	      │
	      ▼
	[Response Encoding] (encode.go)
	      │
	      ├─> Format: JSON / Text
	      ├─> Verbosity: Full / Short
	      ├─> Structure: List / Map
	      │
	      ▼
	[HTTP Response] (Status Code + Body)

# Key Features

  - Component Monitoring: Aggregates health from multiple monitored components.
  - Control Modes: Ignore, Should, Must, AnyOf, Quorum.
  - Caching: Configurable cache duration (default 3s) with atomic operations.
  - Output Formats: JSON (full/short), Plain Text, Map Mode.
  - Gin Middleware: Drop-in integration for the Gin web framework.

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
		ctx := context.NewGlobal()
		sts := status.New(ctx)
		sts.SetInfo("my-app", "v1.0.0", "build-hash")

		// Register monitor pool
		monPool := pool.New(ctx)
		sts.RegisterPool(func() montps.Pool { return monPool })

		// Setup Gin
		r := gin.Default()
		r.GET("/status", func(c *gin.Context) {
			sts.MiddleWare(c)
		})
		r.Run(":8080")
	}

## Configuration

You can configure HTTP return codes and mandatory components, including dynamic
loading from component configurations.

	cfg := status.Config{
		ReturnCode: map[monsts.Status]int{
			monsts.OK:   200,
			monsts.Warn: 207,
			monsts.KO:   503,
		},
		Component: []status.Mandatory{
			{
				Mode: control.Must,
				Keys: []string{"database"}, // Static definition
			},
			{
				Mode:       control.Should,
				ConfigKeys: []string{"cache-component"}, // Dynamic loading
			},
		},
	}
	sts.SetConfig(cfg)

	// Register a resolver for dynamic loading.
	sts.RegisterGetConfigCpt(func(key string) cfgtypes.Component {
		return myConfig.ComponentGet(key)
	})

# Programmatic Health Checks

The package provides several methods to check the system's health programmatically.

## Live vs. Cached Checks

  - Live Checks: `IsHealthy()` and `IsStrictlyHealthy()` perform a real-time health
    assessment by checking all components. These calls can be resource-intensive if
    called frequently.
  - Cached Checks: `IsCacheHealthy()` and `IsCacheStrictlyHealthy()` return a result
    based on a short-lived cache (default 3 seconds). These are extremely fast (<10ns)
    and are ideal for high-frequency checks, such as in a middleware for every
    incoming request.

## Strict vs. Tolerant Checks

  - Tolerant Check (`IsHealthy`, `IsCacheHealthy`): Returns `true` if the global
    status is `OK` or `WARN`. This is useful for readiness probes where a degraded
    service is still considered available.
  - Strict Check (`IsStrictlyHealthy`, `IsCacheStrictlyHealthy`): Returns `true`
    only if the global status is `OK`. This is useful for liveness probes where any
    issue, including warnings, should signal a problem.

## Checking Specific Components

You can also check the health of one or more specific components. The control logic
is still applied.

	// Check if "database" and "cache" are healthy (respects Must, Should, etc.)
	if sts.IsHealthy("database", "cache") {
		// ...
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
