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

// Package startStop provides a thread-safe runner for managing service lifecycle
// with start and stop operations. It handles context cancellation, error tracking,
// and uptime monitoring for long-running services.
//
// The runner executes start/stop functions asynchronously and tracks their execution
// state, errors, and uptime. All operations are thread-safe and can be called concurrently.
//
// Example usage:
//
//	runner := startStop.New(
//	    func(ctx context.Context) error {
//	        // Start your service
//	        return server.ListenAndServe()
//	    },
//	    func(ctx context.Context) error {
//	        // Stop your service
//	        return server.Shutdown(ctx)
//	    },
//	)
//	runner.Start(context.Background())
package startStop

import (
	"context"
	"sync"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	liberr "github.com/nabbar/golib/errors"
	errpol "github.com/nabbar/golib/errors/pool"
	libsrv "github.com/nabbar/golib/runner"
)

// StartStop defines the interface for managing service lifecycle operations.
// It combines server management (Start, Stop, Restart, IsRunning, Uptime) with
// error tracking (ErrorsLast, ErrorsList). All operations are thread-safe.
type StartStop interface {
	// Runner embeds the base server interface providing lifecycle operations:
	// Start, Stop, Restart, IsRunning, Uptime
	libsrv.Runner

	// Errors embeds error tracking operations:
	// ErrorsLast, ErrorsList
	liberr.Errors
}

// New creates a new StartStop runner with the provided start and stop functions.
// The start function is executed asynchronously when Start() is called, and should
// block until the service terminates. The stop function is called to gracefully
// shut down the service.
//
// Parameters:
//   - start: Function to start the service (runs asynchronously, should block)
//   - stop: Function to stop the service (called on Stop/Restart)
//
// Returns:
//   - StartStop: Thread-safe runner instance ready to use
//
// Example:
//
//	runner := startStop.New(
//	    func(ctx context.Context) error {
//	        return httpServer.ListenAndServe()  // Blocks until server stops
//	    },
//	    func(ctx context.Context) error {
//	        return httpServer.Shutdown(ctx)     // Graceful shutdown
//	    },
//	)
func New(start, stop func(ctx context.Context) error) StartStop {
	return &run{
		m: sync.Mutex{},
		e: errpol.New(),
		t: libatm.NewValue[time.Time](),
		n: libatm.NewValue[context.CancelFunc](),

		f: start,
		s: stop,
	}
}
