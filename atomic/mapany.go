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
	"sync"
	"sync/atomic"
)

// ma is the internal implementation of the Map interface.
//
// # IMPLEMENTATION DETAILS
//
// It uses sync.Map as the underlying concurrent storage. While sync.Map handles
// any/any pairs, this wrapper enforces type safety for keys through generics [K].
// Values remain as 'any' to satisfy the Map interface.
//
// Internal State:
//   - l (*atomic.Int64): An atomic counter for the number of entries in the map.
//   - m (sync.Map): The core thread-safe map implementation from the Go standard library.
//
// Thread-safety for the length counter:
// The counter 'l' is updated using atomic operations (Add(1) or Add(-1)) based on the
// results of sync.Map operations (like Swap, LoadAndDelete, etc.) to ensure it accurately
// reflects the number of entries in constant time.
type ma[K comparable] struct {
	l *atomic.Int64 // Current count of entries in the map.
	m sync.Map      // The underlying concurrent-safe map.
}

// Len returns the number of entries currently stored in the map.
// This is a constant-time O(1) operation maintained via atomic increments and decrements.
func (o *ma[K]) Len() uint64 {
	if i := o.l.Load(); i < 0 {
		return 0
	} else {
		return uint64(i)
	}
}

// Load retrieves the value for a key from the map.
func (o *ma[K]) Load(key K) (value any, ok bool) {
	return o.m.Load(key)
}

// Store sets the value for a key in the map.
//
// Logic:
// We use the atomic Swap method of sync.Map to set the value. If Swap indicates
// that the key was not previously present (loaded == false), we increment the length counter.
func (o *ma[K]) Store(key K, value any) {
	if _, loaded := o.m.Swap(key, value); !loaded {
		o.l.Add(1)
	}
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// If a new value is stored, the length counter is incremented.
func (o *ma[K]) LoadOrStore(key K, value any) (actual any, loaded bool) {
	actual, loaded = o.m.LoadOrStore(key, value)
	if !loaded {
		o.l.Add(1)
	}
	return actual, loaded
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// If an entry was found and deleted, the length counter is decremented.
func (o *ma[K]) LoadAndDelete(key K) (value any, loaded bool) {
	value, loaded = o.m.LoadAndDelete(key)
	if loaded {
		o.l.Add(-1)
	}
	return value, loaded
}

// Delete removes the value for a key from the map.
// If an entry was found and deleted, the length counter is decremented.
func (o *ma[K]) Delete(key K) {
	if _, loaded := o.m.LoadAndDelete(key); loaded {
		o.l.Add(-1)
	}
}

// Swap exchanges the value for a key and returns the previous value.
// If the key did not exist before, the length counter is incremented.
func (o *ma[K]) Swap(key K, value any) (previous any, loaded bool) {
	previous, loaded = o.m.Swap(key, value)
	if !loaded {
		o.l.Add(1)
	}
	return previous, loaded
}

// CompareAndSwap performs a CAS operation on a map entry.
// Note: sync.Map.CompareAndSwap does not provide information about whether it was a new entry,
// as CAS by definition only operates on existing matching values.
func (o *ma[K]) CompareAndSwap(key K, old, new any) bool {
	return o.m.CompareAndSwap(key, old, new)
}

// CompareAndDelete deletes the entry if the current value matches 'old'.
// If the deletion succeeds, the length counter is decremented.
func (o *ma[K]) CompareAndDelete(key K, old any) (deleted bool) {
	if deleted = o.m.CompareAndDelete(key, old); deleted {
		o.l.Add(-1)
	}
	return deleted
}

// Range iterates over the map.
//
// # SELF-HEALING MECHANISM
//
// To protect against cases where external code might have directly injected an
// incorrectly typed key into the underlying sync.Map, this implementation performs
// a Cast[K] check on each key during iteration.
//
// If a key cannot be cast to type K:
// 1. The invalid entry is automatically removed from the map.
// 2. The length counter is decremented.
// 3. The invalid entry is skipped, and iteration continues.
func (o *ma[K]) Range(f func(key K, value any) bool) {
	o.m.Range(func(key, value any) bool {
		var (
			l bool
			k K
		)

		// Validate key type before calling the user function.
		if k, l = Cast[K](key); !l {
			// Evict invalid keys and decrement counter.
			o.m.Delete(key)
			o.l.Add(-1)
			return true
		}

		// Execute user-defined function for the typed key.
		return f(k, value)
	})
}
