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

// Package ticker provides a robust, state-managed runner that executes a user-defined function at regular intervals.
//
// This package is part of the 'nabbar' library suite and is specifically designed to handle periodic background
// tasks with a high degree of control over lifecycle, state transitions, and error collection. It leverages
// a Finite State Machine (FSM) to ensure that the runner behaves predictably in highly concurrent environments.
//
// # Architecture Overview
//
// The core component is the Ticker interface, which extends two other fundamental interfaces:
//   - libsrv.Runner: Provides standard lifecycle methods (Start, Stop, Restart, IsRunning, Uptime).
//   - liberr.Errors: Provides a unified way to collect and retrieve errors occurring during periodic executions.
//
// The implementation uses internal Go tickers (time.Ticker) and a main control loop running in its own goroutine.
// This loop monitors for both tick events and state change requests, ensuring responsive management.
//
// # Internal Finite State Machine (FSM)
//
// The runner operates through a set of mutually exclusive states managed atomically:
//   - stopped: The idle state. No goroutines are running, and no ticks are being processed.
//   - started: An intermediate state indicating the Start() procedure has been initiated.
//   - running: The active state. The background goroutine is alive and processing ticks.
//   - reqStart: An internal signal state used to transition from stopped to started/running.
//   - reqStop: A signal state used to notify the background loop that it must terminate gracefully.
//
// # State Transition Diagram
//
//	  [ New / Stop ]
//	        |
//	        v
//	  +----------+
//	  |  stopped | <--------------------------------+
//	  +----------+                                  |
//	        |                                       |
//	  [ Start() ]                                   | (Automatic exit)
//	        |                                       |
//	        v                                       |
//	  +----------+          [ deMuxStart() ]   +----------+
//	  | started  |---------------------------->| reqStop  |
//	  +----------+                             +----------+
//	        |                                       ^
//	(Goroutine Spawn)                               |
//	        |                                  [ Stop() ]
//	        v                                       |
//	  +----------+                                  |
//	  | running  |----------------------------------+
//	  +----------+
//	        |
//	 (Loop: Tick -> Execute -> Collect Error)
//
// # Async Start & Synchronization Flow
//
// The Start() and Stop() methods are synchronous from the caller's perspective but manage an asynchronous
// background process. This is achieved via a "Busy Wait & Poll" pattern:
//
//	Caller Thread (Start/Stop)                 Background Goroutine (Loop)
//	--------------------------                 ---------------------------
//	            |
//	    [1] Lock Mutex
//	            |
//	    [2] Set State (reqStart/reqStop)
//	            |
//	    [3] Spawn Goroutine ---------------------> [A] Initialize Tickers
//	            |                                  [B] Set State (running)
//	    [4] Loop & Sleep (pollState) <----+                |
//	            |                         |        [C] Wait for Tick
//	    [5] Check State (running/stopped) |                |
//	            |                         |        [D] Exec Function
//	    [6] Unlock & Return <-------------+                |
//	                                               [E] Store Errors
//
// # Concurrency and Thread Safety
//
// All public methods (Start, Stop, Restart, IsRunning, Uptime, Errors*) are designed to be thread-safe.
//   - State transitions use 'sync/atomic' for low-latency status checks.
//   - Complex lifecycle changes (Start/Stop) are synchronized using a 'sync.Mutex' to prevent race conditions
//     during initialization or teardown.
//   - Context management (cancellation) uses atomic swaps to ensure that only the most recent context is active,
//     automatically cleaning up previous context resources.
//
// # Error Management
//
// Unlike a simple time.Ticker, this runner collects every error returned by the ticker function.
// It uses an internal error pool ('errpol.Pool') which:
//   - Stores all encountered errors without stopping the loop.
//   - Allows the caller to inspect the last error or the full list of errors at any time.
//   - Clears the error history automatically whenever the runner is restarted.
//
// # Use Cases and Patterns
//
//   - Background Cleanup: Periodically purging expired cache entries or temporary files.
//   - Status Monitoring: Regularly checking the health of external dependencies (DB, API).
//   - Metric Collection: Gathering system statistics at fixed intervals.
//   - Keep-Alive Signals: Sending heartbeats to a coordination service.
//
// # Detailed Quick Start
//
// Below is a comprehensive example demonstrating creation, execution, and error handling:
//
//	package main
//
//	import (
//		"context"
//		"errors"
//		"fmt"
//		"time"
//
//		"github.com/nabbar/golib/runner"
//		"github.com/nabbar/golib/runner/ticker"
//	)
//
//	func main() {
//		// Define a ticker function.
//		// It receives a context (for cancellation) and a TickUpdate (the underlying ticker).
//		fn := func(ctx context.Context, tck runner.TickUpdate) error {
//			fmt.Printf("[%s] Executing periodic task...\n", time.Now().Format(time.RFC3339))
//
//			// Simulate a random error for demonstration
//			if time.Now().Unix() % 5 == 0 {
//				return errors.New("intermittent failure")
//			}
//			return nil
//		}
//
//		// Create a new ticker with a 2-second interval.
//		// If the interval < 1ms, it defaults to 30s.
//		t := ticker.New(2 * time.Second, fn)
//
//		// Start the runner with a timeout context for the startup phase.
//		ctxStart, cancelStart := context.WithTimeout(context.Background(), 5 * time.Second)
//		defer cancelStart()
//
//		if err := t.Start(ctxStart); err != nil {
//			fmt.Printf("Failed to start ticker: %v\n", err)
//			return
//		}
//
//		// Monitor the ticker for a while.
//		fmt.Println("Ticker is running. Uptime:", t.Uptime())
//		time.Sleep(10 * time.Second)
//
//		// Check for errors collected during the run.
//		if err := t.ErrorsLast(); err != nil {
//			fmt.Println("Last error encountered:", err)
//		}
//
//		// Gracefully stop the ticker.
//		ctxStop, cancelStop := context.WithTimeout(context.Background(), 5 * time.Second)
//		defer cancelStop()
//
//		if err := t.Stop(ctxStop); err != nil {
//			fmt.Printf("Error during stop: %v\n", err)
//		}
//
//		fmt.Println("Ticker stopped successfully.")
//	}
//
// # Best Practices
//
//   - Ticker Function Performance: Ensure the ticker function ('fct') does not block for longer
//     than the tick interval itself. If it does, the next tick will be delayed (standard Go Ticker behavior).
//   - Resource Cleanup: Always call Stop() when the ticker is no longer needed to release goroutines and timers.
//   - Context Usage: The context passed to Start() or Stop() is only for the *operation* of starting or stopping.
//     The ticker function receives a separate background context that is canceled only when the ticker stops.
package ticker
