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

package mapCloser

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync/atomic"

	libctx "github.com/nabbar/golib/context"
)

// closer is the internal implementation of the Closer interface.
// It uses atomic operations exclusively for thread-safe counter and state management,
// avoiding mutexes entirely for maximum performance.
//
// Design Philosophy:
//   - Lock-free: All state changes use atomic operations
//   - Context-driven: Lifecycle tied to context cancellation
//   - Fail-safe: Continue operations even when errors occur
//   - Memory-safe: All operations check for nil and closed state
//
// Memory Layout:
//   - c: atomic boolean flag indicating if closer has been closed
//   - f: context cancellation function (called once during Close)
//   - i: atomic counter tracking number of Add() calls (never decrements)
//   - x: thread-safe storage mapping counter values to io.Closer instances
type closer struct {
	c *atomic.Bool          // Closed flag: true after Close() is called
	f func()                // Context cancel function: cancels associated context
	i *atomic.Uint64        // Counter for registered closers: monotonically increasing
	x libctx.Config[uint64] // Storage for closers indexed by counter: thread-safe map
}

// idx returns the current counter value.
// This is a read-only operation that does not modify state.
//
// Performance: O(1) atomic load operation.
func (o *closer) idx() uint64 {
	return o.i.Load()
}

// Add implements the Closer interface.
// See interface documentation for detailed behavior.
func (o *closer) Add(clo ...io.Closer) {
	// Safety checks: nil receiver, nil storage, or closed state
	if o == nil {
		return
	} else if o.x == nil {
		return
	} else if o.x.Err() != nil {
		// Context is done or closer is closed
		return
	}

	// Store each closer with a unique incrementing key
	// Note: nil closers are accepted but will be filtered during Get() and Close()
	for _, c := range clo {
		o.x.Store(o.i.Add(1), c)
	}
}

// Get implements the Closer interface.
// See interface documentation for detailed behavior.
func (o *closer) Get() []io.Closer {
	var res = make([]io.Closer, 0)

	// Safety checks: nil receiver, nil storage, or closed state
	if o == nil {
		return res
	} else if o.x == nil {
		return res
	} else if o.x.Err() != nil {
		// Context is done or closer is closed
		return res
	}

	// Walk through all stored closers and collect non-nil ones
	o.x.Walk(func(key uint64, val interface{}) bool {
		if val == nil {
			// Skip nil closers
			return true
		}
		if v, k := val.(io.Closer); !k {
			// Skip values that are not io.Closer (should never happen)
			return true
		} else {
			res = append(res, v)
			return true
		}
	})
	return res
}

// Len implements the Closer interface.
// See interface documentation for detailed behavior.
func (o *closer) Len() int {
	i := o.idx()

	// Handle overflow case: uint64 > math.MaxInt
	// This is extremely unlikely but documented behavior
	if i > math.MaxInt {
		// overflow: return 0 as documented in interface
		return 0
	} else {
		return int(i)
	}
}

// Len64 returns the counter as uint64 without overflow protection.
// This is an internal helper that provides the raw counter value.
//
// Not exported - use Len() for public API which handles overflow correctly.
func (o *closer) Len64() uint64 {
	return o.idx()
}

// Clean implements the Closer interface.
// See interface documentation for detailed behavior.
func (o *closer) Clean() {
	// Safety checks: nil receiver, nil storage, or closed state
	if o == nil {
		return
	} else if o.x == nil {
		return
	} else if o.x.Err() != nil {
		// Context is done or closer is closed
		return
	}

	// Reset counter to 0 and clear all stored closers
	// Note: Closers are NOT closed, just removed from storage
	o.i.Store(0)
	o.x.Clean()
}

// Clone implements the Closer interface.
// See interface documentation for detailed behavior.
func (o *closer) Clone() Closer {
	// Safety checks: nil receiver, nil storage, or closed state
	if o == nil {
		return nil
	} else if o.x == nil {
		return nil
	} else if o.x.Err() != nil {
		// Context is done or closer is closed
		return nil
	}

	// Create new atomic counter with current value
	i := new(atomic.Uint64)
	i.Store(o.idx())

	// Create new atomic bool with current closed state
	c := new(atomic.Bool)
	c.Store(o.c.Load())

	// Return new closer with copied state but independent storage
	// Note: context cancellation function is shared between original and clone
	return &closer{
		c: c,              // Independent closed flag
		f: o.f,            // Shared context cancel function
		i: i,              // Independent counter
		x: o.x.Clone(o.x), // Deep copy of closer storage
	}
}

// Close implements the Closer interface.
// See interface documentation for detailed behavior.
func (o *closer) Close() error {
	var e = make([]string, 0)

	// Safety check: nil receiver
	if o == nil {
		return fmt.Errorf("not initialized")
	}

	// Atomic compare-and-swap to prevent double close
	// Only the first call to Close() will succeed (swap false to true)
	if !o.c.CompareAndSwap(false, true) {
		// Already closed by another goroutine or previous call
		return fmt.Errorf("closer already closed")
	}

	// Cancel the context after closing all resources
	// This signals to other goroutines that cleanup is complete
	if o.f != nil {
		defer o.f()
	}

	// Safety check: nil storage
	if o.x == nil {
		return fmt.Errorf("not initialized")
	} else if o.x.Err() != nil {
		// Context already done, return its error
		return o.x.Err()
	}

	// Walk through all closers and attempt to close each one
	// Continue even if some closers fail (fail-safe behavior)
	o.x.Walk(func(key uint64, val interface{}) bool {
		if c, k := val.(io.Closer); !k {
			// Not an io.Closer or nil, skip it
			return true
		} else if err := c.Close(); err != nil {
			// Collect error message for aggregation
			e = append(e, err.Error())
		}
		return true // Continue walking even after errors
	})

	// Aggregate all errors into a single error message
	if len(e) > 0 {
		return fmt.Errorf("%s", strings.Join(e, ", "))
	}

	return nil
}
