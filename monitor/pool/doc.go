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
Package pool provides a thread-safe pool implementation for managing multiple health monitors.

# Overview

The pool package extends the monitor functionality by providing a centralized
management system for multiple monitor instances. It offers lifecycle management,
metrics collection, shell command interfaces, and various encoding formats for
monitor data.

# Key Features

  - Thread-safe monitor management (add, get, set, delete, list, walk)
  - Lifecycle operations (start, stop, restart) for all monitors
  - Prometheus metrics integration with comprehensive health metrics
  - Shell command interface for operational management
  - JSON and text encoding support
  - Context-aware operations with proper cancellation handling

# Basic Usage

	import (
		"context"
		"time"

		libctx "github.com/nabbar/golib/context"
		libmon "github.com/nabbar/golib/monitor"
		moninf "github.com/nabbar/golib/monitor/info"
		monpool "github.com/nabbar/golib/monitor/pool"
		libprm "github.com/nabbar/golib/prometheus"
	)

	// Create context provider
	ctx := context.Background()
	ctxFunc := func() context.Context { return ctx }

	// Create a pool
	pool := monpool.New(ctxFunc)

	// Create and add monitors
	info, _ := moninf.New("service-1")
	monitor, _ := libmon.New(ctxFunc, info)
	monitor.SetHealthCheck(func(ctx context.Context) error {
		// Your health check logic
		return nil
	})

	pool.MonitorAdd(monitor)

	// Start all monitors
	pool.Start(ctx)

	// Get monitor status
	mon := pool.MonitorGet("service-1")
	if mon != nil {
		status := mon.Status()
		fmt.Printf("Service status: %s\n", status)
	}

# Metrics Collection

The pool supports comprehensive Prometheus metrics collection:

	// Initialize Prometheus
	prom := libprm.New(ctxFunc)

	// Register metrics with pool
	pool.InitMetrics(func() libprm.Prometheus {
		return prom
	}, loggerFunc)

	// Trigger periodic metrics collection
	go pool.TriggerCollectMetrics(ctx, 30*time.Second)

# Available Metrics

The pool automatically collects the following metrics for all monitors:

  - monitor_latency: Health check execution time (histogram)
  - monitor_uptime: Total uptime in seconds (gauge)
  - monitor_downtime: Total downtime in seconds (gauge)
  - monitor_risetime: Time spent in rising state (gauge)
  - monitor_falltime: Time spent in falling state (gauge)
  - monitor_status: Current health status (gauge)
  - monitor_rise: Rising indicator (gauge, 0 or 1)
  - monitor_fall: Falling indicator (gauge, 0 or 1)
  - monitor_sli: Service Level Indicator with mean/min/max (gauge)

# Shell Commands

The pool provides interactive shell commands for operational management:

	// Get shell commands
	commands := pool.GetShellCommand(ctx)

	// Available commands:
	// - list: Print the monitors' list
	// - info: Print information about monitors
	// - start: Start monitors
	// - stop: Stop monitors
	// - restart: Restart monitors
	// - status: Print status & message for monitors

# Encoding and Serialization

The pool supports multiple encoding formats:

	// Marshal to JSON
	jsonData, err := pool.MarshalJSON()

	// Marshal to text
	textData, err := pool.MarshalText()

# Thread Safety

All pool operations are thread-safe and can be safely called from multiple
goroutines concurrently. The pool uses internal synchronization to protect
shared state.

# Monitor Lifecycle

Monitors in the pool follow this lifecycle:

 1. Created and added to pool (MonitorAdd)
 2. Optionally configured with health checks and config
 3. Started individually or via pool.Start()
 4. Running: periodic health checks executed
 5. Stopped via pool.Stop() or individually
 6. Removed from pool (MonitorDel)

# Context Handling

The pool respects context cancellation and timeouts:

  - All operations accept a context parameter
  - Monitor health checks use the provided context
  - Metrics collection can be cancelled via context
  - Graceful shutdown on context cancellation

# Error Handling

Operations return errors when appropriate:

  - MonitorAdd: returns error if monitor has empty name or fails to start when pool is running
  - MonitorSet: returns error if monitor is nil or has empty name
  - Start/Stop/Restart: returns aggregated errors from all monitor operations
  - InitMetrics: returns error if metric registration fails

# Best Practices

  - Always call Stop() to clean up resources when done
  - Use context with timeout for lifecycle operations
  - Register Prometheus and logger functions before InitMetrics
  - Handle errors from lifecycle operations
  - Use MonitorWalk for bulk operations instead of iterating manually
  - Clean up monitors (defer mon.Stop(ctx)) when adding them

# Performance Considerations

  - The pool is designed for efficient concurrent access
  - Monitor operations scale linearly with the number of monitors
  - Metrics collection is optimized for large monitor counts
  - Use TriggerCollectMetrics with appropriate intervals (30s-60s recommended)

# Integration with Other Packages

The pool integrates with:

  - github.com/nabbar/golib/monitor: Core monitor functionality
  - github.com/nabbar/golib/prometheus: Metrics collection
  - github.com/nabbar/golib/logger: Logging support
  - github.com/nabbar/golib/shell/command: Shell command interface
  - github.com/nabbar/golib/context: Context management

# Examples

See the test files for comprehensive examples:

  - pool_test.go: Basic operations
  - pool_lifecycle_test.go: Lifecycle management
  - pool_metrics_test.go: Metrics integration
  - pool_encoding_test.go: Encoding examples
  - pool_shell_test.go: Shell command usage
  - pool_metrics_collection_test.go: Advanced metrics collection
  - pool_shell_exec_test.go: Shell command execution
  - pool_errors_test.go: Error handling patterns
  - pool_benchmark_test.go: Performance benchmarks
*/
package pool
