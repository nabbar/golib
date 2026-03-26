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

package monitor

import (
	"context"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	montps "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/runner"
	runtck "github.com/nabbar/golib/runner/ticker"
)

const (
	// defaultMonitorName is the fallback identifier used when a monitor is initialized without an explicit name.
	defaultMonitorName = "not named"

	// Internal storage keys for the thread-safe context (libctx.Config).
	keyName        = "keyName"          // keyName stores the monitor's display name (string).
	keyConfig      = "keyConfig"        // keyConfig stores the normalized runtime configuration (*runCfg).
	keyLogger      = "keyLogger"        // keyLogger stores the active structured logger (liblog.Logger).
	keyLoggerDef   = "keyLoggerDefault" // keyLoggerDef stores the fallback logger provider (liblog.FuncLog).
	keyHealthCheck = "keyFct"           // keyHealthCheck stores the registered health check function (montps.HealthCheck).

	// Internal storage keys for metrics-related configuration.
	keyMetricsName = "keyMetricsName" // keyMetricsName stores the slice of registered Prometheus metric names ([]string).
	keyMetricsFunc = "keyMetricsFunc" // keyMetricsFunc stores the Prometheus collection function (libprm.FuncCollectMetrics).

	// Structured logging constants for consistency across monitor log entries.
	LogFieldProcess = "process" // LogFieldProcess is the field key for identifying the monitor process in logs.
	LogValueProcess = "monitor" // LogValueProcess is the constant value for the process field.
	LogFieldName    = "name"    // LogFieldName is the field key for the monitor instance name.
)

// mon is the private concrete implementation of the Monitor interface.
// It serves as the orchestrator for periodic health checks, managing state,
// configuration, and metadata in a thread-safe manner.
//
// Architecture:
// The monitor is designed for high-concurrency environments. It separates frequently accessed
// performance metrics (stored in the 'l' field using atomic types) from less frequent
// configuration data (stored in the 'x' context map).
type mon struct {
	// x is a generic thread-safe configuration and state container keyed by strings.
	// It holds various monitor-related data like the health check function and logger.
	x libctx.Config[string]

	// i is an atomic value container for the monitor's metadata (Info).
	// Atomic storage allows metadata updates (e.g., version changes) without blocking active checks.
	i libatm.Value[montps.Info]

	// r is an atomic value container for the background ticker runner.
	// This allows the monitor to be started, stopped, or restarted (replacing the ticker)
	// in a thread-safe manner without race conditions.
	r libatm.Value[runtck.Ticker]

	// l is a high-performance, atomic-based structure holding the results of the last health check.
	// It stores status, latency, uptime, and downtime using lock-free atomic primitives to
	// ensure that status reads have near-zero overhead.
	l *lastRun
}

// SetHealthCheck registers the core diagnostic function (a health check logic) that will be executed periodically.
//
// Parameters:
//   - fct: A function matching the montps.HealthCheck signature. It must perform the diagnostic
//     and return nil for success or an error for failure.
//
// The provided function (montps.HealthCheck) should perform the necessary diagnostics and return:
// - nil: if the monitored component is healthy (OK).
// - error: if a problem is detected (Warn or KO, depending on the failure count thresholds).
//
// This method is thread-safe and includes a recovery mechanism for unexpected panics during storage.
func (o *mon) SetHealthCheck(fct montps.HealthCheck) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/SetHealthCheck", r)
		}
	}()

	o.x.Store(keyHealthCheck, fct)
}

// GetHealthCheck retrieves the health check function currently registered within the monitor.
// It returns nil if no function has been associated with the monitor yet.
func (o *mon) GetHealthCheck() montps.HealthCheck {
	return o.getFct()
}

// Clone creates a deep copy of the current monitor instance using a new context.
//
// Parameters:
//   - ctx: The new base context for the cloned monitor.
//
// The cloned monitor inherits:
// 1. The configuration and internal state from the original monitor's context.
// 2. The metadata (Info) from the original monitor.
// 3. The last run metrics snapshot.
// 4. The running status: if the original monitor is currently active, the clone will automatically start its ticker.
//
// This method returns the newly created Monitor instance or an error if the start-up process fails for the clone.
// It includes a recovery mechanism to handle potential panics during the cloning process.
func (o *mon) Clone(ctx context.Context) (montps.Monitor, error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/Clone", r)
		}
	}()

	// Initialize new instance with its own thread-safe state.
	n := &mon{
		x: nil,
		i: libatm.NewValue[montps.Info](),
		r: libatm.NewValue[runtck.Ticker](),
		l: nil,
	}

	// Clone the internal state container with the new context.
	n.x = o.x.Clone(ctx)
	// Inherit the metadata by performing an atomic load/store.
	n.i.Store(o.i.Load())
	// Clone the last run metrics snapshot.
	if o.l != nil {
		n.l = o.l.Clone()
	}

	// If the parent monitor is running, ensure the child follows suit.
	if o.IsRunning() {
		if e := n.Start(ctx); e != nil {
			return nil, e
		}
	}

	return n, nil
}
