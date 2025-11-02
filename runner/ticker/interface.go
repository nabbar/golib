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

// Package ticker provides a ticker-based runner implementation that executes
// a function at regular intervals. It combines the github.com/nabbar/golib/runner.Runner
// interface with error collection capabilities.
//
// The ticker automatically manages goroutine lifecycle, context cancellation,
// and error collection. It's designed for use cases requiring periodic execution
// of tasks with proper cleanup and state management.
//
// For more information about the runner package, see github.com/nabbar/golib/runner.
package ticker

import (
	"context"
	"fmt"
	"sync"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	liberr "github.com/nabbar/golib/errors"
	errpol "github.com/nabbar/golib/errors/pool"
	libsrv "github.com/nabbar/golib/runner"
)

// Ticker is the main interface for ticker-based runners. It combines the Runner
// interface from github.com/nabbar/golib/runner with error collection capabilities.
//
// The ticker executes a provided function at regular intervals until stopped or
// until its context is cancelled. All errors returned by the function are collected
// and can be retrieved via the Errors interface methods.
//
// Thread-safety: All methods are safe for concurrent use.
type Ticker interface {
	libsrv.Runner
	liberr.Errors
}

// New creates a new Ticker instance with the specified tick interval and function.
//
// Parameters:
//   - tick: The duration between function executions. If less than 1 millisecond,
//     defaultDuration (30 seconds) will be used instead.
//   - fct: The function to execute on each tick. It receives the ticker's context
//     and the underlying *time.Ticker. If nil, a default error-returning function
//     will be used.
//
// The function is executed in a goroutine and receives:
//   - ctx: A context that will be cancelled when Stop() is called or the parent
//     context expires
//   - tck: The underlying *time.Ticker that can be used for advanced tick control
//
// Returns a Ticker instance that is initially stopped. Call Start() to begin execution.
//
// Example:
//
//	tick := ticker.New(5*time.Second, func(ctx context.Context, tck *time.Ticker) error {
//	    // Perform periodic work
//	    return doWork(ctx)
//	})
//	if err := tick.Start(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
//	defer tick.Stop(context.Background())
func New(tick time.Duration, fct func(ctx context.Context, tck *time.Ticker) error) Ticker {
	if tick < time.Millisecond {
		tick = defaultDuration
	}
	if fct == nil {
		fct = func(ctx context.Context, tck *time.Ticker) error {
			return fmt.Errorf("invalid function ticker")
		}
	}
	return &run{
		m: sync.Mutex{},
		e: errpol.New(),
		t: libatm.NewValue[time.Time](),
		n: libatm.NewValue[context.CancelFunc](),

		f: fct,
		d: tick,
	}
}
