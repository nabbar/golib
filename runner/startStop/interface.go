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

// Package startStop provides a robust, thread-safe runner for managing the lifecycle
// of long-running services or tasks that require explicit start and stop operations.
// It offers a standardized way to control service execution, handle graceful shutdowns,
// track operational status, and collect errors in a concurrent-safe manner.
//
// The core functionality revolves around executing user-defined start and stop functions
// asynchronously. It meticulously manages context cancellation to signal termination
// to the running service, tracks execution state (running/stopped), monitors uptime,
// and aggregates any errors encountered during the service's lifetime.
//
// All operations exposed by this package are designed to be thread-safe, meaning
// they can be safely called concurrently from multiple goroutines without introducing
// race conditions or requiring external synchronization mechanisms.
//
// Example usage demonstrates how to instantiate and use the runner for a typical service:
//
//	runner := startStop.New(
//	    func(ctx context.Context) error {
//	        // This function represents the main logic of your service.
//	        // It should typically block until the service is meant to be terminated.
//	        // The provided context (ctx) will be cancelled when Stop() is called,
//	        // allowing for graceful shutdown procedures within this function.
//	        return server.ListenAndServe() // Example: starting an HTTP server
//	    },
//	    func(ctx context.Context) error {
//	        // This function is responsible for gracefully stopping your service.
//	        // It is invoked when runner.Stop() or runner.Restart() is called.
//	        // The context provided here can be used for shutdown timeouts.
//	        return server.Shutdown(ctx) // Example: shutting down an HTTP server
//	    },
//	)
//
//	// To start the service, typically in a separate goroutine for non-blocking execution:
//	go func() {
//	    if err := runner.Start(context.Background()); err != nil {
//	        log.Printf("Service failed to start or encountered an error: %v", err)
//	    }
//	}()
//
//	// Later, to stop the service:
//	// err := runner.Stop(context.Background())
//	// if err != nil {
//	//     log.Printf("Service failed to stop gracefully: %v", err)
//	// }
package startStop

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	liberr "github.com/nabbar/golib/errors"
	errpol "github.com/nabbar/golib/errors/pool"
	libsrv "github.com/nabbar/golib/runner"
)

// StartStop defines the unified interface for managing service lifecycle operations
// and error tracking. It provides a comprehensive set of methods to control,
// monitor, and inspect the state of a long-running service.
//
// This interface is designed for maximum flexibility and reusability, combining
// functionalities from two distinct interfaces:
//   - libsrv.Runner: for core service lifecycle management (start, stop, restart, status).
//   - liberr.Errors: for robust and centralized error collection and retrieval.
//
// All methods exposed by the StartStop interface are guaranteed to be thread-safe,
// allowing for concurrent interactions from various parts of an application.
type StartStop interface {
	// Runner embeds the base server interface, providing fundamental lifecycle operations.
	// These methods allow for starting, stopping, restarting, and querying the operational
	// status and uptime of the managed service.
	libsrv.Runner

	// Start initiates the execution of the service.
	//   - Parameters: ctx (context.Context) - A context for the start operation itself.
	//   - Returns: error - An error if the start operation fails or the context is already cancelled.
	//   - Behavior: If the service is already running, it will first attempt to stop it
	//     gracefully before initiating a new start sequence. The user-provided 'start'
	//     function will be executed in a new goroutine.
	// libsrv.Runner.Start

	// Stop gracefully terminates the running service.
	//   - Parameters: ctx (context.Context) - A context for the stop operation, potentially
	//     including a timeout for graceful shutdown.
	//   - Returns: error - An error if the stop operation encounters issues.
	//   - Behavior: Signals the running 'start' function to terminate via context cancellation,
	//     executes the user-provided 'stop' function, and waits for the service goroutine
	//     to fully exit.
	// libsrv.Runner.Stop

	// Restart performs a sequential stop and then start operation.
	//   - Parameters: ctx (context.Context) - A context for both the stop and start phases.
	//   - Returns: error - The first error encountered during either the stop or start process.
	//   - Behavior: Ensures a clean shutdown of the current service instance before
	//     launching a new one, useful for configuration reloads or recovery.
	// libsrv.Runner.Restart

	// IsRunning checks the current operational status of the service.
	//   - Parameters: None.
	//   - Returns: bool - True if the service is actively running, false otherwise.
	//   - Behavior: This method is highly optimized, using atomic operations for
	//     fast and thread-safe status checks without locking.
	// libsrv.Runner.IsRunning

	// Uptime calculates the duration since the service was last successfully started.
	//   - Parameters: None.
	//   - Returns: time.Duration - The elapsed time. Returns 0 if the service is not running
	//     or has not been started yet.
	//   - Behavior: Provides a non-blocking, thread-safe way to query service longevity.
	// libsrv.Runner.Uptime

	// Errors embeds error tracking operations, allowing the caller to inspect
	// any errors that occurred during the service's execution or lifecycle transitions.
	liberr.Errors

	// ErrorsLast retrieves the most recent error that occurred within the service.
	//   - Parameters: None.
	//   - Returns: error - The last recorded error, or nil if no errors have occurred.
	//   - Behavior: Useful for quick checks of the service's health status.
	// liberr.Errors.ErrorsLast

	// ErrorsList retrieves a slice of all errors recorded since the last successful Start() call.
	//   - Parameters: None.
	//   - Returns: []error - A slice containing all accumulated errors. The slice
	//     is empty if no errors have occurred.
	//   - Behavior: Provides a historical view of issues, which is cleared upon
	//     each new service start.
	// liberr.Errors.ErrorsList
}

