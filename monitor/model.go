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
	// defaultMonitorName is the default name used when no name is specified.
	defaultMonitorName = "not named"

	// Internal storage keys for monitor configuration
	keyName        = "keyName"
	keyConfig      = "keyConfig"
	keyLogger      = "keyLogger"
	keyLoggerDef   = "keyLoggerDefault"
	keyHealthCheck = "keyFct"
	keyLastRun     = "keyLastRun"

	// Internal storage keys for metrics
	keyMetricsName = "keyMetricsName"
	keyMetricsFunc = "keyMetricsFunc"

	// Log field constants for structured logging
	LogFieldProcess = "process"
	LogValueProcess = "monitor"
	LogFieldName    = "name"
)

// mon is the internal implementation of the Monitor interface.
// It manages the health check lifecycle, configuration, and state tracking.
type mon struct {
	x libctx.Config[string]       // Config stores monitor configuration and state
	i libatm.Value[montps.Info]   // Info provides metadata about the monitored component
	r libatm.Value[runtck.Ticker] // Ticker manages the periodic health check execution
}

// SetHealthCheck registers the health check function to be executed periodically.
// The function should return nil for a healthy state or an error otherwise.
func (o *mon) SetHealthCheck(fct montps.HealthCheck) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/SetHealthCheck", r)
		}
	}()

	o.x.Store(keyHealthCheck, fct)
}

// GetHealthCheck retrieves the currently registered health check function.
// Returns nil if no health check function has been registered.
func (o *mon) GetHealthCheck() montps.HealthCheck {
	return o.getFct()
}

// Clone creates a copy of the monitor with a new context.
// If the original monitor is running, the clone will also be started.
// Returns an error if the cloned monitor fails to start.
func (o *mon) Clone(ctx context.Context) (montps.Monitor, error) {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/monitor/Clone", r)
		}
	}()

	n := &mon{
		x: nil,
		i: libatm.NewValue[montps.Info](),
		r: libatm.NewValue[runtck.Ticker](),
	}

	n.x = o.x.Clone(ctx)
	n.i.Store(o.i.Load())

	if o.IsRunning() {
		if e := n.Start(ctx); e != nil {
			return nil, e
		}
	}

	return n, nil
}
