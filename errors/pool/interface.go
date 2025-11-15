/*
 * MIT License
 *
 * Copyright (c) 2020 Nicolas JUHEL
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

// Package pool provides a thread-safe error pool for collecting and managing multiple errors.
//
// The Pool interface allows concurrent error collection with automatic indexing, providing
// methods to add, retrieve, update, and delete errors by index. It's useful for scenarios
// where multiple operations might fail and all errors need to be collected and reported.
//
// Example usage:
//
//	p := pool.New()
//	defer func() {
//	    if err := p.Error(); err != nil {
//	        log.Printf("Errors occurred: %v", err)
//	    }
//	}()
//
//	// Collect errors from multiple operations
//	for _, item := range items {
//	    if err := process(item); err != nil {
//	        p.Add(err)
//	    }
//	}
//
// Thread Safety:
//   - All operations are thread-safe and can be called concurrently
//   - Uses atomic operations and concurrent-safe maps internally
//   - No external synchronization required
package pool

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
)

// Pool is a thread-safe error collection that manages errors with automatic indexing.
//
// The pool assigns sequential indices starting from 1 when errors are added.
// Errors can be accessed, modified, or deleted by their index. The pool provides
// various methods to query and manipulate the error collection.
type Pool interface {
	// Add appends one or more errors to the pool with automatic sequential indexing.
	// Nil errors are ignored and do not consume an index.
	// This operation is thread-safe and can be called concurrently.
	//
	// Example:
	//   p.Add(err1, err2, err3)
	Add(e ...error)

	// Get retrieves the error at the specified index.
	// Returns nil if the index does not exist or contains a nil error.
	// Index 0 always returns nil as indices start at 1.
	//
	// Example:
	//   err := p.Get(5)
	Get(i uint64) error

	// Set stores an error at the specified index, overwriting any existing error.
	// If the error is nil, the operation is ignored (use Del to remove errors).
	// This allows sparse index assignment and updating existing errors.
	//
	// Example:
	//   p.Set(10, errors.New("error at index 10"))
	Set(i uint64, e error)

	// Del removes the error at the specified index.
	// This operation is idempotent - deleting a non-existent index has no effect.
	//
	// Example:
	//   p.Del(5)
	Del(i uint64)

	// Error returns a combined error containing all errors in the pool.
	// Returns nil if the pool is empty.
	// The returned error implements error unwrapping for compatibility with errors.Is/As.
	//
	// Example:
	//   if err := p.Error(); err != nil {
	//       return err
	//   }
	Error() error

	// Slice returns all errors in the pool as a slice.
	// The order is not guaranteed. Returns an empty slice if no errors exist.
	// Deleted indices are not included in the result.
	//
	// Example:
	//   for _, err := range p.Slice() {
	//       log.Println(err)
	//   }
	Slice() []error

	// Len returns the count of non-nil errors currently in the pool.
	// This count excludes deleted errors and may be less than MaxId.
	//
	// Example:
	//   count := p.Len()
	Len() uint64

	// MaxId returns the highest index currently used in the pool.
	// Returns 0 if the pool is empty.
	//
	// Example:
	//   maxIdx := p.MaxId()
	MaxId() uint64

	// Last returns the error at the highest index (MaxId).
	// Returns nil if the pool is empty or if the last index was deleted.
	//
	// Example:
	//   lastErr := p.Last()
	Last() error

	// Clear removes all errors from the pool, resetting it to empty state.
	// After Clear, Len returns 0, but subsequent Add operations continue
	// from the next sequential index (not reset to 1).
	//
	// Example:
	//   p.Clear()
	Clear()
}

// New creates and returns a new empty error pool.
//
// The returned pool is ready for immediate use and all operations are thread-safe.
// The pool uses atomic operations and concurrent-safe maps internally for
// high-performance concurrent access.
//
// Example:
//
//	p := pool.New()
//	p.Add(errors.New("first error"))
//	p.Add(errors.New("second error"))
//	fmt.Println(p.Len()) // Output: 2
func New() Pool {
	return &mod{
		s: new(atomic.Uint64),
		l: libatm.NewMapTyped[uint64, error](),
	}
}
