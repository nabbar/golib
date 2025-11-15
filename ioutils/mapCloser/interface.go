/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2021 Nicolas JUHEL
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
	"time"

	libctx "github.com/nabbar/golib/context"
)

// Closer is a thread-safe manager for multiple io.Closer instances.
// It provides automatic cleanup when the associated context is cancelled
// and allows manual resource management through Add, Get, Clean, and Close methods.
// All methods are safe for concurrent use.
type Closer interface {
	// Add registers one or more io.Closer instances for management.
	// If the Closer is already closed or the context is done, this is a no-op.
	// Nil closers are accepted but filtered out during Get() and Close().
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	Add(clo ...io.Closer)

	// Get returns a copy of all registered io.Closer instances, excluding nil values.
	// The returned slice is independent and safe to modify.
	// Returns an empty slice if the Closer is closed or no closers are registered.
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	Get() []io.Closer

	// Len returns the total count of closers that have been added.
	// This represents the internal counter, including nil values.
	// Returns 0 if overflow occurs (exceeds math.MaxInt).
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	Len() int

	// Clean removes all registered closers without closing them.
	// Resets the internal counter to zero. Does nothing if already closed.
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	Clean()

	// Clone creates an independent copy of this Closer with the same state.
	// The cloned Closer shares the same context but has independent closer storage.
	// Returns nil if the original Closer is already closed.
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	Clone() Closer

	// Close cancels the context and closes all registered io.Closer instances.
	// Returns an aggregated error if any closer fails to close.
	// Subsequent calls return an error indicating the Closer is already closed.
	//
	// Thread-safe: Can be called concurrently from multiple goroutines.
	Close() error
}

// New creates a new Closer that monitors the provided context.
//
// The returned Closer automatically closes all registered io.Closer instances when:
//   - The context is cancelled
//   - The context times out
//   - Close() is called manually
//
// A background goroutine monitors the context every 100ms and triggers automatic cleanup
// when the context is done. All methods of the returned Closer are thread-safe.
//
// Parameters:
//   - ctx: Context to monitor for cancellation signals
//
// Returns:
//   - Closer: A new thread-safe Closer instance
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	closer := mapCloser.New(ctx)
//	closer.Add(file1, file2, conn)
//	// Resources auto-close when context times out
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

	go func() {
		for !c.c.Load() {
			select {
			case <-c.x.Done():
				_ = c.Close()
				return
			default:
				time.Sleep(time.Millisecond * 100)
			}
		}
	}()

	return c
}
