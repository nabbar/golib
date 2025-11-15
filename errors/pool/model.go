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

package pool

import (
	"sync/atomic"

	libatm "github.com/nabbar/golib/atomic"
	liberr "github.com/nabbar/golib/errors"
)

// mod is the concrete implementation of the Pool interface.
// It uses atomic operations for the sequence counter and a concurrent-safe
// map for storing errors.
type mod struct {
	// s is the sequence counter for automatic index assignment
	// It tracks the next available index for Add operations
	s *atomic.Uint64

	// l is the concurrent-safe map storing errors by their index
	l libatm.MapTyped[uint64, error]
}

// Add implements Pool.Add by appending errors with sequential indices.
// Each non-nil error is assigned the next available index atomically.
// Nil errors are silently ignored.
func (o *mod) Add(e ...error) {
	for _, err := range e {
		if err != nil {
			// Use the return value of Add to get the new index atomically
			// This prevents race conditions where another goroutine might
			// modify the counter between Add and Load operations
			idx := o.s.Add(1)
			o.l.Store(idx, err)
		}
	}
}

// Get implements Pool.Get by retrieving the error at the specified index.
// Returns nil if the index doesn't exist or if the stored error is nil.
func (o *mod) Get(i uint64) error {
	if e, l := o.l.Load(i); !l || e == nil {
		return nil
	} else {
		return e
	}
}

// Set implements Pool.Set by storing an error at a specific index.
// Nil errors are ignored. This allows overwriting existing errors
// and creating sparse index assignments.
func (o *mod) Set(i uint64, e error) {
	if e != nil {
		o.l.Store(i, e)
	}
}

// Del implements Pool.Del by removing the error at the specified index.
// This operation is safe even if the index doesn't exist.
func (o *mod) Del(i uint64) {
	o.l.Delete(i)
}

// Error implements Pool.Error by combining all errors in the pool.
// Uses liberr.UnknownError to create a multi-error that supports
// error unwrapping for compatibility with errors.Is and errors.As.
// Returns nil if the pool is empty.
func (o *mod) Error() error {
	return liberr.UnknownError.IfError(o.Slice()...)
}

// Slice implements Pool.Slice by collecting all non-nil errors.
// The order of errors in the returned slice is not guaranteed.
// Iterates through all stored errors using the concurrent-safe Range method.
func (o *mod) Slice() []error {
	var e = make([]error, 0)
	o.l.Range(func(_ uint64, err error) bool {
		e = append(e, err)
		return true
	})
	return e
}

// Len implements Pool.Len by counting all non-nil errors in the pool.
// Iterates through all stored errors to calculate the count.
// This count may be less than MaxId if errors have been deleted.
func (o *mod) Len() uint64 {
	var i uint64
	o.l.Range(func(_ uint64, err error) bool {
		if err != nil {
			i++
		}
		return true
	})
	return i
}

// MaxId implements Pool.MaxId by finding the highest index with a non-nil error.
// Returns 0 if the pool is empty.
// Scans all stored errors to find the maximum index.
func (o *mod) MaxId() uint64 {
	var i uint64
	o.l.Range(func(k uint64, err error) bool {
		if err != nil && k > i {
			i = k
		}
		return true
	})
	return i
}

// Last implements Pool.Last by returning the error at the highest index.
// This is equivalent to Get(MaxId()).
// Returns nil if the pool is empty or if the last index was deleted.
func (o *mod) Last() error {
	return o.Get(o.MaxId())
}

// Clear implements Pool.Clear by removing all errors from the pool.
// Note: The sequence counter (s) is not reset, so subsequent Add operations
// will continue from the next sequential index, not restart at 1.
// This ensures index uniqueness across the pool's lifetime.
func (o *mod) Clear() {
	o.l.Range(func(k uint64, _ error) bool {
		o.l.Delete(k)
		return true
	})
}
