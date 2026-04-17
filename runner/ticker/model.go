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

package ticker

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	libatm "github.com/nabbar/golib/atomic"
	errpol "github.com/nabbar/golib/errors/pool"
	librun "github.com/nabbar/golib/runner"
)

const (
	// pollChange is the internal duration used to poll the state change request.
	// It ensures that even if the main ticker has a very long duration (e.g., hours),
	// a stop request will be processed within a few milliseconds.
	pollChange = 5 * time.Millisecond

	// pollState is the duration of the sleep interval during busy-wait loops
	// in the Start and Stop methods. It determines the responsiveness of the
	// synchronization between the caller and the background goroutine.
	pollState = 5 * time.Microsecond

	// defaultDuration is used when a ticker is initialized with an invalid
	// or dangerously short duration (less than 1ms).
	defaultDuration = 30 * time.Second
)

// run is the concrete implementation of the Ticker interface.
// It manages the execution lifecycle of a function called at regular intervals
// while maintaining state consistency via an internal Finite State Machine (FSM).
type run struct {
	// m is a mutual exclusion lock that protects the Start and Stop methods.
	// This ensures that only one lifecycle transition occurs at a time,
	// preventing multiple goroutines from starting or stopping the runner simultaneously.
	m sync.Mutex

	// e is an error pool that captures every error returned by the ticker function.
	// It allows for asynchronous error collection without interrupting the ticker loop.
	e errpol.Pool

	// t stores the exact time the runner entered the 'running' state.
	// It is used by the Uptime() method to calculate how long the runner has been active.
	t libatm.Value[time.Time]

	// n stores the context's cancel function for the current execution loop.
	// When the runner is stopped, this function is called to signal the main loop
	// and any long-running ticker functions to terminate.
	n libatm.Value[context.CancelFunc]

	// s is a pointer to an atomic unsigned integer representing the current state.
	// Using atomic operations on this field allows for thread-safe state checks
	// (IsRunning, getState) without the overhead of a mutex.
	s *atomic.Uint32

	// f is the user-provided function that is executed on every tick.
	f librun.FuncTicker

	// d is the configured interval between executions of the ticker function.
	d time.Duration

	// k is the standard Go timer that triggers the execution of the ticker function.
	k *time.Ticker

	// w is a secondary timer used to poll for state changes and context cancellations
	// more frequently than the main tick interval.
	w *time.Ticker
}

// deMuxStop handles the internal sequence to initiate a stop.
// It sets the state to reqStop and cancels the current context.
// This method is called internally by Stop() after acquiring the mutex.
func (o *run) deMuxStop() {
	// If already stopped, no action is needed.
	if o.getState() == stopped {
		return
	}

	// Signal the background loop to stop via the FSM.
	o.setState(reqStop)

	// Signal the background loop to stop via context cancellation.
	o.cancel()
}

// deMuxStart prepares the environment and launches the main ticker loop in a new goroutine.
// It resets errors and timers before starting the background process.
func (o *run) deMuxStart() {
	// Clear any previous errors before a new start.
	o.e.Clear()

	// Set initial starting state.
	o.setState(started)

	// Initialize or reset the execution ticker.
	if o.k == nil {
		o.k = time.NewTicker(o.d)
	} else {
		o.k.Reset(o.d)
	}

	// Initialize or reset the state polling ticker.
	if o.w == nil {
		o.w = time.NewTicker(pollChange)
	} else {
		o.w.Reset(pollChange)
	}

	// Start the main execution loop in its own goroutine.
	go func(ctx context.Context, tck *time.Ticker, chg *time.Ticker) {
		// Panic recovery for the main loop to prevent crashing the entire application.
		defer func() {
			if r := recover(); r != nil {
				librun.RecoveryCaller("golib/server/ticker/deMuxStart", r)
			}
		}()

		// Ensure the state is set to 'stopped' when this goroutine exits.
		defer o.setState(stopped)

		// Final cleanup: reset uptime and ensure context is canceled.
		defer func() {
			o.t.Store(time.Time{})
			o.cancel()
		}()

		// Mark the official start time and transition to 'running'.
		o.t.Store(time.Now())
		o.setState(running)

		// Main control loop.
		for {
			// Check if a stop request was made via the FSM.
			if o.getState() == reqStop {
				return
			}

			select {
			case <-ctx.Done():
				// Exit if the context was canceled.
				return
			case <-chg.C:
				// Regular check for stop request on the polling ticker.
				if o.getState() == reqStop {
					return
				}
			case <-tck.C:
				// A tick occurred: execute the ticker function.
				o.runFunc(ctx, tck)
			}
		}
	}(o.newCancel(), o.k, o.w)
}

// runFunc executes the ticker function 'f' with a recovery wrapper.
// It ensures that a panic in the user-provided function doesn't stop the entire ticker loop.
// All returned errors are added to the internal error pool.
func (o *run) runFunc(ctx context.Context, tck librun.TickUpdate) {
	defer func() {
		// Recovery mechanism to catch panics within the user's ticker function.
		if r := recover(); r != nil {
			librun.RecoveryCaller("golib/server/ticker/runFunc", r)
		}
	}()

	// Execute the function and record any returned error.
	o.e.Add(o.f(ctx, tck))
}

// cancel executes the cancel function of the current execution context.
// It is safe to call even if the runner was never started.
func (o *run) cancel() {
	if o == nil || o.n == nil {
		return
	} else if n := o.n.Load(); n != nil {
		// Trigger the cancellation of the background loop's context.
		n()
	}
}

// newCancel creates a new cancelable context based on context.Background().
// It stores the new cancel function in the 'run' struct atomically.
// If a previous cancel function existed, it is called to ensure resources are freed.
func (o *run) newCancel() context.Context {
	x, n := context.WithCancel(context.Background())

	if o != nil && o.n != nil {
		// Swap the old cancel function with the new one.
		if old := o.n.Swap(n); old != nil {
			// Cancel the previous context if it was still active.
			old()
		}
	} else {
		// Fallback for improperly initialized structures.
		n()
	}

	return x
}
