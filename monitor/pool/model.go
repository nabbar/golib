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

package pool

import (
	"context"
	"sync"
	"time"

	libctx "github.com/nabbar/golib/context"
	liblog "github.com/nabbar/golib/logger"
	libprm "github.com/nabbar/golib/prometheus"
)

// pool is the internal implementation of the Pool interface.
// It manages a collection of monitors with thread-safe operations.
type pool struct {
	m  sync.RWMutex             // Protects Prometheus and logger function access
	fp libprm.FuncGetPrometheus // Function to get Prometheus instance
	fl liblog.FuncLog           // Function to get logger instance
	p  libctx.Config[string]    // Context-aware storage for monitors
}

func (o *pool) setDefaultLog() {
	o.m.Lock()
	defer o.m.Unlock()

	lg := liblog.New(o.p)
	o.fl = func() liblog.Logger {
		return lg
	}
}

func (o *pool) getLog() liblog.Logger {
	o.m.RLock()
	defer o.m.RUnlock()

	if o.fl != nil {
		if l := o.fl(); l != nil {
			return l
		}
	}

	o.m.RUnlock()
	o.setDefaultLog()
	o.m.RLock()

	return o.fl()
}

// RegisterMetrics registers metrics collection for the pool.
// It registers the Prometheus and logger functions and creates necessary metrics.
// Returns an error if metric creation fails.
func (o *pool) RegisterMetrics(prm libprm.FuncGetPrometheus, log liblog.FuncLog) error {
	o.RegisterFctProm(prm)
	o.RegisterFctLogger(log)
	return o.createMetrics()
}

// UnregisterMetrics removes all registered metrics from Prometheus.
// This should be called when shutting down the pool to clean up resources.
func (o *pool) UnregisterMetrics() []error {
	if prm := o.getProm(); prm != nil {
		return prm.ClearMetric(true, true)
	}
	return nil
}

// InitMetrics is deprecated. Use RegisterMetrics instead.
// Deprecated: Use RegisterMetrics instead.
func (o *pool) InitMetrics(prm libprm.FuncGetPrometheus, log liblog.FuncLog) error {
	return o.RegisterMetrics(prm, log)
}

// ShutDown is deprecated. Use UnregisterMetrics instead.
// Deprecated: Use UnregisterMetrics instead.
func (o *pool) ShutDown() {
	o.UnregisterMetrics()
}

// RegisterFctProm registers the function to retrieve the Prometheus instance.
// This function is thread-safe.
func (o *pool) RegisterFctProm(prm libprm.FuncGetPrometheus) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fp = prm
}

// RegisterFctLogger registers the function to retrieve the logger instance.
// This function is thread-safe.
func (o *pool) RegisterFctLogger(log liblog.FuncLog) {
	o.m.Lock()
	defer o.m.Unlock()

	o.fl = log
}

// TriggerCollectMetrics periodically triggers metrics collection for all monitors in the pool.
// It runs until the context is cancelled. The dur parameter specifies the interval between collections.
// This function is designed to be run in a goroutine.
func (o *pool) TriggerCollectMetrics(ctx context.Context, dur time.Duration) {
	var tck = time.NewTicker(dur)
	defer tck.Stop()

	for {
		select {
		case <-tck.C:
			if p := o.getProm(); p != nil {
				p.CollectMetrics(ctx)
			}

		case <-ctx.Done():
			return
		}
	}
}
