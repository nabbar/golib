/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
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
 */

// Package startStop provides a robust, thread-safe mechanism for managing the lifecycle
// of services or long-running tasks that require explicit start and stop operations.
//
// It is designed to handle common patterns in service management:
//   - Asynchronous execution of start functions.
//   - Graceful shutdown via stop functions.
//   - Automatic uptime tracking.
//   - Centralized error collection and reporting.
//   - Thread-safe state management (Running vs. Stopped).
//   - Context-aware operations with cancellation support.
//
// # Architecture & Data Flow
//
// The following diagram illustrates the internal architecture, state transitions,
// and synchronization points during the service lifecycle.
//
//	 [ External Caller ]
//	         |
//	         | (1) Start(ctx)
//	         v
//	+-------------------+
//	|   Start() Logic   | <----------------+
//	+-------------------+                  |
//	         |                             | (Restart calls Stop then Start)
//	         | (2) Mutex Lock              |
//	         | (3) Stop If Running --------+
//	         | (4) Reset Error Pool
//	         | (5) Create 'done' Chan
//	         | (6) Create New Context
//	         | (7) Launch Goroutine (Async)
//	         | (8) Mutex Unlock
//	         v
//	+-----------------------------+
//	| getFctStart() Goroutine     |
//	+-----------------------------+
//	| (A) Set Running = true      |
//	| (B) Set Start Time          |
//	| (C) Execute User Start Func | ----> [ Blocking Work ]
//	| (D) Wait for Completion/Ctx | <---- [ Context Cancelled ]
//	| (E) Capture Error           |
//	| (F) Set Running = false     |
//	| (G) Reset Start Time        |
//	| (H) Close 'done' Chan       | ----> (Signals Stop() to return)
//	+-----------------------------+
//
//	         |
//	         | (1) Stop(ctx)
//	         v
//	+-------------------+
//	|   Stop() Logic    |
//	+-------------------+
//	         |
//	         | (2) Fast Path Check (IsRunning?)
//	         | (3) Mutex Lock
//	         | (4) Signal Ctx Cancel
//	         | (5) Launch getFctStop() Goroutine (Async)
//	         | (6) Mutex Unlock
//	         | (7) Wait for 'done' Chan (Resynchronization)
//	         |     (with Timeout/Safety Wait)
//	         v
//	[ External Caller Resumes ]
//
// # State Transitions
//
//   - STOPPED --(Start)--> STARTING --(Async)--> RUNNING
//   - RUNNING --(Stop)--> STOPPING --(Async)--> STOPPED
//   - RUNNING --(Restart)--> STOPPING --(Async)--> STOPPED --(Start)--> RUNNING
//
// # Structure
//
// The core of the package is the StartStop interface, which combines two main aspects:
//  1. Lifecycle Management: Inherited from libsrv.Runner, providing Start(), Stop(),
//     Restart(), IsRunning(), and Uptime() methods.
//  2. Error Management: Inherited from liberr.Errors, providing access to the history
//     of errors encountered during the service's lifetime.
//
// # Working Principle
//
// When Start() is called:
//  1. It checks if the service is already running. If so, it attempts to stop it first.
//  2. It acquires a mutex to prevent concurrent Start/Stop operations.
//  3. It clears any previous errors from the internal error pool.
//  4. It creates a new "done" channel, which will be closed when the service's goroutine exits.
//  5. It creates a new cancellable context for the service's lifetime.
//  6. It launches the user-provided 'start' function in a new goroutine.
//  7. It releases the mutex.
//
// When the 'start' function returns (either due to completion or error):
//  1. A deferred function captures any returned error and adds it to the internal error pool.
//  2. The service state is updated to "stopped" (atomic flag and start time reset).
//  3. The internal context's cancel function is called.
//  4. The internal wait channel ('done') is closed to signal termination to any waiting Stop() call.
//
// When Stop() is called:
//  1. It performs a fast check: if the service is not running, it returns immediately.
//  2. It acquires a mutex to serialize stop operations.
//  3. It calls the internal context's cancel function to signal the 'start' function to terminate.
//  4. It executes the user-provided 'stop' function in a background goroutine with a timeout.
//  5. It releases the mutex.
//  6. It waits for the 'start' goroutine to signal its full termination via the 'w' channel,
//     subject to secondary timeouts for safety.
//
// # Quick Start
//
// To use this package, you need two functions: one that blocks while the service
// is running (the start function), and one that initiates a shutdown (the stop function).
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "net/http"
//	    "time"
//	    "github.com/nabbar/golib/runner/startStop"
//	)
//
//	func main() {
//	    srv := &http.Server{Addr: ":8080"}
//
//	    runner := startStop.New(
//	        func(ctx context.Context) error {
//	            // The start function usually blocks (e.g., ListenAndServe)
//	            fmt.Println("Service starting...")
//	            if err := srv.ListenAndServe(); err != http.ErrServerClosed {
//	                return err
//	            }
//	            fmt.Println("Service stopped.")
//	            return nil
//	        },
//	        func(ctx context.Context) error {
//	            // The stop function initiates shutdown
//	            fmt.Println("Service shutting down...")
//	            return srv.Shutdown(ctx)
//	        },
//	    )
//
//	    // Start the service in a goroutine
//	    go func() {
//	        if err := runner.Start(context.Background()); err != nil {
//	            fmt.Printf("Service Start error: %v\n", err)
//	        }
//	    }()
//
//	    time.Sleep(2 * time.Second)
//	    fmt.Printf("Is running: %v, Uptime: %v\n", runner.IsRunning(), runner.Uptime())
//
//	    // Stop the service
//	    if err := runner.Stop(context.Background()); err != nil {
//	        fmt.Printf("Service Stop error: %v\n", err)
//	    }
//
//	    // Wait a bit for shutdown to complete
//	    time.Sleep(1 * time.Second)
//	    fmt.Printf("After stop: Is running: %v, Uptime: %v\n", runner.IsRunning(), runner.Uptime())
//	    if err := runner.ErrorsLast(); err != nil {
//	        fmt.Printf("Last error: %v\n", err)
//	    }
//	}
//
// # Use Cases
//
//   - HTTP Servers: Managing the ListenAndServe / Shutdown cycle for web applications.
//   - Workers: Background goroutines processing queues that need clean termination upon application shutdown.
//   - Database Connections: Managing the lifecycle of persistent database connections or connection pools with health checks.
//   - System Daemons: Any long-running component requiring a standard "Service" interface in Go,
//     especially when integrating with system init systems or orchestration tools.
package startStop
