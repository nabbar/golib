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

// Package runner provides thread-safe lifecycle management for long-running
// services and periodic tasks. It offers two execution patterns:
//
//   - startStop: Service lifecycle with blocking start and graceful stop
//   - ticker: Periodic execution at regular intervals
//
// All implementations are thread-safe with atomic operations, context-aware,
// and include automatic error collection and panic recovery.
//
// Example usage with HTTP server:
//
//	runner := startStop.New(
//	    func(ctx context.Context) error {
//	        return httpServer.ListenAndServe()  // Blocks until stopped
//	    },
//	    func(ctx context.Context) error {
//	        return httpServer.Shutdown(ctx)     // Graceful shutdown
//	    },
//	)
//	runner.Start(context.Background())
//	defer runner.Stop(context.Background())
//
// Example usage with periodic task:
//
//	ticker := ticker.New(30*time.Second, func(ctx context.Context, tck *time.Ticker) error {
//	    return performHealthCheck()  // Executes every 30 seconds
//	})
//	ticker.Start(context.Background())
//	defer ticker.Stop(context.Background())
package runner

import (
	"context"
	"time"
)

// FuncAction represents a function that performs an action with context support.
// It is used by the startStop subpackage for both start and stop operations.
// The function should respect context cancellation and return any errors encountered.
//
// The start function typically blocks until the service terminates, while the
// stop function performs cleanup and graceful shutdown operations.
type FuncAction func(ctx context.Context) error

// FuncTicker represents a function executed by the ticker at regular intervals.
// It receives both the context (which will be cancelled on Stop) and the underlying
// time.Ticker for advanced control if needed.
//
// The function is called repeatedly until the ticker is stopped or the context
// is cancelled. Errors are collected and can be retrieved via ErrorsLast() or
// ErrorsList(). Returning an error does not stop the ticker.
//
// Example:
//
//	func(ctx context.Context, tck *time.Ticker) error {
//	    if err := performTask(); err != nil {
//	        return fmt.Errorf("task failed: %w", err)
//	    }
//	    return nil
//	}
type FuncTicker func(ctx context.Context, tck *time.Ticker) error

// Runner defines the core interface for lifecycle management of services and tasks.
// It provides operations to start, stop, restart, and monitor the running state
// of a managed component.
//
// All methods are thread-safe and can be called concurrently from multiple goroutines.
// State transitions are properly synchronized using atomic operations and mutexes.
//
// Implementations include:
//   - startStop.StartStop: For long-running services with blocking start functions
//   - ticker.Ticker: For periodic task execution at regular intervals
type Runner interface {
	// Start launches the runner with the given context. For startStop runners,
	// the start function executes asynchronously in a goroutine. For ticker
	// runners, execution begins at regular intervals.
	//
	// If the runner is already running, it will be stopped first to ensure
	// clean state transition. The context is used to create a cancellable
	// child context for the runner's lifecycle.
	//
	// Returns an error if the context is nil. Operational errors are collected
	// internally and retrievable via the Errors interface (if available).
	Start(ctx context.Context) error

	// Stop gracefully stops the runner if it is currently running. This is an
	// idempotent operation - calling Stop on an already stopped runner is safe
	// and returns nil immediately.
	//
	// The method cancels the runner's context and waits for cleanup to complete
	// using exponential backoff polling (up to 2 seconds). For startStop runners,
	// the stop function is also called asynchronously.
	//
	// Returns an error if the context is nil. The context is used for timeout
	// control but not for the stop operation itself.
	Stop(ctx context.Context) error

	// Restart atomically stops the runner (if running) and starts it again with
	// a fresh state. This is equivalent to calling Stop() followed by Start(),
	// but ensures atomicity through mutex protection.
	//
	// Previous errors are cleared on restart. The context is used for both
	// the stop and start operations.
	//
	// Returns an error if the context is nil or if the stop/start operations fail.
	Restart(ctx context.Context) error

	// IsRunning returns true if the runner is currently active. This is determined
	// by checking if the uptime is greater than zero.
	//
	// This method uses atomic operations for lock-free reads and is safe to call
	// concurrently from multiple goroutines.
	IsRunning() bool

	// Uptime returns the duration since the runner was started. Returns 0 if the
	// runner is not currently running.
	//
	// This method uses atomic operations for lock-free reads and is safe to call
	// concurrently from multiple goroutines. The uptime is calculated from the
	// atomically stored start time.
	Uptime() time.Duration
}

// WaitNotify is an optional interface that can be implemented by runners to support
// waiting for startup completion or receiving notifications. This interface is currently
// not implemented by the standard runner types but is defined for future extensibility.
type WaitNotify interface {
	// StartWaitNotify begins waiting for notifications with the given context.
	StartWaitNotify(ctx context.Context)

	// StopWaitNotify stops waiting for notifications.
	StopWaitNotify()
}
