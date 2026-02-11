/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2025 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

// Package mapCloser provides a thread-safe, context-aware manager for multiple io.Closer instances.
// It automatically closes all registered closers when a context is cancelled or when manually triggered,
// making resource cleanup safe and predictable in concurrent applications.
package mapCloser

import (
	"context"
	"io"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
)

// Closer is a thread-safe manager for multiple io.Closer instances.
//
// It provides automatic cleanup when the associated context is cancelled
// and allows manual resource management through Add, Get, Clean, and Close methods.
// All methods are safe for concurrent use and rely exclusively on atomic operations
// for maximum performance.
//
// Thread-Safety Guarantees:
//   - All methods can be called concurrently from multiple goroutines
//   - No mutexes are used; only atomic operations
//   - Add operations never block other operations
//   - Close is idempotent but returns an error on subsequent calls
//
// Lifecycle:
//  1. Create with New(ctx)
//  2. Add io.Closer instances dynamically
//  3. Automatic close when context is done OR manual Close()
//  4. Post-close: all operations become no-ops
type Closer interface {
	// Add registers one or more io.Closer instances for management.
	//
	// Behavior:
	//   - If the Closer is already closed or the context is done, this is a no-op
	//   - Nil closers are accepted but filtered out during Get() and Close()
	//   - Each Add increments the internal counter (visible via Len())
	//   - No validation is performed on the provided closers
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	// Performance: O(1) atomic increment + O(1) map store.
	Add(clo ...io.Closer)

	// Get returns a copy of all registered io.Closer instances, excluding nil values.
	//
	// Behavior:
	//   - Returns an empty slice if the Closer is closed or no closers are registered
	//   - The returned slice is independent and safe to modify
	//   - Nil closers are automatically filtered out
	//   - The order of closers in the slice is not guaranteed
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	// Performance: O(n) where n is the number of registered closers.
	Get() []io.Closer

	// Len returns the total count of closers that have been added.
	//
	// Behavior:
	//   - This represents the internal counter, including nil values
	//   - Returns 0 if overflow occurs (exceeds math.MaxInt)
	//   - Counter is never decremented, even after Clean()
	//   - Clean() resets the counter to 0
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	// Performance: O(1) atomic load.
	Len() int

	// Clean removes all registered closers without closing them.
	//
	// Behavior:
	//   - Resets the internal counter to zero
	//   - Clears all stored closers from memory
	//   - Does NOT close the closers (use Close() for that)
	//   - Does nothing if already closed
	//   - After Clean(), the Closer can be reused with new closers
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	// Performance: O(1) reset operations.
	Clean()

	// Clone creates an independent copy of this Closer with the same state.
	//
	// Behavior:
	//   - The cloned Closer shares the same context cancellation function
	//   - The cloned Closer has independent closer storage (deep copy)
	//   - The counter value is copied at the time of cloning
	//   - Returns nil if the original Closer is already closed
	//   - Modifications to the clone do not affect the original
	//   - Both closers can be closed independently
	//
	// Use case: Create hierarchical resource managers for sub-contexts.
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	// Performance: O(n) where n is the number of registered closers.
	Clone() Closer

	// Close cancels the context and closes all registered io.Closer instances.
	//
	// Behavior:
	//   - Cancels the associated context immediately
	//   - Closes all registered closers (excluding nil values)
	//   - Continues closing even if some closers return errors
	//   - Returns an aggregated error if any closer fails to close
	//   - First call: performs cleanup and returns result
	//   - Subsequent calls: returns error "closer already closed"
	//   - After Close(), all operations (Add, Get, Clean, Clone) become no-ops
	//
	// Error Format:
	//   Multiple errors are joined with commas: "error 1, error 2, error 3"
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	// The first call to succeed will close the resources; others will fail.
	// Performance: O(n) where n is the number of registered closers.
	Close() error
}

// New creates a new Closer that monitors the provided context.
//
// The returned Closer automatically closes all registered io.Closer instances when:
//   - The context is cancelled (ctx.Done() receives a signal)
//   - The context times out (if created with context.WithTimeout/WithDeadline)
//   - Close() is called manually
//
// Implementation Details:
//   - A background goroutine is spawned to monitor the context
//   - The goroutine blocks on ctx.Done() for immediate response
//   - When the context is done, Close() is called automatically
//   - All methods of the returned Closer are thread-safe
//   - Uses atomic operations exclusively (no mutexes)
//
// Parameters:
//   - ctx: Context to monitor for cancellation signals. Must not be nil.
//
// Returns:
//   - Closer: A new thread-safe Closer instance, never nil.
//
// Example usage:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	closer := mapCloser.New(ctx)
//	closer.Add(file1, file2, conn)
//	// Resources auto-close when context times out or cancel() is called
//
// Performance:
//   - O(1) initialization
//   - Background goroutine has minimal overhead
//   - Goroutine exits after Close() is called
func New(ctx context.Context) Closer {
	var x, n = context.WithCancel(ctx)

	c := &closer{
		f: n,
		i: new(atomic.Uint64),
		c: new(atomic.Bool),
		x: libctx.New[uint64](x),
	}

	c.c.Store(false)
	c.i.Store(0)

	return c
}
