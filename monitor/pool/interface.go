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
	montps "github.com/nabbar/golib/monitor/types"
	libprm "github.com/nabbar/golib/prometheus"
	libsrv "github.com/nabbar/golib/runner"
)

type Pool interface {
	montps.Pool
	libsrv.Runner

	// RegisterMetrics registers the metrics for the pool.
	// It takes a function to get the current Prometheus metrics,
	// and a function to log messages.
	// It returns an error if something went wrong during the registration.
	//
	RegisterMetrics(prm libprm.FuncGetPrometheus, log liblog.FuncLog) error

	// UnregisterMetrics removes all registered metrics from Prometheus.
	// This should be called when the pool is being shut down to clean up resources.
	UnregisterMetrics() []error

	// InitMetrics is deprecated. Use RegisterMetrics instead.
	// Deprecated: Use RegisterMetrics instead.
	InitMetrics(prm libprm.FuncGetPrometheus, log liblog.FuncLog) error

	// ShutDown is deprecated. Use UnregisterMetrics instead.
	// Deprecated: Use UnregisterMetrics instead.
	ShutDown()

	// RegisterFctProm registers a function to get the current Prometheus metrics.
	// The function should return a Prometheus object.
	//
	// This function is used to register a function to get the current Prometheus metrics.
	// The function will be used to initialize the metrics for the pool.
	//
	// It does not return an error.
	//
	// Example:
	// p.RegisterFctProm(libprm.FuncGetPrometheus(func() libprm.Prometheus {
	// 	return libprm.NewPrometheus()
	// }))
	RegisterFctProm(prm libprm.FuncGetPrometheus)
	// RegisterFctLogger registers a function to log messages.
	// The function should take a liblog.Entry and return nothing.
	//
	// This function is used to register a function to log messages.
	// The function will be used to log messages during the initialization of the metrics.
	//
	// It does not return an error.
	RegisterFctLogger(log liblog.FuncLog)
	// TriggerCollectMetrics triggers the collection of metrics for the pool.
	// It takes a context and a duration as parameters.
	// The context is used to stop the collection of metrics if the context is done.
	// The duration is used to specify the interval between two collections of metrics.
	//
	// This function is designed to be used in a goroutine.
	//
	// Example:
	// go func() {
	// 	p.TriggerCollectMetrics(context.Background(), time.Second)
	// }
	TriggerCollectMetrics(ctx context.Context, dur time.Duration)
}

// New returns a new Pool.
//
// The returned pool is initialized with a context provided by the ctx parameter.
// This context is used to initialize the config of the pool.
//
// The returned pool is not initialized with any function to get the current Prometheus metrics.
// The RegisterFctProm function should be used to register such a function.
//
// The returned pool is not initialized with any function to log messages.
// The RegisterFctLogger function should be used to register such a function.
func New(ctx context.Context) Pool {
	return &pool{
		m:  sync.RWMutex{},
		fp: nil,
		p:  libctx.New[string](ctx),
	}
}
