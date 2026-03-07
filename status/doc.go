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
Package status provides a comprehensive health check and status monitoring system for HTTP APIs.
It integrates with the Gin web framework to expose status endpoints that aggregate component health checks
with flexible validation strategies.

# Overview

This package is designed for production-ready microservices, offering:
  - Flexible Validation: Multiple control modes (Must, Should, AnyOf, Quorum).
  - Performance: Built-in caching with atomic operations.
  - Thread-Safety: Full concurrency support.
  - Multi-Format: JSON (default) and plain text output.
  - Integration: Seamlessly works with github.com/nabbar/golib/monitor.

# Architecture

The package is organized into focused subpackages:

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

# Key Features

  - Component Monitoring: Aggregate health from multiple monitored components.
  - Control Modes: Ignore, Should, Must, AnyOf, Quorum.
  - Caching: Configurable cache duration (default 3s) with atomic operations.
  - Output Formats: JSON (full/short), Plain Text, Map Mode.
  - Gin Middleware: Drop-in integration for Gin.

# Usage

## Basic Setup

	import (
	    "github.com/gin-gonic/gin"
	    "github.com/nabbar/golib/context"
	    "github.com/nabbar/golib/status"
	    "github.com/nabbar/golib/monitor/pool"
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

You can configure return codes and mandatory components:

	cfg := status.Config{
	    ReturnCode: map[monsts.Status]int{
	        monsts.OK:   200,
	        monsts.Warn: 207,
	        monsts.KO:   503,
	    },
	    MandatoryComponent: []status.Mandatory{
	        {
	            Mode: control.Must,
	            Keys: []string{"database"},
	        },
	        {
	            Mode: control.Should,
	            Keys: []string{"cache"},
	        },
	    },
	}
	sts.SetConfig(cfg)

# Programmatic Health Checks

The package provides several methods to check the system's health programmatically.

## Live vs. Cached Checks

  - Live Checks: `IsHealthy()` and `IsStrictlyHealthy()` perform a real-time health assessment by checking all components. These calls can be resource-intensive if called frequently.
  - Cached Checks: `IsCacheHealthy()` and `IsCacheStrictlyHealthy()` return a result based on a short-lived cache (default 3 seconds). These are extremely fast (<10ns) and are ideal for high-frequency checks, such as in a middleware for every incoming request.

## Strict vs. Tolerant Checks

  - Tolerant Check (`IsHealthy`, `IsCacheHealthy`): Returns `true` if the global status is `OK` or `WARN`. This is useful for readiness probes where a degraded service is still considered available.
  - Strict Check (`IsStrictlyHealthy`, `IsCacheStrictlyHealthy`): Returns `true` only if the global status is `OK`. This is useful for liveness probes where any issue, including warnings, should signal a problem.

## Checking Specific Components

You can also check the health of one or more specific components. The control logic is still applied.

	// Check if database and cache are healthy (respects Must, Should, etc.)
	if sts.IsHealthy("database", "cache") {
	    // ...
	}

# HTTP API

The exposed endpoint supports several query parameters and headers:

  - Query "short=true" or Header "X-Verbose: false": Returns only overall status.
  - Query "format=text" or Header "Accept: text/plain": Returns plain text output.
  - Query "map=true" or Header "X-MapMode: true": Returns components as a map.

# Subpackages

The package is modularized to separate concerns:

  - control: Defines validation modes (Must, Should, etc.) and their encoding/decoding logic.
  - mandatory: Manages a single group of components associated with a specific validation mode.
  - listmandatory: Manages a collection of mandatory groups, allowing complex validation rules.

# Control Modes & Logic

The health of the application is determined by aggregating the status of monitored components
according to their assigned control mode.

## Available Modes

  - Ignore: The component's status is completely ignored. It does not affect the global status.
  - Should: The component is important but not critical.
  - If KO -> Global Status: WARN.
  - If WARN -> Global Status: WARN.
  - If OK -> Global Status: OK.
  - Must: The component is critical.
  - If KO -> Global Status: KO.
  - If WARN -> Global Status: WARN.
  - If OK -> Global Status: OK.
  - AnyOf: Used for redundant groups (e.g., a cluster).
  - If all components are KO -> Global Status: KO.
  - If at least one is OK/WARN -> Global Status: OK (or WARN if the best status is WARN).
  - Quorum: Used for distributed consensus groups.
  - If healthy count > 50% -> Global Status: OK (or WARN).
  - If healthy count <= 50% -> Global Status: KO.

## Status Transitions

The global status is calculated dynamically based on the worst-case scenario allowed by the configuration.

 1. OK: All `Must` components are OK, `Quorum` is satisfied, `AnyOf` has healthy candidates.
 2. WARN: A `Should` component is KO, or a `Must` component is WARN. The service is degraded but operational.
 3. KO: A `Must` component is KO, `Quorum` is lost, or `AnyOf` has no healthy candidates. The service is unavailable.

Transitions occur automatically as component health changes. The `IsHealthy()` method returns true for both OK and WARN states,
while `IsStrictlyHealthy()` returns true only for OK.
*/
package status
