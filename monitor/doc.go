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
Package monitor provides a robust health check monitoring system with automatic status transitions,
configurable thresholds, and comprehensive metrics tracking.

# Overview

The monitor package implements a sophisticated health monitoring system that periodically executes
health checks and tracks the state of monitored components. It features:

  - Automatic status transitions with configurable thresholds (OK ↔ Warn ↔ KO)
  - Adaptive check intervals based on status (normal, rising, falling)
  - Comprehensive metrics (uptime, downtime, latency, rise/fall times)
  - Thread-safe concurrent operations
  - Prometheus metrics integration
  - Flexible configuration with validation
  - Middleware chain for extensibility

# Status Transitions

The monitor uses a three-state model with hysteresis to prevent flapping:

  - KO: Component is not healthy
  - Warn: Component is degraded but functional
  - OK: Component is fully healthy

Transitions between states require multiple consecutive successes or failures:

	KO --[riseCountKO successes]--> Warn --[riseCountWarn successes]--> OK
	OK --[fallCountWarn failures]--> Warn --[fallCountKO failures]--> KO

# Basic Usage

	import (
		"context"
		"time"
		"github.com/nabbar/golib/monitor"
		"github.com/nabbar/golib/monitor/info"
		"github.com/nabbar/golib/monitor/types"
		"github.com/nabbar/golib/duration"
	)

	// Create info metadata
	inf, err := info.New("database-monitor")
	if err != nil {
		log.Fatal(err)
	}

	// Create monitor
	mon, err := monitor.New(context.Background, inf)
	if err != nil {
		log.Fatal(err)
	}

	// Configure monitor
	cfg := types.Config{
		Name:          "database",
		CheckTimeout:  duration.ParseDuration(5 * time.Second),
		IntervalCheck: duration.ParseDuration(10 * time.Second),
		IntervalFall:  duration.ParseDuration(5 * time.Second),
		IntervalRise:  duration.ParseDuration(5 * time.Second),
		FallCountKO:   3,
		FallCountWarn: 2,
		RiseCountKO:   3,
		RiseCountWarn: 2,
	}
	if err := mon.SetConfig(context.Background, cfg); err != nil {
		log.Fatal(err)
	}

	// Register health check function
	mon.SetHealthCheck(func(ctx context.Context) error {
		// Check database connectivity
		return db.PingContext(ctx)
	})

	// Start monitoring
	if err := mon.Start(context.Background()); err != nil {
		log.Fatal(err)
	}
	defer mon.Stop(context.Background())

	// Query status
	fmt.Printf("Status: %s\n", mon.Status())
	fmt.Printf("Latency: %s\n", mon.Latency())
	fmt.Printf("Uptime: %s\n", mon.Uptime())

# Configuration

The monitor supports extensive configuration:

  - CheckTimeout: Maximum duration for a health check to complete (min: 5s)
  - IntervalCheck: Interval between checks in normal state (min: 1s)
  - IntervalFall: Interval when status is falling (min: 1s, default: IntervalCheck)
  - IntervalRise: Interval when status is rising (min: 1s, default: IntervalCheck)
  - FallCountKO: Failures needed to go from Warn to KO (min: 1)
  - FallCountWarn: Failures needed to go from OK to Warn (min: 1)
  - RiseCountKO: Successes needed to go from KO to Warn (min: 1)
  - RiseCountWarn: Successes needed to go from Warn to OK (min: 1)

All values are automatically normalized to their minimums if set below threshold.

# Metrics Tracking

The monitor tracks comprehensive timing metrics:

  - Latency: Duration of the last health check execution
  - Uptime: Total time in OK status
  - Downtime: Total time in KO or Warn status
  - RiseTime: Total time spent transitioning to better status
  - FallTime: Total time spent transitioning to worse status

# Prometheus Integration

The monitor can export metrics to Prometheus:

	import "github.com/nabbar/golib/prometheus"

	// Register metric names
	mon.RegisterMetricsName("my_service_health")

	// Register collection function
	mon.RegisterCollectMetrics(prometheusCollector)

	// Metrics are automatically collected after each health check

# Encoding Support

The monitor supports multiple encoding formats:

	// Text encoding
	text, _ := mon.MarshalText()
	fmt.Println(string(text))
	// Output: OK: database (version: 1.0) | 5ms / 1h30m / 0s

	// JSON encoding
	json, _ := mon.MarshalJSON()

# Thread Safety

All monitor operations are thread-safe and can be called concurrently from multiple goroutines.
The monitor uses fine-grained locking to minimize contention while ensuring data consistency.

# Best Practices

1. Configure appropriate check intervals to balance responsiveness and resource usage
2. Set fall/rise counts to prevent status flapping during temporary issues
3. Use shorter intervals during transitions (IntervalFall/Rise) for faster detection
4. Set CheckTimeout lower than IntervalCheck to prevent overlapping checks
5. Register a logger for debugging and troubleshooting
6. Always call Stop() when shutting down to clean up resources

# Error Handling

The monitor defines several error codes:

  - ErrorParamEmpty: Empty parameter provided
  - ErrorMissingHealthCheck: No health check function registered
  - ErrorValidatorError: Configuration validation failed
  - ErrorLoggerError: Logger initialization failed
  - ErrorTimeout: Operation timeout
  - ErrorInvalid: Invalid monitor instance

All errors implement the liberr.Error interface for structured error handling.

# Related Packages

  - github.com/nabbar/golib/monitor/info: Dynamic metadata management
  - github.com/nabbar/golib/monitor/status: Health status type
  - github.com/nabbar/golib/monitor/types: Type definitions and interfaces
  - github.com/nabbar/golib/monitor/pool: Monitor pool management
*/
package monitor