// New creates and initializes a new StartStop runner instance. This is the primary
// constructor function for the package, allowing users to define the core logic
// of their service.
//
// Parameters:
//
//   - start: A `runner.FuncAction` function (i.e., `func(ctx context.Context) error`)
//     that encapsulates the main execution logic of the service. This function is
//     expected to be blocking and will be run in its own dedicated goroutine when
//     `Start()` is called. The `context.Context` provided to this function will be
//     cancelled when `Stop()` is invoked, serving as a signal for the service
//     to initiate its graceful shutdown procedures. Any error returned by this
//     function will be captured and stored by the runner.
//
//   - stop: A `runner.FuncAction` function (i.e., `func(ctx context.Context) error`)
//     that defines the graceful shutdown procedure for the service. This function
//     is called when `Stop()` or `Restart()` methods are invoked on the runner.
//     It receives a `context.Context` which can be used to enforce shutdown timeouts
//     or cancellation. Any error returned by this function will also be captured.
//
// Behavior of the returned StartStop instance:
//   - **Single Instance Execution**: The runner ensures that only one instance of
//     the `start` function is actively running at any given time. Subsequent calls
//     to `Start()` while the service is already running will trigger an internal
//     `Stop()` before restarting.
//   - **Thread Safety**: All internal state management, including the running status,
//     start time, context cancellation, and error collection, is handled using
//     `sync.Mutex` and `sync/atomic` primitives to guarantee thread safety across
//     concurrent calls to the runner's methods.
//   - **Uptime Tracking**: The runner automatically records the start time of the
//     service, enabling accurate `Uptime()` reporting without additional user logic.
//   - **Error Aggregation**: Any errors returned by the `start` or `stop` functions
//     are automatically added to an internal error pool, which can be queried
//     via `ErrorsLast()` and `ErrorsList()`. This provides a centralized mechanism
//     for monitoring service health and diagnosing issues.
//
// Returns:
//   - A concrete implementation of the `StartStop` interface (specifically, an
//     instance of the internal `run` struct), configured with the provided
//     start and stop functions.
//
// Example of creating a new runner:
//
//	// Define the service's start and stop logic
//	myServiceStart := func(ctx context.Context) error {
//	    log.Println("My service is starting...")
//	    // Simulate work that blocks until context is cancelled
//	    <-ctx.Done()
//	    log.Println("My service context cancelled, preparing to exit.")
//	    return ctx.Err() // Return context cancellation error or nil
//	}
//
//	myServiceStop := func(ctx context.Context) error {
//	    log.Println("My service is stopping gracefully...")
//	    // Perform cleanup, close connections, etc.
//	    time.Sleep(500 * time.Millisecond) // Simulate cleanup time
//	    log.Println("My service stopped.")
//	    return nil
//	}
//
//	// Create the runner instance
//	runner := startStop.New(myServiceStart, myServiceStop)
//
//	// Now 'runner' can be used to control the lifecycle of 'myService'.
func New(start, stop func(ctx context.Context) error) StartStop {
	// The `run` struct is the concrete implementation of the StartStop interface.
	// It encapsulates all the necessary state and logic for managing the service.
	return &run{
		// m: A mutex to protect critical sections, especially during Start/Stop transitions.
		m: sync.Mutex{},
		// e: An error pool to collect and manage errors from the start/stop functions.
		e: errpol.New(),
		// t: An atomic value to store the service's start time, enabling thread-safe Uptime calculation.
		t: libatm.NewValue[time.Time](),
		// n: An atomic value to hold the context.CancelFunc for the currently running service.
		// This allows for dynamic cancellation of the service's context.
		n: libatm.NewValue[context.CancelFunc](),
		// w: An atomic value to hold a channel that signals when the service's goroutine has fully exited.
		// This is crucial for the Stop() method to wait for complete termination.
		w: libatm.NewValue[chan struct{}](),
		// r: An atomic boolean flag indicating whether the service is currently running.
		// Provides fast, lock-free checks for IsRunning().
		r: new(atomic.Bool),

		// f: The user-provided function to start the service.
		f: start,
		// s: The user-provided function to stop the service.
		s: stop,
	}
}
