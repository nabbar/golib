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
	"sync"
	"sync/atomic"

	liberr "github.com/nabbar/golib/errors"
)

// mod is the internal implementation of the Pool interface.
//
// It uses atomic primitives and a thread-safe map to provide lock-free or
// low-contention concurrent access. This structure is designed to be
// used via the Pool interface.
type mod struct {
	m sync.RWMutex
	l []error
	s *atomic.Uint64
}

// Add appends multiple errors to the pool.
// It iterates through the provided errors and calls the internal addOne helper.
func (o *mod) Add(e ...error) {
	var err = make([]error, 0, len(e))
	for i := range e {
		if e[i] != nil {
			err = append(err, e[i])
		}
	}

	o.m.Lock()
	defer o.m.Unlock()

	o.l = append(o.l, err...)
	o.s.Store(uint64(len(o.l)))
}

// Set assigns an error to a specific index.
// This is used for manual indexing.
// If the index given is not existing, the add function will be called
func (o *mod) Set(i uint64, e error) {
	if j := o.Len(); i >= j {
		o.Add(e)
		return
	}

	if e == nil {
		o.Del(i)
		return
	}

	o.m.Lock()
	defer o.m.Unlock()

	// Double check bounds after lock
	if i < uint64(len(o.l)) {
		o.l[i] = e
	}
}

// Del removes an error from the pool.
// This doesn't affect the sequence counter 's' or the high-water mark 'm'.
func (o *mod) Del(i uint64) {
	o.m.Lock()
	defer o.m.Unlock()

	j := uint64(len(o.l))
	if i >= j {
		return
	}

	// Shift elements to the left
	if i < j-1 {
		copy(o.l[i:], o.l[i+1:])
	}

	// Prevent memory leak by nil-ing the last element
	o.l[j-1] = nil
	o.l = o.l[:j-1]

	o.s.Store(uint64(len(o.l)))
}

// Get retrieves an error by its index.
// If the index is 0 or no error is found, it returns nil.
func (o *mod) Get(i uint64) error {
	o.m.RLock()
	defer o.m.RUnlock()

	if i >= uint64(len(o.l)) {
		return nil
	}

	return o.l[i]
}

// Error combines all errors in the pool into a single Error object.
// It retrieves all errors using Slice() and then uses the main errors package
// to wrap them if there are more than one.
func (o *mod) Error() error {
	o.m.RLock()
	defer o.m.RUnlock()

	switch len(o.l) {
	case 0:
		return nil
	case 1:
		return o.l[0]
	default:
		return liberr.New(0, "", o.l...)
	}
}

// Slice returns a slice containing all non-nil errors currently stored.
// The order of errors in the slice corresponds to the map's iteration order
// and is not guaranteed to be sequential or sorted by index.
func (o *mod) Slice() []error {
	o.m.RLock()
	defer o.m.RUnlock()

	l := len(o.l)
	if l == 0 {
		return nil
	}

	r := make([]error, l)
	copy(r, o.l)
	return r
}

// Len returns the number of items currently in the map.
// This is an O(1) operation.
func (o *mod) Len() uint64 {
	return o.s.Load()
}

// MaxId returns the current value of the high-water mark counter.
func (o *mod) MaxId() uint64 {
	return o.Len()
}

// Last returns the error at the current MaxId index.
func (o *mod) Last() error {
	o.m.RLock()
	defer o.m.RUnlock()

	l := len(o.l)
	if l == 0 {
		return nil
	}

	return o.l[l-1]
}

// Clear empties the entire pool.
// It iterates through all keys to delete them and resets the MaxId.
func (o *mod) Clear() {
	o.m.Lock()
	defer o.m.Unlock()

	// Clear the slice and nil out references
	clear(o.l)
	o.l = o.l[:0]
	o.s.Store(0)
}
