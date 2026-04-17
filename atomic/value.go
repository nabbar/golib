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

// val is the internal implementation of the Value[T] interface.
//
// # PERFORMANCE ARCHITECTURE
//
// This structure wraps sync/atomic.Value to provide type safety via generics. To address
// the overhead of zero-value detection and default value injection, it implements a
// "tiered access" model designed for maximum throughput.
//
// Internal State:
// - av (*atomic.Value): The primary storage for the current value.
// - dl (*atomic.Value): Storage for the default value returned by Load() when empty.
// - ds (*atomic.Value): Storage for the default value to replace zero-values during Store().
// - bl (*atomic.Bool): A high-performance flag; true if a default load value is set.
// - bs (*atomic.Bool): A high-performance flag; true if a default store value is set.
//
// Access Logic:
//   - Fast Path (No Defaults): If bl and bs are false, the implementation uses direct type
//     assertions on the native sync/atomic operations. This path is extremely fast.
//   - Validation Path (Defaults Set): If either flag is true, the implementation invokes
//     the Cast[T] utility to handle zero-value detection and default value replacement.
type val[T any] struct {
	av *atomic.Value // Underlying atomic storage for the main value.
	dl *atomic.Value // Storage for the default value returned by Load().
	ds *atomic.Value // Storage for the default value used during Store().
	bl *atomic.Bool  // Performance flag: true if a default load value is configured.
	bs *atomic.Bool  // Performance flag: true if a default store value is configured.
}

// SetDefaultLoad configures the default value returned by Load() when the container is empty.
// Once called, the internal performance flag 'bl' is set to true, and all subsequent
// Load() calls will include the cost of zero-value validation.
func (o *val[T]) SetDefaultLoad(def T) {
	o.dl.Store(def)
	o.bl.Store(true)
}

// SetDefaultStore configures the default value used to replace zero-values during Store().
// Once called, the internal performance flag 'bs' is set to true, and all subsequent
// Store() calls will perform an IsEmpty check on the input value.
func (o *val[T]) SetDefaultStore(def T) {
	o.ds.Store(def)
	o.bs.Store(true)
}

// getDefaultLoad retrieves the configured default value for Load operations.
// It uses a direct type assertion for maximum performance inside the validation path.
func (o *val[T]) getDefaultLoad() T {
	if v, k := o.dl.Load().(T); k {
		return v
	}
	var tmp T
	return tmp
}

// getDefaultStore retrieves the configured default value for Store operations.
func (o *val[T]) getDefaultStore() T {
	if v, k := o.ds.Load().(T); k {
		return v
	}
	var tmp T
	return tmp
}

// Load retrieves the current value atomically.
//
// Logic Flow:
// 1. Call native atomic.Value.Load().
// 2. Fast Path Check: If no default load is set (o.bl == false), perform direct type
//    assertion and return.
// 3. Validation Path: If a default is set, use Cast[T] to verify if the retrieved value
//    is "empty" (nil, zero scalar, or IsZero() struct). If empty, return the default load value.
func (o *val[T]) Load() (val T) {
	res := o.av.Load()

	// High-performance tiered access: bypass complex logic if no defaults are used.
	if !o.bl.Load() {
		if v, k := res.(T); k {
			return v
		}
		var tmp T
		return tmp
	}

	// Validation path: Ensure the value is not "empty" before returning.
	if v, k := res.(T); !k {
		return o.getDefaultLoad()
	} else if _, casted := Cast[T](v); !casted {
		return o.getDefaultLoad()
	} else {
		return v
	}
}

// Store sets the value atomically.
//
// Logic Flow:
// 1. If a default store value is set (o.bs == true), check if 'val' is empty via IsEmpty[T].
// 2. If 'val' is empty, store the configured default store value instead.
// 3. Otherwise, store 'val' as-is without any additional validation overhead.
func (o *val[T]) Store(val T) {
	if o.bs.Load() && IsEmpty[T](val) {
		o.av.Store(o.getDefaultStore())
	} else {
		o.av.Store(val)
	}
}

// Swap atomically stores the new value and returns the old one.
//
// This operation combines the logic of both Load and Store while maintaining
// atomicity. It respects the tiered performance model for both the input value
// (new) and the returned value (old).
func (o *val[T]) Swap(new T) (old T) {
	// Pre-process new value if default store is enabled.
	if o.bs.Load() && IsEmpty[T](new) {
		new = o.getDefaultStore()
	}

	// Perform atomic swap on the underlying storage.
	res := o.av.Swap(new)

	// Post-process old value if default load is enabled.
	if !o.bl.Load() {
		if v, k := res.(T); k {
			return v
		}
		var tmp T
		return tmp
	}

	if v, k := res.(T); !k {
		return o.getDefaultLoad()
	} else if _, casted := Cast[T](v); !casted {
		return o.getDefaultLoad()
	} else {
		return v
	}
}

// CompareAndSwap performs an atomic compare-and-swap operation.
//
// Logic Flow:
// 1. If default store is enabled, both 'old' and 'new' parameters are validated
//    against their zero-values and replaced by the default if necessary.
// 2. Invoke the native sync/atomic.Value.CompareAndSwap for atomicity.
func (o *val[T]) CompareAndSwap(old, new T) (swapped bool) {
	if o.bs.Load() {
		if IsEmpty[T](old) {
			old = o.getDefaultStore()
		}

		if IsEmpty[T](new) {
			new = o.getDefaultStore()
		}
	}

	return o.av.CompareAndSwap(old, new)
}
