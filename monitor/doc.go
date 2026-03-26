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
Package monitor provides a high-performance, thread-safe health monitoring framework for Go applications.
It is designed to track the operational status of internal and external components (databases, APIs,
microservices) using a robust state machine that handles health transitions with built-in hysteresis
to prevent status flapping.

# Core Philosophy: Performance & Resilience

The monitor is architected for zero-contention status reporting and reliable periodic execution.
It separates configuration management from the performance-critical status reporting path.

Key design principles:
  - Atomic Status Reporting: Reads (Status, Latency, Uptime) use lock-free atomic primitives.
  - Dampened Transitions: Configurable Fall/Rise thresholds prevent noise during transient failures.
  - Dynamic Polling: Intervals automatically adjust based on the current health state (Rise/Fall/Stable).
  - Middleware Pipeline: Extensible execution chain for logging, metrics, and tracing.

# Internal Architecture & Data Flow

The monitor operates as a background orchestrator managed by an atomic ticker runner.

Internal Dataflow Diagram:

	[ Ticker Loop ] <-------------------------------------------+
	      |                                                     |
	      v                                                     |
	[ Interval Resolver ] --(IntervalCheck/Fall/Rise)-----------+
	      |
	      v
	[ Middleware Chain ]
	      |-- (mdlStatus) --+--> [ Start Timer ]
	      |                 |
	      |-- (User Fct) ---+--> [ Health Check Execution ]
	      |                 |
	      |-- (mdlStatus) --+--> [ Stop Timer & Capture Latency ]
	      |                 |
	      |                 +--> [ State Machine Transition Logic ]
	      |                 |
	      |                 +--> [ Atomic Update of Metrics Container ]
	      v
	[ Metrics Dispatch ] --(RegisterCollectMetrics)--> [ Prometheus / Loggers ]

# State Machine & Hysteresis

The monitor implements a 3-state machine (OK, Warn, KO) with directional transition counters.

Transition State Diagram:

	+--------+   (Fail >= fallCountWarn)   +--------+   (Fail >= fallCountKO)   +--------+
	|   OK   | --------------------------> |  Warn  | ------------------------> |   KO   |
	+--------+ <-------------------------- +--------+ <------------------------ +--------+
	           (Succ >= riseCountWarn)                (Succ >= riseCountKO)

Transition Logic Table:

	Current Status | Event   | Counter Logic               | Transition Action
	---------------|---------|-----------------------------|---------------------------
	OK             | Failure | cntFall++                   | If cntFall >= fallCountWarn: -> Warn
	OK             | Success | cntFall=0, cntRise=0        | Stay OK (Uptime++)
	Warn           | Failure | cntFall++                   | If cntFall >= fallCountKO:   -> KO
	Warn           | Success | cntRise++                   | If cntRise >= riseCountWarn: -> OK
	KO             | Success | cntRise++                   | If cntRise >= riseCountKO:   -> Warn
	KO             | Failure | cntRise=0, cntFall=0        | Stay KO (Downtime++)

# High-Performance Read Path

Status retrieval methods (Status, Latency, Uptime, Downtime) are optimized for high-frequency polling
(e.g., thousands of reads per second from a metrics exporter or a load-balancer probe).

Implementation Details:
  - No Mutexes: The "hot path" for reads uses atomic.LoadUint64/Int64 from the metrics container.
  - Zero Allocations: Status and metric reads perform no heap allocations.
  - Near-Zero Latency: Typical read latency is in the single-digit nanosecond range.

# Middleware Extensibility

The health check execution uses a LIFO (Last-In-First-Out) middleware stack. Middlewares can intercept
the execution to perform pre/post actions.

Execution Stack Example:

 1. [ mdlStatus ] (Core: Latency & State logic)
 2. [ CustomLogger ] (Optional: logs failures)
 3. [ OpenTelemetry ] (Optional: traces health check)
 4. [ User HealthCheck Function ] (The actual diagnostic)

# Sub-Packages & Modules

  - info: Metadata management (name, version, environment data).
  - status: Status enumeration and parsing (OK, Warn, KO, Unknown).
  - types: Public interfaces and configuration structures.
  - pool: (Optional) Management for collections of monitors.

# Usage Example: Database Health Monitoring

	import (
		"context"
		"github.com/nabbar/golib/monitor"
		"github.com/nabbar/golib/monitor/info"
		"github.com/nabbar/golib/monitor/types"
		"github.com/nabbar/golib/duration"
	)

	func setupMonitor(db *sql.DB) types.Monitor {
		// 1. Initialize Info with metadata
		inf, _ := info.New("postgres-db")

		// 2. Create the monitor instance
		mon, _ := monitor.New(context.Background(), inf)

		// 3. Configure intervals and thresholds
		cfg := types.Config{
			Name:          "main-database",
			CheckTimeout:  duration.ParseDuration("5s"),
			IntervalCheck: duration.ParseDuration("30s"), // Normal polling frequency
			IntervalFall:  duration.ParseDuration("2s"),  // Aggressive polling when failing
			FallCountWarn: 2,                             // 2 consecutive failures to trigger 'Warn'
			FallCountKO:   3,                             // 3 more failures to trigger 'KO'
		}
		_ = mon.SetConfig(context.Background(), cfg)

		// 4. Register the actual check logic
		mon.SetHealthCheck(func(ctx context.Context) error {
			return db.PingContext(ctx)
		})

		// 5. Start the background runner
		_ = mon.Start(context.Background())

		return mon
	}

# Thread Safety & Concurrency

Every component is safe for concurrent use. Configuration updates (SetConfig) and Metadata updates
(InfoUpd) use atomic swaps or thread-safe containers to ensure that a configuration reload never
causes a race condition or performance degradation for the periodic runner or status readers.

# Prometheus & Metrics Integration

The package is designed to integrate seamlessly with Prometheus via the RegisterCollectMetrics method.
At the end of each diagnostic run, the monitor dispatches its current state to the registered collector,
updating Gauges for latency, status code, and transition timers.
*/
package monitor
