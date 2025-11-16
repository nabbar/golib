/*
 *  MIT License
 *
 *  Copyright (c) 2021 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package queuer

import (
	"context"
	"sync"
	"time"
)

// Counter defines the interface for the internal rate limiting counter.
//
// This interface is used internally by the Pooler to track and enforce
// rate limits. Each counter maintains its own state including the current
// count, maximum allowed operations, time window, and optional callback.
//
// All methods are thread-safe and protected by internal mutex synchronization.
type Counter interface {
	// Pool attempts to acquire permission to perform an operation.
	//
	// This method implements the core throttling logic:
	//  - If quota is available, decrements counter and returns immediately
	//  - If quota is exhausted, sleeps until the next time window
	//  - Checks context cancellation before and after sleeping
	//  - Invokes FuncCaller callback when throttling occurs (if configured)
	//
	// Parameters:
	//   - ctx: Context for cancellation. If the context is cancelled while
	//     waiting, returns ErrorMailPoolerContext.
	//
	// Returns:
	//   - nil if operation is allowed to proceed
	//   - ErrorMailPoolerContext if context is cancelled
	//   - Error from FuncCaller if callback fails
	//
	// Thread-safe: Multiple goroutines can call this concurrently.
	Pool(ctx context.Context) error

	// Reset resets the counter to its maximum value and clears the time window.
	//
	// After reset, the counter allows Max operations immediately without waiting.
	// If throttling is enabled and FuncCaller is configured, invokes the callback.
	//
	// Returns:
	//   - nil on successful reset
	//   - Error from FuncCaller if callback fails
	//
	// Thread-safe: Can be called concurrently with Pool operations.
	Reset() error

	// Clone creates an independent copy of the counter with the same configuration.
	//
	// The cloned counter:
	//   - Has the same Max, Wait, and FuncCaller settings
	//   - Starts with fresh state (num, tim reset)
	//   - Has its own mutex for independent thread-safety
	//
	// Returns a new Counter instance that operates independently.
	//
	// Thread-safe: Reads current state safely under mutex protection.
	Clone() Counter
}

// counter is the internal implementation of the Counter interface.
// It uses a mutex to ensure thread-safe access to its state.
type counter struct {
	m   sync.Mutex    // Protects all fields below
	num int           // Current remaining quota in this time window
	max int           // Maximum operations allowed per time window
	dur time.Duration // Duration of each time window
	tim time.Time     // Start time of current window (zero if not started)
	fct FuncCaller    // Optional callback for throttle events
}

// newCounter creates a new counter with the specified configuration.
//
// Parameters:
//   - max: Maximum number of operations allowed per time window.
//     Set to 0 or negative to disable throttling.
//   - dur: Duration of the time window. Set to 0 or negative to disable throttling.
//   - fct: Optional callback function for throttle events. Can be nil.
//
// The counter starts with full quota (num = max) and no active time window.
//
// Returns a Counter instance ready to use.
func newCounter(max int, dur time.Duration, fct FuncCaller) Counter {
	return &counter{
		m:   sync.Mutex{},
		num: max,
		max: max,
		dur: dur,
		tim: time.Time{},
		fct: fct,
	}
}

// Pool implements the core rate limiting logic with thread-safe quota management.
//
// Behavior:
//  1. If throttling is disabled (max <= 0 or dur <= 0), returns immediately
//  2. Checks if context is already cancelled
//  3. Initializes or resets time window if needed
//  4. If quota available: decrements counter and proceeds
//  5. If quota exhausted: sleeps until next window, then proceeds
//  6. Checks context cancellation after sleep
//  7. Calls FuncCaller callback if configured
//
// The method holds the mutex for the entire duration, including sleep time.
// This ensures that only one goroutine can be in the throttle wait state at a time.
func (c *counter) Pool(ctx context.Context) error {
	c.m.Lock()
	defer c.m.Unlock()

	// Throttling disabled if max or duration is not positive
	if c.max <= 0 || c.dur <= 0 {
		return nil
	}

	// Check if context is already cancelled
	if e := ctx.Err(); e != nil {
		return ErrorMailPoolerContext.Error(e)
	}

	// Initialize or reset time window if needed
	if c.tim.IsZero() {
		// First operation in this counter's lifetime
		c.num = c.max
	} else if time.Since(c.tim) > c.dur {
		// Time window has expired, start a new one
		c.num = c.max
		c.tim = time.Time{}
	}

	if c.num > 0 {
		// Quota available, proceed immediately
		c.num--
		c.tim = time.Now()
	} else {
		// Quota exhausted, wait until next time window
		time.Sleep(c.dur - time.Since(c.tim))

		// Start new window with one operation already consumed
		c.num = c.max - 1
		c.tim = time.Now()

		// Check if context was cancelled during sleep
		if e := ctx.Err(); e != nil {
			return ErrorMailPoolerContext.Error(e)
		} else if c.fct != nil {
			// Invoke callback to notify about throttle event
			if err := c.fct(); err != nil {
				return err
			}
		}
	}

	return nil
}

// Reset resets the counter state to allow immediate operations.
//
// This method:
//  1. Resets the counter to maximum quota
//  2. Clears the time window
//  3. Calls FuncCaller callback if configured and throttling is enabled
//
// Use cases:
//   - Manual reset after temporary rate limit
//   - Resetting state when switching contexts
//   - Testing scenarios requiring fresh state
func (c *counter) Reset() error {
	c.m.Lock()
	defer c.m.Unlock()

	// If throttling is disabled, nothing to reset
	if c.max <= 0 || c.dur <= 0 {
		return nil
	}

	// Reset to full quota with no active time window
	c.num = c.max
	c.tim = time.Time{}

	// Notify callback about reset event if configured
	if c.fct != nil {
		if err := c.fct(); err != nil {
			return err
		}
	}

	return nil
}

// Clone creates an independent copy of the counter for concurrent use.
//
// The cloned counter:
//   - Inherits the same Max, Wait, and FuncCaller configuration
//   - Starts with fresh state (full quota, no time window)
//   - Has its own mutex for independent thread-safety
//
// This is useful when you need multiple independent rate limiters with
// the same settings but separate quota tracking.
//
// Thread-safe: Reads the original counter's state under mutex protection.
func (c *counter) Clone() Counter {
	c.m.Lock()
	defer c.m.Unlock()

	return &counter{
		m:   sync.Mutex{},
		num: c.num, // Copy current quota state
		max: c.max,
		dur: c.dur,
		tim: time.Time{}, // Reset time window for fresh start
		fct: c.fct,
	}
}
