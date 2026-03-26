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

// Package monitor provides a robust, thread-safe framework for periodic health monitoring of components.
// It supports state machine transitions (OK, Warn, KO), middleware execution chains, and integrated
// metrics collection for Prometheus.
package monitor

import (
	"context"
	"fmt"

	libatm "github.com/nabbar/golib/atomic"
	libctx "github.com/nabbar/golib/context"
	montps "github.com/nabbar/golib/monitor/types"
	librun "github.com/nabbar/golib/runner/ticker"
)

// Monitor is the primary interface for managing component health checks.
// It embeds the base Monitor interface from the types package, which defines
// methods for configuration, lifecycle management (Start/Stop), and status retrieval.
type Monitor interface {
	montps.Monitor
}

// New initializes and returns a new Monitor instance.
//
// Parameters:
//   - ctx: The base context used for the monitor's internal state management. If nil, context.Background() is used.
//   - info: An implementation of montps.Info containing metadata (name, version, etc.) for the monitored component.
//
// The returned monitor is initialized in a stopped state with default configuration values.
// It uses an atomic internal structure to ensure thread-safety across all operations.
//
// Returns:
//   - montps.Monitor: A thread-safe monitor instance.
//   - error: Returns an error if the mandatory 'info' parameter is nil.
//
// Example:
//
//	inf, _ := info.New("database-service")
//	mon, err := monitor.New(context.Background(), inf)
//	if err != nil {
//	    log.Fatalf("failed to create monitor: %v", err)
//	}
//	_ = mon.Start(context.Background())
func New(ctx context.Context, info montps.Info) (montps.Monitor, error) {
	if info == nil {
		return nil, fmt.Errorf("info cannot be nil")
	} else if ctx == nil {
		ctx = context.Background()
	}

	// Initialize the private 'mon' structure with thread-safe containers.
	// x: thread-safe configuration and state map.
	// i: atomic value for metadata.
	// r: atomic value for the background ticker runner.
	// l: optimized container for last run results and metrics.
	m := &mon{
		x: libctx.New[string](ctx),
		i: libatm.NewValue[montps.Info](),
		r: libatm.NewValue[librun.Ticker](),
		l: newLastRun(),
	}

	// Store the initial metadata.
	m.i.Store(info)

	// Ensure the monitor starts in a clean, stopped state.
	_ = m.Stop(ctx)

	return m, nil
}
