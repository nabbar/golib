/*
 * MIT License
 *
 * Copyright (c) 2025 Nicolas JUHEL
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
 */

package aggregator

import (
	"context"
	"time"

	"github.com/nabbar/golib/runner"
)

// Deadline returns the time when work done on behalf of this context should be cancelled.
//
// Technical Implementation:
// This method implements the standard context.Context interface. It performs an atomic
// load of the current operational context ('x'). If no operational context exists,
// it indicates that no deadline is set.
//
// Returns:
//   - deadline: The absolute time when the context will be cancelled.
//   - ok: true if a deadline is set, false otherwise.
func (o *agg) Deadline() (deadline time.Time, ok bool) {
	if x := o.x.Load(); x != nil {
		return x.Deadline()
	}
	return time.Time{}, false
}

// Done returns a channel that is closed when the aggregator's work should be cancelled.
//
// Operational Triggers:
// The returned channel is closed under the following conditions:
//  1. The parent context provided to New() or Start() is cancelled.
//  2. Stop() or Close() is explicitly called on the aggregator.
//  3. An internal processing error causes the runner to terminate.
//  4. The operational deadline is exceeded.
//
// Implementation Details:
// It performs a high-speed atomic load of the operational context. If the aggregator is
// currently stopped, it returns a pre-closed channel to ensure that any callers
// immediately detect the inactive state.
//
// Returns:
//   - <-chan struct{}: A channel that signals context completion.
func (o *agg) Done() <-chan struct{} {
	if x := o.x.Load(); x != nil {
		return x.Done()
	}

	// Inactive State: return a pre-closed channel to signal completion.
	c := make(chan struct{})
	close(c)
	return c
}

// Err returns a non-nil error explaining why the aggregator's context was cancelled.
//
// Technical Implementation:
// It implements context.Context. If Done() is not yet closed, it returns nil.
// Otherwise, it returns the cancellation error (e.g., context.Canceled or
// context.DeadlineExceeded) from the underlying operational context.
//
// Returns:
//   - error: nil if active, otherwise the reason for termination.
func (o *agg) Err() error {
	if x := o.x.Load(); x != nil {
		return x.Err()
	}
	return nil
}

// Value retrieves request-scoped data associated with the aggregator's context.
//
// Technical Implementation:
// It implements context.Context by delegating the lookup to the active operational
// context, which in turn inherits from the parent context provided at instantiation.
//
// Parameters:
//   - key: The unique identifier for the context value.
//
// Returns:
//   - any: The value associated with the key, or nil if not found.
func (o *agg) Value(key any) any {
	if x := o.x.Load(); x != nil {
		return x.Value(key)
	}
	return nil
}

// ctxNew initializes a fresh operational context derived from the root parent context.
//
// Lifecycle Orchestration:
//  1. Context Derivation: Creates a new cancellable context using 'context.WithCancel'
//     based on the master context ('m') stored during New().
//  2. Atomic Swap: Atomically updates the current context and its cancel function.
//  3. Resource Safety: If a previous cancel function existed, it is invoked to ensure
//     that any leaked goroutines or timers associated with the old context are terminated.
func (o *agg) ctxNew() {
	defer func() {
		// Recovery from panics during context initialization.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/ctxnew", r)
		}
	}()

	// Derive the new context from the immutable root context.
	x, n := context.WithCancel(o.m.Load())
	o.x.Store(x)

	// Atomically swap the cancel function and finalize the old one if it exists.
	old := o.n.Swap(n)
	if old != nil {
		old()
	}
}

// ctxClose performs an orderly decommissioning of the operational context.
//
// Shutdown Logic:
//  1. Atomic Cancellation: Swaps the current cancel function with a no-op to prevent
//     double-cancellation and then invokes the original function.
//  2. State Consistency: Replaces the operational context with a fresh, pre-cancelled
//     context derived from context.Background(). This ensures that subsequent calls
//     to Done() or Err() correctly report that the aggregator is closed, even after
//     the runner has finalized.
//
// This method is idempotent and safe for concurrent invocation during shutdown.
func (o *agg) ctxClose() {
	defer func() {
		// Recovery from panics during context cleanup.
		if r := recover(); r != nil {
			runner.RecoveryCaller("golib/ioutils/aggregator/ctxclose", r)
		}
	}()

	// Atomically retrieve and deactivate the current cancellation trigger.
	old := o.n.Swap(func() {})
	if old != nil {
		old()
		// If we successfully swapped a real cancel function, the context is already transitioning.
		return
	}

	// If no cancel function was found, ensure the operational context reports as closed.
	x, n := context.WithCancel(context.Background())
	n() // Immediately cancel the sentinel context.

	o.x.Store(x)
}
