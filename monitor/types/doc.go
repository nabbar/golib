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

// Package types provides core type definitions and interfaces for the monitor system.
//
// This package defines the fundamental contracts and types used throughout the
// github.com/nabbar/golib/monitor ecosystem. It serves as the foundation for
// monitor implementations and integrations.
//
// # Core Components
//
// The package provides:
//   - Monitor interface: Primary contract for health monitoring implementations
//   - Config struct: Configuration for monitor behavior and thresholds
//   - Info interface: Metadata about monitored components
//   - HealthCheck function type: Health check implementation contract
//   - Error definitions: Standardized error handling
//
// # Monitor Interface
//
// The Monitor interface combines multiple sub-interfaces:
//   - MonitorInfo: Metadata management (name, version, description)
//   - MonitorStatus: Status querying (OK/Warn/KO, latency, uptime/downtime)
//   - MonitorMetrics: Prometheus metrics integration
//   - libsrv.Runner: Lifecycle management (Start, Stop, IsRunning)
//
// # Configuration
//
// The Config struct defines all monitor behavior parameters:
//   - Check intervals (normal, falling, rising)
//   - Timeout for health checks
//   - Thresholds for status transitions
//   - Logger configuration
//
// Example configuration:
//
//	cfg := types.Config{
//	    Name:          "database",
//	    CheckTimeout:  duration.ParseDuration(5 * time.Second),
//	    IntervalCheck: duration.ParseDuration(10 * time.Second),
//	    FallCountKO:   3,
//	    RiseCountKO:   3,
//	}
//
// # Status Transitions
//
// The monitor uses thresholds to prevent status flapping:
//   - FallCountWarn: Failures needed to go from OK to Warn
//   - FallCountKO: Failures needed to go from Warn to KO
//   - RiseCountKO: Successes needed to go from KO to Warn
//   - RiseCountWarn: Successes needed to go from Warn to OK
//
// # Health Check Function
//
// The HealthCheck function type defines the contract for health check implementations:
//
//	type HealthCheck func(ctx context.Context) error
//
// Implementations should:
//   - Respect context cancellation and timeout
//   - Return nil for healthy status
//   - Return descriptive errors for unhealthy status
//   - Complete quickly (within CheckTimeout)
//
// # Metrics Integration
//
// The MonitorMetrics interface provides Prometheus integration:
//
//	monitor.RegisterMetricsName("service_health", "service_latency")
//	monitor.RegisterCollectMetrics(func(ctx context.Context, names ...string) {
//	    // Update Prometheus metrics
//	})
//
// # Related Packages
//
//   - github.com/nabbar/golib/monitor: Main monitor implementation
//   - github.com/nabbar/golib/monitor/info: Dynamic metadata management
//   - github.com/nabbar/golib/monitor/status: Health status type
//   - github.com/nabbar/golib/monitor/pool: Monitor pool management
//
// # Thread Safety
//
// All interfaces and types in this package are designed for safe concurrent use.
// Implementations must ensure thread safety for all operations.
package types
