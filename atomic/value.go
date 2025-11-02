/*
 * MIT License
 *
 * Copyright (c) 2024 Nicolas JUHEL
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

package atomic

import (
	"sync/atomic"
)

// val is the internal implementation of Value[T] interface.
// It wraps sync/atomic.Value with type-safe operations and default value support.
type val[T any] struct {
	av *atomic.Value // atomic value of T
	dl *atomic.Value // default value for load
	ds *atomic.Value // default value for store
}

// SetDefaultLoad configures the default value returned by Load when the atomic value is empty.
// This allows graceful handling of uninitialized values.
func (o *val[T]) SetDefaultLoad(def T) {
	o.dl.Store(newDefault[T](def))
}

// SetDefaultStore configures the default value used to replace empty values in Store operations.
// This enables automatic substitution of zero/empty values with a meaningful default.
func (o *val[T]) SetDefaultStore(def T) {
	o.ds.Store(newDefault[T](def))
}

// getDefault retrieves and unwraps a default value from the atomic storage.
// Returns the zero value of T if the stored value cannot be cast to defaultValue[T].
func (o *val[T]) getDefault(i any) T {
	if v, k := Cast[defaultValue[T]](i); !k {
		var tmp T
		return tmp
	} else {
		return v.GetDefault()
	}
}

// getDefaultLoad returns the configured default value for Load operations.
func (o *val[T]) getDefaultLoad() T {
	return o.getDefault(o.dl.Load())
}

// getDefaultStore returns the configured default value for Store operations.
func (o *val[T]) getDefaultStore() T {
	return o.getDefault(o.ds.Load())
}

// Load retrieves the current value atomically.
// Returns the configured default load value if the atomic value is empty or cannot be cast to T.
// This operation is lock-free and safe for concurrent access.
func (o *val[T]) Load() (val T) {
	if v, k := Cast[T](o.av.Load()); !k {
		return o.getDefaultLoad()
	} else {
		return v
	}
}

// Store sets the value atomically.
// If the provided value is empty (as determined by IsEmpty), the configured default store value is used instead.
// This operation is lock-free and safe for concurrent access.
func (o *val[T]) Store(val T) {
	if IsEmpty[T](val) {
		o.av.Store(o.getDefaultStore())
	} else {
		o.av.Store(val)
	}
}

// Swap atomically stores the new value and returns the old value.
// If the new value is empty, the configured default store value is used instead.
// Returns the default load value if the old value cannot be cast to T.
// This operation is lock-free and safe for concurrent access.
func (o *val[T]) Swap(new T) (old T) {
	if IsEmpty[T](new) {
		new = o.getDefaultStore()
	}

	if v, k := Cast[T](o.av.Swap(new)); !k {
		return o.getDefaultLoad()
	} else {
		return v
	}
}

// CompareAndSwap atomically compares the current value with old and, if they match, stores new.
// Returns true if the swap was successful, false otherwise.
// Empty values for old or new are replaced with the configured default store value.
// This operation is lock-free and safe for concurrent access.
func (o *val[T]) CompareAndSwap(old, new T) (swapped bool) {
	if IsEmpty[T](old) {
		old = o.getDefaultStore()
	}

	if IsEmpty[T](new) {
		new = o.getDefaultStore()
	}

	return o.av.CompareAndSwap(old, new)
}
